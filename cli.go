package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	N "github.com/IBM/fp-go/v2/number/integer"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	ORD "github.com/IBM/fp-go/v2/ord"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	eithererr "github.com/dictyBase/fp-go-loom/eithererr"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/server"
	"github.com/urfave/cli/v3"
)

type (
	// ServerActionInput is the point-free boundary seed for RunServer.
	ServerActionInput struct {
		Context context.Context
		Command *cli.Command
	}

	serverState struct {
		ctx      context.Context
		cmd      *cli.Command
		cfg      server.Config
		addr     string
		listener net.Listener
	}
)

const (
	defaultGraphQLEndpoint = "https://graphql.dictybase.dev/graphql"
	defaultHost            = "0.0.0.0"
	defaultPort            = 8080
	maxTCPPort             = 65535
	defaultRateLimit       = 120
	readHeaderTimeout      = 10 * time.Second
	shutdownTimeout        = 5 * time.Second

	appName        = "dcr-mcp"
	startServerCmd = "start-server"

	flagGraphQLEndpoint = "graphql-endpoint"
	flagHost            = "host"
	flagPort            = "port"
	flagStateless       = "stateless"
	flagAllowedOrigins  = "allowed-origins"
	flagRateLimit       = "rate-limit"
)

var (
	cfgEndpointLens = L.MakeLens(
		func(c server.Config) string { return c.GraphQLEndpoint },
		func(c server.Config, ep string) server.Config {
			c.GraphQLEndpoint = ep
			return c
		},
	)
	cfgPortLens = L.MakeLens(
		func(c server.Config) int { return c.Port },
		func(c server.Config, p int) server.Config {
			c.Port = p
			return c
		},
	)
	cfgRateLimitLens = L.MakeLens(
		func(c server.Config) int { return c.RateLimit },
		func(c server.Config, r int) server.Config {
			c.RateLimit = r
			return c
		},
	)
	stateCfgLens = L.MakeLens(
		func(s serverState) server.Config { return s.cfg },
		func(s serverState, c server.Config) serverState {
			s.cfg = c
			return s
		},
	)
	stateListenerLens = L.MakeLens(
		func(s serverState) net.Listener { return s.listener },
		func(s serverState, ln net.Listener) serverState {
			s.listener = ln
			return s
		},
	)

	stateEndpointLens = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgEndpointLens))
	statePortLens     = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgPortLens))
	stateRateLens     = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgRateLimitLens))

	isPositiveInt = F.Pipe1(1, ORD.Geq(predord.IntOrd))

	isServerClosedErr = func(err error) bool {
		return errors.Is(err, http.ErrServerClosed)
	}
	normalizeServeErr = F.Flow2(
		O.FromPredicate(P.Not(isServerClosedErr)),
		O.Fold(
			func() error { return nil },
			F.Identity[error],
		),
	)

	errEmptyEndpoint = eithererr.ConstErr[string](
		"graphql-endpoint must not be empty",
	)
	errInvalidPort = eithererr.ConstErr[int](
		"port must be between 0 and 65535",
	)
	errInvalidRateLimit = eithererr.ConstErr[int](
		"rate-limit must be at least 1",
	)

	validateEndpoint = F.Flow2(
		strings.TrimSpace,
		E.FromPredicate(S.IsNonEmpty, errEmptyEndpoint),
	)
	validatePort = E.FromPredicate(
		predord.IntBetweenInclusive(0, maxTCPPort),
		errInvalidPort,
	)
	validateRateLimit   = E.FromPredicate(isPositiveInt, errInvalidRateLimit)
	validateServerState = F.Flow4(
		E.Of[error, serverState],
		E.BindL(stateEndpointLens, validateEndpoint),
		E.BindL(statePortLens, validatePort),
		E.BindL(stateRateLens, validateRateLimit),
	)
)

func seedServerState(in ServerActionInput) serverState {
	host := in.Command.String(flagHost)
	port := in.Command.Int(flagPort)
	return serverState{
		ctx:  in.Context,
		cmd:  in.Command,
		addr: net.JoinHostPort(host, N.ToString(port)),
		cfg: server.Config{
			GraphQLEndpoint: in.Command.String(flagGraphQLEndpoint),
			Host:            host,
			Port:            port,
			Stateless:       in.Command.Bool(flagStateless),
			AllowedOrigins:  in.Command.StringSlice(flagAllowedOrigins),
			RateLimit:       in.Command.Int(flagRateLimit),
		},
	}
}

func bindServerListener(s serverState) IOE.IOEither[error, net.Listener] {
	return F.Pipe1(
		IOE.TryCatchError(func() (net.Listener, error) {
			return net.Listen("tcp", s.addr)
		}),
		IOE.MapLeft[net.Listener](func(err error) error {
			return fmt.Errorf("listen on %s failed: %w", s.addr, err)
		}),
	)
}

func serveHTTP(s serverState) IOE.IOEither[error, struct{}] {
	return F.Pipe1(
		IOE.TryCatchError(func() (struct{}, error) {
			srv := &http.Server{
				Handler:           server.NewHTTPHandler(s.cfg),
				ReadHeaderTimeout: readHeaderTimeout,
			}
			errCh := make(chan error, 1)
			go func() {
				rawErr := srv.Serve(s.listener)
				errCh <- normalizeServeErr(rawErr)
			}()

			select {
			case err := <-errCh:
				return struct{}{}, err
			case <-s.ctx.Done():
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

// RunServer validates CLI flags and runs the Streamable HTTP MCP server.
func RunServer(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe7(
		ServerActionInput{Context: ctx, Command: cmd},
		seedServerState,
		IOE.Of[error, serverState],
		IOE.ChainEitherK(validateServerState),
		IOE.Bind(stateListenerLens.Set, bindServerListener),
		IOE.Chain(serveHTTP),
		ioeutils.ToEither[error, struct{}],
		E.ToError[struct{}],
	)
}

func serverFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flagGraphQLEndpoint,
			Aliases: []string{"endpoint"},
			Value:   defaultGraphQLEndpoint,
			Usage:   "dictyBase GraphQL server endpoint URL",
			Sources: cli.EnvVars("GRAPHQL_ENDPOINT"),
		},
		&cli.StringFlag{
			Name:    flagHost,
			Value:   defaultHost,
			Usage:   "Host address for the HTTP MCP server to bind",
			Sources: cli.EnvVars("MCP_HOST"),
		},
		&cli.IntFlag{
			Name:    flagPort,
			Value:   defaultPort,
			Usage:   "TCP port for the HTTP MCP server",
			Sources: cli.EnvVars("MCP_PORT"),
		},
		&cli.BoolFlag{
			Name:    flagStateless,
			Value:   true,
			Usage:   "Run Streamable HTTP handler in stateless mode for Kubernetes ingress",
			Sources: cli.EnvVars("MCP_STATELESS"),
		},
		&cli.StringSliceFlag{
			Name:    flagAllowedOrigins,
			Value:   []string{"*"},
			Usage:   "Allowed CORS origins for web clients",
			Sources: cli.EnvVars("ALLOWED_ORIGINS"),
		},
		&cli.IntFlag{
			Name:    flagRateLimit,
			Value:   defaultRateLimit,
			Usage:   "Maximum /mcp requests per client IP per minute",
			Sources: cli.EnvVars("MCP_RATE_LIMIT"),
		},
	}
}

// NewApp constructs the root urfave/cli/v3 command for dcr-mcp.
func NewApp() *cli.Command {
	return &cli.Command{
		Name:  appName,
		Usage: "dictyBase Stock Center Model Context Protocol (MCP) server",
		Commands: []*cli.Command{
			{
				Name:    startServerCmd,
				Aliases: []string{"serve", "start"},
				Usage:   "Start the dictyBase Stock Center Streamable HTTP MCP server",
				Flags:   serverFlags(),
				Action:  RunServer,
			},
		},
	}
}
