package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const testGraphQLEndpoint = "https://graphql.dictycr.dev/graphql"

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
