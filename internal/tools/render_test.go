package tools_test

import (
	"testing"

	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// parseKinds parses rendered Markdown so a broken document fails the test
// instead of reaching a client.
func parseKinds(t *testing.T, md string) map[ast.NodeKind]int {
	t.Helper()
	kinds := map[ast.NodeKind]int{}
	doc := goldmark.DefaultParser().Parse(text.NewReader([]byte(md)))
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			kinds[n.Kind()]++
		}
		return ast.WalkContinue, nil
	})
	require.NoError(t, err)
	return kinds
}

func TestRenderStrainMarkdown(t *testing.T) {
	detail := graphql.StrainDetail{
		ID: "DBS0350840", Label: "triA-", Summary: "triA null", InStock: true,
		Species: "Dictyostelium discoideum", Genotypes: []string{"axeA2", "triA-"},
		Phenotypes: []graphql.PhenotypeItem{{
			Phenotype:   "aberrant spore morphology",
			Publication: graphql.PhenotypePublicationRef{ID: "123"},
		}},
		Publications: []graphql.PublicationRef{{
			ID: "17177856", Title: "A paper", Journal: "Differentiation",
			PubDate: "2006-12-01T00:00:00Z", DOI: "10.1111/x",
			Authors: []graphql.AuthorRef{{LastName: "Smith"}, {LastName: "Jones"}},
		}},
	}
	md := F.Pipe2(detail, tools.StrainViewFor("summary"), tools.RenderStrainMarkdown)

	require.Contains(t, md, "## triA- (DBS0350840)")
	require.Contains(t, md, "in stock")
	require.Contains(t, md, "https://dictycr.org/strains/DBS0350840")
	require.Contains(t, md, "axeA2; triA-")
	require.Contains(t, md, "aberrant spore morphology")
	require.Contains(t, md, "https://pubmed.ncbi.nlm.nih.gov/17177856/")
	require.Contains(t, md, "https://doi.org/10.1111/x")
	require.Contains(t, md, "Smith, Jones")
	require.Contains(t, md, "(2006)")
	require.Contains(t, md, "Showing 1 of 1 phenotypes and 1 of 1 publications (detail=summary).")

	kinds := parseKinds(t, md)
	require.Equal(t, 3, kinds[ast.KindHeading], "one h2 header plus two h3 section headers")
	require.NotZero(t, kinds[ast.KindList])
	require.NotZero(t, kinds[ast.KindLink])
}

func TestRenderStrainMarkdownEmptySections(t *testing.T) {
	md := F.Pipe2(graphql.StrainDetail{ID: "DBS1"}, tools.StrainViewFor("summary"), tools.RenderStrainMarkdown)
	require.Contains(t, md, "### Phenotypes\nNone recorded.")
	require.Contains(t, md, "not in stock")
	require.NotNil(t, parseKinds(t, md))
}

func TestRenderSearchMarkdown(t *testing.T) {
	md := tools.RenderSearchMarkdown(tools.ToSearchResultView(graphql.SearchResult{
		Results: []graphql.StockSearchHit{{
			ID: "DBS1", Entity: graphql.StockEntityStrain, Field: "SUMMARY",
			DisplayText: "a | b", StrainLabel: "x-", InStock: true,
		}},
		Limit: 15, TotalCount: 1,
	}))
	require.Contains(t, md, "Showing 1 of 1 matches")
	require.Contains(t, md, "[DBS1](https://dictycr.org/strains/DBS1)")
	require.Contains(t, md, `a \| b`)

	kinds := parseKinds(t, md)
	require.NotZero(t, kinds[ast.KindListItem])
}

func TestRenderPlasmidMarkdown(t *testing.T) {
	md := F.Pipe2(
		graphql.PlasmidDetail{ID: "DBP1", Name: "pX", Sequence: "ATGC", GenbankAccession: "AB123"},
		tools.PlasmidViewFor("summary"),
		tools.RenderPlasmidMarkdown,
	)
	require.Contains(t, md, "## pX (DBP1)")
	require.Contains(t, md, "Sequence: 4 bp (call get_plasmid_sequence)")
	require.Contains(t, md, "AB123")
	require.NotNil(t, parseKinds(t, md))
}

func TestRenderEscapesUserText(t *testing.T) {
	md := F.Pipe2(
		graphql.StrainDetail{ID: "DBS1", Label: "a*b_[c]", Summary: "x | y"},
		tools.StrainViewFor("summary"),
		tools.RenderStrainMarkdown,
	)
	require.Contains(t, md, `a\*b\_\[c\]`)
	require.Contains(t, md, `x \| y`)
}
