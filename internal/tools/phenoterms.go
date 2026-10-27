package tools

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	EQ "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/predicate"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	// PhenotypeTerms is the suggest_phenotype_terms output.
	PhenotypeTerms struct {
		Query string   `json:"query"`
		Terms []string `json:"terms"`
		Note  string   `json:"note"`
	}
)

const (
	phenotypeField = "PHENOTYPE"
	phenoTermsNote = "Pass a term exactly as written to list_strains_with_phenotype."
)

var (
	isPhenotypeHit = F.Pipe2(
		phenotypeField,
		EQ.Equals(predord.StringEq),
		P.ContraMap(func(h graphql.StockSearchHit) string { return h.Field }),
	)
	toPhenoTermsState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in SuggestPhenotypeTermsInput,
		) phenoTermsState {
			return phenoTermsState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
)

func hitDisplayText(h graphql.StockSearchHit) string { return h.DisplayText }

func executePhenoTermSearch(s phenoTermsState) IOE.IOEither[error, graphql.SearchResult] {
	return s.client.SearchStocks(graphql.SearchParams{
		Ctx:    s.ctx,
		Query:  s.input.Query,
		Limit:  phenotypeSampleLimit,
		Entity: graphql.StockEntityStrain,
	})
}

func toPhenotypeTerms(s phenoTermsState) PhenotypeTerms {
	return PhenotypeTerms{
		Query: s.input.Query,
		Terms: F.Pipe4(
			s.result.Results,
			A.Filter(isPhenotypeHit),
			A.Map(hitDisplayText),
			A.StrictUniq[string],
			takeAtMost[string](s.input.Limit),
		),
		Note: phenoTermsNote,
	}
}

func newSuggestPhenotypeTermsHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[SuggestPhenotypeTermsInput, PhenotypeTerms] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input SuggestPhenotypeTermsInput,
	) (*mcp.CallToolResult, PhenotypeTerms, error) {
		buildState := F.Pipe1(client, toPhenoTermsState(ctx))
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateSuggestPhenotypeTermsInput),
			IOE.Map[error](buildState),
			IOE.Bind(phenoTermsResultLens.Set, executePhenoTermSearch),
			IOE.Map[error](toPhenotypeTerms),
			ioeutils.ToEither[error, PhenotypeTerms],
			E.Fold(
				toOutcomeError[PhenotypeTerms],
				toOutcomeSuccess[PhenotypeTerms],
			),
		)
		return out.Result, out.Output, out.Err
	}
}
