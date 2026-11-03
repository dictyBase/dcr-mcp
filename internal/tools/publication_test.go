package tools_test

import (
	"context"
	"testing"

	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
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

func TestRenderPublicationMarkdown(t *testing.T) {
	md := tools.RenderPublicationMarkdown(tools.PublicationView{
		PublicationDetail: graphql.PublicationDetail{
			ID: "1", Title: "A *bold* claim", Journal: "Dev Biol", Volume: "74", Pages: "1-10",
			PubDate: "2006-12-01", Abstract: "line one\nline two",
			Authors: []graphql.FullAuthor{
				{LastName: testAuthorSmith, Initials: "A"},
				{LastName: "Jones"},
				{},
			},
		},
		PubMedURL: "https://pubmed.ncbi.nlm.nih.gov/1/",
	})
	require.Equal(t, "## A \\*bold\\* claim\n\n"+
		"Smith A, Jones\n\n"+
		"*Dev Biol* 74:1-10 (2006). [PubMed](https://pubmed.ncbi.nlm.nih.gov/1/)\n\n"+
		"### Abstract\n\n"+
		"line one line two", md)
}
