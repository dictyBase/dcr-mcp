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

func validateAutocompleteState(
	s autocompleteState,
) E.Either[error, autocompleteState] {
	return F.Pipe1(
		E.Of[error](s),
		E.Bind(
			autoStateInputLens.Set,
			F.Flow2(autoStateInputLens.Get, ValidateAutocompleteInput),
		),
	)
}

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
		out := F.Pipe4(
			IOE.Of[error](autocompleteState{
				ctx:    ctx,
				client: client,
				input:  input,
			}),
			IOE.ChainEitherK(validateAutocompleteState),
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

func validateSearchState(s searchState) E.Either[error, searchState] {
	return F.Pipe1(
		E.Of[error](s),
		E.Bind(
			searchStateInputLens.Set,
			F.Flow2(searchStateInputLens.Get, ValidateSearchInput),
		),
	)
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
) mcp.ToolHandlerFor[SearchStocksInput, graphql.SearchResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input SearchStocksInput,
	) (*mcp.CallToolResult, graphql.SearchResult, error) {
		out := F.Pipe4(
			IOE.Of[error](searchState{
				ctx:    ctx,
				client: client,
				input:  input,
			}),
			IOE.ChainEitherK(validateSearchState),
			IOE.Chain(executeSearch),
			ioeutils.ToEither[error, graphql.SearchResult],
			E.Fold(
				toOutcomeError[graphql.SearchResult],
				toOutcomeSuccess[graphql.SearchResult],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func validateStrainState(s strainState) E.Either[error, strainState] {
	return F.Pipe1(
		E.Of[error](s),
		E.Bind(
			strainStateInputLens.Set,
			F.Flow2(strainStateInputLens.Get, ValidateStrainIDInput),
		),
	)
}

func executeGetStrain(s strainState) IOE.IOEither[error, graphql.StrainDetail] {
	return s.client.GetStrain(graphql.IDParams{
		Ctx: s.ctx,
		ID:  s.input.ID,
	})
}

func newGetStrainHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetStrainInput, graphql.StrainDetail] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetStrainInput,
	) (*mcp.CallToolResult, graphql.StrainDetail, error) {
		out := F.Pipe4(
			IOE.Of[error](strainState{
				ctx:    ctx,
				client: client,
				input:  input,
			}),
			IOE.ChainEitherK(validateStrainState),
			IOE.Chain(executeGetStrain),
			ioeutils.ToEither[error, graphql.StrainDetail],
			E.Fold(
				toOutcomeError[graphql.StrainDetail],
				toOutcomeSuccess[graphql.StrainDetail],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func validatePlasmidState(s plasmidState) E.Either[error, plasmidState] {
	return F.Pipe1(
		E.Of[error](s),
		E.Bind(
			plasmidStateInputLens.Set,
			F.Flow2(plasmidStateInputLens.Get, ValidatePlasmidIDInput),
		),
	)
}

func executeGetPlasmid(s plasmidState) IOE.IOEither[error, graphql.PlasmidDetail] {
	return s.client.GetPlasmid(graphql.IDParams{
		Ctx: s.ctx,
		ID:  s.input.ID,
	})
}

func newGetPlasmidHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetPlasmidInput, graphql.PlasmidDetail] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPlasmidInput,
	) (*mcp.CallToolResult, graphql.PlasmidDetail, error) {
		out := F.Pipe4(
			IOE.Of[error](plasmidState{
				ctx:    ctx,
				client: client,
				input:  input,
			}),
			IOE.ChainEitherK(validatePlasmidState),
			IOE.Chain(executeGetPlasmid),
			ioeutils.ToEither[error, graphql.PlasmidDetail],
			E.Fold(
				toOutcomeError[graphql.PlasmidDetail],
				toOutcomeSuccess[graphql.PlasmidDetail],
			),
		)
		return out.Result, out.Output, out.Err
	}
}

func validatePhenotypeState(
	s phenotypeListState,
) E.Either[error, phenotypeListState] {
	return F.Pipe1(
		E.Of[error](s),
		E.Bind(
			phenoStateInputLens.Set,
			F.Flow2(phenoStateInputLens.Get, ValidatePhenotypeListInput),
		),
	)
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
) mcp.ToolHandlerFor[ListStrainsWithPhenotypeInput, graphql.PhenotypeStrainList] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input ListStrainsWithPhenotypeInput,
	) (*mcp.CallToolResult, graphql.PhenotypeStrainList, error) {
		out := F.Pipe4(
			IOE.Of[error](phenotypeListState{
				ctx:    ctx,
				client: client,
				input:  input,
			}),
			IOE.ChainEitherK(validatePhenotypeState),
			IOE.Chain(executeListPhenotypeStrains),
			ioeutils.ToEither[error, graphql.PhenotypeStrainList],
			E.Fold(
				toOutcomeError[graphql.PhenotypeStrainList],
				toOutcomeSuccess[graphql.PhenotypeStrainList],
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
		Name:        toolListPhenotype,
		Title:       titleListPhenotype,
		Description: descListPhenotype,
		Annotations: readOnlyTool(titleListPhenotype),
	}, newListPhenotypeStrainsHandler(client))
}
