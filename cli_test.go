package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCLIStartServer(t *testing.T) {
	t.Run("start-server starts and shuts down cleanly on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		app := NewApp()
		err := app.Run(ctx, []string{
			appName,
			startServerCmd,
			"--host", "127.0.0.1",
			"--port", "0",
			"--graphql-endpoint", "https://graphql.dictybase.dev/graphql",
		})
		require.NoError(t, err)
	})

	t.Run("start-server rejects empty graphql-endpoint", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--graphql-endpoint", "   ",
		})
		require.Error(t, err)
	})

	t.Run("start-server rejects out-of-range port", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--port", "70000",
		})
		require.Error(t, err)
	})

	t.Run("start-server rejects non-positive rate-limit", func(t *testing.T) {
		app := NewApp()
		err := app.Run(context.Background(), []string{
			appName,
			startServerCmd,
			"--rate-limit", "0",
		})
		require.Error(t, err)
	})
}
