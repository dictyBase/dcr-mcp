package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestChatGPTSearch(t *testing.T) {
	session := newToolsSession(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search",
		Arguments: map[string]any{testQueryKey: testActQuery},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	var payload struct {
		Results []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(text.Text), &payload))
	require.Len(t, payload.Results, 1)
	require.Equal(t, testStrainID, payload.Results[0].ID)
	require.Equal(t, "triA- (DBS0350840)", payload.Results[0].Title)
	require.Equal(t, "https://dictycr.org/strains/DBS0350840", payload.Results[0].URL)
}

func fetchDoc(t *testing.T, id string) (*mcp.CallToolResult, map[string]any) {
	t.Helper()
	session := newToolsSession(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fetch",
		Arguments: map[string]any{"id": id},
	})
	require.NoError(t, err)
	var doc map[string]any
	if !res.IsError {
		text, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		require.NoError(t, json.Unmarshal([]byte(text.Text), &doc))
	}
	return res, doc
}

func TestChatGPTFetch(t *testing.T) {
	t.Run("strain", func(t *testing.T) {
		res, doc := fetchDoc(t, testStrainID)
		require.False(t, res.IsError)
		require.Equal(t, testStrainID, doc["id"])
		require.Equal(t, "triA- (DBS0350840)", doc["title"])
		require.Equal(t, "https://dictycr.org/strains/DBS0350840", doc["url"])
		require.Contains(t, doc["text"], "## triA- (DBS0350840)")
		require.Equal(t, "strain", doc["metadata"].(map[string]any)["entity"])
	})

	t.Run("plasmid, lower-case id", func(t *testing.T) {
		res, doc := fetchDoc(t, "dbp0391806")
		require.False(t, res.IsError)
		require.Equal(t, "https://dictycr.org/plasmids/DBP0391806", doc["url"])
	})

	t.Run("unknown prefix", func(t *testing.T) {
		res, _ := fetchDoc(t, "XYZ123")
		require.True(t, res.IsError)
	})
}
