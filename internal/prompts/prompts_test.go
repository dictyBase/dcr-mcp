package prompts_test

import (
	"context"
	"testing"

	"github.com/dictybase/dcr-mcp/internal/prompts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func newPromptSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	srv := mcp.NewServer(&mcp.Implementation{Name: "dcr-mcp", Version: "test"}, nil)
	prompts.RegisterAll(srv)
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	_, err := srv.Connect(ctx, t1, nil)
	require.NoError(t, err)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "test"}, nil).Connect(ctx, t2, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestPrompts(t *testing.T) {
	session := newPromptSession(t)
	ctx := context.Background()

	list, err := session.ListPrompts(ctx, nil)
	require.NoError(t, err)
	require.Len(t, list.Prompts, 4)
	for _, p := range list.Prompts {
		require.NotEmpty(t, p.Title)
		require.Len(t, p.Arguments, 1)
		require.True(t, p.Arguments[0].Required)
	}

	t.Run("argument is substituted", func(t *testing.T) {
		res, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
			Name:      "find_stocks_for_gene",
			Arguments: map[string]string{"gene": "pkaC"},
		})
		require.NoError(t, err)
		require.Len(t, res.Messages, 1)
		text := res.Messages[0].Content.(*mcp.TextContent).Text
		require.Contains(t, text, "pkaC")
		require.NotContains(t, text, "{{gene}}")
		require.Contains(t, text, "find_stocks_by_gene")
	})

	t.Run("missing argument errors", func(t *testing.T) {
		_, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
			Name:      "find_stocks_for_gene",
			Arguments: map[string]string{"gene": "  "},
		})
		require.Error(t, err)
	})
}
