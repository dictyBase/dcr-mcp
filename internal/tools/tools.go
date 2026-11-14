package tools

import (
	"context"
	A "github.com/IBM/fp-go/v2/array"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	autocompleteParamsCtxLens = L.MakeLens(
		func(x graphql.AutocompleteParams) context.Context { return x.Ctx },
		func(x graphql.AutocompleteParams, value context.Context) graphql.AutocompleteParams {
			x.Ctx = value
			return x
		},
	)
	autocompleteParamsQueryLens = L.MakeLens(
		func(x graphql.AutocompleteParams) string { return x.Query },
		func(x graphql.AutocompleteParams, value string) graphql.AutocompleteParams {
			x.Query = value
			return x
		},
	)
	autocompleteParamsLimitLens = L.MakeLens(
		func(x graphql.AutocompleteParams) int { return x.Limit },
		func(x graphql.AutocompleteParams, value int) graphql.AutocompleteParams {
			x.Limit = value
			return x
		},
	)
	autocompleteParamsEntityLens = L.MakeLens(
		func(x graphql.AutocompleteParams) graphql.StockEntityType { return x.Entity },
		func(x graphql.AutocompleteParams, value graphql.StockEntityType) graphql.AutocompleteParams {
			x.Entity = value
			return x
		},
	)
	phenotypeSearchParamsCtxLens = L.MakeLens(
		func(x graphql.PhenotypeSearchParams) context.Context { return x.Ctx },
		func(x graphql.PhenotypeSearchParams, value context.Context) graphql.PhenotypeSearchParams {
			x.Ctx = value
			return x
		},
	)
	phenotypeSearchParamsQueryLens = L.MakeLens(
		func(x graphql.PhenotypeSearchParams) string { return x.Query },
		func(x graphql.PhenotypeSearchParams, value string) graphql.PhenotypeSearchParams {
			x.Query = value
			return x
		},
	)
	phenotypeSearchParamsLimitLens = L.MakeLens(
		func(x graphql.PhenotypeSearchParams) int { return x.Limit },
		func(x graphql.PhenotypeSearchParams, value int) graphql.PhenotypeSearchParams {
			x.Limit = value
			return x
		},
	)
	phenotypeListParamsCtxLens = L.MakeLens(
		func(x graphql.PhenotypeListParams) context.Context { return x.Ctx },
		func(x graphql.PhenotypeListParams, value context.Context) graphql.PhenotypeListParams {
			x.Ctx = value
			return x
		},
	)
	phenotypeListParamsTermLens = L.MakeLens(
		func(x graphql.PhenotypeListParams) string { return x.PhenotypeTerm },
		func(x graphql.PhenotypeListParams, value string) graphql.PhenotypeListParams {
			x.PhenotypeTerm = value
			return x
		},
	)
	phenotypeListParamsCursorLens = L.MakeLens(
		func(x graphql.PhenotypeListParams) int { return x.Cursor },
		func(x graphql.PhenotypeListParams, value int) graphql.PhenotypeListParams {
			x.Cursor = value
			return x
		},
	)
	phenotypeListParamsLimitLens = L.MakeLens(
		func(x graphql.PhenotypeListParams) int { return x.Limit },
		func(x graphql.PhenotypeListParams, value int) graphql.PhenotypeListParams {
			x.Limit = value
			return x
		},
	)
	callToolResultContentLens = L.MakeLens(
		func(x mcp.CallToolResult) []mcp.Content { return x.Content },
		func(x mcp.CallToolResult, value []mcp.Content) mcp.CallToolResult {
			x.Content = value
			return x
		},
	)
	textContentTextLens = L.MakeLens(
		func(x mcp.TextContent) string { return x.Text },
		func(x mcp.TextContent, value string) mcp.TextContent {
			x.Text = value
			return x
		},
	)
	toolAnnotationsTitleLens = L.MakeLens(
		func(x mcp.ToolAnnotations) string { return x.Title },
		func(x mcp.ToolAnnotations, value string) mcp.ToolAnnotations {
			x.Title = value
			return x
		},
	)
	toolAnnotationsReadOnlyLens = L.MakeLens(
		func(x mcp.ToolAnnotations) bool { return x.ReadOnlyHint },
		func(x mcp.ToolAnnotations, value bool) mcp.ToolAnnotations {
			x.ReadOnlyHint = value
			return x
		},
	)
	toolAnnotationsIdempotentLens = L.MakeLens(
		func(x mcp.ToolAnnotations) bool { return x.IdempotentHint },
		func(x mcp.ToolAnnotations, value bool) mcp.ToolAnnotations {
			x.IdempotentHint = value
			return x
		},
	)
	toolAnnotationsDestructiveLens = L.MakeLens(
		func(x mcp.ToolAnnotations) *bool { return x.DestructiveHint },
		func(x mcp.ToolAnnotations, value *bool) mcp.ToolAnnotations {
			x.DestructiveHint = value
			return x
		},
	)
	toolAnnotationsOpenWorldLens = L.MakeLens(
		func(x mcp.ToolAnnotations) *bool { return x.OpenWorldHint },
		func(x mcp.ToolAnnotations, value *bool) mcp.ToolAnnotations {
			x.OpenWorldHint = value
			return x
		},
	)
	toolNameLens = L.MakeLens(
		func(t mcp.Tool) string { return t.Name },
		func(t mcp.Tool, name string) mcp.Tool {
			t.Name = name
			return t
		},
	)
	toolTitleLens = L.MakeLens(
		func(t mcp.Tool) string { return t.Title },
		func(t mcp.Tool, title string) mcp.Tool {
			t.Title = title
			return t
		},
	)
	toolDescriptionLens = L.MakeLens(
		func(t mcp.Tool) string { return t.Description },
		func(t mcp.Tool, description string) mcp.Tool {
			t.Description = description
			return t
		},
	)
	toolAnnotationsLens = L.MakeLens(
		func(t mcp.Tool) *mcp.ToolAnnotations { return t.Annotations },
		func(t mcp.Tool, annotations *mcp.ToolAnnotations) mcp.Tool {
			t.Annotations = annotations
			return t
		},
	)
	toStrainState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in GetStrainInput,
		) strainState {
			return F.Pipe3(
				strainState{},
				strainStateCtxLens.Set(ctx),
				strainStateClientLens.Set(client),
				strainStateInputLens.Set(in),
			)
		},
	)
	toPlasmidState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in GetPlasmidInput,
		) plasmidState {
			return F.Pipe3(
				plasmidState{},
				plasmidStateCtxLens.Set(ctx),
				plasmidStateClientLens.Set(client),
				plasmidStateInputLens.Set(in),
			)
		},
	)
)

func toAutocompleteParams(ctx context.Context, in AutocompleteStocksInput) graphql.AutocompleteParams {
	return F.Pipe4(
		graphql.AutocompleteParams{},
		autocompleteParamsCtxLens.Set(ctx),
		autocompleteParamsQueryLens.Set(in.Query),
		autocompleteParamsLimitLens.Set(in.Limit),
		autocompleteParamsEntityLens.Set(graphql.StockEntityType(in.Entity)),
	)
}

func toSearchParams(ctx context.Context, in SearchStocksInput) graphql.SearchParams {
	return F.Pipe4(
		graphql.SearchParams{},
		searchParamsCtxLens.Set(ctx),
		searchParamsQueryLens.Set(in.Query),
		searchParamsLimitLens.Set(in.Limit),
		searchParamsEntityLens.Set(graphql.StockEntityType(in.Entity)),
	)
}

func toPhenotypeSearchParams(ctx context.Context, in SearchPhenotypesInput) graphql.PhenotypeSearchParams {
	return F.Pipe3(
		graphql.PhenotypeSearchParams{},
		phenotypeSearchParamsCtxLens.Set(ctx),
		phenotypeSearchParamsQueryLens.Set(in.Query),
		phenotypeSearchParamsLimitLens.Set(in.Limit),
	)
}

func toPhenotypeListParams(ctx context.Context, in ListStrainsWithPhenotypeInput) graphql.PhenotypeListParams {
	return F.Pipe4(
		graphql.PhenotypeListParams{},
		phenotypeListParamsCtxLens.Set(ctx),
		phenotypeListParamsTermLens.Set(in.PhenotypeTerm),
		phenotypeListParamsCursorLens.Set(in.Cursor),
		phenotypeListParamsLimitLens.Set(in.Limit),
	)
}

// outcomeResultLens, outcomeOutputLens, and outcomeErrLens are factories,
// because a package variable cannot hold a generic lens.
func outcomeResultLens[Out any]() L.Lens[toolCallOutcome[Out], *mcp.CallToolResult] {
	return L.MakeLens(
		func(o toolCallOutcome[Out]) *mcp.CallToolResult { return o.Result },
		func(o toolCallOutcome[Out], value *mcp.CallToolResult) toolCallOutcome[Out] {
			o.Result = value
			return o
		},
	)
}

func outcomeOutputLens[Out any]() L.Lens[toolCallOutcome[Out], Out] {
	return L.MakeLens(
		func(o toolCallOutcome[Out]) Out { return o.Output },
		func(o toolCallOutcome[Out], value Out) toolCallOutcome[Out] {
			o.Output = value
			return o
		},
	)
}

func outcomeErrLens[Out any]() L.Lens[toolCallOutcome[Out], error] {
	return L.MakeLens(
		func(o toolCallOutcome[Out]) error { return o.Err },
		func(o toolCallOutcome[Out], value error) toolCallOutcome[Out] {
			o.Err = value
			return o
		},
	)
}

func toOutcomeSuccess[Out any](output Out) toolCallOutcome[Out] {
	return F.Pipe1(toolCallOutcome[Out]{}, outcomeOutputLens[Out]().Set(output))
}

func toOutcomeError[Out any](err error) toolCallOutcome[Out] {
	return F.Pipe1(toolCallOutcome[Out]{}, outcomeErrLens[Out]().Set(err))
}

// toOutcomeWithMarkdown returns render(output) as text content while the SDK
// still fills StructuredContent from output.
func toOutcomeWithMarkdown[Out any](render func(Out) string, output Out) toolCallOutcome[Out] {
	return F.Pipe2(
		toolCallOutcome[Out]{},
		outcomeResultLens[Out]().Set(new(F.Pipe1(
			mcp.CallToolResult{},
			callToolResultContentLens.Set(A.Of[mcp.Content](new(F.Pipe2(
				output,
				render,
				F.Pipe1(mcp.TextContent{}, F.Flip(textContentTextLens.Set)),
			)))),
		))),
		outcomeOutputLens[Out]().Set(output),
	)
}

func newAutocompleteHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[AutocompleteStocksInput, graphql.AutocompleteResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input AutocompleteStocksInput,
	) (*mcp.CallToolResult, graphql.AutocompleteResult, error) {
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateAutocompleteInput),
			IOE.Map[error](F.Bind1st(toAutocompleteParams, ctx)),
			IOE.Chain(client.ListStockSuggestions),
			ioeutils.ToEither[error, graphql.AutocompleteResult],
			E.Fold(
				toOutcomeError[graphql.AutocompleteResult],
				toOutcomeSuccess[graphql.AutocompleteResult],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func newSearchHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[SearchStocksInput, SearchResultView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input SearchStocksInput,
	) (*mcp.CallToolResult, SearchResultView, error) {
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateSearchInput),
			IOE.Map[error](F.Bind1st(toSearchParams, ctx)),
			IOE.Chain(client.SearchStocks),
			IOE.Map[error](F.Bind1st(toSearchResultView, links)),
			ioeutils.ToEither[error, SearchResultView],
			E.Fold(
				toOutcomeError[SearchResultView],
				F.Bind1st(toOutcomeWithMarkdown[SearchResultView], renderSearchMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func newSearchPhenotypesHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[SearchPhenotypesInput, PhenotypeSearchView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input SearchPhenotypesInput,
	) (*mcp.CallToolResult, PhenotypeSearchView, error) {
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateSearchPhenotypesInput),
			IOE.Map[error](F.Bind1st(toPhenotypeSearchParams, ctx)),
			IOE.Chain(client.SearchPhenotypes),
			IOE.Map[error](F.Bind1st(toPhenotypeSearchView, links)),
			ioeutils.ToEither[error, PhenotypeSearchView],
			E.Fold(
				toOutcomeError[PhenotypeSearchView],
				F.Bind1st(toOutcomeWithMarkdown[PhenotypeSearchView], renderPhenotypeSearchMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func executeGetStrain(s strainState) IOE.IOEither[error, graphql.StrainDetail] {
	return F.Pipe3(
		graphql.IDParams{},
		idParamsCtxLens.Set(strainStateCtxLens.Get(s)),
		idParamsIDLens.Set(strainStateInputIDLens.Get(s)),
		strainStateClientLens.Get(s).GetStrain,
	)
}

func newGetStrainHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[GetStrainInput, StrainView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetStrainInput,
	) (*mcp.CallToolResult, StrainView, error) {
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateStrainIDInput),
			IOE.Map[error](F.Pipe1(client, toStrainState(ctx))),
			IOE.Chain(executeGetStrain),
			IOE.Map[error](F.Pipe3(
				input,
				strainDetailLens.Get,
				normalizeDetailString,
				F.Bind1st(strainViewFor, links),
			)),
			ioeutils.ToEither[error, StrainView],
			E.Fold(
				toOutcomeError[StrainView],
				F.Bind1st(toOutcomeWithMarkdown[StrainView], renderStrainMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func executeGetPlasmid(s plasmidState) IOE.IOEither[error, graphql.PlasmidDetail] {
	return F.Pipe3(
		graphql.IDParams{},
		idParamsCtxLens.Set(plasmidStateCtxLens.Get(s)),
		idParamsIDLens.Set(plasmidStateInputIDLens.Get(s)),
		plasmidStateClientLens.Get(s).GetPlasmid,
	)
}

func newGetPlasmidHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[GetPlasmidInput, PlasmidView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPlasmidInput,
	) (*mcp.CallToolResult, PlasmidView, error) {
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePlasmidIDInput),
			IOE.Map[error](F.Pipe1(client, toPlasmidState(ctx))),
			IOE.Chain(executeGetPlasmid),
			IOE.Map[error](F.Pipe3(input, plasmidDetailLens.Get, normalizeDetailString, F.Bind1st(plasmidViewFor, links))),
			ioeutils.ToEither[error, PlasmidView],
			E.Fold(
				toOutcomeError[PlasmidView],
				F.Bind1st(toOutcomeWithMarkdown[PlasmidView], renderPlasmidMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func newGetPlasmidSequenceHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[GetPlasmidSequenceInput, PlasmidSequence] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPlasmidSequenceInput,
	) (*mcp.CallToolResult, PlasmidSequence, error) {
		out := F.Pipe7(
			IOE.Of[error](F.Pipe2(
				GetPlasmidInput{},
				plasmidIDLens.Set(sequenceInputIDLens.Get(input)),
				plasmidDetailLens.Set(detailFull),
			)),
			IOE.ChainEitherK(ValidatePlasmidIDInput),
			IOE.Map[error](F.Pipe1(client, toPlasmidState(ctx))),
			IOE.Chain(executeGetPlasmid),
			IOE.ChainEitherK(E.FromPredicate(hasSequence, errNoSequence)),
			IOE.Map[error](F.Bind1st(toPlasmidSequence, links)),
			ioeutils.ToEither[error, PlasmidSequence],
			E.Fold(
				toOutcomeError[PlasmidSequence],
				toOutcomeSuccess[PlasmidSequence],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func newListPhenotypeStrainsHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[ListStrainsWithPhenotypeInput, PhenotypeStrainListView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input ListStrainsWithPhenotypeInput,
	) (*mcp.CallToolResult, PhenotypeStrainListView, error) {
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePhenotypeListInput),
			IOE.Map[error](F.Bind1st(toPhenotypeListParams, ctx)),
			IOE.Chain(func(p graphql.PhenotypeListParams) IOE.IOEither[error, PhenotypeStrainListView] {
				return F.Pipe1(
					client.ListStrainsWithPhenotype(p),
					IOE.Map[error](F.Bind1st(F.Pipe1(links, F.Bind1of3(phenotypeStrainListView)), phenoTermParamsLens.Get(p))),
				)
			}),
			ioeutils.ToEither[error, PhenotypeStrainListView],
			E.Fold(
				toOutcomeError[PhenotypeStrainListView],
				F.Bind1st(toOutcomeWithMarkdown[PhenotypeStrainListView], renderPhenotypeStrainListMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

// readOnlyTool marks a tool as a side-effect-free lookup against the closed
// dictyBase catalog so web clients can call it without a confirmation prompt.
func readOnlyTool(title string) *mcp.ToolAnnotations {
	return new(F.Pipe5(
		mcp.ToolAnnotations{},
		toolAnnotationsTitleLens.Set(title),
		toolAnnotationsReadOnlyLens.Set(true),
		toolAnnotationsIdempotentLens.Set(true),
		toolAnnotationsDestructiveLens.Set(new(false)),
		toolAnnotationsOpenWorldLens.Set(new(false)),
	))
}

// readOnlyDescriptor builds the read-only tool descriptor for one tool.
func readOnlyDescriptor(name, title, description string) *mcp.Tool {
	return new(F.Pipe4(
		mcp.Tool{},
		toolNameLens.Set(name),
		toolTitleLens.Set(title),
		toolDescriptionLens.Set(description),
		toolAnnotationsLens.Set(readOnlyTool(title)),
	))
}

// RegisterAll registers all stock center MCP tools on srv.
func RegisterAll(srv *mcp.Server, client *graphql.Client, links linkout.Links) {
	mcp.AddTool(srv,
		readOnlyDescriptor(toolAutocomplete, titleAutocomplete, descAutocomplete),
		newAutocompleteHandler(client))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolSearch, titleSearch, descSearch),
		newSearchHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolSearchPhenotypes, titleSearchPhenotypes, descSearchPhenotypes),
		newSearchPhenotypesHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolGetStrain, titleGetStrain, descGetStrain),
		newGetStrainHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolGetPlasmid, titleGetPlasmid, descGetPlasmid),
		newGetPlasmidHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolGetPlasmidSequence, titleGetPlasmidSequence, descGetPlasmidSequence),
		newGetPlasmidSequenceHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolListPhenotype, titleListPhenotype, descListPhenotype),
		newListPhenotypeStrainsHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolChatGPTSearch, titleChatGPTSearch, descChatGPTSearch),
		newChatGPTSearchHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolChatGPTFetch, titleChatGPTFetch, descChatGPTFetch),
		newChatGPTFetchHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolFindByGene, titleFindByGene, descFindByGene),
		newFindStocksByGeneHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolCheckAvailability, titleCheckAvailability, descCheckAvailability),
		newCheckAvailabilityHandler(client, links))

	mcp.AddTool(srv,
		readOnlyDescriptor(toolGetPublication, titleGetPublication, descGetPublication),
		newGetPublicationHandler(client, links))
}
