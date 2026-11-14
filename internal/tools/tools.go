package tools

import (
	"context"

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
	toAutocompleteParams = F.Curry2(
		func(ctx context.Context, in AutocompleteStocksInput) graphql.AutocompleteParams {
			return graphql.AutocompleteParams{
				Ctx:    ctx,
				Query:  in.Query,
				Limit:  in.Limit,
				Entity: graphql.StockEntityType(in.Entity),
			}
		},
	)
	toSearchParams = F.Curry2(
		func(ctx context.Context, in SearchStocksInput) graphql.SearchParams {
			return graphql.SearchParams{
				Ctx:    ctx,
				Query:  in.Query,
				Limit:  in.Limit,
				Entity: graphql.StockEntityType(in.Entity),
			}
		},
	)
	toPhenotypeSearchParams = F.Curry2(
		func(ctx context.Context, in SearchPhenotypesInput) graphql.PhenotypeSearchParams {
			return graphql.PhenotypeSearchParams{
				Ctx:   ctx,
				Query: in.Query,
				Limit: in.Limit,
			}
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
	toPhenotypeListParams = F.Curry2(
		func(ctx context.Context, in ListStrainsWithPhenotypeInput) graphql.PhenotypeListParams {
			return graphql.PhenotypeListParams{
				Ctx:           ctx,
				PhenotypeTerm: in.PhenotypeTerm,
				Cursor:        in.Cursor,
				Limit:         in.Limit,
			}
		},
	)
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
			IOE.Map[error](toAutocompleteParams(ctx)),
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
			IOE.Map[error](toSearchParams(ctx)),
			IOE.Chain(client.SearchStocks),
			IOE.Map[error](F.Bind1st(toSearchResultView, links)),
			ioeutils.ToEither[error, SearchResultView],
			E.Fold(
				toOutcomeError[SearchResultView],
				toOutcomeWithMarkdown(renderSearchMarkdown),
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
			IOE.Map[error](toPhenotypeSearchParams(ctx)),
			IOE.Chain(client.SearchPhenotypes),
			IOE.Map[error](F.Bind1st(toPhenotypeSearchView, links)),
			ioeutils.ToEither[error, PhenotypeSearchView],
			E.Fold(
				toOutcomeError[PhenotypeSearchView],
				toOutcomeWithMarkdown(renderPhenotypeSearchMarkdown),
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
				toOutcomeWithMarkdown(renderStrainMarkdown),
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
				toOutcomeWithMarkdown(renderPlasmidMarkdown),
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
			IOE.Map[error](toPhenotypeListParams(ctx)),
			IOE.Chain(func(p graphql.PhenotypeListParams) IOE.IOEither[error, PhenotypeStrainListView] {
				return F.Pipe1(
					client.ListStrainsWithPhenotype(p),
					IOE.Map[error](F.Bind1st(F.Pipe1(links, F.Bind1of3(phenotypeStrainListView)), phenoTermParamsLens.Get(p))),
				)
			}),
			ioeutils.ToEither[error, PhenotypeStrainListView],
			E.Fold(
				toOutcomeError[PhenotypeStrainListView],
				toOutcomeWithMarkdown(renderPhenotypeStrainListMarkdown),
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
