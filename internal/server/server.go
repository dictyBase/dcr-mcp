// Package server provides the Streamable HTTP MCP server, chi router, CORS middleware, and health endpoint.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName        = "dcr-mcp"
	serverVersion     = "1.0.0"
	wildcardOrigin    = "*"
	healthzBody       = `{"status":"ok"}`
	corsMaxAgeSeconds = 300
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 5 * time.Second
)

// Config holds HTTP and GraphQL configuration for the MCP server.
type Config struct {
	GraphQLEndpoint string
	Host            string
	Port            int
	Stateless       bool
	AllowedOrigins  []string
}

var isServerClosedErr = func(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}

// NewMCPServer creates an mcp.Server registered with all dictyBase stock tools.
func NewMCPServer(client *graphql.Client) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)
	tools.RegisterAll(srv, client)
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
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware)
	router.Get("/healthz", healthzHandler)
	router.Handle("/mcp", streamHandler)
	return router
}

func normalizeServeErr(err error) error {
	return F.Pipe2(
		err,
		O.FromPredicate(P.Not(isServerClosedErr)),
		O.Fold(
			func() error { return nil },
			F.Identity[error],
		),
	)
}

func bindListener(addr string) IOE.IOEither[error, net.Listener] {
	return F.Pipe1(
		IOE.TryCatchError(func() (net.Listener, error) {
			return net.Listen("tcp", addr)
		}),
		IOE.MapLeft[net.Listener](func(err error) error {
			return fmt.Errorf("listen on %s failed: %w", addr, err)
		}),
	)
}

func serveListener(
	ctx context.Context,
	cfg Config,
	ln net.Listener,
) IOE.IOEither[error, struct{}] {
	return F.Pipe1(
		IOE.TryCatchError(func() (struct{}, error) {
			srv := &http.Server{
				Handler:           NewHTTPHandler(cfg),
				ReadHeaderTimeout: readHeaderTimeout,
			}
			errCh := make(chan error, 1)
			go func() {
				rawErr := srv.Serve(ln)
				errCh <- normalizeServeErr(rawErr)
			}()

			select {
			case err := <-errCh:
				return struct{}{}, err
			case <-ctx.Done():
				shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
				defer cancel()
				_ = srv.Shutdown(shutCtx)
				return struct{}{}, <-errCh
			}
		}),
		IOE.MapLeft[struct{}](func(err error) error {
			return fmt.Errorf("http server error: %w", err)
		}),
	)
}

// ListenAndServe starts the HTTP server for cfg and shuts down when ctx is canceled.
func ListenAndServe(ctx context.Context, cfg Config) IOE.IOEither[error, struct{}] {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	return F.Pipe2(
		addr,
		bindListener,
		IOE.Chain(func(ln net.Listener) IOE.IOEither[error, struct{}] {
			return serveListener(ctx, cfg, ln)
		}),
	)
}
