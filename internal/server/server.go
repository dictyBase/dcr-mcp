// Package server provides the Streamable HTTP MCP server, chi router, CORS middleware, and health endpoint.
package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictybase/dcr-mcp/internal/graphql"
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

// NewMCPServer creates an mcp.Server registered with all dictyBase stock tools.
func NewMCPServer(client *graphql.Client) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, &mcp.ServerOptions{
		Instructions: serverInstructions,
	})
	srv.AddReceivingMiddleware(ToolCallLogger(slog.Default()))
	tools.RegisterAll(srv, client)
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
	gqlClient := graphql.NewClient(cfg.GraphQLEndpoint, nil)
	mcpSrv := NewMCPServer(gqlClient)
	streamHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpSrv },
		&mcp.StreamableHTTPOptions{
			Stateless:                    cfg.Stateless,
			JSONResponse:                 true,
			DisableLocalhostProtection:   true,
			PropagateRequestCancellation: true,
		},
	)

	corsMiddleware := newCORSHandler(cfg.AllowedOrigins)
	rateLimiter := httprate.LimitBy(resolveRateLimit(cfg.RateLimit), rateLimitWindow, clientIPKey)
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
