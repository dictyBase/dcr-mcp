package tools_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type toolsStubRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

const (
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

	t.Run("ValidatePhenotypeListInput applies defaults and checks bounds", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: " increased growth rate ",
		}))
		require.NoError(t, err)
		require.Equal(t, testGrowthPheno, got.Phenotype)
		require.Equal(t, 0, got.Cursor)
		require.Equal(t, 10, got.Limit)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: testBlankQuery,
		}))
		require.Error(t, err)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: testGrowthPheno,
			Cursor:    -1,
		}))
		require.Error(t, err)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: testGrowthPheno,
			Limit:     101,
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
		req.Variables["gene"] == "pkaC":
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
						"strain_label": "",
						"in_stock": true
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
							"strain_label": "",
							"in_stock": true
						},
						{
							"id": "DBS2",
							"entity": "STRAIN",
							"field": "PHENOTYPE",
							"display_text": "increased growth rate",
							"score": 8,
							"summary": "",
							"strain_label": "",
							"in_stock": true
						},
						{
							"id": "DBS3",
							"entity": "STRAIN",
							"field": "SUMMARY",
							"display_text": "grows fast",
							"score": 7,
							"summary": "",
							"strain_label": "",
							"in_stock": true
						},
						{
							"id": "DBS4",
							"entity": "STRAIN",
							"field": "PHENOTYPE",
							"display_text": "decreased growth rate",
							"score": 6,
							"summary": "",
							"strain_label": "",
							"in_stock": true
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
							"strain_label": "triA-",
							"in_stock": true
						},
						{
							"id": "DBS0350840",
							"entity": "STRAIN",
							"field": "SUMMARY",
							"display_text": "triA null",
							"score": 900.0,
							"summary": "triA null",
							"strain_label": "triA-",
							"in_stock": true
						}
					]
				}
			}
		}`)
	}
}

func serveToolsStockDetailStub(w http.ResponseWriter, req toolsStubRequest) {
	switch {
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
					"publications": [],
					"phenotypes": []
				}
			}
		}`)
	case strings.Contains(req.Query, "plasmid(") &&
		req.Variables["id"] == "DBP0000001":
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
					"publications": []
				}
			}
		}`)
	case strings.Contains(req.Query, "listStrainsWithAnnotation"):
		writeToolsJSON(w, `{
			"data": {
				"listStrainsWithAnnotation": {
					"totalCount": 1,
					"nextCursor": 1728057313873,
					"strains": [{
						"id": "DBS0349965",
						"label": "bzpN-",
						"summary": "bzpN null",
						"in_stock": true,
						"genes": [{"name": "bzpN"}],
						"publications": []
					}]
				}
			}
		}`)
	}
}

func newToolsSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	gqlSrv := newToolsStubGraphQLServer(t)
	t.Cleanup(gqlSrv.Close)

	gqlClient := graphql.NewClient(gqlSrv.URL, gqlSrv.Client())
	mcpSrv := mcp.NewServer(&mcp.Implementation{Name: "dcr-mcp", Version: "1.0.0"}, nil)
	tools.RegisterAll(mcpSrv, gqlClient)

	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	_, err := mcpSrv.Connect(ctx, t1, nil)
	require.NoError(t, err)

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, t2, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestMCPToolsSession(t *testing.T) {
	session := newToolsSession(t)
	ctx := context.Background()

	listRes, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, listRes.Tools, 12)

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
		require.Equal(t, "https://dictycr.org/strains/DBS0350840", structured["url"])
		require.Equal(t, "summary", structured["detail"])

		require.NotEmpty(t, res.Content)
		text, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		require.Contains(t, text.Text, "https://dictycr.org/strains/DBS0350840")
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
			Arguments: map[string]any{"phenotype": testGrowthPheno},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("get_plasmid_sequence returns FASTA", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_plasmid_sequence",
			Arguments: map[string]any{"id": "DBP0000001"},
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
