package tools_test

import (
	"testing"

	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/stretchr/testify/require"
)

func strainFixture() graphql.StrainDetail {
	return graphql.StrainDetail{
		ID:           testStrainID,
		Label:        testStrainLabel,
		Phenotypes:   make([]graphql.PhenotypeItem, 12),
		Publications: make([]graphql.PublicationRef, 7),
	}
}

func TestStrainViewFor(t *testing.T) {
	t.Run("summary caps phenotypes and publications but reports full counts", func(t *testing.T) {
		view := F.Pipe1(strainFixture(), tools.StrainViewFor(linkout.Links{})("summary"))
		require.Len(t, view.Phenotypes, 10)
		require.Len(t, view.Publications, 5)
		require.Equal(t, 12, view.PhenotypeCount)
		require.Equal(t, 7, view.PublicationCount)
		require.Equal(t, "summary", view.Detail)
		require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS0350840", view.URL)
	})

	t.Run("full keeps everything", func(t *testing.T) {
		view := F.Pipe1(strainFixture(), tools.StrainViewFor(linkout.Links{})("full"))
		require.Len(t, view.Phenotypes, 12)
		require.Len(t, view.Publications, 7)
		require.Equal(t, "full", view.Detail)
	})

	t.Run("empty slices stay non-nil", func(t *testing.T) {
		view := F.Pipe1(graphql.StrainDetail{ID: testShortStrainID}, tools.StrainViewFor(linkout.Links{})("summary"))
		require.NotNil(t, view.Phenotypes)
		require.NotNil(t, view.Publications)
	})
}

func TestPlasmidViewFor(t *testing.T) {
	fixture := graphql.PlasmidDetail{
		ID:           testPlasmidID,
		Sequence:     "ATGCATGCAT",
		Publications: make([]graphql.PublicationRef, 6),
	}

	t.Run("summary drops sequence and keeps its length", func(t *testing.T) {
		view := F.Pipe1(fixture, tools.PlasmidViewFor(linkout.Links{})("summary"))
		require.Empty(t, view.Sequence)
		require.Equal(t, 10, view.SequenceLength)
		require.Len(t, view.Publications, 5)
		require.Equal(t, 6, view.PublicationCount)
		require.Equal(t, "https://dictybase.dev/stockcenter/plasmids/DBP0391806", view.URL)
	})

	t.Run("full keeps sequence", func(t *testing.T) {
		view := F.Pipe1(fixture, tools.PlasmidViewFor(linkout.Links{})("full"))
		require.Equal(t, "ATGCATGCAT", view.Sequence)
		require.Len(t, view.Publications, 6)
	})
}

func TestSearchAndPhenotypeViews(t *testing.T) {
	search := tools.ToSearchResultView(linkout.Links{})(graphql.SearchResult{
		Results: []graphql.StockSearchHit{
			{ID: testShortStrainID, Entity: graphql.StockEntityStrain},
			{ID: "DBP1", Entity: graphql.StockEntityPlasmid},
		},
		Limit: 15, TotalCount: 2,
	})
	require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS1", search.Results[0].URL)
	require.Equal(t, "https://dictybase.dev/stockcenter/plasmids/DBP1", search.Results[1].URL)
	require.Equal(t, 2, search.TotalCount)

	pheno := F.Pipe1(
		graphql.PhenotypeStrainList{
			Strains:    []graphql.PhenotypeStrainItem{{ID: "DBS9"}},
			NextCursor: 42,
		},
		tools.ToPhenotypeStrainListView(linkout.Links{})(testGrowthPheno),
	)
	require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS9", pheno.Strains[0].URL)
	require.EqualValues(t, 42, pheno.NextCursor)
	require.True(t, pheno.HasMore)
	require.Equal(t, 1, pheno.PageCount)
	require.Equal(t, testGrowthPheno, pheno.PhenotypeTerm)
}
