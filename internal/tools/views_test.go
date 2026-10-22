package tools_test

import (
	"testing"

	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/stretchr/testify/require"
)

func strainFixture() graphql.StrainDetail {
	return graphql.StrainDetail{
		ID:           "DBS0350840",
		Label:        "triA-",
		Phenotypes:   make([]graphql.PhenotypeItem, 12),
		Publications: make([]graphql.PublicationRef, 7),
	}
}

func TestStrainViewFor(t *testing.T) {
	t.Run("summary caps phenotypes and publications but reports full counts", func(t *testing.T) {
		view := F.Pipe1(strainFixture(), tools.StrainViewFor("summary"))
		require.Len(t, view.Phenotypes, 10)
		require.Len(t, view.Publications, 5)
		require.Equal(t, 12, view.PhenotypeCount)
		require.Equal(t, 7, view.PublicationCount)
		require.Equal(t, "summary", view.Detail)
		require.Equal(t, "https://dictycr.org/strains/DBS0350840", view.URL)
	})

	t.Run("full keeps everything", func(t *testing.T) {
		view := F.Pipe1(strainFixture(), tools.StrainViewFor("full"))
		require.Len(t, view.Phenotypes, 12)
		require.Len(t, view.Publications, 7)
		require.Equal(t, "full", view.Detail)
	})

	t.Run("empty slices stay non-nil", func(t *testing.T) {
		view := F.Pipe1(graphql.StrainDetail{ID: "DBS1"}, tools.StrainViewFor("summary"))
		require.NotNil(t, view.Phenotypes)
		require.NotNil(t, view.Publications)
	})
}

func TestPlasmidViewFor(t *testing.T) {
	fixture := graphql.PlasmidDetail{
		ID:           "DBP0391806",
		Sequence:     "ATGCATGCAT",
		Publications: make([]graphql.PublicationRef, 6),
	}

	t.Run("summary drops sequence and keeps its length", func(t *testing.T) {
		view := F.Pipe1(fixture, tools.PlasmidViewFor("summary"))
		require.Empty(t, view.Sequence)
		require.Equal(t, 10, view.SequenceLength)
		require.Len(t, view.Publications, 5)
		require.Equal(t, 6, view.PublicationCount)
		require.Equal(t, "https://dictycr.org/plasmids/DBP0391806", view.URL)
	})

	t.Run("full keeps sequence", func(t *testing.T) {
		view := F.Pipe1(fixture, tools.PlasmidViewFor("full"))
		require.Equal(t, "ATGCATGCAT", view.Sequence)
		require.Len(t, view.Publications, 6)
	})
}
