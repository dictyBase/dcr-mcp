package main

import (
	"context"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	ORD "github.com/IBM/fp-go/v2/ord"
	S "github.com/IBM/fp-go/v2/string"
	eithererr "github.com/dictyBase/fp-go-loom/eithererr"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/server"
	"github.com/urfave/cli/v3"
)

const (
	defaultGraphQLEndpoint = "https://graphql.dictybase.dev/graphql"
	defaultHost            = "0.0.0.0"
	defaultPort            = 8080
	maxTCPPort             = 65535
	defaultRateLimit       = 120

	flagGraphQLEndpoint = "graphql-endpoint"
	flagHost            = "host"
	flagPort            = "port"
	flagStateless       = "stateless"
	flagAllowedOrigins  = "allowed-origins"
	flagRateLimit       = "rate-limit"
)

// ServerActionInput is the point-free boundary seed for RunServer.
type ServerActionInput struct {
	Context context.Context
	Command *cli.Command
}

type serverState struct {
	ctx context.Context
	cmd *cli.Command
	cfg server.Config
}

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

	isPositiveInt = F.Pipe1(1, ORD.Geq(predord.IntOrd))

	errEmptyEndpoint = eithererr.ConstErr[string](
		"graphql-endpoint must not be empty",
	)
	errInvalidPort = eithererr.ConstErr[int](
		"port must be between 0 and 65535",
	)
	errInvalidRateLimit = eithererr.ConstErr[int](
		"rate-limit must be at least 1",
	)
)

func seedServerState(in ServerActionInput) serverState {
	return serverState{
		ctx: in.Context,
		cmd: in.Command,
		cfg: server.Config{
			GraphQLEndpoint: in.Command.String(flagGraphQLEndpoint),
			Host:            in.Command.String(flagHost),
			Port:            in.Command.Int(flagPort),
			Stateless:       in.Command.Bool(flagStateless),
			AllowedOrigins:  in.Command.StringSlice(flagAllowedOrigins),
			RateLimit:       in.Command.Int(flagRateLimit),
		},
	}
}

func validateEndpoint(cfg server.Config) E.Either[error, string] {
	return F.Pipe2(
		cfg.GraphQLEndpoint,
		strings.TrimSpace,
		E.FromPredicate(S.IsNonEmpty, errEmptyEndpoint),
	)
}

func validatePort(cfg server.Config) E.Either[error, int] {
	return F.Pipe1(
		cfg.Port,
		E.FromPredicate(
			predord.IntBetweenInclusive(0, maxTCPPort),
			errInvalidPort,
		),
	)
}

func validateRateLimit(cfg server.Config) E.Either[error, int] {
	return F.Pipe1(
		cfg.RateLimit,
		E.FromPredicate(isPositiveInt, errInvalidRateLimit),
	)
}

func validateConfig(cfg server.Config) E.Either[error, server.Config] {
	return F.Pipe3(
		E.Of[error](cfg),
		E.Bind(cfgEndpointLens.Set, validateEndpoint),
		E.Bind(cfgPortLens.Set, validatePort),
		E.Bind(cfgRateLimitLens.Set, validateRateLimit),
	)
}

func validateServerState(s serverState) E.Either[error, serverState] {
	return F.Pipe1(
		E.Of[error](s),
		E.Bind(
			stateCfgLens.Set,
			F.Flow2(stateCfgLens.Get, validateConfig),
		),
	)
}

// RunServer validates CLI flags and runs the Streamable HTTP MCP server.
func RunServer(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		ServerActionInput{Context: ctx, Command: cmd},
		seedServerState,
		IOE.Of[error, serverState],
		IOE.ChainEitherK(validateServerState),
		IOE.Chain(func(s serverState) IOE.IOEither[error, struct{}] {
			return server.ListenAndServe(s.ctx, s.cfg)
		}),
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
		Name:  "dcr-mcp",
		Usage: "dictyBase Stock Center Model Context Protocol (MCP) server",
		Commands: []*cli.Command{
			{
				Name:    "start-server",
				Aliases: []string{"serve", "start"},
				Usage:   "Start the dictyBase Stock Center Streamable HTTP MCP server",
				Flags:   serverFlags(),
				Action:  RunServer,
			},
		},
	}
}
