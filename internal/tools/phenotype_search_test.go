package tools_test

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/stretchr/testify/require"
)

const (
	testVegGrowth  = "abolished vegetative growth"
	testFieldTerm  = "TERM_LABEL"
	testGrowthWord = "growth"
)

func TestValidateSearchPhenotypesInput(t *testing.T) {
	tests := []struct {
		name      string
		in        tools.SearchPhenotypesInput
		wantQuery string
		wantLimit int
		wantErr   bool
	}{
		{name: "trims and applies default limit", in: tools.SearchPhenotypesInput{Query: "  growth "},
			wantQuery: testGrowthWord, wantLimit: 15},
		{name: "limit 50 accepted", in: tools.SearchPhenotypesInput{Query: testGrowthWord, Limit: 50},
			wantQuery: testGrowthWord, wantLimit: 50},
		{name: "limit 51 rejected", in: tools.SearchPhenotypesInput{Query: testGrowthWord, Limit: 51}, wantErr: true},
		{name: "negative limit rejected", in: tools.SearchPhenotypesInput{Query: testGrowthWord, Limit: -1}, wantErr: true},
		{name: "one character rejected", in: tools.SearchPhenotypesInput{Query: "a"}, wantErr: true},
		{name: "blank rejected", in: tools.SearchPhenotypesInput{Query: testBlankQuery}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := E.Unwrap(tools.ValidateSearchPhenotypesInput(tt.in))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantQuery, got.Query)
			require.Equal(t, tt.wantLimit, got.Limit)
		})
	}
}

func phenoHit(id, field, displayText string) graphql.StockSearchHit {
	return graphql.StockSearchHit{
		ID: id, Entity: graphql.StockEntityStrain, Field: field, DisplayText: displayText,
		StockLabel: id + "-", InStock: true, StockType: "regular",
	}
}

func TestPhenotypeSearchViewGroupsMatchedTexts(t *testing.T) {
	view := tools.ToPhenotypeSearchView(linkout.Links{})(graphql.SearchResult{
		Results: []graphql.StockSearchHit{
			phenoHit("DBS1", testFieldTerm, "growth curve"),
			phenoHit("DBS2", testFieldTerm, testVegGrowth),
			phenoHit("DBS3", testFieldTerm, "growth curve"),
			phenoHit("DBS4", "TERM_DEFINITION", "growth curve"),
		},
		Limit: 15, TotalCount: 4,
	})
	require.Equal(t, 4, view.Count)
	require.Equal(t, []tools.MatchedTextView{
		{Field: testFieldTerm, DisplayText: "growth curve", StrainCount: 2},
		{Field: testFieldTerm, DisplayText: testVegGrowth, StrainCount: 1},
		{Field: "TERM_DEFINITION", DisplayText: "growth curve", StrainCount: 1},
	}, view.MatchedTexts, "first-seen order, grouped by field and text")
	require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS1", view.Results[0].URL)
}

func TestPhenotypeSearchViewEmpty(t *testing.T) {
	view := tools.ToPhenotypeSearchView(linkout.Links{})(graphql.SearchResult{Results: []graphql.StockSearchHit{}})
	require.NotNil(t, view.MatchedTexts)
	require.NotNil(t, view.Results)
	require.Zero(t, view.Count)

	md := tools.RenderPhenotypeSearchMarkdown(view)
	require.Contains(t, md, "Top 0 strain matches (maximum 50, no further pages).")
	require.Contains(t, md, "None recorded.")
}

func TestRenderPhenotypeSearchMarkdown(t *testing.T) {
	gwdi := phenoHit("DBS2", testFieldTerm, testVegGrowth)
	gwdi.StockType = "gwdi"
	md := F.Pipe2(
		graphql.SearchResult{Results: []graphql.StockSearchHit{
			phenoHit("DBS1", testFieldTerm, testVegGrowth),
			gwdi,
		}},
		tools.ToPhenotypeSearchView(linkout.Links{}),
		tools.RenderPhenotypeSearchMarkdown,
	)
	require.Contains(t, md, "Top 2 strain matches")
	require.Contains(t, md, "abolished vegetative growth (TERM\\_LABEL) — 2 strains")
	require.Contains(t, md, "[DBS1](https://dictybase.dev/stockcenter/strains/DBS1)")
	require.Contains(t, md, "· gwdi")
	require.NotContains(t, md, "· regular")
}

func TestRenderPhenotypeStrainListMarkdown(t *testing.T) {
	last := F.Pipe2(
		graphql.PhenotypeStrainList{Strains: []graphql.PhenotypeStrainItem{{ID: "DBS9", Label: "x-"}}},
		tools.ToPhenotypeStrainListView(linkout.Links{})(testVegGrowth),
		tools.RenderPhenotypeStrainListMarkdown,
	)
	require.Contains(t, last, `1 strains with phenotype term "abolished vegetative growth" on this page. More pages: no.`)
	require.Contains(t, last, "[DBS9](https://dictybase.dev/stockcenter/strains/DBS9)")

	more := F.Pipe2(
		graphql.PhenotypeStrainList{Strains: []graphql.PhenotypeStrainItem{}, NextCursor: 7},
		tools.ToPhenotypeStrainListView(linkout.Links{})(testVegGrowth),
		tools.RenderPhenotypeStrainListMarkdown,
	)
	require.Contains(t, more, "More pages: yes")
}
