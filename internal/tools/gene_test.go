package tools_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestFindStocksByGene(t *testing.T) {
	session := newToolsSession(t)
	ctx := context.Background()

	t.Run("known gene", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "find_stocks_by_gene",
			Arguments: map[string]any{"gene": "pkaC"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		out, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok)
		require.Equal(t, "DDB_G0283907", out["gene"].(map[string]any)["id"])
		strains := out["strains"].([]any)
		require.Len(t, strains, 1)
		require.Equal(t, "https://dictycr.org/strains/DBS0236783", strains[0].(map[string]any)["url"])
		require.Len(t, out["plasmids"].([]any), 1)
		require.NotEmpty(t, out["plasmid_match_note"])
	})

	t.Run("unknown gene is a tool error", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "find_stocks_by_gene",
			Arguments: map[string]any{"gene": "zzz"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}
