package main

import (
	"context"
	"errors"
	"fmt"
	A "github.com/IBM/fp-go/v2/array"
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
	actionInputContextLens = L.MakeLens(
		func(x ServerActionInput) context.Context { return x.Context },
		func(x ServerActionInput, value context.Context) ServerActionInput {
			x.Context = value
			return x
		},
	)
	actionInputCommandLens = L.MakeLens(
		func(x ServerActionInput) *cli.Command { return x.Command },
		func(x ServerActionInput, value *cli.Command) ServerActionInput {
			x.Command = value
			return x
		},
	)
	stringFlagNameLens = L.MakeLens(
		func(x cli.StringFlag) string { return x.Name },
		func(x cli.StringFlag, value string) cli.StringFlag {
			x.Name = value
			return x
		},
	)
	stringFlagValueLens = L.MakeLens(
		func(x cli.StringFlag) string { return x.Value },
		func(x cli.StringFlag, value string) cli.StringFlag {
			x.Value = value
			return x
		},
	)
	stringFlagUsageLens = L.MakeLens(
		func(x cli.StringFlag) string { return x.Usage },
		func(x cli.StringFlag, value string) cli.StringFlag {
			x.Usage = value
			return x
		},
	)
	stringFlagSourcesLens = L.MakeLens(
		func(x cli.StringFlag) cli.ValueSourceChain { return x.Sources },
		func(x cli.StringFlag, value cli.ValueSourceChain) cli.StringFlag {
			x.Sources = value
			return x
		},
	)
	intFlagNameLens = L.MakeLens(
		func(x cli.IntFlag) string { return x.Name },
		func(x cli.IntFlag, value string) cli.IntFlag {
			x.Name = value
			return x
		},
	)
	intFlagValueLens = L.MakeLens(
		func(x cli.IntFlag) int { return x.Value },
		func(x cli.IntFlag, value int) cli.IntFlag {
			x.Value = value
			return x
		},
	)
	intFlagUsageLens = L.MakeLens(
		func(x cli.IntFlag) string { return x.Usage },
		func(x cli.IntFlag, value string) cli.IntFlag {
			x.Usage = value
			return x
		},
	)
	intFlagSourcesLens = L.MakeLens(
		func(x cli.IntFlag) cli.ValueSourceChain { return x.Sources },
		func(x cli.IntFlag, value cli.ValueSourceChain) cli.IntFlag {
			x.Sources = value
			return x
		},
	)
	boolFlagNameLens = L.MakeLens(
		func(x cli.BoolFlag) string { return x.Name },
		func(x cli.BoolFlag, value string) cli.BoolFlag {
			x.Name = value
			return x
		},
	)
	boolFlagValueLens = L.MakeLens(
		func(x cli.BoolFlag) bool { return x.Value },
		func(x cli.BoolFlag, value bool) cli.BoolFlag {
			x.Value = value
			return x
		},
	)
	boolFlagUsageLens = L.MakeLens(
		func(x cli.BoolFlag) string { return x.Usage },
		func(x cli.BoolFlag, value string) cli.BoolFlag {
			x.Usage = value
			return x
		},
	)
	boolFlagSourcesLens = L.MakeLens(
		func(x cli.BoolFlag) cli.ValueSourceChain { return x.Sources },
		func(x cli.BoolFlag, value cli.ValueSourceChain) cli.BoolFlag {
			x.Sources = value
			return x
		},
	)
	stringSliceFlagNameLens = L.MakeLens(
		func(x cli.StringSliceFlag) string { return x.Name },
		func(x cli.StringSliceFlag, value string) cli.StringSliceFlag {
			x.Name = value
			return x
		},
	)
	stringSliceFlagValueLens = L.MakeLens(
		func(x cli.StringSliceFlag) []string { return x.Value },
		func(x cli.StringSliceFlag, value []string) cli.StringSliceFlag {
			x.Value = value
			return x
		},
	)
	stringSliceFlagUsageLens = L.MakeLens(
		func(x cli.StringSliceFlag) string { return x.Usage },
		func(x cli.StringSliceFlag, value string) cli.StringSliceFlag {
			x.Usage = value
			return x
		},
	)
	stringSliceFlagSourcesLens = L.MakeLens(
		func(x cli.StringSliceFlag) cli.ValueSourceChain { return x.Sources },
		func(x cli.StringSliceFlag, value cli.ValueSourceChain) cli.StringSliceFlag {
			x.Sources = value
			return x
		},
	)
	stringFlagAliasesLens = L.MakeLens(
		func(x cli.StringFlag) []string { return x.Aliases },
		func(x cli.StringFlag, value []string) cli.StringFlag {
			x.Aliases = value
			return x
		},
	)
	commandNameLens = L.MakeLens(
		func(x cli.Command) string { return x.Name },
		func(x cli.Command, value string) cli.Command {
			x.Name = value
			return x
		},
	)
	commandUsageLens = L.MakeLens(
		func(x cli.Command) string { return x.Usage },
		func(x cli.Command, value string) cli.Command {
			x.Usage = value
			return x
		},
	)
	commandAliasesLens = L.MakeLens(
		func(x cli.Command) []string { return x.Aliases },
		func(x cli.Command, value []string) cli.Command {
			x.Aliases = value
			return x
		},
	)
	commandCommandsLens = L.MakeLens(
		func(x cli.Command) []*cli.Command { return x.Commands },
		func(x cli.Command, value []*cli.Command) cli.Command {
			x.Commands = value
			return x
		},
	)
	commandFlagsLens = L.MakeLens(
		func(x cli.Command) []cli.Flag { return x.Flags },
		func(x cli.Command, value []cli.Flag) cli.Command {
			x.Flags = value
			return x
		},
	)
	commandActionLens = L.MakeLens(
		func(x cli.Command) cli.ActionFunc { return x.Action },
		func(x cli.Command, value cli.ActionFunc) cli.Command {
			x.Action = value
			return x
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

	stateHostLens      = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigHostLens))
	stateStatelessLens = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigStatelessLens))
	stateOriginsLens   = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigOriginsLens))
	stateLinkoutsLens  = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigLinkoutsLens))
	stateCfgLens       = L.MakeLens(
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

	stateEndpointLens = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigEndpointLens))
	statePortLens     = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigPortLens))
	stateRateLens     = F.Pipe1(stateCfgLens, L.Compose[serverState](server.ConfigRateLimitLens))

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

// fp-go:allow-struct-literal http.Server holds a mutex, so a value lens cannot copy it
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
	return F.Pipe9(
		ServerActionInput{},
		actionInputContextLens.Set(ctx),
		actionInputCommandLens.Set(cmd),
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
			stringFlagNameLens.Set(flagLinkoutBaseURL),
			stringFlagValueLens.Set(linkout.DefaultBaseURL),
			stringFlagUsageLens.Set("Base URL for gene, strain, plasmid, and publication links"),
		)),
		new(F.Pipe4(
			cli.StringFlag{},
			stringFlagNameLens.Set(flagGraphQLEndpoint),
			stringFlagAliasesLens.Set([]string{"endpoint"}),
			stringFlagUsageLens.Set("dictyBase GraphQL server endpoint URL"),
			stringFlagSourcesLens.Set(cli.EnvVars("GRAPHQL_ENDPOINT")),
		)),
		new(F.Pipe3(
			cli.StringFlag{},
			stringFlagNameLens.Set(flagGraphQLHost),
			stringFlagUsageLens.Set("Kubernetes GraphQL service host"),
			stringFlagSourcesLens.Set(cli.EnvVars("GRAPHQL_API_SERVICE_HOST", "GRAPHQL_API_PORT_8080_TCP_ADDR")),
		)),
		new(F.Pipe4(
			cli.StringFlag{},
			stringFlagNameLens.Set(flagGraphQLPort),
			stringFlagValueLens.Set(defaultGraphQLPort),
			stringFlagUsageLens.Set("Kubernetes GraphQL service port"),
			stringFlagSourcesLens.Set(cli.EnvVars("GRAPHQL_API_SERVICE_PORT", "GRAPHQL_API_PORT_8080_TCP_PORT")),
		)),
		new(F.Pipe4(
			cli.StringFlag{},
			stringFlagNameLens.Set(flagHost),
			stringFlagValueLens.Set(defaultHost),
			stringFlagUsageLens.Set("Host address for the HTTP MCP server to bind"),
			stringFlagSourcesLens.Set(cli.EnvVars("MCP_HOST")),
		)),
		new(F.Pipe4(
			cli.IntFlag{},
			intFlagNameLens.Set(flagPort),
			intFlagValueLens.Set(defaultPort),
			intFlagUsageLens.Set("TCP port for the HTTP MCP server"),
			intFlagSourcesLens.Set(cli.EnvVars("MCP_PORT")),
		)),
		new(F.Pipe4(
			cli.BoolFlag{},
			boolFlagNameLens.Set(flagStateless),
			boolFlagValueLens.Set(true),
			boolFlagUsageLens.Set("Run Streamable HTTP handler in stateless mode for Kubernetes ingress"),
			boolFlagSourcesLens.Set(cli.EnvVars("MCP_STATELESS")),
		)),
		new(F.Pipe4(
			cli.StringSliceFlag{},
			stringSliceFlagNameLens.Set(flagAllowedOrigins),
			stringSliceFlagValueLens.Set([]string{"*"}),
			stringSliceFlagUsageLens.Set("Allowed CORS origins for web clients"),
			stringSliceFlagSourcesLens.Set(cli.EnvVars("ALLOWED_ORIGINS")),
		)),
		new(F.Pipe4(
			cli.IntFlag{},
			intFlagNameLens.Set(flagRateLimit),
			intFlagValueLens.Set(defaultRateLimit),
			intFlagUsageLens.Set("Maximum /mcp requests per client IP per minute"),
			intFlagSourcesLens.Set(cli.EnvVars("MCP_RATE_LIMIT")),
		)),
	}
}

// NewApp constructs the root urfave/cli/v3 command for dcr-mcp.
func NewApp() *cli.Command {
	return new(F.Pipe3(
		cli.Command{},
		commandNameLens.Set(appName),
		commandUsageLens.Set("dictyBase Stock Center Model Context Protocol (MCP) server"),
		commandCommandsLens.Set(A.Of(new(F.Pipe5(
			cli.Command{},
			commandNameLens.Set(startServerCmd),
			commandAliasesLens.Set([]string{"serve", "start"}),
			commandUsageLens.Set("Start the dictyBase Stock Center Streamable HTTP MCP server"),
			commandFlagsLens.Set(serverFlags()),
			commandActionLens.Set(RunServer),
		)))),
	))
}
