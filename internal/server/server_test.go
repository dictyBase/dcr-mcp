package server_test

import (
	"context"
	"encoding/json"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

const testVersion = "test"

var (
	transportEndpointLens = L.MakeLens(
		func(tr mcp.StreamableClientTransport) string { return tr.Endpoint },
		func(tr mcp.StreamableClientTransport, endpoint string) mcp.StreamableClientTransport {
			tr.Endpoint = endpoint
			return tr
		},
	)
	transportClientLens = L.MakeLens(
		func(tr mcp.StreamableClientTransport) *http.Client { return tr.HTTPClient },
		func(tr mcp.StreamableClientTransport, client *http.Client) mcp.StreamableClientTransport {
			tr.HTTPClient = client
			return tr
		},
	)
	transportNoSSELens = L.MakeLens(
		func(tr mcp.StreamableClientTransport) bool { return tr.DisableStandaloneSSE },
		func(tr mcp.StreamableClientTransport, disable bool) mcp.StreamableClientTransport {
			tr.DisableStandaloneSSE = disable
			return tr
		},
	)
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
	callParamsNameLens = L.MakeLens(
		func(p mcp.CallToolParams) string { return p.Name },
		func(p mcp.CallToolParams, name string) mcp.CallToolParams {
			p.Name = name
			return p
		},
	)
	callParamsArgsLens = L.MakeLens(
		func(p mcp.CallToolParams) any { return p.Arguments },
		func(p mcp.CallToolParams, args any) mcp.CallToolParams {
			p.Arguments = args
			return p
		},
	)
)

func serverLinks(t *testing.T, raw string) linkout.Links {
	t.Helper()
	links, err := E.Unwrap(linkout.Parse(raw))
	require.NoError(t, err)
	return links
}

func linkoutHandler(t *testing.T, endpoint string, links linkout.Links) *httptest.Server {
	t.Helper()
	httpSrv := httptest.NewServer(server.NewHTTPHandler(F.Pipe4(
		server.Config{},
		server.ConfigEndpointLens.Set(endpoint),
		server.ConfigStatelessLens.Set(true),
		server.ConfigOriginsLens.Set([]string{"*"}),
		server.ConfigLinkoutsLens.Set(links),
	)))
	t.Cleanup(httpSrv.Close)
	return httpSrv
}

func linkoutSession(t *testing.T, httpSrv *httptest.Server) *mcp.ClientSession {
	t.Helper()
	transport := new(F.Pipe3(
		mcp.StreamableClientTransport{},
		transportEndpointLens.Set(httpSrv.URL+"/mcp"),
		transportClientLens.Set(httpSrv.Client()),
		transportNoSSELens.Set(true),
	))
	client := mcp.NewClient(new(F.Pipe2(
		mcp.Implementation{},
		implNameLens.Set("web-agent"),
		implVersionLens.Set(testVersion),
	)), nil)
	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func strainCall(session *mcp.ClientSession) (*mcp.CallToolResult, error) {
	return session.CallTool(context.Background(), new(F.Pipe2(
		mcp.CallToolParams{},
		callParamsNameLens.Set("get_strain"),
		callParamsArgsLens.Set(map[string]any{"id": "DBS0350840"}),
	)))
}

func strainURLAndText(t *testing.T, res *mcp.CallToolResult) (string, string) {
	t.Helper()
	require.False(t, res.IsError)
	structured, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	url, ok := structured["url"].(string)
	require.True(t, ok)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	return url, text.Text
}

func TestHTTPHandlerLinkouts(t *testing.T) {
	gqlSrv := newServerStubGraphQL(t)
	defer gqlSrv.Close()

	t.Run("configured base", func(t *testing.T) {
		httpSrv := linkoutHandler(t, gqlSrv.URL, serverLinks(t, "https://catalog.example.org"))
		res, err := strainCall(linkoutSession(t, httpSrv))
		require.NoError(t, err)
		url, text := strainURLAndText(t, res)
		require.Equal(t, "https://catalog.example.org/stockcenter/strains/DBS0350840", url)
		require.Contains(t, text, url)
	})

	t.Run("zero value uses the default website", func(t *testing.T) {
		httpSrv := linkoutHandler(t, gqlSrv.URL, linkout.Links{})
		res, err := strainCall(linkoutSession(t, httpSrv))
		require.NoError(t, err)
		url, text := strainURLAndText(t, res)
		require.Equal(t, "https://dictybase.dev/stockcenter/strains/DBS0350840", url)
		require.Contains(t, text, url)
	})
}

func TestHTTPHandlersKeepIndependentLinkouts(t *testing.T) {
	gqlSrv := newServerStubGraphQL(t)
	defer gqlSrv.Close()

	alphaSession := linkoutSession(t, linkoutHandler(t, gqlSrv.URL, serverLinks(t, "https://alpha.example.org")))
	betaSession := linkoutSession(t, linkoutHandler(t, gqlSrv.URL, serverLinks(t, "https://beta.example.org")))

	results := make([]*mcp.CallToolResult, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		results[0], errs[0] = strainCall(alphaSession)
	}()
	go func() {
		defer wg.Done()
		results[1], errs[1] = strainCall(betaSession)
	}()
	wg.Wait()
	require.NoError(t, errs[0])
	require.NoError(t, errs[1])

	alphaAgain, err := strainCall(alphaSession)
	require.NoError(t, err)

	alphaURL, alphaText := strainURLAndText(t, results[0])
	betaURL, betaText := strainURLAndText(t, results[1])
	alphaAgainURL, _ := strainURLAndText(t, alphaAgain)

	require.Equal(t, "https://alpha.example.org/stockcenter/strains/DBS0350840", alphaURL)
	require.Equal(t, "https://beta.example.org/stockcenter/strains/DBS0350840", betaURL)
	require.Equal(t, alphaURL, alphaAgainURL)
	require.Contains(t, alphaText, alphaURL)
	require.Contains(t, betaText, betaURL)
	require.NotContains(t, alphaText, betaURL)
}

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
			writeServerStubJSON(w, `{
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
		case strings.Contains(req.Query, "searchStocks"):
			writeServerStubJSON(w, `{
				"data": {
					"searchStocks": {
						"limit": 50,
						"totalCount": 1,
						"results": [{
							"id": "DBS0350840",
							"entity": "STRAIN",
							"field": "GENOTYPE",
							"display_text": "act15:triA",
							"score": 1011.4,
							"summary": "triA null",
							"stock_label": "triA-",
							"in_stock": true,
							"stock_type": "regular"
						}]
					}
				}
			}`)
		case strings.Contains(req.Query, "strain(") && !strings.Contains(req.Query, "label"):
			writeServerStubJSON(w, `{
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
			writeServerStubJSON(w, `{
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
		case strings.Contains(req.Query, "plasmid(") && req.Variables["id"] == "DBP0000001":
			writeServerStubJSON(w, `{
				"data": {
					"plasmid": {
						"id": "DBP0000001",
						"name": "pTest",
						"summary": "",
						"image_map": "",
						"sequence": "ATGC",
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
			writeServerStubJSON(w, `{
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
			writeServerStubJSON(w, `{
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
	}))
}

func writeServerStubJSON(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
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

	t.Run("StreamableClientTransport connects and calls all 6 tools over HTTP", func(t *testing.T) {
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
		require.Len(t, toolsRes.Tools, 12)

		for _, tc := range []struct {
			name string
			args map[string]any
		}{
			{name: "autocomplete_stocks", args: map[string]any{"query": "act"}},
			{name: "search_stocks", args: map[string]any{"query": "act"}},
			{name: "get_strain", args: map[string]any{"id": "DBS0350840"}},
			{name: "get_plasmid", args: map[string]any{"id": "DBP0391806"}},
			{name: "get_plasmid_sequence", args: map[string]any{"id": "DBP0000001"}},
			{name: "list_strains_with_phenotype", args: map[string]any{"phenotype_term": "increased growth rate"}},
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

func TestServerInstructions(t *testing.T) {
	srv := server.NewMCPServer(graphql.NewClient("http://127.0.0.1:0", nil), linkout.Links{})
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	_, err := srv.Connect(ctx, t1, nil)
	require.NoError(t, err)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "c", Version: testVersion}, nil).Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	instructions := session.InitializeResult().Instructions
	require.Contains(t, instructions, "find_stocks_by_gene")
	require.Contains(t, instructions, "search_phenotypes")
	require.NotContains(t, instructions, "suggest_phenotype_terms")
	require.Contains(t, instructions, "url")
}

func TestRateLimit(t *testing.T) {
	gqlSrv := newServerStubGraphQL(t)
	defer gqlSrv.Close()

	handler := server.NewHTTPHandler(server.Config{
		GraphQLEndpoint: gqlSrv.URL,
		Stateless:       true,
		AllowedOrigins:  []string{"*"},
		RateLimit:       2,
	})
	httpSrv := httptest.NewServer(handler)
	defer httpSrv.Close()

	post := func() int {
		body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
		req, err := http.NewRequest(http.MethodPost, httpSrv.URL+"/mcp", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		resp, err := httpSrv.Client().Do(req)
		require.NoError(t, err)
		_ = resp.Body.Close()
		return resp.StatusCode
	}

	require.NotEqual(t, http.StatusTooManyRequests, post())
	require.NotEqual(t, http.StatusTooManyRequests, post())
	require.Equal(t, http.StatusTooManyRequests, post())

	resp, err := http.Get(httpSrv.URL + "/healthz")
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "healthz is not rate limited")
}
