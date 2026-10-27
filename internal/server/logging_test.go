package server_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/dictybase/dcr-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type echoInput struct {
	Msg string `json:"msg"`
}

func TestToolCallLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	srv := mcp.NewServer(&mcp.Implementation{Name: "t", Version: testVersion}, nil)
	srv.AddReceivingMiddleware(server.ToolCallLogger(logger))
	mcp.AddTool(srv, &mcp.Tool{Name: "echo"}, func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		_ echoInput,
	) (*mcp.CallToolResult, struct{}, error) {
		return nil, struct{}{}, nil
	})

	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	_, err := srv.Connect(ctx, t1, nil)
	require.NoError(t, err)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "test"}, nil).Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	_, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "echo", Arguments: map[string]any{"msg": "secret"}})
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, `"method":"tools/call"`)
	require.Contains(t, out, `"tool":"echo"`)
	require.Contains(t, out, `"duration_ms"`)
	require.NotContains(t, out, "secret", "arguments must not be logged")
}
