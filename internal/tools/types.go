// Package tools defines and registers the MCP stock center tools on an mcp.Server.
package tools

import (
	"context"

	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
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

	// SearchPhenotypesInput defines the JSON Schema parameters for search_phenotypes.
	SearchPhenotypesInput struct {
		Query string `json:"query" jsonschema:"Words from a phenotype statement (at least 2 characters)"`
		Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of strain rows to return (default 15, 1-50)"`
	}

	// ListStrainsWithPhenotypeInput defines the parameters for list_strains_with_phenotype.
	ListStrainsWithPhenotypeInput struct {
		PhenotypeTerm string `json:"phenotype_term" jsonschema:"Exact, case-sensitive phenotype term label from get_strain"`
		Cursor        int    `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page (0 for the first page)"`
		Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of strains per page (default 10, 1-100)"`
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

	fetchState struct {
		links  linkout.Links
		ctx    context.Context
		client *graphql.Client
		input  FetchInput
	}

	geneStocksState struct {
		links    linkout.Links
		ctx      context.Context
		client   *graphql.Client
		input    FindStocksByGeneInput
		gene     graphql.GeneInfo
		strains  []graphql.GeneStrainItem
		plasmids graphql.SearchResult
	}
)

const (
	defaultAutocompleteLimit = 5
	maxAutocompleteLimit     = 50
	minAutocompleteQueryLen  = 3

	defaultSearchLimit      = 15
	maxSearchLimit          = 100
	minSearchQueryLen       = 2
	maxPhenotypeSearchLimit = 50

	defaultPhenotypeLimit = 10
	maxPhenotypeLimit     = 100

	detailSummary         = "summary"
	detailFull            = "full"
	summaryPhenotypeCap   = 10
	summaryPublicationCap = 5

	minGeneQueryLen  = 2
	genePlasmidLimit = 25

	maxAvailabilityIDs = 25
)

var (
	phenoTermParamsLens = L.MakeLens(
		func(p graphql.PhenotypeListParams) string { return p.PhenotypeTerm },
		func(p graphql.PhenotypeListParams, term string) graphql.PhenotypeListParams {
			p.PhenotypeTerm = term
			return p
		},
	)
	idParamsCtxLens = L.MakeLens(
		func(p graphql.IDParams) context.Context { return p.Ctx },
		func(p graphql.IDParams, ctx context.Context) graphql.IDParams {
			p.Ctx = ctx
			return p
		},
	)
	idParamsIDLens = L.MakeLens(
		func(p graphql.IDParams) string { return p.ID },
		func(p graphql.IDParams, id string) graphql.IDParams {
			p.ID = id
			return p
		},
	)
	sequenceInputIDLens = L.MakeLens(
		func(in GetPlasmidSequenceInput) string { return in.ID },
		func(in GetPlasmidSequenceInput, id string) GetPlasmidSequenceInput {
			in.ID = id
			return in
		},
	)
	chatgptQueryLens = L.MakeLens(
		func(in ChatGPTSearchInput) string { return in.Query },
		func(in ChatGPTSearchInput, query string) ChatGPTSearchInput {
			in.Query = query
			return in
		},
	)
	strainStateCtxLens = L.MakeLens(
		func(s strainState) context.Context { return s.ctx },
		func(s strainState, ctx context.Context) strainState {
			s.ctx = ctx
			return s
		},
	)
	strainStateClientLens = L.MakeLens(
		func(s strainState) *graphql.Client { return s.client },
		func(s strainState, client *graphql.Client) strainState {
			s.client = client
			return s
		},
	)
	strainStateInputLens = L.MakeLens(
		func(s strainState) GetStrainInput { return s.input },
		func(s strainState, in GetStrainInput) strainState {
			s.input = in
			return s
		},
	)
	plasmidStateCtxLens = L.MakeLens(
		func(s plasmidState) context.Context { return s.ctx },
		func(s plasmidState, ctx context.Context) plasmidState {
			s.ctx = ctx
			return s
		},
	)
	plasmidStateClientLens = L.MakeLens(
		func(s plasmidState) *graphql.Client { return s.client },
		func(s plasmidState, client *graphql.Client) plasmidState {
			s.client = client
			return s
		},
	)
	plasmidStateInputLens = L.MakeLens(
		func(s plasmidState) GetPlasmidInput { return s.input },
		func(s plasmidState, in GetPlasmidInput) plasmidState {
			s.input = in
			return s
		},
	)
	fetchStateCtxLens = L.MakeLens(
		func(s fetchState) context.Context { return s.ctx },
		func(s fetchState, ctx context.Context) fetchState {
			s.ctx = ctx
			return s
		},
	)
	fetchStateClientLens = L.MakeLens(
		func(s fetchState) *graphql.Client { return s.client },
		func(s fetchState, client *graphql.Client) fetchState {
			s.client = client
			return s
		},
	)
	fetchStateInputLens = L.MakeLens(
		func(s fetchState) FetchInput { return s.input },
		func(s fetchState, in FetchInput) fetchState {
			s.input = in
			return s
		},
	)
	fetchStateLinksLens = L.MakeLens(
		func(s fetchState) linkout.Links { return s.links },
		func(s fetchState, links linkout.Links) fetchState {
			s.links = links
			return s
		},
	)
	geneStocksCtxLens = L.MakeLens(
		func(s geneStocksState) context.Context { return s.ctx },
		func(s geneStocksState, ctx context.Context) geneStocksState {
			s.ctx = ctx
			return s
		},
	)
	geneStocksClientLens = L.MakeLens(
		func(s geneStocksState) *graphql.Client { return s.client },
		func(s geneStocksState, client *graphql.Client) geneStocksState {
			s.client = client
			return s
		},
	)
	geneStocksInputLens = L.MakeLens(
		func(s geneStocksState) FindStocksByGeneInput { return s.input },
		func(s geneStocksState, in FindStocksByGeneInput) geneStocksState {
			s.input = in
			return s
		},
	)
	geneStocksLinksLens = L.MakeLens(
		func(s geneStocksState) linkout.Links { return s.links },
		func(s geneStocksState, links linkout.Links) geneStocksState {
			s.links = links
			return s
		},
	)

	strainStateInputIDLens = F.Pipe1(
		strainStateInputLens,
		L.Compose[strainState](strainIDLens),
	)
	plasmidStateInputIDLens = F.Pipe1(
		plasmidStateInputLens,
		L.Compose[plasmidState](plasmidIDLens),
	)
	fetchStateInputIDLens = F.Pipe1(
		fetchStateInputLens,
		L.Compose[fetchState](fetchIDLens),
	)
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

	phenoSearchQueryLens = L.MakeLens(
		func(in SearchPhenotypesInput) string { return in.Query },
		func(in SearchPhenotypesInput, q string) SearchPhenotypesInput {
			in.Query = q
			return in
		},
	)
	phenoSearchLimitLens = L.MakeLens(
		func(in SearchPhenotypesInput) int { return in.Limit },
		func(in SearchPhenotypesInput, l int) SearchPhenotypesInput {
			in.Limit = l
			return in
		},
	)

	phenoTermLens = L.MakeLens(
		func(in ListStrainsWithPhenotypeInput) string { return in.PhenotypeTerm },
		func(in ListStrainsWithPhenotypeInput, p string) ListStrainsWithPhenotypeInput {
			in.PhenotypeTerm = p
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
