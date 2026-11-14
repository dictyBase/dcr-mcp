package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/dictybase/dcr-mcp/internal/tools"
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
	require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS0350840", payload.Results[0].URL)
}

func searchHit(id, label, field string, entity graphql.StockEntityType) graphql.StockSearchHit {
	return graphql.StockSearchHit{
		ID:          id,
		Entity:      entity,
		Field:       field,
		DisplayText: "matched " + field,
		StockLabel:  label,
	}
}

func searchDocs(hits ...graphql.StockSearchHit) tools.SearchDocuments {
	return F.Pipe1(graphql.SearchResult{Results: hits}, tools.ToSearchDocuments(linkout.Links{}))
}

func TestToSearchDocuments(t *testing.T) {
	t.Run("label present gives label and id title", func(t *testing.T) {
		docs := searchDocs(searchHit(testStrainID, testStrainLabel, "label", graphql.StockEntityStrain))
		require.Equal(t, []tools.SearchDocument{{
			ID:    testStrainID,
			Title: "triA- (DBS0350840)",
			URL:   "https://dictybase.dev/stockcenter/strains/DBS0350840",
		}}, docs.Results)
	})

	t.Run("empty label gives the id as title, not the matched text", func(t *testing.T) {
		docs := searchDocs(searchHit(testPlasmidID, "", "summary", graphql.StockEntityPlasmid))
		require.Len(t, docs.Results, 1)
		require.Equal(t, testPlasmidID, docs.Results[0].Title)
		require.Equal(t, "https://dictybase.dev/stockcenter/plasmids/DBP0391806", docs.Results[0].URL)
	})

	t.Run("unknown entity gives an empty url", func(t *testing.T) {
		docs := searchDocs(searchHit(testStrainID, testStrainLabel, "label", graphql.StockEntityAll))
		require.Len(t, docs.Results, 1)
		require.Empty(t, docs.Results[0].URL)
	})

	t.Run("duplicate ids collapse to the first hit", func(t *testing.T) {
		docs := searchDocs(
			searchHit(testStrainID, testStrainLabel, "label", graphql.StockEntityStrain),
			searchHit(testPlasmidID, "pA15", "name", graphql.StockEntityPlasmid),
			searchHit(testStrainID, "other", "summary", graphql.StockEntityStrain),
		)
		require.Len(t, docs.Results, 2)
		require.Equal(t, testStrainID, docs.Results[0].ID)
		require.Equal(t, "triA- (DBS0350840)", docs.Results[0].Title)
		require.Equal(t, testPlasmidID, docs.Results[1].ID)
	})

	t.Run("no hits gives a non-nil empty slice", func(t *testing.T) {
		docs := searchDocs()
		require.NotNil(t, docs.Results)
		require.Empty(t, docs.Results)
	})
}

func TestSearchProjectionsUseConfiguredBase(t *testing.T) {
	links, err := E.Unwrap(linkout.Parse("https://example.org"))
	require.NoError(t, err)
	result := graphql.SearchResult{Results: []graphql.StockSearchHit{
		searchHit(testStrainID, testStrainLabel, "label", graphql.StockEntityStrain),
		searchHit(testPlasmidID, "", "name", graphql.StockEntityPlasmid),
	}}
	strainURL := "https://example.org/stockcenter/strains/DBS0350840"
	plasmidURL := "https://example.org/stockcenter/plasmids/DBP0391806"

	docs := F.Pipe1(result, tools.ToSearchDocuments(links))
	require.Equal(t, strainURL, docs.Results[0].URL)
	require.Equal(t, plasmidURL, docs.Results[1].URL)

	view := F.Pipe1(result, tools.ToSearchResultView(links))
	require.Equal(t, strainURL, view.Results[0].URL)
	require.Equal(t, plasmidURL, view.Results[1].URL)
}

func TestLabelWithID(t *testing.T) {
	require.Equal(t, "triA- (DBS0350840)", tools.LabelWithID(testStrainLabel, testStrainID))
	require.Equal(t, testStrainID, tools.LabelWithID("", testStrainID))
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
		require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS0350840", doc["url"])
		require.Contains(t, doc["text"], "## triA- (DBS0350840)")
		require.Equal(t, "strain", doc["metadata"].(map[string]any)["entity"])
	})

	t.Run("plasmid, lower-case id", func(t *testing.T) {
		res, doc := fetchDoc(t, "dbp0391806")
		require.False(t, res.IsError)
		require.Equal(t, "https://dictybase.dev/stockcenter/plasmids/DBP0391806", doc["url"])
	})

	t.Run("unknown prefix", func(t *testing.T) {
		res, _ := fetchDoc(t, "XYZ123")
		require.True(t, res.IsError)
	})
}
