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

func autocompletePipeline(
	s autocompleteState,
) toolCallOutcome[graphql.AutocompleteResult] {
	return F.Pipe4(
		IOE.Of[error](s),
		IOE.ChainEitherK(validateAutocompleteState),
		IOE.Chain(executeAutocomplete),
		ioeutils.ToEither[error, graphql.AutocompleteResult],
		E.Fold(
			toOutcomeError[graphql.AutocompleteResult],
			toOutcomeSuccess[graphql.AutocompleteResult],
		),
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
		out := autocompletePipeline(autocompleteState{
			ctx:    ctx,
			client: client,
			input:  input,
		})
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

func searchPipeline(s searchState) toolCallOutcome[graphql.SearchResult] {
	return F.Pipe4(
		IOE.Of[error](s),
		IOE.ChainEitherK(validateSearchState),
		IOE.Chain(executeSearch),
		ioeutils.ToEither[error, graphql.SearchResult],
		E.Fold(
			toOutcomeError[graphql.SearchResult],
			toOutcomeSuccess[graphql.SearchResult],
		),
	)
}

func newSearchHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[SearchStocksInput, graphql.SearchResult] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input SearchStocksInput,
	) (*mcp.CallToolResult, graphql.SearchResult, error) {
		out := searchPipeline(searchState{
			ctx:    ctx,
			client: client,
			input:  input,
		})
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

func getStrainPipeline(s strainState) toolCallOutcome[graphql.StrainDetail] {
	return F.Pipe4(
		IOE.Of[error](s),
		IOE.ChainEitherK(validateStrainState),
		IOE.Chain(executeGetStrain),
		ioeutils.ToEither[error, graphql.StrainDetail],
		E.Fold(
			toOutcomeError[graphql.StrainDetail],
			toOutcomeSuccess[graphql.StrainDetail],
		),
	)
}

func newGetStrainHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetStrainInput, graphql.StrainDetail] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetStrainInput,
	) (*mcp.CallToolResult, graphql.StrainDetail, error) {
		out := getStrainPipeline(strainState{
			ctx:    ctx,
			client: client,
			input:  input,
		})
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

func getPlasmidPipeline(s plasmidState) toolCallOutcome[graphql.PlasmidDetail] {
	return F.Pipe4(
		IOE.Of[error](s),
		IOE.ChainEitherK(validatePlasmidState),
		IOE.Chain(executeGetPlasmid),
		ioeutils.ToEither[error, graphql.PlasmidDetail],
		E.Fold(
			toOutcomeError[graphql.PlasmidDetail],
			toOutcomeSuccess[graphql.PlasmidDetail],
		),
	)
}

func newGetPlasmidHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetPlasmidInput, graphql.PlasmidDetail] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPlasmidInput,
	) (*mcp.CallToolResult, graphql.PlasmidDetail, error) {
		out := getPlasmidPipeline(plasmidState{
			ctx:    ctx,
			client: client,
			input:  input,
		})
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

func listPhenotypeStrainsPipeline(
	s phenotypeListState,
) toolCallOutcome[graphql.PhenotypeStrainList] {
	return F.Pipe4(
		IOE.Of[error](s),
		IOE.ChainEitherK(validatePhenotypeState),
		IOE.Chain(executeListPhenotypeStrains),
		ioeutils.ToEither[error, graphql.PhenotypeStrainList],
		E.Fold(
			toOutcomeError[graphql.PhenotypeStrainList],
			toOutcomeSuccess[graphql.PhenotypeStrainList],
		),
	)
}

func newListPhenotypeStrainsHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[ListStrainsWithPhenotypeInput, graphql.PhenotypeStrainList] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input ListStrainsWithPhenotypeInput,
	) (*mcp.CallToolResult, graphql.PhenotypeStrainList, error) {
		out := listPhenotypeStrainsPipeline(phenotypeListState{
			ctx:    ctx,
			client: client,
			input:  input,
		})
		return out.Result, out.Output, out.Err
	}
}

// RegisterAll registers all five stock center MCP tools on srv.
func RegisterAll(srv *mcp.Server, client *graphql.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "autocomplete_stocks",
		Description: "Return type-ahead autocomplete suggestions for dictyBase strains and plasmids.",
	}, newAutocompleteHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search_stocks",
		Description: "Run a full-text search across dictyBase strains and plasmids.",
	}, newSearchHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_strain",
		Description: "Fetch a single dictyBase strain by its ID (e.g. DBS0350840).",
	}, newGetStrainHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_plasmid",
		Description: "Fetch a single dictyBase plasmid by its ID (e.g. DBP0391806).",
	}, newGetPlasmidHandler(client))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_strains_with_phenotype",
		Description: "List dictyBase strains annotated with a specific phenotype term.",
	}, newListPhenotypeStrainsHandler(client))
}
