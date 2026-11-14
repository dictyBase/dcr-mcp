package tools_test

import (
	"context"
	A "github.com/IBM/fp-go/v2/array"
	O "github.com/IBM/fp-go/v2/option"
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

const (
	testDetailKey        = "detail"
	testFullDetail       = "full"
	testGeneKey          = "gene"
	testIDsKey           = "ids"
	testGeneSymbol       = "pkaC"
	testSequenceID       = "DBP0000001"
	testPhenotypeKey     = "phenotype_term"
	testCustomBase       = "https://catalog.example.org:8443"
	testCustomStrainURL  = testCustomBase + "/stockcenter/strains/DBS0350840"
	testCustomPlasmidURL = testCustomBase + "/stockcenter/plasmids/DBP0391806"
	testDefaultBase      = "https://dictybase.dev"
	testObsoleteBase     = "https://dictycr.org/"
	testURLKey           = "url"
)

var (
	stubStrainIDLens = L.MakeLens(
		func(d graphql.StrainDetail) string { return d.ID },
		func(d graphql.StrainDetail, id string) graphql.StrainDetail {
			d.ID = id
			return d
		},
	)
	stubStrainPapersLens = L.MakeLens(
		func(d graphql.StrainDetail) []graphql.PublicationRef { return d.Publications },
		func(d graphql.StrainDetail, papers []graphql.PublicationRef) graphql.StrainDetail {
			d.Publications = papers
			return d
		},
	)
	stubPaperTitleLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.Title },
		func(p graphql.PublicationRef, title string) graphql.PublicationRef {
			p.Title = title
			return p
		},
	)
	stubPaperJournalLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.Journal },
		func(p graphql.PublicationRef, journal string) graphql.PublicationRef {
			p.Journal = journal
			return p
		},
	)
	stubPaperPubDateLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.PubDate },
		func(p graphql.PublicationRef, pubDate string) graphql.PublicationRef {
			p.PubDate = pubDate
			return p
		},
	)
	stubPaperAuthorsLens = L.MakeLens(
		func(p graphql.PublicationRef) []graphql.AuthorRef { return p.Authors },
		func(p graphql.PublicationRef, authors []graphql.AuthorRef) graphql.PublicationRef {
			p.Authors = authors
			return p
		},
	)
	stubAuthorLastNameLens = L.MakeLens(
		func(a graphql.AuthorRef) string { return a.LastName },
		func(a graphql.AuthorRef, lastName string) graphql.AuthorRef {
			a.LastName = lastName
			return a
		},
	)
	stubViewPapersLens = L.MakeLens(
		func(v tools.StrainView) []tools.PublicationRefView { return v.Publications },
		func(v tools.StrainView, papers []tools.PublicationRefView) tools.StrainView {
			v.Publications = papers
			return v
		},
	)
	stubRefViewURLLens = L.MakeLens(
		func(v tools.PublicationRefView) string { return v.URL },
		func(v tools.PublicationRefView, url string) tools.PublicationRefView {
			v.URL = url
			return v
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

func TestGeneLinkoutUsesResolvedID(t *testing.T) {
	session := newToolsSessionWithLinks(t, customLinks(t))
	out, _ := callTool(t, session, "find_stocks_by_gene", map[string]any{"gene": "pkaC"})

	gene, ok := out["gene"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "DDB_G0283907", gene["id"])
	require.Equal(t, testCustomBase+"/gene/DDB_G0283907", gene[testURLKey])
	require.NotContains(t, gene[testURLKey], "pkaC")
	require.Equal(t, "PKA catalytic subunit", gene["gene_product"])
	require.Equal(t, []any{"PKA"}, gene["synonyms"])
	require.Equal(t, "", gene["description"])

	defaultSession := newToolsSession(t)
	defaultOut, _ := callTool(t, defaultSession, "find_stocks_by_gene", map[string]any{"gene": "pkaC"})
	defaultGene, ok := defaultOut["gene"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, testDefaultBase+"/gene/DDB_G0283907", defaultGene[testURLKey])
}

func TestPublicationLinkouts(t *testing.T) {
	session := newToolsSessionWithLinks(t, customLinks(t))
	out, text := callTool(t, session, "get_publication", map[string]any{"id": "17177856"})

	require.Equal(t, testCustomBase+"/publication/17177856", out[testURLKey])
	require.Equal(t, "https://pubmed.ncbi.nlm.nih.gov/17177856/", out["pubmed_url"])
	require.Equal(t, "https://doi.org/10.1111/x", out["doi_url"])
	require.Contains(t, text, "[Publication]("+testCustomBase+"/publication/17177856)")
	require.Contains(t, text, "[PubMed](https://pubmed.ncbi.nlm.nih.gov/17177856/)")
	require.Contains(t, text, "[DOI](https://doi.org/10.1111/x)")

	defaultSession := newToolsSession(t)
	defaultOut, defaultText := callTool(t, defaultSession, "get_publication", map[string]any{"id": "17177856"})
	require.Equal(t, testDefaultBase+"/publication/17177856", defaultOut[testURLKey])
	require.Contains(t, defaultText, "[Publication]("+testDefaultBase+"/publication/17177856)")
}

func TestNestedPublicationLinkouts(t *testing.T) {
	session := newToolsSessionWithLinks(t, customLinks(t))
	localPaper := testCustomBase + "/publication/17177856"
	localPhenotypePaper := testCustomBase + "/publication/21760904"

	for _, detail := range []string{"summary", testFullDetail} {
		t.Run("get_strain detail="+detail, func(t *testing.T) {
			out, text := callTool(t, session, "get_strain", map[string]any{
				"id":          testStrainID,
				testDetailKey: detail,
			})
			papers := mapsAt(t, out, "publications")
			require.Len(t, papers, 1)
			require.Equal(t, "17177856", papers[0]["id"])
			require.Equal(t, localPaper, papers[0][testURLKey])
			require.Equal(t, "Differentiation", papers[0]["journal"])

			phenotypes := mapsAt(t, out, "phenotypes")
			require.Len(t, phenotypes, 1)
			nested, ok := phenotypes[0]["publication"].(map[string]any)
			require.True(t, ok)
			require.Equal(t, localPhenotypePaper, nested[testURLKey])
			require.Equal(t, "BzpN slows proliferation", nested["title"])

			require.Contains(t, text, "[Publication]("+localPaper+")")
			require.Contains(t, text, "[Publication]("+localPhenotypePaper+")")
			require.Contains(t, text, "[PubMed](https://pubmed.ncbi.nlm.nih.gov/17177856/)")
			require.Contains(t, text, "[DOI](https://doi.org/10.1111/x)")
			require.NotContains(t, text, testObsoleteBase)
		})
	}

	t.Run("get_plasmid", func(t *testing.T) {
		out, text := callTool(t, session, "get_plasmid", map[string]any{"id": testPlasmidID})
		papers := mapsAt(t, out, "publications")
		require.Len(t, papers, 1)
		require.Equal(t, localPaper, papers[0][testURLKey])
		require.Contains(t, text, "[Publication]("+localPaper+")")
	})

	t.Run("fetch keeps the same destinations", func(t *testing.T) {
		strain, _ := callTool(t, session, "fetch", map[string]any{"id": testStrainID})
		require.Contains(t, strain["text"], "[Publication]("+localPaper+")")
		require.Contains(t, strain["text"], "[Publication]("+localPhenotypePaper+")")

		plasmid, _ := callTool(t, session, "fetch", map[string]any{"id": testPlasmidID})
		require.Contains(t, plasmid["text"], "[Publication]("+localPaper+")")
	})
}

func TestNameOnlyGenesUnchanged(t *testing.T) {
	session := newToolsSessionWithLinks(t, customLinks(t))
	out, _ := callTool(t, session, "get_strain", map[string]any{"id": testStrainID})

	genes := mapsAt(t, out, "genes")
	require.Len(t, genes, 1)
	require.Equal(t, "triA", genes[0]["name"])
	require.NotContains(t, genes[0], "id")
	require.NotContains(t, genes[0], testURLKey)
	require.Len(t, genes[0], 1, "a name-only gene gains no field")
}

func TestPublicationViewsWithoutIDOrDOI(t *testing.T) {
	links := customLinks(t)
	view := F.Pipe1(
		F.Pipe2(
			graphql.StrainDetail{},
			stubStrainIDLens.Set(testStrainID),
			stubStrainPapersLens.Set([]graphql.PublicationRef{F.Pipe4(
				graphql.PublicationRef{},
				stubPaperTitleLens.Set("No identifier"),
				stubPaperJournalLens.Set("J"),
				stubPaperPubDateLens.Set("2001"),
				stubPaperAuthorsLens.Set([]graphql.AuthorRef{
					F.Pipe1(graphql.AuthorRef{}, stubAuthorLastNameLens.Set("Lee")),
				}),
			)}),
		),
		tools.StrainViewFor(links)(testFullDetail),
	)
	require.Empty(t, F.Pipe2(view, stubViewPapersLens.Get, stubRefViewURLOfFirst))

	text := tools.RenderStrainMarkdown(view)
	require.NotContains(t, text, "[Publication]()")
	require.NotContains(t, text, "[DOI]")
	require.Contains(t, text, "[PubMed]")
}

func stubRefViewURLOfFirst(papers []tools.PublicationRefView) string {
	return F.Pipe3(
		papers,
		A.First[tools.PublicationRefView],
		O.Map(stubRefViewURLLens.Get),
		O.GetOrElse(F.Constant("")),
	)
}

func customLinks(t *testing.T) linkout.Links {
	t.Helper()
	links, err := E.Unwrap(linkout.Parse(testCustomBase))
	require.NoError(t, err)
	return links
}

func toCallParams(name string, args map[string]any) *mcp.CallToolParams {
	return new(F.Pipe2(
		mcp.CallToolParams{},
		callParamsNameLens.Set(name),
		callParamsArgsLens.Set(args),
	))
}

func callTool(
	t *testing.T,
	session *mcp.ClientSession,
	name string,
	args map[string]any,
) (map[string]any, string) {
	t.Helper()
	res, err := session.CallTool(context.Background(), toCallParams(name, args))
	require.NoError(t, err)
	require.False(t, res.IsError, toolText(res))
	structured, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	return structured, toolText(res)
}

func mapsAt(t *testing.T, parent map[string]any, key string) []map[string]any {
	t.Helper()
	raw, ok := parent[key].([]any)
	require.True(t, ok, key)
	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		entry, isMap := item.(map[string]any)
		require.True(t, isMap, key)
		items = append(items, entry)
	}
	return items
}

func TestStockLinkoutsCustomBase(t *testing.T) {
	session := newToolsSessionWithLinks(t, customLinks(t))

	t.Run("search_stocks", func(t *testing.T) {
		out, text := callTool(t, session, "search_stocks", map[string]any{testQueryKey: testActQuery})
		results := mapsAt(t, out, "results")
		require.Len(t, results, 2)
		require.Equal(t, testCustomStrainURL, results[0][testURLKey])
		require.Equal(t, testCustomStrainURL, results[1][testURLKey])
		require.Contains(t, text, testCustomStrainURL)
		require.NotContains(t, text, testObsoleteBase)
	})

	t.Run("search_phenotypes", func(t *testing.T) {
		out, text := callTool(t, session, "search_phenotypes", map[string]any{
			testQueryKey: "vegetative growth",
		})
		results := mapsAt(t, out, "results")
		require.Len(t, results, 2)
		require.Equal(t, testCustomBase+"/stockcenter/strains/DBS0235587", results[0][testURLKey])
		require.Equal(t, testCustomBase+"/stockcenter/strains/DBS0236827", results[1][testURLKey])
		require.Contains(t, text, testCustomBase+"/stockcenter/strains/DBS0235587")
		require.NotContains(t, text, testObsoleteBase)
	})

	t.Run("get_strain", func(t *testing.T) {
		summary, summaryText := callTool(t, session, "get_strain", map[string]any{"id": testStrainID})
		require.Equal(t, testCustomStrainURL, summary[testURLKey])
		require.Equal(t, "summary", summary[testDetailKey])
		require.Contains(t, summaryText, testCustomStrainURL)

		full, fullText := callTool(t, session, "get_strain", map[string]any{
			"id":          testStrainID,
			testDetailKey: testFullDetail,
		})
		require.Equal(t, testCustomStrainURL, full[testURLKey])
		require.Contains(t, fullText, testCustomStrainURL)
		require.NotContains(t, fullText, testObsoleteBase)
	})

	t.Run("get_plasmid", func(t *testing.T) {
		summary, summaryText := callTool(t, session, "get_plasmid", map[string]any{"id": testPlasmidID})
		require.Equal(t, testCustomPlasmidURL, summary[testURLKey])
		require.Contains(t, summaryText, testCustomPlasmidURL)

		full, _ := callTool(t, session, "get_plasmid", map[string]any{
			"id":          testPlasmidID,
			testDetailKey: testFullDetail,
		})
		require.Equal(t, testCustomPlasmidURL, full[testURLKey])
	})

	t.Run("get_plasmid_sequence", func(t *testing.T) {
		out, _ := callTool(t, session, "get_plasmid_sequence", map[string]any{"id": testSequenceID})
		require.Equal(t, testCustomBase+"/stockcenter/plasmids/DBP0000001", out[testURLKey])
		require.Equal(t, ">DBP0000001 pTest\nATGCATGCAAAA", out["fasta"])
		require.EqualValues(t, 12, out["length"])
	})

	t.Run("list_strains_with_phenotype", func(t *testing.T) {
		out, text := callTool(t, session, "list_strains_with_phenotype", map[string]any{
			testPhenotypeKey: testGrowthPheno,
		})
		strains := mapsAt(t, out, "strains")
		require.Len(t, strains, 1)
		require.Equal(t, testCustomBase+"/stockcenter/strains/DBS0349965", strains[0][testURLKey])
		require.EqualValues(t, 1728057313873, out["next_cursor"])
		require.EqualValues(t, 1, out["page_count"])
		require.Contains(t, text, testCustomBase+"/stockcenter/strains/DBS0349965")
	})

	t.Run("find_stocks_by_gene", func(t *testing.T) {
		out, _ := callTool(t, session, "find_stocks_by_gene", map[string]any{testGeneKey: testGeneSymbol})
		strains := mapsAt(t, out, "strains")
		plasmids := mapsAt(t, out, "plasmids")
		require.Len(t, strains, 1)
		require.Len(t, plasmids, 1)
		require.Equal(t, testCustomBase+"/stockcenter/strains/DBS0236783", strains[0][testURLKey])
		require.Equal(t, testCustomBase+"/stockcenter/plasmids/DBP0000585", plasmids[0][testURLKey])
	})

	t.Run("check_availability", func(t *testing.T) {
		out, _ := callTool(t, session, "check_availability", map[string]any{
			testIDsKey: []string{testStrainID, testPlasmidID, "XYZ1"},
		})
		items := mapsAt(t, out, "items")
		require.Len(t, items, 3)
		require.Equal(t, testCustomStrainURL, items[0][testURLKey])
		require.Equal(t, testCustomPlasmidURL, items[1][testURLKey])
		require.NotContains(t, items[2], testURLKey)
		require.Contains(t, items[2]["error"], "unrecognized stock id")
	})

	t.Run("search", func(t *testing.T) {
		out, _ := callTool(t, session, "search", map[string]any{testQueryKey: testActQuery})
		results := mapsAt(t, out, "results")
		require.Len(t, results, 1, "duplicate stock ids collapse to one document")
		require.Equal(t, testCustomStrainURL, results[0][testURLKey])
	})

	t.Run("fetch", func(t *testing.T) {
		strain, _ := callTool(t, session, "fetch", map[string]any{"id": testStrainID})
		require.Equal(t, testCustomStrainURL, strain[testURLKey])
		require.Contains(t, strain["text"], testCustomStrainURL)

		plasmid, _ := callTool(t, session, "fetch", map[string]any{"id": "dbp0391806"})
		require.Equal(t, testCustomPlasmidURL, plasmid[testURLKey])
		require.Contains(t, plasmid["text"], testCustomPlasmidURL)
	})
}

func TestStockLinkoutsDefaultBase(t *testing.T) {
	session := newToolsSession(t)

	strain, strainText := callTool(t, session, "get_strain", map[string]any{"id": testStrainID})
	require.Equal(t, testDefaultBase+"/stockcenter/strains/DBS0350840", strain[testURLKey])
	require.Contains(t, strainText, testDefaultBase+"/stockcenter/strains/DBS0350840")
	require.NotContains(t, strainText, testObsoleteBase)

	plasmid, _ := callTool(t, session, "get_plasmid", map[string]any{"id": testPlasmidID})
	require.Equal(t, testDefaultBase+"/stockcenter/plasmids/DBP0391806", plasmid[testURLKey])
}

func TestStockURLEntityDispatch(t *testing.T) {
	links := customLinks(t)
	require.Equal(t, testCustomBase+"/stockcenter/strains/DBS1",
		tools.StockURL(links, graphql.StockEntityStrain, "DBS1"))
	require.Equal(t, testCustomBase+"/stockcenter/plasmids/DBP1",
		tools.StockURL(links, graphql.StockEntityPlasmid, "DBP1"))
	require.Empty(t, tools.StockURL(links, graphql.StockEntityAll, "DBS1"))
	require.Empty(t, tools.StockURL(links, graphql.StockEntityType("UNKNOWN"), "DBS1"))
	require.Empty(t, tools.StockURL(links, graphql.StockEntityStrain, ""))
}
