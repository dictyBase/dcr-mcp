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
								"doi": "10.1/x",
								"authors": [{"last_name": "Phillips"}]
							}]
						}]
					}
				}
			}`))
		case strings.Contains(req.Query, "geneGeneralInformation") && req.Variables["gene"] == "pkaC":
			_, _ = w.Write([]byte(`{"data":{"geneGeneralInformation":{"id":"DDB_G0283907","name_description":["pkaC = Protein Kinase A, Catalytic subunit"],"gene_product":"cAMP-dependent protein kinase A, catalytic subunit","synonyms":["PKA","DdPK3"],"description":""}}}`))
		case strings.Contains(req.Query, "geneGeneralInformation"):
			_, _ = w.Write([]byte(`{"data":{"geneGeneralInformation":null}}`))
		case strings.Contains(req.Query, "listStrainsWithGene") && req.Variables["gene"] == "DDB_G0283907":
			_, _ = w.Write([]byte(`{"data":{"listStrainsWithGene":[{"id":"DBS0236783","label":"pkaC-","summary":"pkaC null","in_stock":true}]}}`))
		case strings.Contains(req.Query, "listStrainsWithGene"):
			_, _ = w.Write([]byte(`{"errors":[{"message":"could not find any strains","path":["listStrainsWithGene"],"extensions":{"code":"NotFound"}}],"data":{"listStrainsWithGene":[]}}`))
		case strings.Contains(req.Query, "publication("):
			_, _ = w.Write([]byte(`{"data":{"publication":{"id":"17177856","doi":"10.1111/x","title":"A paper","abstract":"Abstract text","journal":"Differentiation","pub_date":"2006-12-01T00:00:00.000Z","volume":"74","pages":"1-10","pub_type":"journal_article","authors":[{"last_name":"Smith","first_name":"A","initials":"A","rank":"1"}]}}}`))
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
		require.Equal(t, "10.1/x", res.Strains[0].Publications[0].DOI)
	})
}

func TestGeneQueries(t *testing.T) {
	srv := newStubGraphQLServer(t)
	defer srv.Close()
	client := graphql.NewClient(srv.URL, srv.Client())
	ctx := context.Background()

	t.Run("GetGeneInfo resolves a gene name", func(t *testing.T) {
		got, err := E.Unwrap(ioeutils.ToEither(client.GetGeneInfo(graphql.GeneParams{Ctx: ctx, Gene: "pkaC"})))
		require.NoError(t, err)
		require.Equal(t, "DDB_G0283907", got.ID)
		require.Equal(t, []string{"PKA", "DdPK3"}, got.Synonyms)
	})

	t.Run("GetGeneInfo returns Left for unknown gene", func(t *testing.T) {
		_, err := E.Unwrap(ioeutils.ToEither(client.GetGeneInfo(graphql.GeneParams{Ctx: ctx, Gene: "zzz"})))
		require.ErrorContains(t, err, `gene "zzz" not found`)
	})

	t.Run("ListStrainsWithGene returns strains", func(t *testing.T) {
		got, err := E.Unwrap(ioeutils.ToEither(client.ListStrainsWithGene(graphql.GeneParams{Ctx: ctx, Gene: "DDB_G0283907"})))
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, "DBS0236783", got[0].ID)
	})

	t.Run("ListStrainsWithGene maps NotFound to empty list", func(t *testing.T) {
		got, err := E.Unwrap(ioeutils.ToEither(client.ListStrainsWithGene(graphql.GeneParams{Ctx: ctx, Gene: "DDB_G0000000"})))
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got)
	})

	t.Run("GetPublication", func(t *testing.T) {
		got, err := E.Unwrap(ioeutils.ToEither(client.GetPublication(graphql.IDParams{Ctx: ctx, ID: "17177856"})))
		require.NoError(t, err)
		require.Equal(t, "Abstract text", got.Abstract)
		require.Equal(t, "Smith", got.Authors[0].LastName)
	})
}
