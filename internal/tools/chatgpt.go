package tools

import (
	"context"
	"strconv"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	S "github.com/IBM/fp-go/v2/string"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/fp-go-loom/strutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	// SearchDocument is one result in the ChatGPT search contract.
	SearchDocument struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url"`
	}

	// SearchDocuments is the ChatGPT search output.
	SearchDocuments struct {
		Results []SearchDocument `json:"results"`
	}

	// FetchDocument is the ChatGPT fetch output.
	FetchDocument struct {
		ID       string            `json:"id"`
		Title    string            `json:"title"`
		Text     string            `json:"text"`
		URL      string            `json:"url"`
		Metadata map[string]string `json:"metadata"`
	}

	fetchRoute = func(fetchState) IOE.IOEither[error, FetchDocument]
)

const chatgptSearchLimit = 20

var (
	fetchRoutes = map[string]fetchRoute{
		strainIDPrefix:  fetchStrainDocument,
		plasmidIDPrefix: fetchPlasmidDocument,
	}

	searchDocumentIDLens = L.MakeLens(
		func(doc SearchDocument) string { return doc.ID },
		func(doc SearchDocument, id string) SearchDocument {
			doc.ID = id
			return doc
		},
	)
	searchDocumentTitleLens = L.MakeLens(
		func(doc SearchDocument) string { return doc.Title },
		func(doc SearchDocument, title string) SearchDocument {
			doc.Title = title
			return doc
		},
	)
	searchDocumentURLLens = L.MakeLens(
		func(doc SearchDocument) string { return doc.URL },
		func(doc SearchDocument, url string) SearchDocument {
			doc.URL = url
			return doc
		},
	)
	searchDocumentsResultsLens = L.MakeLens(
		func(docs SearchDocuments) []SearchDocument { return docs.Results },
		func(docs SearchDocuments, results []SearchDocument) SearchDocuments {
			docs.Results = results
			return docs
		},
	)
	fetchDocumentIDLens = L.MakeLens(
		func(doc FetchDocument) string { return doc.ID },
		func(doc FetchDocument, id string) FetchDocument {
			doc.ID = id
			return doc
		},
	)
	fetchDocumentTitleLens = L.MakeLens(
		func(doc FetchDocument) string { return doc.Title },
		func(doc FetchDocument, title string) FetchDocument {
			doc.Title = title
			return doc
		},
	)
	fetchDocumentTextLens = L.MakeLens(
		func(doc FetchDocument) string { return doc.Text },
		func(doc FetchDocument, text string) FetchDocument {
			doc.Text = text
			return doc
		},
	)
	fetchDocumentURLLens = L.MakeLens(
		func(doc FetchDocument) string { return doc.URL },
		func(doc FetchDocument, url string) FetchDocument {
			doc.URL = url
			return doc
		},
	)
	fetchDocumentMetadataLens = L.MakeLens(
		func(doc FetchDocument) map[string]string { return doc.Metadata },
		func(doc FetchDocument, metadata map[string]string) FetchDocument {
			doc.Metadata = metadata
			return doc
		},
	)
)

func fetchStrainDocument(s fetchState) IOE.IOEither[error, FetchDocument] {
	return F.Pipe5(
		GetStrainInput{},
		strainIDLens.Set(fetchStateInputIDLens.Get(s)),
		strainDetailLens.Set(detailFull),
		F.Pipe1(fetchStateClientLens.Get(s), toStrainState(fetchStateCtxLens.Get(s))),
		executeGetStrain,
		IOE.Map[error](F.Flow2(
			F.Bind1st(fullStrainView, fetchStateLinksLens.Get(s)),
			strainDocument,
		)),
	)
}

func fetchPlasmidDocument(s fetchState) IOE.IOEither[error, FetchDocument] {
	return F.Pipe5(
		GetPlasmidInput{},
		plasmidIDLens.Set(fetchStateInputIDLens.Get(s)),
		plasmidDetailLens.Set(detailSummary),
		F.Pipe1(fetchStateClientLens.Get(s), toPlasmidState(fetchStateCtxLens.Get(s))),
		executeGetPlasmid,
		IOE.Map[error](F.Flow2(
			F.Bind1st(summaryPlasmidView, fetchStateLinksLens.Get(s)),
			plasmidDocument,
		)),
	)
}

// labelWithID gives "label (id)", or id alone when label is empty.
func labelWithID(label, id string) string {
	return F.Pipe3(
		label,
		O.FromPredicate(S.IsNonEmpty),
		O.Map(F.Bind2nd(strutils.Space.Concat, F.Pipe1(id, strutils.Parens))),
		O.GetOrElse(F.Constant(id)),
	)
}

// toSearchDocument titles a hit with its stock label. It does not use
// DisplayText because that is the text of the matched field, which changes
// with the query.
func toSearchDocument(links linkout.Links, h graphql.StockSearchHit) SearchDocument {
	return F.Pipe3(
		SearchDocument{},
		searchDocumentIDLens.Set(F.Pipe1(h, searchHitIDLens.Get)),
		searchDocumentTitleLens.Set(F.Pipe2(
			h,
			searchHitStockLabelLens.Get,
			F.Bind2nd(labelWithID, F.Pipe1(h, searchHitIDLens.Get)),
		)),
		searchDocumentURLLens.Set(F.Pipe1(h, F.Bind1st(searchHitURL, links))),
	)
}

// toSearchDocuments keeps one document per stock. searchStocks returns one
// hit per matched field, sorted by score descending, so A.Uniq keeps the
// best hit of each stock.
func toSearchDocuments(links linkout.Links, r graphql.SearchResult) SearchDocuments {
	return F.Pipe1(
		SearchDocuments{},
		searchDocumentsResultsLens.Set(F.Pipe4(
			r,
			searchResultResultsLens.Get,
			A.Uniq(searchHitIDLens.Get),
			A.Map(F.Bind1st(toSearchDocument, links)),
			notNilSlice[SearchDocument],
		)),
	)
}

func newChatGPTSearchHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[ChatGPTSearchInput, SearchDocuments] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,

		input ChatGPTSearchInput,
	) (*mcp.CallToolResult, SearchDocuments, error) {
		out := F.Pipe6(
			IOE.Of[error](F.Pipe2(
				SearchStocksInput{},
				searchQueryLens.Set(chatgptQueryLens.Get(input)),
				searchLimitLens.Set(chatgptSearchLimit),
			)),
			IOE.ChainEitherK(ValidateSearchInput),
			IOE.Map[error](toSearchParams(ctx)),
			IOE.Chain(client.SearchStocks),
			IOE.Map[error](F.Bind1st(toSearchDocuments, links)),
			ioeutils.ToEither[error, SearchDocuments],
			E.Fold(
				toOutcomeError[SearchDocuments],
				toOutcomeSuccess[SearchDocuments],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func strainDocument(v StrainView) FetchDocument {
	id := strainViewIDLens.Get(v)

	return F.Pipe5(
		FetchDocument{},
		fetchDocumentIDLens.Set(id),
		fetchDocumentTitleLens.Set(F.Pipe2(v, strainViewLabelLens.Get, F.Bind2nd(labelWithID, id))),
		fetchDocumentTextLens.Set(renderStrainMarkdown(v)),
		fetchDocumentURLLens.Set(strainViewURLLens.Get(v)),
		fetchDocumentMetadataLens.Set(map[string]string{
			"entity":   strainEntityName,
			"in_stock": strconv.FormatBool(strainViewInStockLens.Get(v)),
			"species":  strainViewSpeciesLens.Get(v),
		}),
	)
}

func plasmidDocument(v PlasmidView) FetchDocument {
	id := plasmidViewIDLens.Get(v)

	return F.Pipe5(
		FetchDocument{},
		fetchDocumentIDLens.Set(id),
		fetchDocumentTitleLens.Set(F.Pipe2(v, plasmidViewNameLens.Get, F.Bind2nd(labelWithID, id))),
		fetchDocumentTextLens.Set(renderPlasmidMarkdown(v)),
		fetchDocumentURLLens.Set(plasmidViewURLLens.Get(v)),
		fetchDocumentMetadataLens.Set(map[string]string{
			"entity":   plasmidEntityName,
			"in_stock": strconv.FormatBool(plasmidViewInStockLens.Get(v)),
			"genbank":  plasmidViewGenbankLens.Get(v),
		}),
	)
}

func newChatGPTFetchHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[FetchInput, FetchDocument] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input FetchInput,
	) (*mcp.CallToolResult, FetchDocument, error) {
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateFetchInput),
			IOE.Map[error](F.Flow4(
				F.Pipe1(fetchState{}, F.Flip(fetchStateInputLens.Set)),
				fetchStateCtxLens.Set(ctx),
				fetchStateClientLens.Set(client),
				fetchStateLinksLens.Set(links),
			)),
			IOE.Chain(F.Bind1st(
				F.Pipe1(fetchRoutes, F.Bind1of3(routeEffect[fetchState, FetchDocument])),
				fetchStateInputIDLens.Get,
			)),
			ioeutils.ToEither[error, FetchDocument],
			E.Fold(
				toOutcomeError[FetchDocument],
				toOutcomeSuccess[FetchDocument],
			),
		)
		return out.Result, out.Output, out.Err
	}
}
