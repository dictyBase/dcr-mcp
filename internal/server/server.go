// Package server provides the Streamable HTTP MCP server, CORS middleware, and health endpoint.
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
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName        = "dcr-mcp"
	serverVersion     = "1.0.0"
	wildcardOrigin    = "*"
	allowedMethods    = "GET, POST, DELETE, OPTIONS"
	allowedHeaders    = "Content-Type, Accept, Authorization, Mcp-Session-Id, MCP-Protocol-Version, Last-Event-ID"
	exposedHeaders    = "Mcp-Session-Id, MCP-Protocol-Version"
	healthzBody       = `{"status":"ok"}`
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

type corsExchange struct {
	writer  http.ResponseWriter
	request *http.Request
	next    http.Handler
}

var (
	isWildcardOrigin = predord.StrEq(wildcardOrigin)
	isOptionsMethod  = predord.StrEq(http.MethodOptions)
	isOptionsRequest = F.Pipe1(
		isOptionsMethod,
		P.ContraMap(func(ex corsExchange) string { return ex.request.Method }),
	)
	isServerClosedErr = func(err error) bool {
		return errors.Is(err, http.ErrServerClosed)
	}
	corsArms = P.Fold(
		func(ex corsExchange) struct{} {
			ex.next.ServeHTTP(ex.writer, ex.request)
			return struct{}{}
		},
		func(ex corsExchange) struct{} {
			ex.writer.WriteHeader(http.StatusNoContent)
			return struct{}{}
		},
	)
	dispatchCORSRequest = F.Pipe1(isOptionsRequest, corsArms)
)

// NewMCPServer creates an mcp.Server registered with all dictyBase stock tools.
func NewMCPServer(client *graphql.Client) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)
	tools.RegisterAll(srv, client)
	return srv
}

func matchOrigin(allowed []string, reqOrigin string) string {
	hasWildcard := F.Pipe1(allowed, A.Any(isWildcardOrigin))
	matchesReq := F.Pipe2(
		reqOrigin,
		predord.StrEq,
		func(pred P.Predicate[string]) bool {
			return F.Pipe1(allowed, A.Any(pred))
		},
	)
	return F.Pipe2(
		allowed,
		O.FromPredicate(func([]string) bool {
			return !hasWildcard && matchesReq
		}),
		O.Fold(
			F.Constant(wildcardOrigin),
			F.Constant1[[]string](reqOrigin),
		),
	)
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

func withCORS(allowedOrigins []string, next http.Handler) http.Handler {
	normalized := normalizeOrigins(allowedOrigins)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := matchOrigin(normalized, r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
		_ = dispatchCORSRequest(corsExchange{
			writer:  w,
			request: r,
			next:    next,
		})
	})
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(healthzBody))
}

// NewHTTPHandler constructs the HTTP mux serving /healthz and /mcp (Streamable HTTP + CORS).
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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)
	mux.Handle("/mcp", withCORS(cfg.AllowedOrigins, streamHandler))
	return mux
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

// Ensure E is available for extensions.
var _ = E.Of[error, struct{}]
