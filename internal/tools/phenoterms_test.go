package tools_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestSuggestPhenotypeTerms(t *testing.T) {
	session := newToolsSession(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "suggest_phenotype_terms",
		Arguments: map[string]any{testQueryKey: "growth rate"},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	out := res.StructuredContent.(map[string]any)
	require.Equal(t, []any{testGrowthPheno, "decreased growth rate"}, out["terms"])
}
