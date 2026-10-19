package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dictybase/dcr-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func newServerStubGraphQL(t *testing.T) *httptest.Server {
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
		case strings.Contains(req.Query, "plasmid("):
			_, _ = w.Write([]byte(`{"data":{"plasmid":{"id":"DBP0391806","name":"pDGB_O2_0006","summary":"GoldenBraid","image_map":"","sequence":"","keywords":[],"genbank_accession":"","in_stock":true,"dbxrefs":[],"depositor":{"first_name":"Gadi","last_name":"Shaulsky"},"genes":[],"publications":[]}}}`))
		case strings.Contains(req.Query, "listStrainsWithAnnotation"):
			_, _ = w.Write([]byte(`{"data":{"listStrainsWithAnnotation":{"totalCount":1,"nextCursor":1728057313873,"strains":[{"id":"DBS0349965","label":"bzpN-","summary":"bzpN null","in_stock":true,"genes":[{"name":"bzpN"}],"publications":[]}]}}}`))
		}
	}))
}

func TestHTTPHandler(t *testing.T) {
	gqlSrv := newServerStubGraphQL(t)
	defer gqlSrv.Close()

	handler := server.NewHTTPHandler(server.Config{
		GraphQLEndpoint: gqlSrv.URL,
		Host:            "127.0.0.1",
		Port:            8080,
		Stateless:       true,
		AllowedOrigins:  []string{"*"},
	})
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	t.Run("GET /healthz returns 200 OK JSON", func(t *testing.T) {
		resp, err := httpSrv.Client().Get(httpSrv.URL + "/healthz")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]string
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		require.Equal(t, "ok", body["status"])
	})

	t.Run("OPTIONS /mcp preflight returns 200 with chi CORS headers", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodOptions, httpSrv.URL+"/mcp", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "https://stockcenter.dictybase.org")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Mcp-Session-Id")

		resp, err := httpSrv.Client().Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		require.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Mcp-Session-Id")
	})

	t.Run("GET /healthz with Origin returns ExposedHeaders from chi CORS", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, httpSrv.URL+"/healthz", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "https://stockcenter.dictybase.org")

		resp, err := httpSrv.Client().Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		require.Contains(t, resp.Header.Get("Access-Control-Expose-Headers"), "Mcp-Session-Id")
	})

	t.Run("StreamableClientTransport connects and calls all 5 tools over HTTP", func(t *testing.T) {
		ctx := context.Background()
		transport := &mcp.StreamableClientTransport{
			Endpoint:             httpSrv.URL + "/mcp",
			HTTPClient:           httpSrv.Client(),
			DisableStandaloneSSE: true,
		}
		client := mcp.NewClient(&mcp.Implementation{Name: "web-agent", Version: "1.0.0"}, nil)
		session, err := client.Connect(ctx, transport, nil)
		require.NoError(t, err)
		defer func() { _ = session.Close() }()

		toolsRes, err := session.ListTools(ctx, nil)
		require.NoError(t, err)
		require.Len(t, toolsRes.Tools, 5)

		for _, tc := range []struct {
			name string
			args map[string]any
		}{
			{name: "autocomplete_stocks", args: map[string]any{"query": "act"}},
			{name: "search_stocks", args: map[string]any{"query": "act"}},
			{name: "get_strain", args: map[string]any{"id": "DBS0350840"}},
			{name: "get_plasmid", args: map[string]any{"id": "DBP0391806"}},
			{name: "list_strains_with_phenotype", args: map[string]any{"phenotype": "increased growth rate"}},
		} {
			res, callErr := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      tc.name,
				Arguments: tc.args,
			})
			require.NoError(t, callErr, tc.name)
			require.False(t, res.IsError, tc.name)
			require.NotNil(t, res.StructuredContent, tc.name)
		}
	})
}
