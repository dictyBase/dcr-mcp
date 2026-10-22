package tools_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestCheckAvailability(t *testing.T) {
	session := newToolsSession(t)
	ctx := context.Background()

	t.Run("mixed ids with one bad prefix", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "check_availability",
			Arguments: map[string]any{"ids": []string{"DBS0350840", " DBP0391806 ", "XYZ1", "DBS0350840"}},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		items := res.StructuredContent.(map[string]any)["items"].([]any)
		require.Len(t, items, 3, "duplicates are removed")

		first := items[0].(map[string]any)
		require.Equal(t, "strain", first["entity"])
		require.Equal(t, true, first["in_stock"])
		require.Equal(t, "https://dictycr.org/strains/DBS0350840", first["url"])

		require.Equal(t, "plasmid", items[1].(map[string]any)["entity"])
		require.Contains(t, items[2].(map[string]any)["error"], "unrecognized stock id")
	})

	t.Run("empty list rejected", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "check_availability",
			Arguments: map[string]any{"ids": []string{" "}},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}
