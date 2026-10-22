// Package graphql provides functional IOEither wrappers around dictyBase Stock Center GraphQL queries.
package graphql

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	hasuragql "github.com/hasura/go-graphql-client"
)

// StockEntityType selects the kind of stock covered by autocomplete or search.
type StockEntityType string

// GetGraphQLType returns the GraphQL enum name for hasura/go-graphql-client.
func (StockEntityType) GetGraphQLType() string { return "StockEntityType" }

const (
	// StockEntityAll covers both strains and plasmids.
	StockEntityAll StockEntityType = "ALL"
	// StockEntityStrain restricts to strains.
	StockEntityStrain StockEntityType = "STRAIN"
	// StockEntityPlasmid restricts to plasmids.
	StockEntityPlasmid StockEntityType = "PLASMID"
)

// AutocompleteParams holds validated inputs for ListStockSuggestions.
type AutocompleteParams struct {
	Ctx    context.Context
	Query  string
	Limit  int
	Entity StockEntityType
}

// SearchParams holds validated inputs for SearchStocks.
type SearchParams struct {
	Ctx    context.Context
	Query  string
	Limit  int
	Entity StockEntityType
}

// IDParams holds a validated stock ID for GetStrain or GetPlasmid.
type IDParams struct {
	Ctx context.Context
	ID  string
}

// PhenotypeListParams holds validated inputs for ListStrainsWithPhenotype.
type PhenotypeListParams struct {
	Ctx       context.Context
	Phenotype string
	Cursor    int
	Limit     int
}

// GeneParams holds a validated gene name or DDB_G identifier.
type GeneParams struct {
	Ctx  context.Context
	Gene string
}

// GeneInfo holds general information for one gene.
type GeneInfo struct {
	ID              string   `graphql:"id" json:"id"`
	NameDescription []string `graphql:"name_description" json:"name_description"`
	GeneProduct     string   `graphql:"gene_product" json:"gene_product"`
	Synonyms        []string `graphql:"synonyms" json:"synonyms"`
	Description     string   `graphql:"description" json:"description"`
}

// GeneStrainItem is one strain returned by listStrainsWithGene.
type GeneStrainItem struct {
	ID      string `graphql:"id" json:"id"`
	Label   string `graphql:"label" json:"label"`
	Summary string `graphql:"summary" json:"summary"`
	InStock bool   `graphql:"in_stock" json:"in_stock"`
}

// FullAuthor holds a publication author.
type FullAuthor struct {
	LastName  string `graphql:"last_name" json:"last_name"`
	FirstName string `graphql:"first_name" json:"first_name"`
	Initials  string `graphql:"initials" json:"initials"`
	Rank      string `graphql:"rank" json:"rank"`
}

// PublicationDetail holds one publication record.
type PublicationDetail struct {
	ID       string       `graphql:"id" json:"id"`
	DOI      string       `graphql:"doi" json:"doi,omitempty"`
	Title    string       `graphql:"title" json:"title"`
	Abstract string       `graphql:"abstract" json:"abstract"`
	Journal  string       `graphql:"journal" json:"journal"`
	PubDate  string       `graphql:"pub_date" json:"pub_date"`
	Volume   string       `graphql:"volume" json:"volume"`
	Pages    string       `graphql:"pages" json:"pages"`
	PubType  string       `graphql:"pub_type" json:"pub_type"`
	Authors  []FullAuthor `graphql:"authors" json:"authors"`
}

// StockSuggestion represents one type-ahead autocomplete suggestion.
type StockSuggestion struct {
	ID          string          `graphql:"id" json:"id"`
	Entity      StockEntityType `graphql:"entity" json:"entity"`
	Field       string          `graphql:"field" json:"field"`
	DisplayText string          `graphql:"display_text" json:"display_text"`
	Score       float64         `graphql:"score" json:"score"`
}

// AutocompleteResult holds the response from listStockSuggestions.
type AutocompleteResult struct {
	Suggestions []StockSuggestion `json:"suggestions"`
	Limit       int               `json:"limit"`
	TotalCount  int               `json:"total_count"`
}

// StockSearchHit represents one full-search match.
type StockSearchHit struct {
	ID          string          `graphql:"id" json:"id"`
	Entity      StockEntityType `graphql:"entity" json:"entity"`
	Field       string          `graphql:"field" json:"field"`
	DisplayText string          `graphql:"display_text" json:"display_text"`
	Score       float64         `graphql:"score" json:"score"`
	Summary     string          `graphql:"summary" json:"summary"`
	StrainLabel string          `graphql:"strain_label" json:"strain_label"`
	InStock     bool            `graphql:"in_stock" json:"in_stock"`
}

// SearchResult holds the response from searchStocks.
type SearchResult struct {
	Results    []StockSearchHit `json:"results"`
	Limit      int              `json:"limit"`
	TotalCount int              `json:"total_count"`
}

// ParentStrain holds the parent strain reference of a strain.
type ParentStrain struct {
	ID    string `graphql:"id" json:"id"`
	Label string `graphql:"label" json:"label"`
}

// Depositor holds the first and last name of a stock depositor.
type Depositor struct {
	FirstName string `graphql:"first_name" json:"first_name"`
	LastName  string `graphql:"last_name" json:"last_name"`
}

// GeneRef holds a gene name associated with a stock.
type GeneRef struct {
	Name string `graphql:"name" json:"name"`
}

// AuthorRef holds a publication author's last name.
type AuthorRef struct {
	LastName string `graphql:"last_name" json:"last_name"`
}

// PublicationRef holds publication metadata associated with a stock.
type PublicationRef struct {
	ID      string      `graphql:"id" json:"id"`
	PubDate string      `graphql:"pub_date" json:"pub_date"`
	Title   string      `graphql:"title" json:"title"`
	Journal string      `graphql:"journal" json:"journal"`
	Volume  string      `graphql:"volume" json:"volume"`
	Pages   string      `graphql:"pages" json:"pages"`
	DOI     string      `graphql:"doi" json:"doi,omitempty"`
	Authors []AuthorRef `graphql:"authors" json:"authors"`
}

// PhenotypePublicationRef holds publication metadata inside a phenotype annotation.
type PhenotypePublicationRef struct {
	ID      string      `graphql:"id" json:"id"`
	PubDate string      `graphql:"pub_date" json:"pub_date"`
	Title   string      `graphql:"title" json:"title"`
	Journal string      `graphql:"journal" json:"journal"`
	Volume  string      `graphql:"volume" json:"volume"`
	Pages   string      `graphql:"pages" json:"pages"`
	DOI     string      `graphql:"doi" json:"doi,omitempty"`
	Authors []AuthorRef `graphql:"authors" json:"authors"`
}

// PhenotypeItem represents one phenotype annotation on a strain.
type PhenotypeItem struct {
	Phenotype   string                  `graphql:"phenotype" json:"phenotype"`
	Note        string                  `graphql:"note" json:"note"`
	Assay       string                  `graphql:"assay" json:"assay"`
	Environment string                  `graphql:"environment" json:"environment"`
	Publication PhenotypePublicationRef `graphql:"publication" json:"publication"`
}

// StrainDetail holds the complete record for a single strain.
type StrainDetail struct {
	ID                  string           `graphql:"id" json:"id"`
	Label               string           `graphql:"label" json:"label"`
	Summary             string           `graphql:"summary" json:"summary"`
	Species             string           `graphql:"species" json:"species"`
	SystematicName      string           `graphql:"systematic_name" json:"systematic_name"`
	Plasmid             string           `graphql:"plasmid" json:"plasmid"`
	Dbxrefs             []string         `graphql:"dbxrefs" json:"dbxrefs"`
	InStock             bool             `graphql:"in_stock" json:"in_stock"`
	Genotypes           []string         `graphql:"genotypes" json:"genotypes"`
	MutagenesisMethod   string           `graphql:"mutagenesis_method" json:"mutagenesis_method"`
	GeneticModification string           `graphql:"genetic_modification" json:"genetic_modification"`
	Names               []string         `graphql:"names" json:"names"`
	Characteristics     []string         `graphql:"characteristics" json:"characteristics"`
	Parent              ParentStrain     `graphql:"parent" json:"parent"`
	Depositor           Depositor        `graphql:"depositor" json:"depositor"`
	Genes               []GeneRef        `graphql:"genes" json:"genes"`
	Publications        []PublicationRef `graphql:"publications" json:"publications"`
	Phenotypes          []PhenotypeItem  `graphql:"phenotypes" json:"phenotypes"`
}

// PlasmidDetail holds the complete record for a single plasmid.
type PlasmidDetail struct {
	ID               string           `graphql:"id" json:"id"`
	Name             string           `graphql:"name" json:"name"`
	Summary          string           `graphql:"summary" json:"summary"`
	ImageMap         string           `graphql:"image_map" json:"image_map"`
	Sequence         string           `graphql:"sequence" json:"sequence"`
	Keywords         []string         `graphql:"keywords" json:"keywords"`
	GenbankAccession string           `graphql:"genbank_accession" json:"genbank_accession"`
	InStock          bool             `graphql:"in_stock" json:"in_stock"`
	Dbxrefs          []string         `graphql:"dbxrefs" json:"dbxrefs"`
	Depositor        Depositor        `graphql:"depositor" json:"depositor"`
	Genes            []GeneRef        `graphql:"genes" json:"genes"`
	Publications     []PublicationRef `graphql:"publications" json:"publications"`
}

// PhenotypeStrainItem holds a strain returned by listStrainsWithAnnotation.
type PhenotypeStrainItem struct {
	ID           string                    `graphql:"id" json:"id"`
	Label        string                    `graphql:"label" json:"label"`
	Summary      string                    `graphql:"summary" json:"summary"`
	InStock      bool                      `graphql:"in_stock" json:"in_stock"`
	Genes        []GeneRef                 `graphql:"genes" json:"genes"`
	Publications []PhenotypePublicationRef `graphql:"publications" json:"publications"`
}

// PhenotypeStrainList holds the paginated response from listStrainsWithAnnotation.
type PhenotypeStrainList struct {
	Strains    []PhenotypeStrainItem `json:"strains"`
	NextCursor int64                 `json:"next_cursor"`
	TotalCount int                   `json:"total_count"`
}

// listStockSuggestionsQuery is the GraphQL query structure for listStockSuggestions.
type listStockSuggestionsQuery struct {
	ListStockSuggestions struct {
		Suggestions []StockSuggestion `graphql:"suggestions"`
		Limit       int               `graphql:"limit"`
		TotalCount  int               `graphql:"totalCount"`
	} `graphql:"listStockSuggestions(query: $query, limit: $limit, entity: $entity)"`
}

// searchStocksQuery is the GraphQL query structure for searchStocks.
type searchStocksQuery struct {
	SearchStocks struct {
		Results    []StockSearchHit `graphql:"results"`
		Limit      int              `graphql:"limit"`
		TotalCount int              `graphql:"totalCount"`
	} `graphql:"searchStocks(query: $query, limit: $limit, entity: $entity)"`
}

// getStrainQuery is the GraphQL query structure for strain(id: $id).
type getStrainQuery struct {
	Strain StrainDetail `graphql:"strain(id: $id)"`
}

// getPlasmidQuery is the GraphQL query structure for plasmid(id: $id).
type getPlasmidQuery struct {
	Plasmid PlasmidDetail `graphql:"plasmid(id: $id)"`
}

// listStrainsWithPhenotypeQuery is the GraphQL query structure for listStrainsWithAnnotation.
type listStrainsWithPhenotypeQuery struct {
	ListStrainsWithAnnotation struct {
		TotalCount int                   `graphql:"totalCount"`
		NextCursor int64                 `graphql:"nextCursor"`
		Strains    []PhenotypeStrainItem `graphql:"strains"`
	} `graphql:"listStrainsWithAnnotation(cursor: $cursor, limit: $limit, type: $type, annotation: $annotation)"`
}

type geneInfoQuery struct {
	GeneGeneralInformation GeneInfo `graphql:"geneGeneralInformation(gene: $gene)"`
}

type strainsWithGeneQuery struct {
	ListStrainsWithGene []GeneStrainItem `graphql:"listStrainsWithGene(gene: $gene)"`
}

type publicationQuery struct {
	Publication PublicationDetail `graphql:"publication(id: $id)"`
}

func emptySlice[T any]() []T {
	return make([]T, 0)
}

func nonNilSlice[T any](items []T) []T {
	hasElem := predarrays.IsNonEmpty[T]()
	return F.Pipe2(
		items,
		O.FromPredicate(hasElem),
		O.GetOrElse(emptySlice[T]),
	)
}

func toAutocompleteResult(q *listStockSuggestionsQuery) AutocompleteResult {
	return AutocompleteResult{
		Suggestions: nonNilSlice(q.ListStockSuggestions.Suggestions),
		Limit:       q.ListStockSuggestions.Limit,
		TotalCount:  q.ListStockSuggestions.TotalCount,
	}
}

func toSearchResult(q *searchStocksQuery) SearchResult {
	return SearchResult{
		Results:    nonNilSlice(q.SearchStocks.Results),
		Limit:      q.SearchStocks.Limit,
		TotalCount: q.SearchStocks.TotalCount,
	}
}

func normalizePublicationRef(pub PublicationRef) PublicationRef {
	pub.Authors = nonNilSlice(pub.Authors)
	return pub
}

func normalizePhenotypePubRef(pub PhenotypePublicationRef) PhenotypePublicationRef {
	pub.Authors = nonNilSlice(pub.Authors)
	return pub
}

func normalizePhenotypeItem(item PhenotypeItem) PhenotypeItem {
	item.Publication = normalizePhenotypePubRef(item.Publication)
	return item
}

func toStrainDetail(q *getStrainQuery) StrainDetail {
	s := q.Strain
	s.Dbxrefs = nonNilSlice(s.Dbxrefs)
	s.Genotypes = nonNilSlice(s.Genotypes)
	s.Names = nonNilSlice(s.Names)
	s.Characteristics = nonNilSlice(s.Characteristics)
	s.Genes = nonNilSlice(s.Genes)
	s.Publications = F.Pipe2(
		s.Publications,
		A.Map(normalizePublicationRef),
		nonNilSlice[PublicationRef],
	)
	s.Phenotypes = F.Pipe2(
		s.Phenotypes,
		A.Map(normalizePhenotypeItem),
		nonNilSlice[PhenotypeItem],
	)
	return s
}

func toPlasmidDetail(q *getPlasmidQuery) PlasmidDetail {
	p := q.Plasmid
	p.Keywords = nonNilSlice(p.Keywords)
	p.Dbxrefs = nonNilSlice(p.Dbxrefs)
	p.Genes = nonNilSlice(p.Genes)
	p.Publications = F.Pipe2(
		p.Publications,
		A.Map(normalizePublicationRef),
		nonNilSlice[PublicationRef],
	)
	return p
}

func normalizePhenotypeStrainItem(item PhenotypeStrainItem) PhenotypeStrainItem {
	item.Genes = nonNilSlice(item.Genes)
	item.Publications = F.Pipe2(
		item.Publications,
		A.Map(normalizePhenotypePubRef),
		nonNilSlice[PhenotypePublicationRef],
	)
	return item
}

func toPhenotypeStrainList(q *listStrainsWithPhenotypeQuery) PhenotypeStrainList {
	return PhenotypeStrainList{
		Strains: F.Pipe2(
			q.ListStrainsWithAnnotation.Strains,
			A.Map(normalizePhenotypeStrainItem),
			nonNilSlice[PhenotypeStrainItem],
		),
		NextCursor: q.ListStrainsWithAnnotation.NextCursor,
		TotalCount: q.ListStrainsWithAnnotation.TotalCount,
	}
}

func toGeneInfo(q *geneInfoQuery) GeneInfo {
	g := q.GeneGeneralInformation
	g.NameDescription = nonNilSlice(g.NameDescription)
	g.Synonyms = nonNilSlice(g.Synonyms)
	return g
}

var toGeneStrainItems = F.Flow2(
	func(q *strainsWithGeneQuery) []GeneStrainItem { return q.ListStrainsWithGene },
	nonNilSlice[GeneStrainItem],
)

func toPublicationDetail(q *publicationQuery) PublicationDetail {
	p := q.Publication
	p.Authors = nonNilSlice(p.Authors)
	return p
}

// Ensure hasuragql is referenced when needed by schema types.
var _ = hasuragql.ID("")
