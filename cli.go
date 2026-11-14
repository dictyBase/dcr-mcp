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
	MO "github.com/dictyBase/fp-go-loom/matchopt"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/linkout"
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
		linkoutBaseURL string

		ctx      context.Context
		cmd      *cli.Command
		cfg      server.Config
		addr     string
		listener net.Listener
	}
)

const (
	flagLinkoutBaseURL = "linkout-base-url"
	defaultGraphQLPort = "8080"
	defaultHost        = "0.0.0.0"
	defaultPort        = 8080
	maxTCPPort         = 65535
	defaultRateLimit   = 120
	readHeaderTimeout  = 10 * time.Second
	shutdownTimeout    = 5 * time.Second

	appName        = "dcr-mcp"
	startServerCmd = "start-server"

	flagGraphQLEndpoint = "graphql-endpoint"
	flagGraphQLHost     = "graphql-host"
	flagGraphQLPort     = "graphql-port"
	flagHost            = "host"
	flagPort            = "port"
	flagStateless       = "stateless"
	flagAllowedOrigins  = "allowed-origins"
	flagRateLimit       = "rate-limit"
)

var (
	cfgHostLens = L.MakeLens(
		func(c server.Config) string { return c.Host },
		func(c server.Config, host string) server.Config {
			c.Host = host
			return c
		},
	)
	cfgStatelessLens = L.MakeLens(
		func(c server.Config) bool { return c.Stateless },
		func(c server.Config, stateless bool) server.Config {
			c.Stateless = stateless
			return c
		},
	)
	cfgOriginsLens = L.MakeLens(
		func(c server.Config) []string { return c.AllowedOrigins },
		func(c server.Config, origins []string) server.Config {
			c.AllowedOrigins = origins
			return c
		},
	)
	cfgLinkoutsLens = L.MakeLens(
		func(c server.Config) linkout.Links { return c.Linkouts },
		func(c server.Config, links linkout.Links) server.Config {
			c.Linkouts = links
			return c
		},
	)
	stateCtxLens = L.MakeLens(
		func(s serverState) context.Context { return s.ctx },
		func(s serverState, ctx context.Context) serverState {
			s.ctx = ctx
			return s
		},
	)
	stateCmdLens = L.MakeLens(
		func(s serverState) *cli.Command { return s.cmd },
		func(s serverState, cmd *cli.Command) serverState {
			s.cmd = cmd
			return s
		},
	)
	stateAddrLens = L.MakeLens(
		func(s serverState) string { return s.addr },
		func(s serverState, addr string) serverState {
			s.addr = addr
			return s
		},
	)
	stateLinkoutRawLens = L.MakeLens(
		func(s serverState) string { return s.linkoutBaseURL },
		func(s serverState, raw string) serverState {
			s.linkoutBaseURL = raw
			return s
		},
	)

	stateHostLens      = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgHostLens))
	stateStatelessLens = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgStatelessLens))
	stateOriginsLens   = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgOriginsLens))
	stateLinkoutsLens  = F.Pipe1(stateCfgLens, L.Compose[serverState](cfgLinkoutsLens))
	flagNameLens       = L.MakeLens(
		func(f cli.StringFlag) string { return f.Name },
		func(f cli.StringFlag, name string) cli.StringFlag {
			f.Name = name
			return f
		},
	)
	flagValueLens = L.MakeLens(
		func(f cli.StringFlag) string { return f.Value },
		func(f cli.StringFlag, value string) cli.StringFlag {
			f.Value = value
			return f
		},
	)
	flagUsageLens = L.MakeLens(
		func(f cli.StringFlag) string { return f.Usage },
		func(f cli.StringFlag, usage string) cli.StringFlag {
			f.Usage = usage
			return f
		},
	)
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
		"graphql-endpoint or GRAPHQL_API_SERVICE_HOST must be set",
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
	validateServerState = F.Flow5(
		E.Of[error, serverState],
		E.BindL(stateEndpointLens, validateEndpoint),
		E.BindL(statePortLens, validatePort),
		E.BindL(stateRateLens, validateRateLimit),
		E.Bind(
			stateLinkoutsLens.Set,
			F.Flow2(stateLinkoutRawLens.Get, linkout.Parse),
		),
	)

	isEndpointSet = func(cmd *cli.Command) bool {
		return cmd.IsSet(flagGraphQLEndpoint)
	}
	hasGraphQLHost = F.Flow3(
		func(cmd *cli.Command) string { return cmd.String(flagGraphQLHost) },
		strings.TrimSpace,
		S.IsNonEmpty,
	)
	resolveGraphQLEndpoint = MO.Match("", []O.Kleisli[*cli.Command, string]{
		MO.Case(isEndpointSet, func(cmd *cli.Command) string {
			return cmd.String(flagGraphQLEndpoint)
		}),
		MO.Case(hasGraphQLHost, buildK8sGraphQLEndpoint),
	})
)

func buildK8sGraphQLEndpoint(cmd *cli.Command) string {
	host := F.Pipe1(cmd.String(flagGraphQLHost), strings.TrimSpace)
	port := F.Pipe3(
		cmd.String(flagGraphQLPort),
		strings.TrimSpace,
		O.FromPredicate(S.IsNonEmpty),
		O.GetOrElse(F.Constant(defaultGraphQLPort)),
	)
	return fmt.Sprintf("http://%s/graphql", net.JoinHostPort(host, port))
}

func seedServerState(in ServerActionInput) serverState {
	host := in.Command.String(flagHost)
	port := in.Command.Int(flagPort)

	return F.Pipe10(
		serverState{},
		stateCtxLens.Set(in.Context),
		stateCmdLens.Set(in.Command),
		stateAddrLens.Set(net.JoinHostPort(host, N.ToString(port))),
		stateLinkoutRawLens.Set(in.Command.String(flagLinkoutBaseURL)),
		stateEndpointLens.Set(resolveGraphQLEndpoint(in.Command)),
		stateHostLens.Set(host),
		statePortLens.Set(port),
		stateStatelessLens.Set(in.Command.Bool(flagStateless)),
		stateOriginsLens.Set(in.Command.StringSlice(flagAllowedOrigins)),
		stateRateLens.Set(in.Command.Int(flagRateLimit)),
	)
}

func bindServerListener(s serverState) IOE.IOEither[error, net.Listener] {
	addr := stateAddrLens.Get(s)

	return F.Pipe1(
		IOE.TryCatchError(func() (net.Listener, error) {
			return net.Listen("tcp", addr)
		}),
		IOE.MapLeft[net.Listener](func(err error) error {
			return fmt.Errorf("listen on %s failed: %w", addr, err)
		}),
	)
}

func serveHTTP(s serverState) IOE.IOEither[error, struct{}] {
	return F.Pipe1(
		IOE.TryCatchError(func() (struct{}, error) {
			srv := &http.Server{
				Handler:           server.NewHTTPHandler(stateCfgLens.Get(s)),
				ReadHeaderTimeout: readHeaderTimeout,
			}
			errCh := make(chan error, 1)
			go func() {
				rawErr := srv.Serve(stateListenerLens.Get(s))
				errCh <- normalizeServeErr(rawErr)
			}()

			select {
			case err := <-errCh:
				return struct{}{}, err
			case <-stateCtxLens.Get(s).Done():
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
		new(F.Pipe3(
			cli.StringFlag{},
			flagNameLens.Set(flagLinkoutBaseURL),
			flagValueLens.Set(linkout.DefaultBaseURL),
			flagUsageLens.Set("Base URL for gene, strain, plasmid, and publication links"),
		)),
		&cli.StringFlag{
			Name:    flagGraphQLEndpoint,
			Aliases: []string{"endpoint"},
			Usage:   "dictyBase GraphQL server endpoint URL",
			Sources: cli.EnvVars("GRAPHQL_ENDPOINT"),
		},
		&cli.StringFlag{
			Name:    flagGraphQLHost,
			Usage:   "Kubernetes GraphQL service host",
			Sources: cli.EnvVars("GRAPHQL_API_SERVICE_HOST", "GRAPHQL_API_PORT_8080_TCP_ADDR"),
		},
		&cli.StringFlag{
			Name:    flagGraphQLPort,
			Value:   defaultGraphQLPort,
			Usage:   "Kubernetes GraphQL service port",
			Sources: cli.EnvVars("GRAPHQL_API_SERVICE_PORT", "GRAPHQL_API_PORT_8080_TCP_PORT"),
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
