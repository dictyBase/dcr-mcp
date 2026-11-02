//go:build live

package graphql_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/stretchr/testify/require"
)

const defaultLiveEndpoint = "https://graphql.dictybase.dev/graphql"

func liveClient(t *testing.T) (*graphql.Client, context.Context) {
	t.Helper()
	endpoint := os.Getenv("DCR_GRAPHQL_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultLiveEndpoint
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	return graphql.NewClient(endpoint, nil), ctx
}

// TestLiveSchema fails when the upstream schema drifts from the client query
// structs, for example a renamed field (strain_label → stock_label).
func TestLiveSchema(t *testing.T) {
	client, ctx := liveClient(t)

	t.Run("searchStocks", func(t *testing.T) {
		_, err := E.Unwrap(ioeutils.ToEither(client.SearchStocks(graphql.SearchParams{
			Ctx: ctx, Query: "pkaC", Limit: 2, Entity: graphql.StockEntityAll,
		})))
		require.NoError(t, err)
	})

	t.Run("listStockSuggestions", func(t *testing.T) {
		_, err := E.Unwrap(ioeutils.ToEither(client.ListStockSuggestions(graphql.AutocompleteParams{
			Ctx: ctx, Query: "act", Limit: 2, Entity: graphql.StockEntityAll,
		})))
		require.NoError(t, err)
	})

	t.Run("searchPhenotypes", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.SearchPhenotypes(graphql.PhenotypeSearchParams{
			Ctx: ctx, Query: "vegetative growth", Limit: 5,
		})))
		require.NoError(t, err)
		require.NotEmpty(t, res.Results)
	})

	t.Run("listStrainsWithAnnotation accepts a phenotype term", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.ListStrainsWithPhenotype(graphql.PhenotypeListParams{
			Ctx: ctx, PhenotypeTerm: "abolished vegetative growth", Limit: 3,
		})))
		require.NoError(t, err)
		require.NotEmpty(t, res.Strains)
	})

	for _, term := range []string{"in suspension", "motility assay"} {
		t.Run("listStrainsWithAnnotation rejects "+term, func(t *testing.T) {
			_, err := E.Unwrap(ioeutils.ToEither(client.ListStrainsWithPhenotype(graphql.PhenotypeListParams{
				Ctx: ctx, PhenotypeTerm: term, Limit: 3,
			})))
			require.True(t, errors.Is(err, graphql.ErrNotPhenotypeTerm), "%v", err)
		})
	}

	t.Run("strain", func(t *testing.T) {
		_, err := E.Unwrap(ioeutils.ToEither(client.GetStrain(graphql.IDParams{Ctx: ctx, ID: "DBS0235587"})))
		require.NoError(t, err)
	})

	t.Run("plasmid", func(t *testing.T) {
		_, err := E.Unwrap(ioeutils.ToEither(client.GetPlasmid(graphql.IDParams{Ctx: ctx, ID: "DBP0391806"})))
		require.NoError(t, err)
	})
}
