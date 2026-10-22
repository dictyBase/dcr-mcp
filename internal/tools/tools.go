package tools

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func toOutcomeSuccess[Out any](output Out) toolCallOutcome[Out] {
	return toolCallOutcome[Out]{Output: output}
}

func toOutcomeError[Out any](err error) toolCallOutcome[Out] {
	return toolCallOutcome[Out]{Err: err}
}

func textCallResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// toOutcomeWithMarkdown returns render(output) as text content while the SDK
// still fills StructuredContent from output.
func toOutcomeWithMarkdown[Out any](
	render func(Out) string,
) func(Out) toolCallOutcome[Out] {
	return func(output Out) toolCallOutcome[Out] {
		return toolCallOutcome[Out]{
			Result: F.Pipe2(output, render, textCallResult),
			Output: output,
		}
	}
}

var (
	toAutocompleteState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in AutocompleteStocksInput,
		) autocompleteState {
			return autocompleteState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
	toSearchState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in SearchStocksInput,
		) searchState {
			return searchState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
	toStrainState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in GetStrainInput,
		) strainState {
			return strainState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
	toPlasmidState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in GetPlasmidInput,
		) plasmidState {
			return plasmidState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
	toPhenotypeListState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in ListStrainsWithPhenotypeInput,
		) phenotypeListState {
			return phenotypeListState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
)

func executeAutocomplete(
	s autocompleteState,
) IOE.IOEither[error, graphql.AutocompleteResult] {
	return s.client.ListStockSuggestions(graphql.AutocompleteParams{
		Ctx:    s.ctx,
		Query:  s.input.Query,
		Limit:  s.input.Limit,
		Entity: graphql.StockEntityType(s.input.Entity),
	})
}

func newAutocompleteHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[AutocompleteStocksInput, graphql.AutocompleteResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input AutocompleteStocksInput,
	) (*mcp.CallToolResult, graphql.AutocompleteResult, error) {
		buildState := F.Pipe1(client, toAutocompleteState(ctx))
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateAutocompleteInput),
			IOE.Map[error](buildState),
			IOE.Chain(executeAutocomplete),
			ioeutils.ToEither[error, graphql.AutocompleteResult],
			E.Fold(
				toOutcomeError[graphql.AutocompleteResult],
				toOutcomeSuccess[graphql.AutocompleteResult],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func executeSearch(s searchState) IOE.IOEither[error, graphql.SearchResult] {
	return s.client.SearchStocks(graphql.SearchParams{
		Ctx:    s.ctx,
		Query:  s.input.Query,
		Limit:  s.input.Limit,
		Entity: graphql.StockEntityType(s.input.Entity),
	})
}

func newSearchHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[SearchStocksInput, SearchResultView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input SearchStocksInput,
	) (*mcp.CallToolResult, SearchResultView, error) {
		buildState := F.Pipe1(client, toSearchState(ctx))
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateSearchInput),
			IOE.Map[error](buildState),
			IOE.Chain(executeSearch),
			IOE.Map[error](toSearchResultView),
			ioeutils.ToEither[error, SearchResultView],
			E.Fold(
				toOutcomeError[SearchResultView],
				toOutcomeWithMarkdown(renderSearchMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func executeGetStrain(s strainState) IOE.IOEither[error, graphql.StrainDetail] {
	return s.client.GetStrain(graphql.IDParams{
		Ctx: s.ctx,
		ID:  s.input.ID,
	})
}

func newGetStrainHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetStrainInput, StrainView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetStrainInput,
	) (*mcp.CallToolResult, StrainView, error) {
		view := strainViewFor(normalizeDetailString(input.Detail))
		buildState := F.Pipe1(client, toStrainState(ctx))
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateStrainIDInput),
			IOE.Map[error](buildState),
			IOE.Chain(F.Flow2(executeGetStrain, IOE.Map[error](view))),
			ioeutils.ToEither[error, StrainView],
			E.Fold(
				toOutcomeError[StrainView],
				toOutcomeWithMarkdown(renderStrainMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func executeGetPlasmid(s plasmidState) IOE.IOEither[error, graphql.PlasmidDetail] {
	return s.client.GetPlasmid(graphql.IDParams{
		Ctx: s.ctx,
		ID:  s.input.ID,
	})
}

func newGetPlasmidHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetPlasmidInput, PlasmidView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPlasmidInput,
	) (*mcp.CallToolResult, PlasmidView, error) {
		view := plasmidViewFor(normalizeDetailString(input.Detail))
		buildState := F.Pipe1(client, toPlasmidState(ctx))
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePlasmidIDInput),
			IOE.Map[error](buildState),
			IOE.Chain(F.Flow2(executeGetPlasmid, IOE.Map[error](view))),
			ioeutils.ToEither[error, PlasmidView],
			E.Fold(
				toOutcomeError[PlasmidView],
				toOutcomeWithMarkdown(renderPlasmidMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func newGetPlasmidSequenceHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetPlasmidSequenceInput, PlasmidSequence] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPlasmidSequenceInput,
	) (*mcp.CallToolResult, PlasmidSequence, error) {
		toSequence := F.Flow3(
			executeGetPlasmid,
			IOE.ChainEitherK(E.FromPredicate(hasSequence, errNoSequence)),
			IOE.Map[error](toPlasmidSequence),
		)
		buildState := F.Pipe1(client, toPlasmidState(ctx))
		out := F.Pipe5(
			IOE.Of[error](GetPlasmidInput{ID: input.ID, Detail: detailFull}),
			IOE.ChainEitherK(ValidatePlasmidIDInput),
			IOE.Map[error](buildState),
			IOE.Chain(toSequence),
			ioeutils.ToEither[error, PlasmidSequence],
			E.Fold(
				toOutcomeError[PlasmidSequence],
				toOutcomeSuccess[PlasmidSequence],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func executeListPhenotypeStrains(
	s phenotypeListState,
) IOE.IOEither[error, graphql.PhenotypeStrainList] {
	return s.client.ListStrainsWithPhenotype(graphql.PhenotypeListParams{
		Ctx:       s.ctx,
		Phenotype: s.input.Phenotype,
		Cursor:    s.input.Cursor,
		Limit:     s.input.Limit,
	})
}

func newListPhenotypeStrainsHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[ListStrainsWithPhenotypeInput, PhenotypeStrainListView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input ListStrainsWithPhenotypeInput,
	) (*mcp.CallToolResult, PhenotypeStrainListView, error) {
		buildState := F.Pipe1(client, toPhenotypeListState(ctx))
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePhenotypeListInput),
			IOE.Map[error](buildState),
			IOE.Chain(executeListPhenotypeStrains),
			IOE.Map[error](toPhenotypeStrainListView),
			ioeutils.ToEither[error, PhenotypeStrainListView],
			E.Fold(
				toOutcomeError[PhenotypeStrainListView],
				toOutcomeSuccess[PhenotypeStrainListView],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

// readOnlyTool marks a tool as a side-effect-free lookup against the closed
// dictyBase catalog so web clients can call it without a confirmation prompt.
func readOnlyTool(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:           title,
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: new(false),
		OpenWorldHint:   new(false),
	}
}

// RegisterAll registers all stock center MCP tools on srv.
func RegisterAll(srv *mcp.Server, client *graphql.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        toolAutocomplete,
		Title:       titleAutocomplete,
		Description: descAutocomplete,
		Annotations: readOnlyTool(titleAutocomplete),
	}, newAutocompleteHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        toolSearch,
		Title:       titleSearch,
		Description: descSearch,
		Annotations: readOnlyTool(titleSearch),
	}, newSearchHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        toolGetStrain,
		Title:       titleGetStrain,
		Description: descGetStrain,
		Annotations: readOnlyTool(titleGetStrain),
	}, newGetStrainHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        toolGetPlasmid,
		Title:       titleGetPlasmid,
		Description: descGetPlasmid,
		Annotations: readOnlyTool(titleGetPlasmid),
	}, newGetPlasmidHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        toolGetPlasmidSequence,
		Title:       titleGetPlasmidSequence,
		Description: descGetPlasmidSequence,
		Annotations: readOnlyTool(titleGetPlasmidSequence),
	}, newGetPlasmidSequenceHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        toolListPhenotype,
		Title:       titleListPhenotype,
		Description: descListPhenotype,
		Annotations: readOnlyTool(titleListPhenotype),
	}, newListPhenotypeStrainsHandler(client))
}
