// Package server provides the Streamable HTTP MCP server, chi router, CORS middleware, and health endpoint.
package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/dictybase/dcr-mcp/internal/prompts"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds HTTP and GraphQL configuration for the MCP server.
type Config struct {
	GraphQLEndpoint string
	Host            string
	Port            int
	Stateless       bool
	AllowedOrigins  []string
	// RateLimit is the max /mcp requests per client IP per minute; 0 uses the default.
	RateLimit int
	// Linkouts builds public website links. The zero value uses
	// linkout.DefaultBaseURL.
	Linkouts linkout.Links
}

const (
	serverName        = "dcr-mcp"
	serverVersion     = "1.1.0"
	wildcardOrigin    = "*"
	healthzBody       = `{"status":"ok"}`
	corsMaxAgeSeconds = 300
	defaultRateLimit  = 120
	rateLimitWindow   = time.Minute

	serverInstructions = "dcr-mcp gives read-only access to the Dicty Stock Center catalog of Dictyostelium " +
		"strains (DBS ids) and plasmids (DBP ids). Guidance: " +
		"use find_stocks_by_gene when the user names a gene; " +
		"use search_stocks for other discovery questions and autocomplete_stocks for partial names or ids; " +
		"use search_phenotypes to find strains by phenotype statement (phenotype, environment, or assay) " +
		"from free words; list_strains_with_phenotype accepts only an exact phenotype term, so copy it from " +
		"the phenotype field of get_strain first and never pass an environment or an assay term; " +
		"use get_strain and get_plasmid with the default detail=summary and switch to detail=full only when needed; " +
		"use check_availability before recommending stocks to order; " +
		"use get_publication to read a paper cited on a record. " +
		"Always cite stocks with the url field from results. The catalog cannot be modified through this server."
)

var (
	implNameLens = L.MakeLens(
		func(impl mcp.Implementation) string { return impl.Name },
		func(impl mcp.Implementation, name string) mcp.Implementation {
			impl.Name = name
			return impl
		},
	)
	implVersionLens = L.MakeLens(
		func(impl mcp.Implementation) string { return impl.Version },
		func(impl mcp.Implementation, version string) mcp.Implementation {
			impl.Version = version
			return impl
		},
	)
	serverOptionsInstructionsLens = L.MakeLens(
		func(opts mcp.ServerOptions) string { return opts.Instructions },
		func(opts mcp.ServerOptions, instructions string) mcp.ServerOptions {
			opts.Instructions = instructions
			return opts
		},
	)
	cfgEndpointLens = L.MakeLens(
		func(c Config) string { return c.GraphQLEndpoint },
		func(c Config, endpoint string) Config {
			c.GraphQLEndpoint = endpoint
			return c
		},
	)
	cfgStatelessLens = L.MakeLens(
		func(c Config) bool { return c.Stateless },
		func(c Config, stateless bool) Config {
			c.Stateless = stateless
			return c
		},
	)
	cfgOriginsLens = L.MakeLens(
		func(c Config) []string { return c.AllowedOrigins },
		func(c Config, origins []string) Config {
			c.AllowedOrigins = origins
			return c
		},
	)
	cfgRateLimitLens = L.MakeLens(
		func(c Config) int { return c.RateLimit },
		func(c Config, limit int) Config {
			c.RateLimit = limit
			return c
		},
	)
	cfgLinkoutsLens = L.MakeLens(
		func(c Config) linkout.Links { return c.Linkouts },
		func(c Config, links linkout.Links) Config {
			c.Linkouts = links
			return c
		},
	)
	streamStatelessLens = L.MakeLens(
		func(o mcp.StreamableHTTPOptions) bool { return o.Stateless },
		func(o mcp.StreamableHTTPOptions, stateless bool) mcp.StreamableHTTPOptions {
			o.Stateless = stateless
			return o
		},
	)
	streamJSONResponseLens = L.MakeLens(
		func(o mcp.StreamableHTTPOptions) bool { return o.JSONResponse },
		func(o mcp.StreamableHTTPOptions, jsonResponse bool) mcp.StreamableHTTPOptions {
			o.JSONResponse = jsonResponse
			return o
		},
	)
	streamLocalhostLens = L.MakeLens(
		func(o mcp.StreamableHTTPOptions) bool { return o.DisableLocalhostProtection },
		func(o mcp.StreamableHTTPOptions, disable bool) mcp.StreamableHTTPOptions {
			o.DisableLocalhostProtection = disable
			return o
		},
	)
	streamCancellationLens = L.MakeLens(
		func(o mcp.StreamableHTTPOptions) bool { return o.PropagateRequestCancellation },
		func(o mcp.StreamableHTTPOptions, propagate bool) mcp.StreamableHTTPOptions {
			o.PropagateRequestCancellation = propagate
			return o
		},
	)
)

// NewMCPServer creates an mcp.Server registered with all dictyBase stock tools.
func NewMCPServer(client *graphql.Client, links linkout.Links) *mcp.Server {
	srv := mcp.NewServer(
		new(F.Pipe2(
			mcp.Implementation{},
			implNameLens.Set(serverName),
			implVersionLens.Set(serverVersion),
		)),
		new(F.Pipe1(
			mcp.ServerOptions{},
			serverOptionsInstructionsLens.Set(serverInstructions),
		)),
	)
	srv.AddReceivingMiddleware(ToolCallLogger(slog.Default()))
	tools.RegisterAll(srv, client, links)
	prompts.RegisterAll(srv)
	return srv
}

func normalizeOrigins(origins []string) []string {
	hasOrigins := predarrays.IsNonEmpty[string]()
	return F.Pipe3(
		origins,
		A.Map(strings.TrimSpace),
		O.FromPredicate(hasOrigins),
		O.GetOrElse(F.Constant([]string{wildcardOrigin})),
	)
}

func newCORSHandler(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: normalizeOrigins(allowedOrigins),
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Last-Event-ID",
			"Mcp-Session-Id",
			"MCP-Protocol-Version",
		},
		ExposedHeaders: []string{
			"Mcp-Session-Id",
			"MCP-Protocol-Version",
		},
		AllowCredentials: true,
		MaxAge:           corsMaxAgeSeconds,
	})
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(healthzBody))
}

func resolveRateLimit(limit int) int {
	return F.Pipe2(
		limit,
		O.FromPredicate(P.IsNonZero[int]()),
		O.GetOrElse(F.Constant(defaultRateLimit)),
	)
}

func clientIPKey(r *http.Request) (string, error) {
	return F.Pipe2(
		r.Context(),
		middleware.GetClientIP,
		httprate.CanonicalizeIP,
	), nil
}

// NewHTTPHandler constructs the chi router serving /healthz and /mcp (Streamable HTTP + chi middlewares).
func NewHTTPHandler(cfg Config) http.Handler {
	gqlClient := graphql.NewClient(cfgEndpointLens.Get(cfg), nil)

	mcpSrv := NewMCPServer(gqlClient, cfgLinkoutsLens.Get(cfg))
	streamHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpSrv },
		new(F.Pipe4(
			mcp.StreamableHTTPOptions{},
			streamStatelessLens.Set(cfgStatelessLens.Get(cfg)),
			streamJSONResponseLens.Set(true),
			streamLocalhostLens.Set(true),
			streamCancellationLens.Set(true),
		)),
	)

	corsMiddleware := newCORSHandler(cfgOriginsLens.Get(cfg))
	rateLimiter := httprate.LimitBy(
		resolveRateLimit(cfgRateLimitLens.Get(cfg)),
		rateLimitWindow,
		clientIPKey,
	)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.ClientIPFromXFF(
		"127.0.0.0/8",
		"::1/128",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	))
	router.Use(corsMiddleware)
	router.Get("/healthz", healthzHandler)
	router.With(rateLimiter).Handle("/mcp", streamHandler)
	return router
}
