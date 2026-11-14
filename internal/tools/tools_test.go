package tools_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type toolsStubRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

const (
	testVersion       = "1.0.0"
	testActQuery      = "act"
	testBlankQuery    = "   "
	testGrowthPheno   = "increased growth rate"
	testGrowthQuery   = "growth rate"
	testStrainID      = "DBS0350840"
	testPlasmidID     = "DBP0391806"
	testQueryKey      = "query"
	testShortStrainID = "DBS1"
	testStrainLabel   = "triA-"
)

var (
	implNameLens = L.MakeLens(
		func(impl mcp.Implementation) string { return impl.Name },
		func(impl mcp.Implementation, name string) mcp.Implementation {
			impl.Name = name
			return impl
		},
	)
	implVersionLens = L.MakeLens(
		func(impl mcp.Implementation) string { return impl.Version },
		func(impl mcp.Implementation, version string) mcp.Implementation {
			impl.Version = version
			return impl
		},
	)
)

func toImplementation(name, version string) *mcp.Implementation {
	return new(F.Pipe2(
		mcp.Implementation{},
		implNameLens.Set(name),
		implVersionLens.Set(version),
	))
}

func TestValidateAutocompleteInput(t *testing.T) {
	tests := []struct {
		name       string
		in         tools.AutocompleteStocksInput
		wantQuery  string
		wantLimit  int
		wantEntity graphql.StockEntityType
		wantErr    bool
	}{
		{
			name:       "valid minimal query applies defaults",
			in:         tools.AutocompleteStocksInput{Query: "  act  "},
			wantQuery:  testActQuery,
			wantLimit:  5,
			wantEntity: graphql.StockEntityAll,
		},
		{
			name:       "case-insensitive entity and explicit limit",
			in:         tools.AutocompleteStocksInput{Query: "dbs035", Limit: 15, Entity: "strain"},
			wantQuery:  "dbs035",
			wantLimit:  15,
			wantEntity: graphql.StockEntityStrain,
		},
		{
			name:    "whitespace-only query rejected",
			in:      tools.AutocompleteStocksInput{Query: "    "},
			wantErr: true,
		},
		{
			name:    "two-char query rejected for autocomplete",
			in:      tools.AutocompleteStocksInput{Query: "ab"},
			wantErr: true,
		},
		{
			name:    "limit over 50 rejected",
			in:      tools.AutocompleteStocksInput{Query: testActQuery, Limit: 51},
			wantErr: true,
		},
		{
			name:    "negative limit rejected",
			in:      tools.AutocompleteStocksInput{Query: testActQuery, Limit: -1},
			wantErr: true,
		},
		{
			name:    "invalid entity rejected",
			in:      tools.AutocompleteStocksInput{Query: testActQuery, Entity: "GENE"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := E.Unwrap(tools.ValidateAutocompleteInput(tt.in))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantQuery, got.Query)
			require.Equal(t, tt.wantLimit, got.Limit)
			require.Equal(t, string(tt.wantEntity), got.Entity)
		})
	}
}

func TestValidateSearchInput(t *testing.T) {
	tests := []struct {
		name       string
		in         tools.SearchStocksInput
		wantQuery  string
		wantLimit  int
		wantEntity graphql.StockEntityType
		wantErr    bool
	}{
		{
			name:       "two-char query allowed with defaults",
			in:         tools.SearchStocksInput{Query: " cs "},
			wantQuery:  "cs",
			wantLimit:  15,
			wantEntity: graphql.StockEntityAll,
		},
		{
			name:       "plasmid entity normalized",
			in:         tools.SearchStocksInput{Query: "pDGB", Limit: 100, Entity: "plasmid"},
			wantQuery:  "pDGB",
			wantLimit:  100,
			wantEntity: graphql.StockEntityPlasmid,
		},
		{
			name:    "one-char query rejected",
			in:      tools.SearchStocksInput{Query: "a"},
			wantErr: true,
		},
		{
			name:    "whitespace-only query rejected",
			in:      tools.SearchStocksInput{Query: testBlankQuery},
			wantErr: true,
		},
		{
			name:    "limit over 100 rejected",
			in:      tools.SearchStocksInput{Query: testActQuery, Limit: 101},
			wantErr: true,
		},
		{
			name:    "invalid entity rejected",
			in:      tools.SearchStocksInput{Query: testActQuery, Entity: "PROTEIN"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := E.Unwrap(tools.ValidateSearchInput(tt.in))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantQuery, got.Query)
			require.Equal(t, tt.wantLimit, got.Limit)
			require.Equal(t, string(tt.wantEntity), got.Entity)
		})
	}
}

func TestValidateIDAndPhenotypeInputs(t *testing.T) {
	t.Run("ValidateStrainIDInput trims and requires non-empty", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: " DBS0350840 "}))
		require.NoError(t, err)
		require.Equal(t, testStrainID, got.ID)

		_, err = E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: testBlankQuery}))
		require.Error(t, err)
	})

	t.Run("ValidatePlasmidIDInput trims and requires non-empty", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: " DBP0391806 "}))
		require.NoError(t, err)
		require.Equal(t, testPlasmidID, got.ID)

		_, err = E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: ""}))
		require.Error(t, err)
	})

	t.Run("ValidatePhenotypeListInput normalizes the term and checks bounds", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			PhenotypeTerm: "  increased   growth rate ",
		}))
		require.NoError(t, err)
		require.Equal(t, testGrowthPheno, got.PhenotypeTerm)
		require.Equal(t, 0, got.Cursor)
		require.Equal(t, 10, got.Limit)

		kept, err := E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			PhenotypeTerm: "decreased cAMP Response",
		}))
		require.NoError(t, err)
		require.Equal(t, "decreased cAMP Response", kept.PhenotypeTerm, "case is kept")

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			PhenotypeTerm: testBlankQuery,
		}))
		require.Error(t, err)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			PhenotypeTerm: testGrowthPheno,
			Cursor:        -1,
		}))
		require.Error(t, err)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			PhenotypeTerm: testGrowthPheno,
			Limit:         101,
		}))
		require.Error(t, err)
	})

	t.Run("ValidateStrainIDInput defaults detail to summary and rejects unknown", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: testStrainID}))
		require.NoError(t, err)
		require.Equal(t, "summary", got.Detail)

		got, err = E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: testStrainID, Detail: " FULL "}))
		require.NoError(t, err)
		require.Equal(t, "full", got.Detail)

		_, err = E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: testStrainID, Detail: "everything"}))
		require.Error(t, err)
	})

	t.Run("ValidatePlasmidIDInput defaults detail to summary", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: testPlasmidID}))
		require.NoError(t, err)
		require.Equal(t, "summary", got.Detail)

		_, err = E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: testPlasmidID, Detail: "huge"}))
		require.Error(t, err)
	})
}

func newToolsStubGraphQLServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req toolsStubRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")
		serveToolsGraphQLStub(w, req)
	}))
}

func writeToolsJSON(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

func serveToolsGraphQLStub(w http.ResponseWriter, req toolsStubRequest) {
	switch {
	case strings.Contains(req.Query, "listStockSuggestions"):
		writeToolsJSON(w, `{
			"data": {
				"listStockSuggestions": {
					"limit": 5,
					"totalCount": 1,
					"suggestions": [{
						"id": "DBS0350026",
						"entity": "STRAIN",
						"field": "GENOTYPE",
						"display_text": "act15:mip1",
						"score": 1013.5
					}]
				}
			}
		}`)
	case strings.Contains(req.Query, "geneGeneralInformation"),
		strings.Contains(req.Query, "listStrainsWithGene"):
		serveToolsGeneStub(w, req)
	case strings.Contains(req.Query, "searchPhenotypes"):
		writeToolsJSON(w, `{
			"data": {
				"searchPhenotypes": {
					"limit": 15,
					"totalCount": 2,
					"results": [
						{"id": "DBS0235587", "entity": "STRAIN", "field": "TERM_LABEL",
						 "display_text": "abolished vegetative growth", "score": 503.4,
						 "summary": "chcA null", "stock_label": "chcA-", "in_stock": true,
						 "stock_type": "regular"},
						{"id": "DBS0236827", "entity": "STRAIN", "field": "TERM_LABEL",
						 "display_text": "abolished vegetative growth", "score": 503.4,
						 "summary": "proA/proB", "stock_label": "[AS]proA/proB-", "in_stock": true,
						 "stock_type": "gwdi"}
					]
				}
			}
		}`)
	case strings.Contains(req.Query, "searchStocks"):
		serveToolsSearchStub(w, req)
	case strings.Contains(req.Query, "publication("):
		writeToolsJSON(w, `{
			"data": {
				"publication": {
					"id": "17177856",
					"doi": "10.1111/x",
					"title": "A paper",
					"abstract": "Abstract text",
					"journal": "Differentiation",
					"pub_date": "2006-12-01T00:00:00.000Z",
					"volume": "74",
					"pages": "1-10",
					"pub_type": "journal_article",
					"authors": [{
						"last_name": "Smith",
						"first_name": "A",
						"initials": "A",
						"rank": "1"
					}]
				}
			}
		}`)
	default:
		serveToolsStockDetailStub(w, req)
	}
}

func serveToolsGeneStub(w http.ResponseWriter, req toolsStubRequest) {
	switch {
	case strings.Contains(req.Query, "geneGeneralInformation") &&
		req.Variables[testGeneKey] == testGeneSymbol:
		writeToolsJSON(w, `{
			"data": {
				"geneGeneralInformation": {
					"id": "DDB_G0283907",
					"name_description": ["pkaC = Protein Kinase A"],
					"gene_product": "PKA catalytic subunit",
					"synonyms": ["PKA"],
					"description": ""
				}
			}
		}`)
	case strings.Contains(req.Query, "geneGeneralInformation"):
		writeToolsJSON(w, `{"data":{"geneGeneralInformation":null}}`)
	case strings.Contains(req.Query, "listStrainsWithGene"):
		writeToolsJSON(w, `{
			"data": {
				"listStrainsWithGene": [{
					"id": "DBS0236783",
					"label": "pkaC-",
					"summary": "pkaC null",
					"in_stock": true
				}]
			}
		}`)
	}
}

func serveToolsSearchStub(w http.ResponseWriter, req toolsStubRequest) {
	switch {
	case strings.Contains(req.Query, "searchStocks") &&
		req.Variables["entity"] == "PLASMID":
		writeToolsJSON(w, `{
			"data": {
				"searchStocks": {
					"limit": 25,
					"totalCount": 1,
					"results": [{
						"id": "DBP0000585",
						"entity": "PLASMID",
						"field": "NAME",
						"display_text": "pDXA-YFP-MCS1-PKA-Fret",
						"score": 10,
						"summary": "",
						"stock_label": "",
						"in_stock": true,
						"stock_type": "regular"
					}]
				}
			}
		}`)
	case strings.Contains(req.Query, "searchStocks") &&
		req.Variables["query"] == testGrowthQuery:
		writeToolsJSON(w, `{
			"data": {
				"searchStocks": {
					"limit": 100,
					"totalCount": 4,
					"results": [
						{
							"id": "DBS1",
							"entity": "STRAIN",
							"field": "PHENOTYPE",
							"display_text": "increased growth rate",
							"score": 9,
							"summary": "",
							"stock_label": "",
							"in_stock": true,
							"stock_type": "regular"
						},
						{
							"id": "DBS2",
							"entity": "STRAIN",
							"field": "PHENOTYPE",
							"display_text": "increased growth rate",
							"score": 8,
							"summary": "",
							"stock_label": "",
							"in_stock": true,
							"stock_type": "regular"
						},
						{
							"id": "DBS3",
							"entity": "STRAIN",
							"field": "SUMMARY",
							"display_text": "grows fast",
							"score": 7,
							"summary": "",
							"stock_label": "",
							"in_stock": true,
							"stock_type": "regular"
						},
						{
							"id": "DBS4",
							"entity": "STRAIN",
							"field": "PHENOTYPE",
							"display_text": "decreased growth rate",
							"score": 6,
							"summary": "",
							"stock_label": "",
							"in_stock": true,
							"stock_type": "regular"
						}
					]
				}
			}
		}`)
	case strings.Contains(req.Query, "searchStocks"):
		writeToolsJSON(w, `{
			"data": {
				"searchStocks": {
					"limit": 50,
					"totalCount": 2,
					"results": [
						{
							"id": "DBS0350840",
							"entity": "STRAIN",
							"field": "GENOTYPE",
							"display_text": "act15:triA",
							"score": 1011.4,
							"summary": "triA null",
							"stock_label": "triA-",
							"in_stock": true,
							"stock_type": "regular"
						},
						{
							"id": "DBS0350840",
							"entity": "STRAIN",
							"field": "SUMMARY",
							"display_text": "triA null",
							"score": 900.0,
							"summary": "triA null",
							"stock_label": "triA-",
							"in_stock": true,
							"stock_type": "regular"
						}
					]
				}
			}
		}`)
	}
}

func serveToolsStockDetailStub(w http.ResponseWriter, req toolsStubRequest) {
	switch {
	case strings.Contains(req.Query, "strain(") && !strings.Contains(req.Query, "label"):
		writeToolsJSON(w, `{
			"data": {
				"strain": {
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
				}
			}
		}`)
	case strings.Contains(req.Query, "strain("):
		writeToolsJSON(w, `{
			"data": {
				"strain": {
					"id": "DBS0350840",
					"label": "triA-",
					"summary": "triA null",
					"species": "Dictyostelium discoideum",
					"systematic_name": "DBS0350840",
					"plasmid": "",
					"dbxrefs": [],
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
						"title": "A paper",
						"journal": "Differentiation",
						"volume": "74",
						"pages": "1-10",
						"doi": "10.1111/x",
						"authors": [{"last_name": "Smith"}]
					}],
					"phenotypes": [{
						"phenotype": "aberrant spore morphology",
						"note": "",
						"assay": "",
						"environment": "",
						"publication": {
							"id": "21760904",
							"pub_date": "2011-07-07T00:00:00.000Z",
							"title": "BzpN slows proliferation",
							"journal": "PloS one",
							"volume": "6",
							"pages": "e21765",
							"doi": "",
							"authors": [{"last_name": "Phillips"}]
						}
					}]
				}
			}
		}`)
	case strings.Contains(req.Query, "plasmid(") &&
		req.Variables["id"] == testSequenceID:
		writeToolsJSON(w, `{
			"data": {
				"plasmid": {
					"id": "DBP0000001",
					"name": "pTest",
					"summary": "",
					"image_map": "",
					"sequence": "ATGC ATGC\nAAAA",
					"keywords": [],
					"genbank_accession": "",
					"in_stock": true,
					"dbxrefs": [],
					"depositor": {"first_name": "", "last_name": ""},
					"genes": [],
					"publications": []
				}
			}
		}`)
	case strings.Contains(req.Query, "plasmid("):
		writeToolsJSON(w, `{
			"data": {
				"plasmid": {
					"id": "DBP0391806",
					"name": "pDGB_O2_0006",
					"summary": "GoldenBraid",
					"image_map": "",
					"sequence": "",
					"keywords": [],
					"genbank_accession": "",
					"in_stock": true,
					"dbxrefs": [],
					"depositor": {"first_name": "Gadi", "last_name": "Shaulsky"},
					"genes": [],
					"publications": [{
						"id": "17177856",
						"pub_date": "2006-12-01T00:00:00.000Z",
						"title": "A paper",
						"journal": "Differentiation",
						"volume": "74",
						"pages": "1-10",
						"doi": "10.1111/x",
						"authors": [{"last_name": "Smith"}]
					}]
				}
			}
		}`)
	case strings.Contains(req.Query, "listStrainsWithAnnotation") &&
		req.Variables["annotation"] == "in suspension":
		writeToolsJSON(w, `{
			"errors": [{
				"message": "annotation list not found",
				"path": ["listStrainsWithAnnotation"],
				"extensions": {"code": "NotFound"}
			}],
			"data": {"listStrainsWithAnnotation": null}
		}`)
	case strings.Contains(req.Query, "listStrainsWithAnnotation"):
		writeToolsJSON(w, `{
			"data": {
				"listStrainsWithAnnotation": {
					"nextCursor": 1728057313873,
					"strains": [{
						"id": "DBS0349965",
						"label": "bzpN-",
						"summary": "bzpN null",
						"in_stock": true,
						"genes": [{"name": "bzpN"}]
					}]
				}
			}
		}`)
	}
}

func newToolsSessionWithLinks(t *testing.T, links linkout.Links) *mcp.ClientSession {
	t.Helper()
	gqlSrv := newToolsStubGraphQLServer(t)
	t.Cleanup(gqlSrv.Close)

	gqlClient := graphql.NewClient(gqlSrv.URL, gqlSrv.Client())
	mcpSrv := mcp.NewServer(toImplementation("dcr-mcp", testVersion), nil)
	tools.RegisterAll(mcpSrv, gqlClient, links)

	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	_, err := mcpSrv.Connect(ctx, t1, nil)
	require.NoError(t, err)

	mcpClient := mcp.NewClient(toImplementation("test-client", testVersion), nil)
	session, err := mcpClient.Connect(ctx, t2, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func newToolsSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	return newToolsSessionWithLinks(t, linkout.Links{})
}

func toolText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func objectAt(t *testing.T, property any) map[string]any {
	t.Helper()
	schema, ok := property.(map[string]any)
	require.True(t, ok)
	return schema
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	return props
}

func itemsOf(t *testing.T, property any) map[string]any {
	t.Helper()
	schema := objectAt(t, property)
	items, ok := schema["items"].(map[string]any)
	require.True(t, ok)
	return items
}

func TestMCPToolsSession(t *testing.T) {
	session := newToolsSession(t)
	ctx := context.Background()

	listRes, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, listRes.Tools, 12)

	t.Run("output schemas hold one publications property", func(t *testing.T) {
		res, err := session.ListTools(ctx, nil)
		require.NoError(t, err)
		schemas := map[string]map[string]any{}
		for _, tool := range res.Tools {
			schema, ok := tool.OutputSchema.(map[string]any)
			require.True(t, ok, tool.Name)
			schemas[tool.Name] = schema
		}

		strainProps := schemaProperties(t, schemas["get_strain"])
		require.Contains(t, strainProps, "publications")
		require.Contains(t, strainProps, "phenotypes")

		paperProps := schemaProperties(t, itemsOf(t, strainProps["publications"]))
		require.Contains(t, paperProps, "url")
		require.Contains(t, paperProps, "journal")
		require.Contains(t, paperProps, "authors")

		phenotypeProps := schemaProperties(t, itemsOf(t, strainProps["phenotypes"]))
		require.Contains(t, phenotypeProps, "publication")
		nestedProps := schemaProperties(t, objectAt(t, phenotypeProps["publication"]))
		require.Contains(t, nestedProps, "url")
		require.Contains(t, nestedProps, "journal")

		plasmidProps := schemaProperties(t, schemas["get_plasmid"])
		require.Contains(t, plasmidProps, "publications")
		require.Contains(t, schemaProperties(t, itemsOf(t, plasmidProps["publications"])), "url")

		geneProps := schemaProperties(t, schemas["find_stocks_by_gene"])
		geneRecordProps := schemaProperties(t, objectAt(t, geneProps["gene"]))
		require.Contains(t, geneRecordProps, "url")
		require.Contains(t, geneRecordProps, "gene_product")
	})

	t.Run("autocomplete_stocks", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "autocomplete_stocks",
			Arguments: map[string]any{testQueryKey: testActQuery},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("search_stocks", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "search_stocks",
			Arguments: map[string]any{testQueryKey: testActQuery, "entity": "strain"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("get_strain", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_strain",
			Arguments: map[string]any{"id": testStrainID},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)

		structured, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok)
		require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS0350840", structured["url"])
		require.Equal(t, "summary", structured["detail"])

		require.NotEmpty(t, res.Content)
		text, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		require.Contains(t, text.Text, "https://dictybase.dev/stockcenter/strains/DBS0350840")
		require.Contains(t, text.Text, "## triA- (DBS0350840)")
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("get_plasmid", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_plasmid",
			Arguments: map[string]any{"id": testPlasmidID},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("list_strains_with_phenotype", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "list_strains_with_phenotype",
			Arguments: map[string]any{testPhenotypeKey: testGrowthPheno},
		})
		require.NoError(t, err)
		require.False(t, res.IsError, "%s", toolText(res))
		structured, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok)
		require.Equal(t, testGrowthPheno, structured[testPhenotypeKey])
		require.EqualValues(t, 1, structured["page_count"])
		require.Equal(t, true, structured["has_more"])
		require.Contains(t, toolText(res), "More pages: yes")
	})

	t.Run("list_strains_with_phenotype rejects an environment term", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "list_strains_with_phenotype",
			Arguments: map[string]any{testPhenotypeKey: "in suspension"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
		require.Contains(t, toolText(res), "not an exact phenotype term")
		require.Contains(t, toolText(res), "search_phenotypes")
	})

	t.Run("search_phenotypes", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "search_phenotypes",
			Arguments: map[string]any{testQueryKey: "vegetative growth"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError, "%s", toolText(res))
		structured, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok)
		require.EqualValues(t, 2, structured["count"])
		texts, ok := structured["matched_texts"].([]any)
		require.True(t, ok)
		require.Len(t, texts, 1)
		require.Contains(t, toolText(res), "Top 2 strain matches")
		require.Contains(t, toolText(res), "· gwdi")
	})

	t.Run("get_plasmid_sequence returns FASTA", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_plasmid_sequence",
			Arguments: map[string]any{"id": testSequenceID},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		structured, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok)
		require.Equal(t, ">DBP0000001 pTest\nATGCATGCAAAA", structured["fasta"])
		require.EqualValues(t, 12, structured["length"])
	})

	t.Run("get_plasmid_sequence errors when no sequence on record", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_plasmid_sequence",
			Arguments: map[string]any{"id": testPlasmidID},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("invalid input returns tool IsError true", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "autocomplete_stocks",
			Arguments: map[string]any{testQueryKey: "a"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}

func TestToolMetadata(t *testing.T) {
	session := newToolsSession(t)
	listRes, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	for _, tool := range listRes.Tools {
		t.Run(tool.Name, func(t *testing.T) {
			require.NotEmpty(t, tool.Title)
			require.NotNil(t, tool.Annotations)
			require.Equal(t, tool.Title, tool.Annotations.Title)
			require.True(t, tool.Annotations.ReadOnlyHint)
			require.True(t, tool.Annotations.IdempotentHint)
			require.NotNil(t, tool.Annotations.DestructiveHint)
			require.False(t, *tool.Annotations.DestructiveHint)
			require.NotNil(t, tool.Annotations.OpenWorldHint)
			require.False(t, *tool.Annotations.OpenWorldHint)
			require.GreaterOrEqual(t, len(tool.Description), 80, "description must explain when to use the tool")
		})
	}
}
