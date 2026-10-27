// Package tools defines and registers the MCP stock center tools on an mcp.Server.
package tools

import (
	"context"

	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	// AutocompleteStocksInput defines the JSON Schema parameters for autocomplete_stocks.
	AutocompleteStocksInput struct {
		Query  string `json:"query" jsonschema:"Partial stock search text (at least 3 characters)"`
		Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of suggestions to return (default 5, 1-50)"`
		Entity string `json:"entity,omitempty" jsonschema:"Stock entity filter: ALL, STRAIN, or PLASMID (default ALL)"`
	}

	// SearchStocksInput defines the JSON Schema parameters for search_stocks.
	SearchStocksInput struct {
		Query  string `json:"query" jsonschema:"Full stock search text (at least 2 characters)"`
		Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of search results to return (default 15, 1-100)"`
		Entity string `json:"entity,omitempty" jsonschema:"Stock entity filter: ALL, STRAIN, or PLASMID (default ALL)"`
	}

	// GetStrainInput defines the JSON Schema parameters for get_strain.
	// The summary detail caps the response at 10 phenotypes and 5 publications.
	GetStrainInput struct {
		ID     string `json:"id" jsonschema:"dictyBase strain ID (for example DBS0350840)"`
		Detail string `json:"detail,omitempty" jsonschema:"Response size: summary (default) or full"`
	}

	// GetPlasmidInput defines the JSON Schema parameters for get_plasmid.
	// The summary detail omits the DNA sequence and caps publications at 5.
	GetPlasmidInput struct {
		ID     string `json:"id" jsonschema:"dictyBase plasmid ID (for example DBP0391806)"`
		Detail string `json:"detail,omitempty" jsonschema:"Response size: summary (default) or full"`
	}

	// GetPlasmidSequenceInput defines the JSON Schema parameters for get_plasmid_sequence.
	GetPlasmidSequenceInput struct {
		ID string `json:"id" jsonschema:"dictyBase plasmid ID (for example DBP0391806)"`
	}

	// ListStrainsWithPhenotypeInput defines the parameters for list_strains_with_phenotype.
	ListStrainsWithPhenotypeInput struct {
		Phenotype string `json:"phenotype" jsonschema:"Phenotype annotation name (for example increased growth rate)"`
		Cursor    int    `json:"cursor,omitempty" jsonschema:"Pagination cursor (default 0)"`
		Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of strains to return (default 10, 1-100)"`
	}

	// ChatGPTSearchInput defines the JSON Schema parameters for the ChatGPT-compatible search tool.
	ChatGPTSearchInput struct {
		Query string `json:"query" jsonschema:"Search text, for example a gene name, strain label, or phenotype"`
	}

	// FetchInput defines the JSON Schema parameters for the ChatGPT-compatible fetch tool.
	FetchInput struct {
		ID string `json:"id" jsonschema:"Stock id returned by search: DBS... (strain) or DBP... (plasmid)"`
	}

	// FindStocksByGeneInput defines the JSON Schema parameters for find_stocks_by_gene.
	FindStocksByGeneInput struct {
		Gene string `json:"gene" jsonschema:"Gene name (for example pkaC) or dictyBase gene id (for example DDB_G0283907)"`
	}

	// SuggestPhenotypeTermsInput defines the JSON Schema parameters for suggest_phenotype_terms.
	SuggestPhenotypeTermsInput struct {
		Query string `json:"query" jsonschema:"Phenotype keywords (at least 3 characters), for example growth rate"`
		Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of terms (default 10, 1-25)"`
	}

	// CheckAvailabilityInput defines the JSON Schema parameters for check_availability.
	CheckAvailabilityInput struct {
		IDs []string `json:"ids" jsonschema:"1 to 25 strain (DBS...) and/or plasmid (DBP...) ids"`
	}

	// GetPublicationInput defines the JSON Schema parameters for get_publication.
	GetPublicationInput struct {
		ID string `json:"id" jsonschema:"PubMed id of a publication listed on a strain or plasmid (for example 17177856)"`
	}

	// toolCallOutcome holds the 3-tuple returned to mcp.ToolHandlerFor.
	toolCallOutcome[Out any] struct {
		Result *mcp.CallToolResult
		Output Out
		Err    error
	}

	autocompleteState struct {
		ctx    context.Context
		client *graphql.Client
		input  AutocompleteStocksInput
	}

	searchState struct {
		ctx    context.Context
		client *graphql.Client
		input  SearchStocksInput
	}

	strainState struct {
		ctx    context.Context
		client *graphql.Client
		input  GetStrainInput
	}

	plasmidState struct {
		ctx    context.Context
		client *graphql.Client
		input  GetPlasmidInput
	}

	phenotypeListState struct {
		ctx    context.Context
		client *graphql.Client
		input  ListStrainsWithPhenotypeInput
	}

	fetchState struct {
		ctx    context.Context
		client *graphql.Client
		input  FetchInput
	}

	geneStocksState struct {
		ctx      context.Context
		client   *graphql.Client
		input    FindStocksByGeneInput
		gene     graphql.GeneInfo
		strains  []graphql.GeneStrainItem
		plasmids graphql.SearchResult
	}

	phenoTermsState struct {
		ctx    context.Context
		client *graphql.Client
		input  SuggestPhenotypeTermsInput
		result graphql.SearchResult
	}

	availabilityState struct {
		ctx    context.Context
		client *graphql.Client
		input  CheckAvailabilityInput
	}

	publicationState struct {
		ctx    context.Context
		client *graphql.Client
		input  GetPublicationInput
	}
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

	minGeneQueryLen  = 2
	genePlasmidLimit = 25

	phenotypeSampleLimit  = 100
	defaultPhenoTermLimit = 10
	maxPhenoTermLimit     = 25
	minPhenoTermQueryLen  = 3

	maxAvailabilityIDs = 25
)

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

	fetchIDLens = L.MakeLens(
		func(in FetchInput) string { return in.ID },
		func(in FetchInput, id string) FetchInput {
			in.ID = id
			return in
		},
	)

	geneNameLens = L.MakeLens(
		func(in FindStocksByGeneInput) string { return in.Gene },
		func(in FindStocksByGeneInput, g string) FindStocksByGeneInput {
			in.Gene = g
			return in
		},
	)
	geneStocksGeneLens = L.MakeLens(
		func(s geneStocksState) graphql.GeneInfo { return s.gene },
		func(s geneStocksState, g graphql.GeneInfo) geneStocksState {
			s.gene = g
			return s
		},
	)
	geneStocksStrainsLens = L.MakeLens(
		func(s geneStocksState) []graphql.GeneStrainItem { return s.strains },
		func(s geneStocksState, items []graphql.GeneStrainItem) geneStocksState {
			s.strains = items
			return s
		},
	)
	geneStocksPlasmidsLens = L.MakeLens(
		func(s geneStocksState) graphql.SearchResult { return s.plasmids },
		func(s geneStocksState, r graphql.SearchResult) geneStocksState {
			s.plasmids = r
			return s
		},
	)

	phenoTermQueryLens = L.MakeLens(
		func(in SuggestPhenotypeTermsInput) string { return in.Query },
		func(in SuggestPhenotypeTermsInput, q string) SuggestPhenotypeTermsInput {
			in.Query = q
			return in
		},
	)
	phenoTermLimitLens = L.MakeLens(
		func(in SuggestPhenotypeTermsInput) int { return in.Limit },
		func(in SuggestPhenotypeTermsInput, l int) SuggestPhenotypeTermsInput {
			in.Limit = l
			return in
		},
	)
	phenoTermsResultLens = L.MakeLens(
		func(s phenoTermsState) graphql.SearchResult { return s.result },
		func(s phenoTermsState, r graphql.SearchResult) phenoTermsState {
			s.result = r
			return s
		},
	)

	availabilityIDsLens = L.MakeLens(
		func(in CheckAvailabilityInput) []string { return in.IDs },
		func(in CheckAvailabilityInput, ids []string) CheckAvailabilityInput {
			in.IDs = ids
			return in
		},
	)

	publicationIDLens = L.MakeLens(
		func(in GetPublicationInput) string { return in.ID },
		func(in GetPublicationInput, id string) GetPublicationInput {
			in.ID = id
			return in
		},
	)
)
