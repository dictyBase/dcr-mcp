package main

import (
	"context"
	E "github.com/IBM/fp-go/v2/either"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const testGraphQLEndpoint = "https://graphql.dictycr.dev/graphql"

func validatedServerState(t *testing.T, args ...string) (serverState, error) {
	t.Helper()
	var state serverState
	var validateErr error
	app := NewApp()
	app.Commands[0].Action = func(ctx context.Context, cmd *cli.Command) error {
		seeded := seedServerState(ServerActionInput{Context: ctx, Command: cmd})
		state, validateErr = E.Unwrap(validateServerState(seeded))
		return nil
	}
	require.NoError(t, app.Run(context.Background(), args))
	return state, validateErr
}

func TestCLILinkoutBaseURL(t *testing.T) {
	t.Run("default base and independent endpoint and host", func(t *testing.T) {
		state, err := validatedServerState(t,
			appName,
			startServerCmd,
			"--"+flagGraphQLEndpoint, testGraphQLEndpoint,
			"--"+flagHost, "127.0.0.1",
			"--"+flagPort, "0",
		)
		require.NoError(t, err)
		require.Equal(t, "https://dictybase.dev/gene/DDB_G0283907",
			state.cfg.Linkouts.GeneURL("DDB_G0283907"))
		require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS0350840",
			state.cfg.Linkouts.StrainURL("DBS0350840"))
		require.Equal(t, testGraphQLEndpoint, state.cfg.GraphQLEndpoint)
		require.Equal(t, "127.0.0.1", state.cfg.Host)
	})

	t.Run("custom origin with a terminal slash", func(t *testing.T) {
		state, err := validatedServerState(t,
			appName,
			startServerCmd,
			"--"+flagGraphQLEndpoint, testGraphQLEndpoint,
			"--"+flagLinkoutBaseURL, "https://catalog.example.org/",
		)
		require.NoError(t, err)
		require.Equal(t, "https://catalog.example.org/gene/DDB_G0283907",
			state.cfg.Linkouts.GeneURL("DDB_G0283907"))
		require.Equal(t, "https://catalog.example.org/publication/17177856",
			state.cfg.Linkouts.PublicationURL("17177856"))
	})

	for _, alias := range []string{"serve", "start"} {
		t.Run("alias "+alias, func(t *testing.T) {
			state, err := validatedServerState(t,
				appName,
				alias,
				"--"+flagGraphQLEndpoint, testGraphQLEndpoint,
				"--"+flagLinkoutBaseURL, "https://catalog.example.org:8443",
			)
			require.NoError(t, err)
			require.Equal(t, "https://catalog.example.org:8443/stockcenter/plasmids/DBP0391806",
				state.cfg.Linkouts.PlasmidURL("DBP0391806"))
		})
	}
}

func TestCLIRejectsInvalidLinkoutBeforeListen(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "explicit empty", raw: ""},
		{name: "whitespace", raw: "   "},
		{name: "path prefix", raw: "https://catalog.example.org/prefix"},
		{name: "query", raw: "https://catalog.example.org?x=1"},
		{name: "fragment", raw: "https://catalog.example.org#part"},
		{name: "unsupported scheme", raw: "ftp://catalog.example.org"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			err := app.Run(context.Background(), []string{
				appName,
				startServerCmd,
				"--" + flagGraphQLEndpoint, testGraphQLEndpoint,
				"--" + flagHost, "203.0.113.1",
				"--" + flagPort, "0",
				"--" + flagLinkoutBaseURL, tt.raw,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, flagLinkoutBaseURL)
			require.NotContains(t, err.Error(), "listen on")
		})
	}
}

func TestCLIStartServer(t *testing.T) {
	t.Run("start-server starts and shuts down cleanly on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		app := NewApp()
		err := app.Run(ctx, []string{
			appName,
			startServerCmd,
			"--" + flagHost, "127.0.0.1",
			"--" + flagPort, "0",
			"--" + flagGraphQLEndpoint, testGraphQLEndpoint,
		})
		require.NoError(t, err)
	})

	t.Run("start-server rejects missing graphql endpoint and kubernetes service host", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--" + flagHost, "127.0.0.1",
			"--" + flagPort, "0",
		})
		require.Error(t, err)
	})

	t.Run("resolveGraphQLEndpoint builds URL from GRAPHQL_API_SERVICE_HOST and PORT", func(t *testing.T) {
		t.Setenv("GRAPHQL_API_SERVICE_HOST", "10.96.42.10")
		t.Setenv("GRAPHQL_API_SERVICE_PORT", "8080")

		var got string
		app := NewApp()
		app.Commands[0].Action = func(_ context.Context, cmd *cli.Command) error {
			got = resolveGraphQLEndpoint(cmd)
			return nil
		}
		err := app.Run(context.Background(), []string{appName, startServerCmd})
		require.NoError(t, err)
		require.Equal(t, "http://10.96.42.10:8080/graphql", got)
	})

	t.Run("resolveGraphQLEndpoint builds URL from GRAPHQL_API_PORT_8080_TCP_ADDR and PORT", func(t *testing.T) {
		t.Setenv("GRAPHQL_API_PORT_8080_TCP_ADDR", "10.96.42.11")
		t.Setenv("GRAPHQL_API_PORT_8080_TCP_PORT", "9090")

		var got string
		app := NewApp()
		app.Commands[0].Action = func(_ context.Context, cmd *cli.Command) error {
			got = resolveGraphQLEndpoint(cmd)
			return nil
		}
		err := app.Run(context.Background(), []string{appName, startServerCmd})
		require.NoError(t, err)
		require.Equal(t, "http://10.96.42.11:9090/graphql", got)
	})

	t.Run("resolveGraphQLEndpoint prefers --graphql-endpoint over k8s service env vars", func(t *testing.T) {
		t.Setenv("GRAPHQL_API_SERVICE_HOST", "10.96.42.10")
		t.Setenv("GRAPHQL_API_SERVICE_PORT", "8080")

		var got string
		app := NewApp()
		app.Commands[0].Action = func(_ context.Context, cmd *cli.Command) error {
			got = resolveGraphQLEndpoint(cmd)
			return nil
		}
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--" + flagGraphQLEndpoint, "https://custom.example.org/graphql",
		})
		require.NoError(t, err)
		require.Equal(t, "https://custom.example.org/graphql", got)
	})

	t.Run("resolveGraphQLEndpoint prefers GRAPHQL_ENDPOINT over k8s service env vars", func(t *testing.T) {
		t.Setenv("GRAPHQL_ENDPOINT", "https://env.example.org/graphql")
		t.Setenv("GRAPHQL_API_SERVICE_HOST", "10.96.42.10")
		t.Setenv("GRAPHQL_API_SERVICE_PORT", "8080")

		var got string
		app := NewApp()
		app.Commands[0].Action = func(_ context.Context, cmd *cli.Command) error {
			got = resolveGraphQLEndpoint(cmd)
			return nil
		}
		err := app.Run(context.Background(), []string{appName, startServerCmd})
		require.NoError(t, err)
		require.Equal(t, "https://env.example.org/graphql", got)
	})

	t.Run("start-server rejects empty graphql-endpoint", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--" + flagGraphQLEndpoint, "   ",
		})
		require.Error(t, err)
	})

	t.Run("start-server rejects out-of-range port", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--" + flagGraphQLEndpoint, testGraphQLEndpoint,
			"--" + flagPort, "70000",
		})
		require.Error(t, err)
	})

	t.Run("start-server rejects non-positive rate-limit", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--" + flagGraphQLEndpoint, testGraphQLEndpoint,
			"--" + flagRateLimit, "0",
		})
		require.Error(t, err)
	})
}
