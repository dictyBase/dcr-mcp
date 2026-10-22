package tools_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestGetPublication(t *testing.T) {
	session := newToolsSession(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_publication",
		Arguments: map[string]any{"id": "17177856"},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	out := res.StructuredContent.(map[string]any)
	require.Equal(t, "https://pubmed.ncbi.nlm.nih.gov/17177856/", out["pubmed_url"])
	require.Equal(t, "https://doi.org/10.1111/x", out["doi_url"])

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "Abstract text")
	require.Contains(t, text, "Smith A")
}
