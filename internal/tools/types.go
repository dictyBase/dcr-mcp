// Package tools defines and registers the MCP stock center tools on an mcp.Server.
package tools

import (
	"context"

	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultAutocompleteLimit = 5
	maxAutocompleteLimit     = 50
	minAutocompleteQueryLen  = 3

	defaultSearchLimit = 15
	maxSearchLimit     = 100
	minSearchQueryLen  = 2

	defaultPhenotypeLimit = 10
	maxPhenotypeLimit     = 100

	detailSummary         = "summary"
	detailFull            = "full"
	summaryPhenotypeCap   = 10
	summaryPublicationCap = 5
)

// AutocompleteStocksInput defines the JSON Schema parameters for autocomplete_stocks.
type AutocompleteStocksInput struct {
	Query  string `json:"query" jsonschema:"Partial stock search text (at least 3 characters)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of suggestions to return (default 5, 1-50)"`
	Entity string `json:"entity,omitempty" jsonschema:"Stock entity filter: ALL, STRAIN, or PLASMID (default ALL)"`
}

// SearchStocksInput defines the JSON Schema parameters for search_stocks.
type SearchStocksInput struct {
	Query  string `json:"query" jsonschema:"Full stock search text (at least 2 characters)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of search results to return (default 15, 1-100)"`
	Entity string `json:"entity,omitempty" jsonschema:"Stock entity filter: ALL, STRAIN, or PLASMID (default ALL)"`
}

// GetStrainInput defines the JSON Schema parameters for get_strain.
type GetStrainInput struct {
	ID     string `json:"id" jsonschema:"dictyBase strain ID (for example DBS0350840)"`
	Detail string `json:"detail,omitempty" jsonschema:"Response size: summary (default; at most 10 phenotypes and 5 publications) or full"`
}

// GetPlasmidInput defines the JSON Schema parameters for get_plasmid.
type GetPlasmidInput struct {
	ID     string `json:"id" jsonschema:"dictyBase plasmid ID (for example DBP0391806)"`
	Detail string `json:"detail,omitempty" jsonschema:"Response size: summary (default; omits the DNA sequence, at most 5 publications) or full"`
}

// ListStrainsWithPhenotypeInput defines the JSON Schema parameters for list_strains_with_phenotype.
type ListStrainsWithPhenotypeInput struct {
	Phenotype string `json:"phenotype" jsonschema:"Phenotype annotation name (for example increased growth rate)"`
	Cursor    int    `json:"cursor,omitempty" jsonschema:"Pagination cursor (default 0)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of strains to return (default 10, 1-100)"`
}

// toolCallOutcome holds the 3-tuple returned to mcp.ToolHandlerFor.
type toolCallOutcome[Out any] struct {
	Result *mcp.CallToolResult
	Output Out
	Err    error
}

type autocompleteState struct {
	ctx    context.Context
	client *graphql.Client
	input  AutocompleteStocksInput
}

type searchState struct {
	ctx    context.Context
	client *graphql.Client
	input  SearchStocksInput
}

type strainState struct {
	ctx    context.Context
	client *graphql.Client
	input  GetStrainInput
}

type plasmidState struct {
	ctx    context.Context
	client *graphql.Client
	input  GetPlasmidInput
}

type phenotypeListState struct {
	ctx    context.Context
	client *graphql.Client
	input  ListStrainsWithPhenotypeInput
}

var (
	autoQueryLens = L.MakeLens(
		func(in AutocompleteStocksInput) string { return in.Query },
		func(in AutocompleteStocksInput, q string) AutocompleteStocksInput {
			in.Query = q
			return in
		},
	)
	autoLimitLens = L.MakeLens(
		func(in AutocompleteStocksInput) int { return in.Limit },
		func(in AutocompleteStocksInput, l int) AutocompleteStocksInput {
			in.Limit = l
			return in
		},
	)
	autoEntityLens = L.MakeLens(
		func(in AutocompleteStocksInput) string { return in.Entity },
		func(in AutocompleteStocksInput, e string) AutocompleteStocksInput {
			in.Entity = e
			return in
		},
	)
	autoStateInputLens = L.MakeLens(
		func(s autocompleteState) AutocompleteStocksInput { return s.input },
		func(s autocompleteState, in AutocompleteStocksInput) autocompleteState {
			s.input = in
			return s
		},
	)

	searchQueryLens = L.MakeLens(
		func(in SearchStocksInput) string { return in.Query },
		func(in SearchStocksInput, q string) SearchStocksInput {
			in.Query = q
			return in
		},
	)
	searchLimitLens = L.MakeLens(
		func(in SearchStocksInput) int { return in.Limit },
		func(in SearchStocksInput, l int) SearchStocksInput {
			in.Limit = l
			return in
		},
	)
	searchEntityLens = L.MakeLens(
		func(in SearchStocksInput) string { return in.Entity },
		func(in SearchStocksInput, e string) SearchStocksInput {
			in.Entity = e
			return in
		},
	)
	searchStateInputLens = L.MakeLens(
		func(s searchState) SearchStocksInput { return s.input },
		func(s searchState, in SearchStocksInput) searchState {
			s.input = in
			return s
		},
	)

	strainIDLens = L.MakeLens(
		func(in GetStrainInput) string { return in.ID },
		func(in GetStrainInput, id string) GetStrainInput {
			in.ID = id
			return in
		},
	)
	strainDetailLens = L.MakeLens(
		func(in GetStrainInput) string { return in.Detail },
		func(in GetStrainInput, d string) GetStrainInput {
			in.Detail = d
			return in
		},
	)
	strainStateInputLens = L.MakeLens(
		func(s strainState) GetStrainInput { return s.input },
		func(s strainState, in GetStrainInput) strainState {
			s.input = in
			return s
		},
	)

	plasmidIDLens = L.MakeLens(
		func(in GetPlasmidInput) string { return in.ID },
		func(in GetPlasmidInput, id string) GetPlasmidInput {
			in.ID = id
			return in
		},
	)
	plasmidDetailLens = L.MakeLens(
		func(in GetPlasmidInput) string { return in.Detail },
		func(in GetPlasmidInput, d string) GetPlasmidInput {
			in.Detail = d
			return in
		},
	)
	plasmidStateInputLens = L.MakeLens(
		func(s plasmidState) GetPlasmidInput { return s.input },
		func(s plasmidState, in GetPlasmidInput) plasmidState {
			s.input = in
			return s
		},
	)

	phenoNameLens = L.MakeLens(
		func(in ListStrainsWithPhenotypeInput) string { return in.Phenotype },
		func(in ListStrainsWithPhenotypeInput, p string) ListStrainsWithPhenotypeInput {
			in.Phenotype = p
			return in
		},
	)
	phenoCursorLens = L.MakeLens(
		func(in ListStrainsWithPhenotypeInput) int { return in.Cursor },
		func(in ListStrainsWithPhenotypeInput, c int) ListStrainsWithPhenotypeInput {
			in.Cursor = c
			return in
		},
	)
	phenoLimitLens = L.MakeLens(
		func(in ListStrainsWithPhenotypeInput) int { return in.Limit },
		func(in ListStrainsWithPhenotypeInput, l int) ListStrainsWithPhenotypeInput {
			in.Limit = l
			return in
		},
	)
	phenoStateInputLens = L.MakeLens(
		func(s phenotypeListState) ListStrainsWithPhenotypeInput { return s.input },
		func(s phenotypeListState, in ListStrainsWithPhenotypeInput) phenotypeListState {
			s.input = in
			return s
		},
	)
)
