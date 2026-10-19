package graphql_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/stretchr/testify/require"
)

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func newStubGraphQLServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "listStockSuggestions"):
			_, _ = w.Write([]byte(`{
				"data": {
					"listStockSuggestions": {
						"limit": 2,
						"totalCount": 1,
						"suggestions": [
							{
								"id": "DBS0350026",
								"entity": "STRAIN",
								"field": "GENOTYPE",
								"display_text": "act15:mip1",
								"score": 1013.5
							}
						]
					}
				}
			}`))
		case strings.Contains(req.Query, "searchStocks"):
			_, _ = w.Write([]byte(`{
				"data": {
					"searchStocks": {
						"limit": 2,
						"totalCount": 1,
						"results": [
							{
								"id": "DBS0350840",
								"entity": "STRAIN",
								"field": "GENOTYPE",
								"display_text": "act15:triA",
								"score": 1011.4,
								"summary": "triA null cells",
								"strain_label": "triA-",
								"in_stock": true
							}
						]
					}
				}
			}`))
		case strings.Contains(req.Query, "strain("):
			if req.Variables["id"] == "MISSING" {
				_, _ = w.Write([]byte(`{
					"errors": [{"message": "strain not found"}]
				}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"data": {
					"strain": {
						"id": "DBS0350840",
						"label": "triA-",
						"summary": "triA null cells",
						"species": "Dictyostelium discoideum",
						"systematic_name": "DBS0350840",
						"plasmid": "",
						"dbxrefs": null,
						"in_stock": true,
						"genotypes": ["axeA2"],
						"mutagenesis_method": "",
						"genetic_modification": "exogenous mutation",
						"names": ["triA-"],
						"characteristics": ["axenic"],
						"parent": {"id": "", "label": ""},
						"depositor": {"first_name": "Dicty", "last_name": "Stock Center"},
						"genes": [{"name": "triA"}],
						"publications": [{
							"id": "17177856",
							"pub_date": "2006-12-01T00:00:00.000Z",
							"title": "Trishanku",
							"journal": "Differentiation",
							"volume": "74",
							"pages": "596-607",
							"doi": "10.1111/test",
							"authors": [{"last_name": "Jaiswal"}]
						}],
						"phenotypes": []
					}
				}
			}`))
		case strings.Contains(req.Query, "plasmid("):
			_, _ = w.Write([]byte(`{
				"data": {
					"plasmid": {
						"id": "DBP0391806",
						"name": "pDGB_O2_0006",
						"summary": "GoldenBraid cloning system",
						"image_map": "",
						"sequence": "ATGC",
						"keywords": null,
						"genbank_accession": "",
						"in_stock": true,
						"dbxrefs": null,
						"depositor": {"first_name": "Gadi", "last_name": "Shaulsky"},
						"genes": [],
						"publications": []
					}
				}
			}`))
		case strings.Contains(req.Query, "listStrainsWithAnnotation"):
			_, _ = w.Write([]byte(`{
				"data": {
					"listStrainsWithAnnotation": {
						"totalCount": 1,
						"nextCursor": 1728057313873,
						"strains": [{
							"id": "DBS0349965",
							"label": "bzpN-",
							"summary": "bzpN null mutant",
							"in_stock": true,
							"genes": [{"name": "bzpN"}],
							"publications": [{
								"id": "21760904",
								"pub_date": "2011-07-07T00:00:00.000Z",
								"title": "BzpN slows proliferation",
								"journal": "PloS one",
								"volume": "6",
								"pages": "e21765",
								"authors": [{"last_name": "Phillips"}]
							}]
						}]
					}
				}
			}`))
		}
	}))
}

func TestClientOperations(t *testing.T) {
	srv := newStubGraphQLServer(t)
	defer srv.Close()

	client := graphql.NewClient(srv.URL, srv.Client())
	ctx := context.Background()

	t.Run("ListStockSuggestions returns suggestions", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.ListStockSuggestions(graphql.AutocompleteParams{
			Ctx:    ctx,
			Query:  "act",
			Limit:  2,
			Entity: graphql.StockEntityAll,
		})))
		require.NoError(t, err)
		require.Equal(t, 1, res.TotalCount)
		require.Len(t, res.Suggestions, 1)
		require.Equal(t, "DBS0350026", res.Suggestions[0].ID)
		require.Equal(t, "act15:mip1", res.Suggestions[0].DisplayText)
	})

	t.Run("SearchStocks returns search hits", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.SearchStocks(graphql.SearchParams{
			Ctx:    ctx,
			Query:  "act",
			Limit:  2,
			Entity: graphql.StockEntityStrain,
		})))
		require.NoError(t, err)
		require.Equal(t, 1, res.TotalCount)
		require.Len(t, res.Results, 1)
		require.Equal(t, "DBS0350840", res.Results[0].ID)
		require.True(t, res.Results[0].InStock)
	})

	t.Run("GetStrain normalizes nil slices and returns strain detail", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.GetStrain(graphql.IDParams{
			Ctx: ctx,
			ID:  "DBS0350840",
		})))
		require.NoError(t, err)
		require.Equal(t, "DBS0350840", res.ID)
		require.Equal(t, "triA-", res.Label)
		require.NotNil(t, res.Dbxrefs)
		require.Empty(t, res.Dbxrefs)
		require.NotNil(t, res.Phenotypes)
		require.Empty(t, res.Phenotypes)
		require.Len(t, res.Publications, 1)
	})

	t.Run("GetStrain propagates GraphQL error as Left", func(t *testing.T) {
		_, err := E.Unwrap(ioeutils.ToEither(client.GetStrain(graphql.IDParams{
			Ctx: ctx,
			ID:  "MISSING",
		})))
		require.Error(t, err)
		require.ErrorContains(t, err, "strain not found")
	})

	t.Run("GetPlasmid normalizes nil slices and returns plasmid detail", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.GetPlasmid(graphql.IDParams{
			Ctx: ctx,
			ID:  "DBP0391806",
		})))
		require.NoError(t, err)
		require.Equal(t, "DBP0391806", res.ID)
		require.Equal(t, "pDGB_O2_0006", res.Name)
		require.NotNil(t, res.Keywords)
		require.Empty(t, res.Keywords)
		require.NotNil(t, res.Dbxrefs)
		require.Empty(t, res.Dbxrefs)
	})

	t.Run("ListStrainsWithPhenotype returns strains and nextCursor", func(t *testing.T) {
		res, err := E.Unwrap(ioeutils.ToEither(client.ListStrainsWithPhenotype(graphql.PhenotypeListParams{
			Ctx:       ctx,
			Phenotype: "increased growth rate",
			Cursor:    0,
			Limit:     2,
		})))
		require.NoError(t, err)
		require.Equal(t, 1, res.TotalCount)
		require.Equal(t, int64(1728057313873), res.NextCursor)
		require.Len(t, res.Strains, 1)
		require.Equal(t, "DBS0349965", res.Strains[0].ID)
	})
}
