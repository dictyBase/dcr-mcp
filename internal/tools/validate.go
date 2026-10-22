package tools

import (
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	ORD "github.com/IBM/fp-go/v2/ord"
	P "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	eithererr "github.com/dictyBase/fp-go-loom/eithererr"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/graphql"
)

var (
	knownDetails = map[string]string{
		detailSummary: detailSummary,
		detailFull:    detailFull,
	}

	knownEntities = map[string]graphql.StockEntityType{
		string(graphql.StockEntityAll):     graphql.StockEntityAll,
		string(graphql.StockEntityStrain):  graphql.StockEntityStrain,
		string(graphql.StockEntityPlasmid): graphql.StockEntityPlasmid,
	}

	isNonNegInt = F.Pipe1(0, ORD.Geq(predord.IntOrd))

	errAutocompleteQueryShort = eithererr.ConstErr[string](
		"query must be at least 3 characters",
	)
	errAutocompleteLimitRange = eithererr.ConstErr[int](
		"limit must be between 1 and 50",
	)
	errSearchQueryShort = eithererr.ConstErr[string](
		"query must be at least 2 characters",
	)
	errSearchLimitRange = eithererr.ConstErr[int](
		"limit must be between 1 and 100",
	)
	errEmptyStrainID = eithererr.ConstErr[string](
		"strain id is required",
	)
	errEmptyPlasmidID = eithererr.ConstErr[string](
		"plasmid id is required",
	)
	errEmptyPhenotype = eithererr.ConstErr[string](
		"phenotype is required",
	)
	errNegativeCursor = eithererr.ConstErr[int](
		"cursor must be non-negative",
	)
	errPhenotypeLimitRange = eithererr.ConstErr[int](
		"limit must be between 1 and 100",
	)
	errFetchIDShort = eithererr.ConstErr[string](
		"id is required: a DBS (strain) or DBP (plasmid) id",
	)
	errGeneShort = eithererr.ConstErr[string](
		"gene must be at least 2 characters",
	)
	errPhenoTermQueryShort = eithererr.ConstErr[string](
		"query must be at least 3 characters",
	)
	errPhenoTermLimitRange = eithererr.ConstErr[int](
		"limit must be between 1 and 25",
	)
	errAvailabilityIDCount = eithererr.ConstErr[[]string](
		"ids must contain between 1 and 25 non-empty ids",
	)
	errEmptyPublicationID = eithererr.ConstErr[string](
		"publication id is required",
	)

	validIDCount = F.Pipe1(
		predarrays.MaxLen[string](maxAvailabilityIDs),
		P.And(predarrays.MinLen[string](1)),
	)
)

func defaultIfZeroInt(dflt int) func(int) int {
	return F.Flow2(
		O.FromPredicate(P.IsNonZero[int]()),
		O.GetOrElse(F.Constant(dflt)),
	)
}

const defaultEntityStr = string(graphql.StockEntityAll)

func normalizeEntityString(raw string) string {
	return F.Pipe3(
		raw,
		strings.TrimSpace,
		O.FromPredicate(S.IsNonEmpty),
		O.Fold(
			F.Constant(defaultEntityStr),
			strings.ToUpper,
		),
	)
}

func parseEntity(raw string) E.Either[error, string] {
	norm := normalizeEntityString(raw)
	lookup := R.Lookup[graphql.StockEntityType](norm)
	toErr := F.Constant(fmt.Errorf(
		"invalid entity %q: expected ALL, STRAIN, or PLASMID",
		raw,
	))
	return F.Pipe3(
		knownEntities,
		lookup,
		E.FromOption[graphql.StockEntityType](toErr),
		E.Map[error](func(et graphql.StockEntityType) string {
			return string(et)
		}),
	)
}

func validateAutocompleteQuery(in AutocompleteStocksInput) E.Either[error, string] {
	return F.Pipe2(
		in.Query,
		strings.TrimSpace,
		E.FromPredicate(
			predord.MinStrLen(minAutocompleteQueryLen),
			errAutocompleteQueryShort,
		),
	)
}

func validateAutocompleteLimit(in AutocompleteStocksInput) E.Either[error, int] {
	resolveDefault := defaultIfZeroInt(defaultAutocompleteLimit)
	return F.Pipe2(
		in.Limit,
		resolveDefault,
		E.FromPredicate(
			predord.IntBetweenInclusive(1, maxAutocompleteLimit),
			errAutocompleteLimitRange,
		),
	)
}

func validateAutocompleteEntity(in AutocompleteStocksInput) E.Either[error, string] {
	return parseEntity(in.Entity)
}

// ValidateAutocompleteInput trims, defaults, and validates AutocompleteStocksInput.
func ValidateAutocompleteInput(
	in AutocompleteStocksInput,
) E.Either[error, AutocompleteStocksInput] {
	return F.Pipe3(
		E.Of[error](in),
		E.Bind(autoQueryLens.Set, validateAutocompleteQuery),
		E.Bind(autoLimitLens.Set, validateAutocompleteLimit),
		E.Bind(autoEntityLens.Set, validateAutocompleteEntity),
	)
}

func validateSearchQuery(in SearchStocksInput) E.Either[error, string] {
	return F.Pipe2(
		in.Query,
		strings.TrimSpace,
		E.FromPredicate(
			predord.MinStrLen(minSearchQueryLen),
			errSearchQueryShort,
		),
	)
}

func validateSearchLimit(in SearchStocksInput) E.Either[error, int] {
	resolveDefault := defaultIfZeroInt(defaultSearchLimit)
	return F.Pipe2(
		in.Limit,
		resolveDefault,
		E.FromPredicate(
			predord.IntBetweenInclusive(1, maxSearchLimit),
			errSearchLimitRange,
		),
	)
}

func validateSearchEntity(in SearchStocksInput) E.Either[error, string] {
	return parseEntity(in.Entity)
}

// ValidateSearchInput trims, defaults, and validates SearchStocksInput.
func ValidateSearchInput(in SearchStocksInput) E.Either[error, SearchStocksInput] {
	return F.Pipe3(
		E.Of[error](in),
		E.Bind(searchQueryLens.Set, validateSearchQuery),
		E.Bind(searchLimitLens.Set, validateSearchLimit),
		E.Bind(searchEntityLens.Set, validateSearchEntity),
	)
}

func validateStrainID(in GetStrainInput) E.Either[error, string] {
	return F.Pipe2(
		in.ID,
		strings.TrimSpace,
		E.FromPredicate(S.IsNonEmpty, errEmptyStrainID),
	)
}

func normalizeDetailString(raw string) string {
	return F.Pipe3(
		raw,
		strings.TrimSpace,
		O.FromPredicate(S.IsNonEmpty),
		O.Fold(
			F.Constant(detailSummary),
			strings.ToLower,
		),
	)
}

func parseDetail(raw string) E.Either[error, string] {
	toErr := F.Constant(fmt.Errorf(
		"invalid detail %q: expected summary or full",
		raw,
	))
	return F.Pipe2(
		knownDetails,
		R.Lookup[string](normalizeDetailString(raw)),
		E.FromOption[string](toErr),
	)
}

func validateStrainDetail(in GetStrainInput) E.Either[error, string] {
	return parseDetail(in.Detail)
}

// ValidateStrainIDInput trims and validates GetStrainInput and defaults detail to summary.
func ValidateStrainIDInput(in GetStrainInput) E.Either[error, GetStrainInput] {
	return F.Pipe2(
		E.Of[error](in),
		E.Bind(strainIDLens.Set, validateStrainID),
		E.Bind(strainDetailLens.Set, validateStrainDetail),
	)
}

func validatePlasmidID(in GetPlasmidInput) E.Either[error, string] {
	return F.Pipe2(
		in.ID,
		strings.TrimSpace,
		E.FromPredicate(S.IsNonEmpty, errEmptyPlasmidID),
	)
}

func validatePlasmidDetail(in GetPlasmidInput) E.Either[error, string] {
	return parseDetail(in.Detail)
}

// ValidatePlasmidIDInput trims and validates GetPlasmidInput and defaults detail to summary.
func ValidatePlasmidIDInput(in GetPlasmidInput) E.Either[error, GetPlasmidInput] {
	return F.Pipe2(
		E.Of[error](in),
		E.Bind(plasmidIDLens.Set, validatePlasmidID),
		E.Bind(plasmidDetailLens.Set, validatePlasmidDetail),
	)
}

func validatePhenotypeName(in ListStrainsWithPhenotypeInput) E.Either[error, string] {
	return F.Pipe2(
		in.Phenotype,
		strings.TrimSpace,
		E.FromPredicate(S.IsNonEmpty, errEmptyPhenotype),
	)
}

func validatePhenotypeCursor(in ListStrainsWithPhenotypeInput) E.Either[error, int] {
	return F.Pipe1(
		in.Cursor,
		E.FromPredicate(isNonNegInt, errNegativeCursor),
	)
}

func validatePhenotypeLimit(in ListStrainsWithPhenotypeInput) E.Either[error, int] {
	resolveDefault := defaultIfZeroInt(defaultPhenotypeLimit)
	return F.Pipe2(
		in.Limit,
		resolveDefault,
		E.FromPredicate(
			predord.IntBetweenInclusive(1, maxPhenotypeLimit),
			errPhenotypeLimitRange,
		),
	)
}

// ValidatePhenotypeListInput trims, defaults, and validates ListStrainsWithPhenotypeInput.
func ValidatePhenotypeListInput(
	in ListStrainsWithPhenotypeInput,
) E.Either[error, ListStrainsWithPhenotypeInput] {
	return F.Pipe3(
		E.Of[error](in),
		E.Bind(phenoNameLens.Set, validatePhenotypeName),
		E.Bind(phenoCursorLens.Set, validatePhenotypeCursor),
		E.Bind(phenoLimitLens.Set, validatePhenotypeLimit),
	)
}

func validateFetchID(in FetchInput) E.Either[error, string] {
	return F.Pipe3(
		in.ID,
		strings.TrimSpace,
		strings.ToUpper,
		E.FromPredicate(predord.MinStrLen(stockIDPrefixLen+1), errFetchIDShort),
	)
}

// ValidateFetchInput trims, upper-cases, and validates FetchInput.
func ValidateFetchInput(in FetchInput) E.Either[error, FetchInput] {
	return F.Pipe1(
		E.Of[error](in),
		E.Bind(fetchIDLens.Set, validateFetchID),
	)
}

func validateGeneName(in FindStocksByGeneInput) E.Either[error, string] {
	return F.Pipe2(
		in.Gene,
		strings.TrimSpace,
		E.FromPredicate(predord.MinStrLen(minGeneQueryLen), errGeneShort),
	)
}

// ValidateFindStocksByGeneInput trims and validates FindStocksByGeneInput.
func ValidateFindStocksByGeneInput(in FindStocksByGeneInput) E.Either[error, FindStocksByGeneInput] {
	return F.Pipe1(
		E.Of[error](in),
		E.Bind(geneNameLens.Set, validateGeneName),
	)
}

func validatePhenoTermQuery(in SuggestPhenotypeTermsInput) E.Either[error, string] {
	return F.Pipe2(
		in.Query,
		strings.TrimSpace,
		E.FromPredicate(predord.MinStrLen(minPhenoTermQueryLen), errPhenoTermQueryShort),
	)
}

func validatePhenoTermLimit(in SuggestPhenotypeTermsInput) E.Either[error, int] {
	resolveDefault := defaultIfZeroInt(defaultPhenoTermLimit)
	return F.Pipe2(
		in.Limit,
		resolveDefault,
		E.FromPredicate(
			predord.IntBetweenInclusive(1, maxPhenoTermLimit),
			errPhenoTermLimitRange,
		),
	)
}

// ValidateSuggestPhenotypeTermsInput trims, defaults, and validates SuggestPhenotypeTermsInput.
func ValidateSuggestPhenotypeTermsInput(
	in SuggestPhenotypeTermsInput,
) E.Either[error, SuggestPhenotypeTermsInput] {
	return F.Pipe2(
		E.Of[error](in),
		E.Bind(phenoTermQueryLens.Set, validatePhenoTermQuery),
		E.Bind(phenoTermLimitLens.Set, validatePhenoTermLimit),
	)
}

func validateAvailabilityIDs(in CheckAvailabilityInput) E.Either[error, []string] {
	return F.Pipe5(
		in.IDs,
		A.Map(strings.TrimSpace),
		A.Map(strings.ToUpper),
		A.Filter(S.IsNonEmpty),
		A.StrictUniq[string],
		E.FromPredicate(validIDCount, errAvailabilityIDCount),
	)
}

// ValidateCheckAvailabilityInput normalizes, de-duplicates, and bounds the id list.
func ValidateCheckAvailabilityInput(in CheckAvailabilityInput) E.Either[error, CheckAvailabilityInput] {
	return F.Pipe1(
		E.Of[error](in),
		E.Bind(availabilityIDsLens.Set, validateAvailabilityIDs),
	)
}

func validatePublicationID(in GetPublicationInput) E.Either[error, string] {
	return F.Pipe2(
		in.ID,
		strings.TrimSpace,
		E.FromPredicate(S.IsNonEmpty, errEmptyPublicationID),
	)
}

// ValidatePublicationInput trims and validates GetPublicationInput.
func ValidatePublicationInput(in GetPublicationInput) E.Either[error, GetPublicationInput] {
	return F.Pipe1(
		E.Of[error](in),
		E.Bind(publicationIDLens.Set, validatePublicationID),
	)
}
