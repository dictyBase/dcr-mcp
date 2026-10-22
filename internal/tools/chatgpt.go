package tools

import (
	"context"
	"fmt"
	"strconv"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const chatgptSearchLimit = 20

// SearchDocument is one result in the ChatGPT search contract.
type SearchDocument struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// SearchDocuments is the ChatGPT search output.
type SearchDocuments struct {
	Results []SearchDocument `json:"results"`
}

func hitID(h graphql.StockSearchHit) string { return h.ID }

func hitLabel(h graphql.StockSearchHit) string {
	return F.Pipe2(
		h.StrainLabel,
		O.FromPredicate(S.IsNonEmpty),
		O.GetOrElse(F.Constant(h.DisplayText)),
	)
}

func toSearchDocument(h graphql.StockSearchHit) SearchDocument {
	return SearchDocument{
		ID:    h.ID,
		Title: fmt.Sprintf("%s (%s)", hitLabel(h), h.ID),
		URL:   StockURL(h.Entity, h.ID),
	}
}

// toSearchDocuments keeps one document per stock; searchStocks returns one hit per matched field.
func toSearchDocuments(r graphql.SearchResult) SearchDocuments {
	return SearchDocuments{
		Results: F.Pipe2(
			r.Results,
			A.Uniq(hitID),
			A.Map(toSearchDocument),
		),
	}
}

func newChatGPTSearchHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[ChatGPTSearchInput, SearchDocuments] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input ChatGPTSearchInput,
	) (*mcp.CallToolResult, SearchDocuments, error) {
		buildState := F.Pipe1(client, toSearchState(ctx))
		out := F.Pipe6(
			IOE.Of[error](SearchStocksInput{Query: input.Query, Limit: chatgptSearchLimit}),
			IOE.ChainEitherK(ValidateSearchInput),
			IOE.Map[error](buildState),
			IOE.Chain(executeSearch),
			IOE.Map[error](toSearchDocuments),
			ioeutils.ToEither[error, SearchDocuments],
			E.Fold(
				toOutcomeError[SearchDocuments],
				toOutcomeSuccess[SearchDocuments],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

// FetchDocument is the ChatGPT fetch output.
type FetchDocument struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Text     string            `json:"text"`
	URL      string            `json:"url"`
	Metadata map[string]string `json:"metadata"`
}

type fetchRoute func(fetchState) IOE.IOEither[error, FetchDocument]

var fetchRoutes = map[string]fetchRoute{
	strainIDPrefix:  fetchStrainDocument,
	plasmidIDPrefix: fetchPlasmidDocument,
}

func strainDocument(v StrainView) FetchDocument {
	return FetchDocument{
		ID:    v.ID,
		Title: fmt.Sprintf("%s (%s)", v.Label, v.ID),
		Text:  renderStrainMarkdown(v),
		URL:   v.URL,
		Metadata: map[string]string{
			"entity":   "strain",
			"in_stock": strconv.FormatBool(v.InStock),
			"species":  v.Species,
		},
	}
}

func plasmidDocument(v PlasmidView) FetchDocument {
	return FetchDocument{
		ID:    v.ID,
		Title: fmt.Sprintf("%s (%s)", v.Name, v.ID),
		Text:  renderPlasmidMarkdown(v),
		URL:   v.URL,
		Metadata: map[string]string{
			"entity":   "plasmid",
			"in_stock": strconv.FormatBool(v.InStock),
			"genbank":  v.GenbankAccession,
		},
	}
}

var (
	fetchStrainDocument = F.Flow3(
		func(s fetchState) strainState {
			return strainState{
				ctx:    s.ctx,
				client: s.client,
				input:  GetStrainInput{ID: s.input.ID, Detail: detailFull},
			}
		},
		executeGetStrain,
		IOE.Map[error](F.Flow2(fullStrainView, strainDocument)),
	)
	fetchPlasmidDocument = F.Flow3(
		func(s fetchState) plasmidState {
			return plasmidState{
				ctx:    s.ctx,
				client: s.client,
				input:  GetPlasmidInput{ID: s.input.ID, Detail: detailSummary},
			}
		},
		executeGetPlasmid,
		IOE.Map[error](F.Flow2(summaryPlasmidView, plasmidDocument)),
	)
)

func routeFetch(s fetchState) IOE.IOEither[error, FetchDocument] {
	return F.Pipe2(
		fetchRoutes,
		R.Lookup[fetchRoute](IDPrefix(s.input.ID)),
		O.Fold(
			func() IOE.IOEither[error, FetchDocument] {
				return IOE.Left[FetchDocument](unknownIDError(s.input.ID))
			},
			func(route fetchRoute) IOE.IOEither[error, FetchDocument] { return route(s) },
		),
	)
}

func newChatGPTFetchHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[FetchInput, FetchDocument] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input FetchInput,
	) (*mcp.CallToolResult, FetchDocument, error) {
		buildState := F.Pipe1(client, toFetchState(ctx))
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateFetchInput),
			IOE.Map[error](buildState),
			IOE.Chain(routeFetch),
			ioeutils.ToEither[error, FetchDocument],
			E.Fold(
				toOutcomeError[FetchDocument],
				toOutcomeSuccess[FetchDocument],
			),
		)
		return out.Result, out.Output, out.Err
	}
}
