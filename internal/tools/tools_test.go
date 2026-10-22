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
			wantQuery:  "act",
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
			in:      tools.AutocompleteStocksInput{Query: "act", Limit: 51},
			wantErr: true,
		},
		{
			name:    "negative limit rejected",
			in:      tools.AutocompleteStocksInput{Query: "act", Limit: -1},
			wantErr: true,
		},
		{
			name:    "invalid entity rejected",
			in:      tools.AutocompleteStocksInput{Query: "act", Entity: "GENE"},
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
			in:      tools.SearchStocksInput{Query: "   "},
			wantErr: true,
		},
		{
			name:    "limit over 100 rejected",
			in:      tools.SearchStocksInput{Query: "act", Limit: 101},
			wantErr: true,
		},
		{
			name:    "invalid entity rejected",
			in:      tools.SearchStocksInput{Query: "act", Entity: "PROTEIN"},
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
		require.Equal(t, "DBS0350840", got.ID)

		_, err = E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: "   "}))
		require.Error(t, err)
	})

	t.Run("ValidatePlasmidIDInput trims and requires non-empty", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: " DBP0391806 "}))
		require.NoError(t, err)
		require.Equal(t, "DBP0391806", got.ID)

		_, err = E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: ""}))
		require.Error(t, err)
	})

	t.Run("ValidatePhenotypeListInput applies defaults and checks bounds", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: " increased growth rate ",
		}))
		require.NoError(t, err)
		require.Equal(t, "increased growth rate", got.Phenotype)
		require.Equal(t, 0, got.Cursor)
		require.Equal(t, 10, got.Limit)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: "   ",
		}))
		require.Error(t, err)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: "increased growth rate",
			Cursor:    -1,
		}))
		require.Error(t, err)

		_, err = E.Unwrap(tools.ValidatePhenotypeListInput(tools.ListStrainsWithPhenotypeInput{
			Phenotype: "increased growth rate",
			Limit:     101,
		}))
		require.Error(t, err)
	})

	t.Run("ValidateStrainIDInput defaults detail to summary and rejects unknown", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: "DBS0350840"}))
		require.NoError(t, err)
		require.Equal(t, "summary", got.Detail)

		got, err = E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: "DBS0350840", Detail: " FULL "}))
		require.NoError(t, err)
		require.Equal(t, "full", got.Detail)

		_, err = E.Unwrap(tools.ValidateStrainIDInput(tools.GetStrainInput{ID: "DBS0350840", Detail: "everything"}))
		require.Error(t, err)
	})

	t.Run("ValidatePlasmidIDInput defaults detail to summary", func(t *testing.T) {
		got, err := E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: "DBP0391806"}))
		require.NoError(t, err)
		require.Equal(t, "summary", got.Detail)

		_, err = E.Unwrap(tools.ValidatePlasmidIDInput(tools.GetPlasmidInput{ID: "DBP0391806", Detail: "huge"}))
		require.Error(t, err)
	})
}

func newToolsStubGraphQLServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "listStockSuggestions"):
			_, _ = w.Write([]byte(`{"data":{"listStockSuggestions":{"limit":5,"totalCount":1,"suggestions":[{"id":"DBS0350026","entity":"STRAIN","field":"GENOTYPE","display_text":"act15:mip1","score":1013.5}]}}}`))
		case strings.Contains(req.Query, "searchStocks"):
			_, _ = w.Write([]byte(`{"data":{"searchStocks":{"limit":50,"totalCount":1,"results":[{"id":"DBS0350840","entity":"STRAIN","field":"GENOTYPE","display_text":"act15:triA","score":1011.4,"summary":"triA null","strain_label":"triA-","in_stock":true}]}}}`))
		case strings.Contains(req.Query, "strain("):
			_, _ = w.Write([]byte(`{"data":{"strain":{"id":"DBS0350840","label":"triA-","summary":"triA null","species":"Dictyostelium discoideum","systematic_name":"DBS0350840","plasmid":"","dbxrefs":[],"in_stock":true,"genotypes":["axeA2"],"mutagenesis_method":"","genetic_modification":"exogenous mutation","names":["triA-"],"characteristics":["axenic"],"parent":{"id":"","label":""},"depositor":{"first_name":"Dicty","last_name":"Stock Center"},"genes":[{"name":"triA"}],"publications":[],"phenotypes":[]}}}`))
		case strings.Contains(req.Query, "plasmid(") && req.Variables["id"] == "DBP0000001":
			_, _ = w.Write([]byte(`{"data":{"plasmid":{"id":"DBP0000001","name":"pTest","summary":"","image_map":"","sequence":"ATGC ATGC\nAAAA","keywords":[],"genbank_accession":"","in_stock":true,"dbxrefs":[],"depositor":{"first_name":"","last_name":""},"genes":[],"publications":[]}}}`))
		case strings.Contains(req.Query, "plasmid("):
			_, _ = w.Write([]byte(`{"data":{"plasmid":{"id":"DBP0391806","name":"pDGB_O2_0006","summary":"GoldenBraid","image_map":"","sequence":"","keywords":[],"genbank_accession":"","in_stock":true,"dbxrefs":[],"depositor":{"first_name":"Gadi","last_name":"Shaulsky"},"genes":[],"publications":[]}}}`))
		case strings.Contains(req.Query, "listStrainsWithAnnotation"):
			_, _ = w.Write([]byte(`{"data":{"listStrainsWithAnnotation":{"totalCount":1,"nextCursor":1728057313873,"strains":[{"id":"DBS0349965","label":"bzpN-","summary":"bzpN null","in_stock":true,"genes":[{"name":"bzpN"}],"publications":[]}]}}}`))
		}
	}))
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
	require.Len(t, listRes.Tools, 6)

	t.Run("autocomplete_stocks", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "autocomplete_stocks",
			Arguments: map[string]any{"query": "act"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("search_stocks", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "search_stocks",
			Arguments: map[string]any{"query": "act", "entity": "strain"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("get_strain", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_strain",
			Arguments: map[string]any{"id": "DBS0350840"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)

		structured, ok := res.StructuredContent.(map[string]any)
		require.True(t, ok)
		require.Equal(t, "https://dictycr.org/strains/DBS0350840", structured["url"])
		require.Equal(t, "summary", structured["detail"])
	})

	t.Run("get_plasmid", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_plasmid",
			Arguments: map[string]any{"id": "DBP0391806"},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		require.NotNil(t, res.StructuredContent)
	})

	t.Run("list_strains_with_phenotype", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "list_strains_with_phenotype",
			Arguments: map[string]any{"phenotype": "increased growth rate"},
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
			Arguments: map[string]any{"id": "DBP0391806"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("invalid input returns tool IsError true", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "autocomplete_stocks",
			Arguments: map[string]any{"query": "a"},
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
