package tools

import (
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	I "github.com/IBM/fp-go/v2/identity"
	M "github.com/IBM/fp-go/v2/monoid"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
)

type (
	// PublicationRefView is one publication on a stock plus its local page url.
	PublicationRefView struct {
		graphql.PublicationRef
		URL string `json:"url,omitempty"`
	}

	// PhenotypePublicationRefView is the publication inside a phenotype plus its url.
	PhenotypePublicationRefView struct {
		graphql.PhenotypePublicationRef
		URL string `json:"url,omitempty"`
	}

	// PhenotypeItemView is one phenotype whose publication carries a local url.
	PhenotypeItemView struct {
		graphql.PhenotypeItem
		Publication PhenotypePublicationRefView `json:"publication"`
	}
	// strainViewFunc and plasmidViewFunc take the links first and the record second.
	strainViewFunc = func(linkout.Links) func(graphql.StrainDetail) StrainView

	plasmidViewFunc = func(linkout.Links) func(graphql.PlasmidDetail) PlasmidView
	// StrainView is the get_strain output: the GraphQL record plus presentation fields.
	StrainView struct {
		Publications []PublicationRefView `json:"publications"`
		Phenotypes   []PhenotypeItemView  `json:"phenotypes"`
		graphql.StrainDetail
		URL              string `json:"url"`
		Detail           string `json:"detail"`
		PhenotypeCount   int    `json:"phenotype_count"`
		PublicationCount int    `json:"publication_count"`
	}

	// PlasmidView is the get_plasmid output: the GraphQL record plus presentation fields.
	PlasmidView struct {
		Publications []PublicationRefView `json:"publications"`
		graphql.PlasmidDetail
		URL              string `json:"url"`
		Detail           string `json:"detail"`
		SequenceLength   int    `json:"sequence_length"`
		PublicationCount int    `json:"publication_count"`
	}

	// PlasmidSequence is the get_plasmid_sequence output.
	PlasmidSequence struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Length int    `json:"length"`
		Format string `json:"format"`
		FASTA  string `json:"fasta"`
		URL    string `json:"url"`
	}

	// SearchHitView is one search_stocks hit plus its stock page url.
	SearchHitView struct {
		graphql.StockSearchHit
		URL string `json:"url"`
	}

	// SearchResultView is the search_stocks output.
	SearchResultView struct {
		Results    []SearchHitView `json:"results"`
		Limit      int             `json:"limit"`
		TotalCount int             `json:"total_count"`
	}

	// PhenotypeStrainView is one list_strains_with_phenotype item plus its stock page url.
	PhenotypeStrainView struct {
		graphql.PhenotypeStrainItem
		URL string `json:"url"`
	}

	// MatchedTextView groups search_phenotypes rows by the text that matched.
	// The text can come from a phenotype, an environment, or an assay term, or
	// from a term definition or comment. It is not always a phenotype term.
	MatchedTextView struct {
		Field       string `json:"field"`
		DisplayText string `json:"display_text"`
		StrainCount int    `json:"strain_count"`
	}

	// PhenotypeSearchView is the search_phenotypes output. Count is the number
	// of rows. There are no further pages.
	PhenotypeSearchView struct {
		MatchedTexts []MatchedTextView `json:"matched_texts"`
		Results      []SearchHitView   `json:"results"`
		Limit        int               `json:"limit"`
		Count        int               `json:"count"`
	}

	// PhenotypeStrainListView is the list_strains_with_phenotype output.
	// PageCount is the number of unique strains on this page, not a total.
	PhenotypeStrainListView struct {
		PhenotypeTerm string                `json:"phenotype_term"`
		Strains       []PhenotypeStrainView `json:"strains"`
		NextCursor    int64                 `json:"next_cursor"`
		HasMore       bool                  `json:"has_more"`
		PageCount     int                   `json:"page_count"`
	}

	matchKey struct {
		field       string
		displayText string
	}
)

const fastaFormat = "fasta"

var (
	searchHitFieldLens = L.MakeLens(
		func(x graphql.StockSearchHit) string { return x.Field },
		func(x graphql.StockSearchHit, value string) graphql.StockSearchHit {
			x.Field = value
			return x
		},
	)
	searchHitDisplayTextLens = L.MakeLens(
		func(x graphql.StockSearchHit) string { return x.DisplayText },
		func(x graphql.StockSearchHit, value string) graphql.StockSearchHit {
			x.DisplayText = value
			return x
		},
	)
	matchKeyFieldLens = L.MakeLens(
		func(x matchKey) string { return x.field },
		func(x matchKey, value string) matchKey {
			x.field = value
			return x
		},
	)
	matchKeyDisplayTextLens = L.MakeLens(
		func(x matchKey) string { return x.displayText },
		func(x matchKey, value string) matchKey {
			x.displayText = value
			return x
		},
	)
	matchedTextViewFieldLens = L.MakeLens(
		func(x MatchedTextView) string { return x.Field },
		func(x MatchedTextView, value string) MatchedTextView {
			x.Field = value
			return x
		},
	)
	matchedTextViewDisplayTextLens = L.MakeLens(
		func(x MatchedTextView) string { return x.DisplayText },
		func(x MatchedTextView, value string) MatchedTextView {
			x.DisplayText = value
			return x
		},
	)
	matchedTextViewStrainCountLens = L.MakeLens(
		func(x MatchedTextView) int { return x.StrainCount },
		func(x MatchedTextView, value int) MatchedTextView {
			x.StrainCount = value
			return x
		},
	)
	strainViewPublicationsLens = L.MakeLens(
		func(v StrainView) []PublicationRefView { return v.Publications },
		func(v StrainView, items []PublicationRefView) StrainView {
			v.Publications = items
			return v
		},
	)
	strainViewPhenotypesLens = L.MakeLens(
		func(v StrainView) []PhenotypeItemView { return v.Phenotypes },
		func(v StrainView, items []PhenotypeItemView) StrainView {
			v.Phenotypes = items
			return v
		},
	)
	plasmidViewPublicationsLens = L.MakeLens(
		func(v PlasmidView) []PublicationRefView { return v.Publications },

		func(v PlasmidView, items []PublicationRefView) PlasmidView {
			v.Publications = items
			return v
		},
	)
	publicationRefViewRecordLens = L.MakeLens(
		func(v PublicationRefView) graphql.PublicationRef { return v.PublicationRef },
		func(v PublicationRefView, p graphql.PublicationRef) PublicationRefView {
			v.PublicationRef = p
			return v
		},
	)
	publicationRefViewURLLens = L.MakeLens(
		func(v PublicationRefView) string { return v.URL },
		func(v PublicationRefView, url string) PublicationRefView {
			v.URL = url
			return v
		},
	)
	phenotypePubRefViewRecordLens = L.MakeLens(
		func(v PhenotypePublicationRefView) graphql.PhenotypePublicationRef {
			return v.PhenotypePublicationRef
		},
		func(v PhenotypePublicationRefView, p graphql.PhenotypePublicationRef) PhenotypePublicationRefView {
			v.PhenotypePublicationRef = p
			return v
		},
	)
	phenotypePubRefViewURLLens = L.MakeLens(
		func(v PhenotypePublicationRefView) string { return v.URL },
		func(v PhenotypePublicationRefView, url string) PhenotypePublicationRefView {
			v.URL = url
			return v
		},
	)
	phenotypeItemViewRecordLens = L.MakeLens(
		func(v PhenotypeItemView) graphql.PhenotypeItem { return v.PhenotypeItem },
		func(v PhenotypeItemView, p graphql.PhenotypeItem) PhenotypeItemView {
			v.PhenotypeItem = p
			return v
		},
	)
	phenotypeItemViewPublicationLens = L.MakeLens(
		func(v PhenotypeItemView) PhenotypePublicationRefView { return v.Publication },
		func(v PhenotypeItemView, p PhenotypePublicationRefView) PhenotypeItemView {
			v.Publication = p
			return v
		},
	)
	publicationRefIDLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.ID },
		func(p graphql.PublicationRef, id string) graphql.PublicationRef {
			p.ID = id
			return p
		},
	)
	publicationRefTitleLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.Title },
		func(p graphql.PublicationRef, title string) graphql.PublicationRef {
			p.Title = title
			return p
		},
	)
	publicationRefJournalLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.Journal },
		func(p graphql.PublicationRef, journal string) graphql.PublicationRef {
			p.Journal = journal
			return p
		},
	)
	publicationRefPubDateLens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.PubDate },
		func(p graphql.PublicationRef, pubDate string) graphql.PublicationRef {
			p.PubDate = pubDate
			return p
		},
	)
	publicationRefDOILens = L.MakeLens(
		func(p graphql.PublicationRef) string { return p.DOI },
		func(p graphql.PublicationRef, doi string) graphql.PublicationRef {
			p.DOI = doi
			return p
		},
	)
	publicationRefAuthorsLens = L.MakeLens(
		func(p graphql.PublicationRef) []graphql.AuthorRef { return p.Authors },
		func(p graphql.PublicationRef, authors []graphql.AuthorRef) graphql.PublicationRef {
			p.Authors = authors
			return p
		},
	)
	phenotypePubRefIDLens = L.MakeLens(
		func(p graphql.PhenotypePublicationRef) string { return p.ID },
		func(p graphql.PhenotypePublicationRef, id string) graphql.PhenotypePublicationRef {
			p.ID = id
			return p
		},
	)
	phenotypeItemPhenotypeLens = L.MakeLens(
		func(p graphql.PhenotypeItem) string { return p.Phenotype },
		func(p graphql.PhenotypeItem, phenotype string) graphql.PhenotypeItem {
			p.Phenotype = phenotype
			return p
		},
	)
	phenotypeItemEnvironmentLens = L.MakeLens(
		func(p graphql.PhenotypeItem) string { return p.Environment },
		func(p graphql.PhenotypeItem, environment string) graphql.PhenotypeItem {
			p.Environment = environment
			return p
		},
	)
	phenotypeItemAssayLens = L.MakeLens(
		func(p graphql.PhenotypeItem) string { return p.Assay },
		func(p graphql.PhenotypeItem, assay string) graphql.PhenotypeItem {
			p.Assay = assay
			return p
		},
	)
	phenotypeItemPaperLens = L.MakeLens(
		func(p graphql.PhenotypeItem) graphql.PhenotypePublicationRef { return p.Publication },
		func(p graphql.PhenotypeItem, paper graphql.PhenotypePublicationRef) graphql.PhenotypeItem {
			p.Publication = paper
			return p
		},
	)

	publicationRefViewIDLens = F.Pipe1(
		publicationRefViewRecordLens,
		L.Compose[PublicationRefView](publicationRefIDLens),
	)
	publicationRefViewTitleLens = F.Pipe1(
		publicationRefViewRecordLens,
		L.Compose[PublicationRefView](publicationRefTitleLens),
	)
	publicationRefViewJournalLens = F.Pipe1(
		publicationRefViewRecordLens,
		L.Compose[PublicationRefView](publicationRefJournalLens),
	)
	publicationRefViewPubDateLens = F.Pipe1(
		publicationRefViewRecordLens,
		L.Compose[PublicationRefView](publicationRefPubDateLens),
	)
	publicationRefViewDOILens = F.Pipe1(
		publicationRefViewRecordLens,
		L.Compose[PublicationRefView](publicationRefDOILens),
	)
	publicationRefViewAuthorsLens = F.Pipe1(
		publicationRefViewRecordLens,
		L.Compose[PublicationRefView](publicationRefAuthorsLens),
	)
	phenotypePubRefViewIDLens = F.Pipe1(
		phenotypePubRefViewRecordLens,
		L.Compose[PhenotypePublicationRefView](phenotypePubRefIDLens),
	)
	phenotypeItemViewPhenotypeLens = F.Pipe1(
		phenotypeItemViewRecordLens,
		L.Compose[PhenotypeItemView](phenotypeItemPhenotypeLens),
	)
	phenotypeItemViewEnvironmentLens = F.Pipe1(
		phenotypeItemViewRecordLens,
		L.Compose[PhenotypeItemView](phenotypeItemEnvironmentLens),
	)
	phenotypeItemViewAssayLens = F.Pipe1(
		phenotypeItemViewRecordLens,
		L.Compose[PhenotypeItemView](phenotypeItemAssayLens),
	)
	phenotypeItemViewPaperURLLens = F.Pipe1(
		phenotypeItemViewPublicationLens,
		L.Compose[PhenotypeItemView](phenotypePubRefViewURLLens),
	)
	phenotypeItemViewPaperIDLens = F.Pipe1(
		phenotypeItemViewPublicationLens,
		L.Compose[PhenotypeItemView](phenotypePubRefViewIDLens),
	)

	strainDetailIDLens = L.MakeLens(
		func(d graphql.StrainDetail) string { return d.ID },
		func(d graphql.StrainDetail, id string) graphql.StrainDetail {
			d.ID = id
			return d
		},
	)
	strainDetailLabelLens = L.MakeLens(
		func(d graphql.StrainDetail) string { return d.Label },
		func(d graphql.StrainDetail, label string) graphql.StrainDetail {
			d.Label = label
			return d
		},
	)
	strainDetailSpeciesLens = L.MakeLens(
		func(d graphql.StrainDetail) string { return d.Species },
		func(d graphql.StrainDetail, species string) graphql.StrainDetail {
			d.Species = species
			return d
		},
	)
	strainDetailInStockLens = L.MakeLens(
		func(d graphql.StrainDetail) bool { return d.InStock },
		func(d graphql.StrainDetail, inStock bool) graphql.StrainDetail {
			d.InStock = inStock
			return d
		},
	)
	strainDetailPhenotypesLens = L.MakeLens(
		func(d graphql.StrainDetail) []graphql.PhenotypeItem { return d.Phenotypes },
		func(d graphql.StrainDetail, items []graphql.PhenotypeItem) graphql.StrainDetail {
			d.Phenotypes = items
			return d
		},
	)
	strainDetailPublicationsLens = L.MakeLens(
		func(d graphql.StrainDetail) []graphql.PublicationRef { return d.Publications },
		func(d graphql.StrainDetail, items []graphql.PublicationRef) graphql.StrainDetail {
			d.Publications = items
			return d
		},
	)
	plasmidDetailIDLens = L.MakeLens(
		func(d graphql.PlasmidDetail) string { return d.ID },
		func(d graphql.PlasmidDetail, id string) graphql.PlasmidDetail {
			d.ID = id
			return d
		},
	)
	plasmidDetailNameLens = L.MakeLens(
		func(d graphql.PlasmidDetail) string { return d.Name },
		func(d graphql.PlasmidDetail, name string) graphql.PlasmidDetail {
			d.Name = name
			return d
		},
	)
	plasmidDetailInStockLens = L.MakeLens(
		func(d graphql.PlasmidDetail) bool { return d.InStock },
		func(d graphql.PlasmidDetail, inStock bool) graphql.PlasmidDetail {
			d.InStock = inStock
			return d
		},
	)
	plasmidDetailGenbankLens = L.MakeLens(
		func(d graphql.PlasmidDetail) string { return d.GenbankAccession },
		func(d graphql.PlasmidDetail, accession string) graphql.PlasmidDetail {
			d.GenbankAccession = accession
			return d
		},
	)
	plasmidDetailSequenceLens = L.MakeLens(
		func(d graphql.PlasmidDetail) string { return d.Sequence },
		func(d graphql.PlasmidDetail, seq string) graphql.PlasmidDetail {
			d.Sequence = seq
			return d
		},
	)
	plasmidDetailPublicationsLens = L.MakeLens(
		func(d graphql.PlasmidDetail) []graphql.PublicationRef { return d.Publications },
		func(d graphql.PlasmidDetail, items []graphql.PublicationRef) graphql.PlasmidDetail {
			d.Publications = items
			return d
		},
	)
	searchHitIDLens = L.MakeLens(
		func(h graphql.StockSearchHit) string { return h.ID },
		func(h graphql.StockSearchHit, id string) graphql.StockSearchHit {
			h.ID = id
			return h
		},
	)
	searchHitEntityLens = L.MakeLens(
		func(h graphql.StockSearchHit) graphql.StockEntityType { return h.Entity },
		func(h graphql.StockSearchHit, entity graphql.StockEntityType) graphql.StockSearchHit {
			h.Entity = entity
			return h
		},
	)
	searchHitStockLabelLens = L.MakeLens(
		func(h graphql.StockSearchHit) string { return h.StockLabel },
		func(h graphql.StockSearchHit, label string) graphql.StockSearchHit {
			h.StockLabel = label
			return h
		},
	)
	searchResultResultsLens = L.MakeLens(
		func(r graphql.SearchResult) []graphql.StockSearchHit { return r.Results },
		func(r graphql.SearchResult, hits []graphql.StockSearchHit) graphql.SearchResult {
			r.Results = hits
			return r
		},
	)
	searchResultLimitLens = L.MakeLens(
		func(r graphql.SearchResult) int { return r.Limit },
		func(r graphql.SearchResult, limit int) graphql.SearchResult {
			r.Limit = limit
			return r
		},
	)
	searchResultTotalCountLens = L.MakeLens(
		func(r graphql.SearchResult) int { return r.TotalCount },
		func(r graphql.SearchResult, total int) graphql.SearchResult {
			r.TotalCount = total
			return r
		},
	)
	phenotypeStrainItemIDLens = L.MakeLens(
		func(item graphql.PhenotypeStrainItem) string { return item.ID },
		func(item graphql.PhenotypeStrainItem, id string) graphql.PhenotypeStrainItem {
			item.ID = id
			return item
		},
	)
	phenotypeStrainListStrainsLens = L.MakeLens(
		func(l graphql.PhenotypeStrainList) []graphql.PhenotypeStrainItem { return l.Strains },
		func(l graphql.PhenotypeStrainList, items []graphql.PhenotypeStrainItem) graphql.PhenotypeStrainList {
			l.Strains = items
			return l
		},
	)
	phenotypeStrainListCursorLens = L.MakeLens(
		func(l graphql.PhenotypeStrainList) int64 { return l.NextCursor },
		func(l graphql.PhenotypeStrainList, cursor int64) graphql.PhenotypeStrainList {
			l.NextCursor = cursor
			return l
		},
	)
	geneStrainItemIDLens = L.MakeLens(
		func(item graphql.GeneStrainItem) string { return item.ID },
		func(item graphql.GeneStrainItem, id string) graphql.GeneStrainItem {
			item.ID = id
			return item
		},
	)
	strainViewRecordLens = L.MakeLens(
		func(v StrainView) graphql.StrainDetail { return v.StrainDetail },
		func(v StrainView, d graphql.StrainDetail) StrainView {
			v.StrainDetail = d
			return v
		},
	)
	strainViewURLLens = L.MakeLens(
		func(v StrainView) string { return v.URL },
		func(v StrainView, url string) StrainView {
			v.URL = url
			return v
		},
	)
	strainViewDetailLens = L.MakeLens(
		func(v StrainView) string { return v.Detail },
		func(v StrainView, detail string) StrainView {
			v.Detail = detail
			return v
		},
	)
	strainViewPhenotypeCountLens = L.MakeLens(
		func(v StrainView) int { return v.PhenotypeCount },
		func(v StrainView, count int) StrainView {
			v.PhenotypeCount = count
			return v
		},
	)

	strainViewPublicationCountLens = L.MakeLens(
		func(v StrainView) int { return v.PublicationCount },
		func(v StrainView, count int) StrainView {
			v.PublicationCount = count
			return v
		},
	)
	plasmidViewRecordLens = L.MakeLens(
		func(v PlasmidView) graphql.PlasmidDetail { return v.PlasmidDetail },
		func(v PlasmidView, d graphql.PlasmidDetail) PlasmidView {
			v.PlasmidDetail = d
			return v
		},
	)
	plasmidViewURLLens = L.MakeLens(
		func(v PlasmidView) string { return v.URL },
		func(v PlasmidView, url string) PlasmidView {
			v.URL = url
			return v
		},
	)
	plasmidViewDetailLens = L.MakeLens(
		func(v PlasmidView) string { return v.Detail },
		func(v PlasmidView, detail string) PlasmidView {
			v.Detail = detail
			return v
		},
	)
	plasmidViewSequenceLengthLens = L.MakeLens(
		func(v PlasmidView) int { return v.SequenceLength },
		func(v PlasmidView, length int) PlasmidView {
			v.SequenceLength = length
			return v
		},
	)
	plasmidViewPublicationCountLens = L.MakeLens(
		func(v PlasmidView) int { return v.PublicationCount },
		func(v PlasmidView, count int) PlasmidView {
			v.PublicationCount = count
			return v
		},
	)
	searchHitViewRecordLens = L.MakeLens(
		func(v SearchHitView) graphql.StockSearchHit { return v.StockSearchHit },
		func(v SearchHitView, h graphql.StockSearchHit) SearchHitView {
			v.StockSearchHit = h
			return v
		},
	)
	searchHitViewURLLens = L.MakeLens(
		func(v SearchHitView) string { return v.URL },
		func(v SearchHitView, url string) SearchHitView {
			v.URL = url
			return v
		},
	)
	searchResultViewResultsLens = L.MakeLens(
		func(v SearchResultView) []SearchHitView { return v.Results },
		func(v SearchResultView, hits []SearchHitView) SearchResultView {
			v.Results = hits
			return v
		},
	)
	searchResultViewLimitLens = L.MakeLens(
		func(v SearchResultView) int { return v.Limit },
		func(v SearchResultView, limit int) SearchResultView {
			v.Limit = limit
			return v
		},
	)
	searchResultViewTotalLens = L.MakeLens(
		func(v SearchResultView) int { return v.TotalCount },
		func(v SearchResultView, total int) SearchResultView {
			v.TotalCount = total
			return v
		},
	)
	phenotypeStrainViewRecordLens = L.MakeLens(
		func(v PhenotypeStrainView) graphql.PhenotypeStrainItem { return v.PhenotypeStrainItem },
		func(v PhenotypeStrainView, item graphql.PhenotypeStrainItem) PhenotypeStrainView {
			v.PhenotypeStrainItem = item
			return v
		},
	)
	phenotypeStrainViewURLLens = L.MakeLens(
		func(v PhenotypeStrainView) string { return v.URL },
		func(v PhenotypeStrainView, url string) PhenotypeStrainView {
			v.URL = url
			return v
		},
	)
	phenotypeSearchViewMatchedLens = L.MakeLens(
		func(v PhenotypeSearchView) []MatchedTextView { return v.MatchedTexts },
		func(v PhenotypeSearchView, texts []MatchedTextView) PhenotypeSearchView {
			v.MatchedTexts = texts
			return v
		},
	)
	phenotypeSearchViewResultsLens = L.MakeLens(
		func(v PhenotypeSearchView) []SearchHitView { return v.Results },
		func(v PhenotypeSearchView, hits []SearchHitView) PhenotypeSearchView {
			v.Results = hits
			return v
		},
	)
	phenotypeSearchViewLimitLens = L.MakeLens(
		func(v PhenotypeSearchView) int { return v.Limit },
		func(v PhenotypeSearchView, limit int) PhenotypeSearchView {
			v.Limit = limit
			return v
		},
	)
	phenotypeSearchViewCountLens = L.MakeLens(
		func(v PhenotypeSearchView) int { return v.Count },
		func(v PhenotypeSearchView, count int) PhenotypeSearchView {
			v.Count = count
			return v
		},
	)
	phenotypeListViewTermLens = L.MakeLens(
		func(v PhenotypeStrainListView) string { return v.PhenotypeTerm },
		func(v PhenotypeStrainListView, term string) PhenotypeStrainListView {
			v.PhenotypeTerm = term
			return v
		},
	)
	phenotypeListViewStrainsLens = L.MakeLens(
		func(v PhenotypeStrainListView) []PhenotypeStrainView { return v.Strains },
		func(v PhenotypeStrainListView, items []PhenotypeStrainView) PhenotypeStrainListView {
			v.Strains = items
			return v
		},
	)
	phenotypeListViewCursorLens = L.MakeLens(
		func(v PhenotypeStrainListView) int64 { return v.NextCursor },
		func(v PhenotypeStrainListView, cursor int64) PhenotypeStrainListView {
			v.NextCursor = cursor
			return v
		},
	)
	phenotypeListViewHasMoreLens = L.MakeLens(
		func(v PhenotypeStrainListView) bool { return v.HasMore },
		func(v PhenotypeStrainListView, hasMore bool) PhenotypeStrainListView {
			v.HasMore = hasMore
			return v
		},
	)
	phenotypeListViewPageCountLens = L.MakeLens(
		func(v PhenotypeStrainListView) int { return v.PageCount },
		func(v PhenotypeStrainListView, count int) PhenotypeStrainListView {
			v.PageCount = count
			return v
		},
	)
	plasmidSequenceIDLens = L.MakeLens(
		func(s PlasmidSequence) string { return s.ID },
		func(s PlasmidSequence, id string) PlasmidSequence {
			s.ID = id
			return s
		},
	)
	plasmidSequenceNameLens = L.MakeLens(
		func(s PlasmidSequence) string { return s.Name },
		func(s PlasmidSequence, name string) PlasmidSequence {
			s.Name = name
			return s
		},
	)
	plasmidSequenceLengthLens = L.MakeLens(
		func(s PlasmidSequence) int { return s.Length },
		func(s PlasmidSequence, length int) PlasmidSequence {
			s.Length = length
			return s
		},
	)
	plasmidSequenceFormatLens = L.MakeLens(
		func(s PlasmidSequence) string { return s.Format },
		func(s PlasmidSequence, format string) PlasmidSequence {
			s.Format = format
			return s
		},
	)
	plasmidSequenceFASTALens = L.MakeLens(
		func(s PlasmidSequence) string { return s.FASTA },
		func(s PlasmidSequence, fasta string) PlasmidSequence {
			s.FASTA = fasta
			return s
		},
	)
	plasmidSequenceURLLens = L.MakeLens(
		func(s PlasmidSequence) string { return s.URL },
		func(s PlasmidSequence, url string) PlasmidSequence {
			s.URL = url
			return s
		},
	)

	strainViewIDLens = F.Pipe1(
		strainViewRecordLens,
		L.Compose[StrainView](strainDetailIDLens),
	)
	strainViewLabelLens = F.Pipe1(
		strainViewRecordLens,
		L.Compose[StrainView](strainDetailLabelLens),
	)
	strainViewSpeciesLens = F.Pipe1(
		strainViewRecordLens,
		L.Compose[StrainView](strainDetailSpeciesLens),
	)
	strainViewInStockLens = F.Pipe1(
		strainViewRecordLens,
		L.Compose[StrainView](strainDetailInStockLens),
	)
	plasmidViewIDLens = F.Pipe1(
		plasmidViewRecordLens,
		L.Compose[PlasmidView](plasmidDetailIDLens),
	)
	plasmidViewNameLens = F.Pipe1(
		plasmidViewRecordLens,
		L.Compose[PlasmidView](plasmidDetailNameLens),
	)
	plasmidViewInStockLens = F.Pipe1(
		plasmidViewRecordLens,
		L.Compose[PlasmidView](plasmidDetailInStockLens),
	)
	plasmidViewGenbankLens = F.Pipe1(
		plasmidViewRecordLens,
		L.Compose[PlasmidView](plasmidDetailGenbankLens),
	)

	plasmidViewRecordSequenceLens = F.Pipe1(
		plasmidViewRecordLens,
		L.Compose[PlasmidView](plasmidDetailSequenceLens),
	)

	strainViews = map[string]strainViewFunc{
		detailSummary: F.Curry2(summaryStrainView),
		detailFull:    F.Curry2(fullStrainView),
	}
	plasmidViews = map[string]plasmidViewFunc{
		detailSummary: F.Curry2(summaryPlasmidView),
		detailFull:    F.Curry2(fullPlasmidView),
	}

	hasSequence = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(plasmidDetailSequenceLens.Get),
	)

	hasNextPage = P.IsNonZero[int64]()

	strainViewGenotypesLens = L.MakeLens(
		func(v StrainView) []string { return v.Genotypes },
		func(v StrainView, value []string) StrainView {
			v.Genotypes = value
			return v
		},
	)
	strainViewDepositorLens = L.MakeLens(
		func(v StrainView) graphql.Depositor { return v.Depositor },
		func(v StrainView, value graphql.Depositor) StrainView {
			v.Depositor = value
			return v
		},
	)
	strainViewSummaryLens = L.MakeLens(
		func(v StrainView) string { return v.Summary },
		func(v StrainView, value string) StrainView {
			v.Summary = value
			return v
		},
	)
	plasmidViewKeywordsLens = L.MakeLens(
		func(v PlasmidView) []string { return v.Keywords },
		func(v PlasmidView, value []string) PlasmidView {
			v.Keywords = value
			return v
		},
	)
	plasmidViewDepositorLens = L.MakeLens(
		func(v PlasmidView) graphql.Depositor { return v.Depositor },
		func(v PlasmidView, value graphql.Depositor) PlasmidView {
			v.Depositor = value
			return v
		},
	)
	plasmidViewSummaryLens = L.MakeLens(
		func(v PlasmidView) string { return v.Summary },
		func(v PlasmidView, value string) PlasmidView {
			v.Summary = value
			return v
		},
	)
	searchHitViewIDLens = L.MakeLens(
		func(v SearchHitView) string { return v.ID },
		func(v SearchHitView, value string) SearchHitView {
			v.ID = value
			return v
		},
	)
	searchHitViewStockLabelLens = L.MakeLens(
		func(v SearchHitView) string { return v.StockLabel },
		func(v SearchHitView, value string) SearchHitView {
			v.StockLabel = value
			return v
		},
	)
	searchHitViewFieldLens = L.MakeLens(
		func(v SearchHitView) string { return v.Field },
		func(v SearchHitView, value string) SearchHitView {
			v.Field = value
			return v
		},
	)
	searchHitViewDisplayTextLens = L.MakeLens(
		func(v SearchHitView) string { return v.DisplayText },
		func(v SearchHitView, value string) SearchHitView {
			v.DisplayText = value
			return v
		},
	)
	searchHitViewInStockLens = L.MakeLens(
		func(v SearchHitView) bool { return v.InStock },
		func(v SearchHitView, value bool) SearchHitView {
			v.InStock = value
			return v
		},
	)
	searchHitViewStockTypeLens = L.MakeLens(
		func(v SearchHitView) string { return v.StockType },
		func(v SearchHitView, value string) SearchHitView {
			v.StockType = value
			return v
		},
	)
	phenotypeStrainViewIDLens = L.MakeLens(
		func(v PhenotypeStrainView) string { return v.ID },
		func(v PhenotypeStrainView, value string) PhenotypeStrainView {
			v.ID = value
			return v
		},
	)
	phenotypeStrainViewLabelLens = L.MakeLens(
		func(v PhenotypeStrainView) string { return v.Label },
		func(v PhenotypeStrainView, value string) PhenotypeStrainView {
			v.Label = value
			return v
		},
	)
	phenotypeStrainViewInStockLens = L.MakeLens(
		func(v PhenotypeStrainView) bool { return v.InStock },
		func(v PhenotypeStrainView, value bool) PhenotypeStrainView {
			v.InStock = value
			return v
		},
	)
)

func toPublicationRefView(links linkout.Links, p graphql.PublicationRef) PublicationRefView {
	return F.Pipe2(
		PublicationRefView{},
		publicationRefViewRecordLens.Set(p),
		publicationRefViewURLLens.Set(
			F.Pipe2(p, publicationRefIDLens.Get, links.PublicationURL),
		),
	)
}

func toPhenotypeItemView(links linkout.Links, p graphql.PhenotypeItem) PhenotypeItemView {
	paper := phenotypeItemPaperLens.Get(p)

	return F.Pipe2(
		PhenotypeItemView{},
		phenotypeItemViewRecordLens.Set(p),
		phenotypeItemViewPublicationLens.Set(F.Pipe2(
			PhenotypePublicationRefView{},
			phenotypePubRefViewRecordLens.Set(paper),
			phenotypePubRefViewURLLens.Set(
				F.Pipe2(paper, phenotypePubRefIDLens.Get, links.PublicationURL),
			),
		)),
	)
}

// publicationRefViewsOrEmpty converts papers to views with local URLs. It
// returns [], not nil, for no papers.
func publicationRefViewsOrEmpty(
	links linkout.Links,
	papers []graphql.PublicationRef,
) []PublicationRefView {
	return F.Pipe2(
		papers,
		A.Map(F.Bind1st(toPublicationRefView, links)),
		notNilSlice[PublicationRefView],
	)
}

// phenotypeItemViewsOrEmpty converts phenotypes to views with local URLs. It
// returns [], not nil, for no phenotypes.
func phenotypeItemViewsOrEmpty(
	links linkout.Links,
	phenotypes []graphql.PhenotypeItem,
) []PhenotypeItemView {
	return F.Pipe2(
		phenotypes,
		A.Map(F.Bind1st(toPhenotypeItemView, links)),
		notNilSlice[PhenotypeItemView],
	)
}

// takeAtMost keeps the first n items and never returns nil, so JSON emits [] not null.
func takeAtMost[T any](n int) func([]T) []T {
	return F.Flow2(A.Slice[T](0, n), notNilSlice[T])
}

func emptySlice[T any]() []T { return make([]T, 0) }

// notNilSlice keeps JSON output [] instead of null, because fp-go A.Map and
// A.Uniq return nil for an empty input.
func notNilSlice[T any](items []T) []T {
	return F.Pipe2(
		items,
		O.FromPredicate(predarrays.IsNonEmpty[T]()),
		O.GetOrElse(emptySlice[T]),
	)
}

func fullStrainView(links linkout.Links, d graphql.StrainDetail) StrainView {
	phenotypes := strainDetailPhenotypesLens.Get(d)
	papers := strainDetailPublicationsLens.Get(d)

	return F.Pipe7(
		StrainView{},
		strainViewRecordLens.Set(d),
		strainViewURLLens.Set(F.Pipe2(d, strainDetailIDLens.Get, links.StrainURL)),
		strainViewDetailLens.Set(detailFull),
		strainViewPhenotypeCountLens.Set(A.Size(phenotypes)),
		strainViewPublicationCountLens.Set(A.Size(papers)),
		strainViewPhenotypesLens.Set(F.Pipe1(phenotypes, F.Bind1st(phenotypeItemViewsOrEmpty, links))),
		strainViewPublicationsLens.Set(F.Pipe1(papers, F.Bind1st(publicationRefViewsOrEmpty, links))),
	)
}

func summaryStrainView(links linkout.Links, d graphql.StrainDetail) StrainView {
	return F.Pipe3(
		fullStrainView(links, d),
		strainViewDetailLens.Set(detailSummary),
		strainViewPhenotypesLens.Set(F.Pipe3(
			d,
			strainDetailPhenotypesLens.Get,
			takeAtMost[graphql.PhenotypeItem](summaryPhenotypeCap),
			F.Bind1st(phenotypeItemViewsOrEmpty, links),
		)),
		strainViewPublicationsLens.Set(F.Pipe3(
			d,
			strainDetailPublicationsLens.Get,
			takeAtMost[graphql.PublicationRef](summaryPublicationCap),
			F.Bind1st(publicationRefViewsOrEmpty, links),
		)),
	)
}

func strainViewFor(links linkout.Links, detail string) func(graphql.StrainDetail) StrainView {
	return F.Pipe3(
		strainViews,
		R.Lookup[strainViewFunc](detail),
		O.GetOrElse(F.Constant(F.Curry2(summaryStrainView))),
		// Apply the selected view to links. The result takes the strain record.
		I.Flap[func(graphql.StrainDetail) StrainView](links),
	)
}

func fullPlasmidView(links linkout.Links, d graphql.PlasmidDetail) PlasmidView {
	papers := plasmidDetailPublicationsLens.Get(d)

	return F.Pipe6(
		PlasmidView{},
		plasmidViewRecordLens.Set(d),
		plasmidViewURLLens.Set(F.Pipe2(d, plasmidDetailIDLens.Get, links.PlasmidURL)),
		plasmidViewDetailLens.Set(detailFull),
		plasmidViewSequenceLengthLens.Set(F.Pipe2(d, plasmidDetailSequenceLens.Get, S.Size)),
		plasmidViewPublicationCountLens.Set(A.Size(papers)),
		plasmidViewPublicationsLens.Set(F.Pipe1(papers, F.Bind1st(publicationRefViewsOrEmpty, links))),
	)
}

func summaryPlasmidView(links linkout.Links, d graphql.PlasmidDetail) PlasmidView {
	return F.Pipe3(
		fullPlasmidView(links, d),
		plasmidViewDetailLens.Set(detailSummary),
		plasmidViewRecordSequenceLens.Set(S.Monoid.Empty()),
		plasmidViewPublicationsLens.Set(F.Pipe3(
			d,
			plasmidDetailPublicationsLens.Get,
			takeAtMost[graphql.PublicationRef](summaryPublicationCap),
			F.Bind1st(publicationRefViewsOrEmpty, links),
		)),
	)
}

func plasmidViewFor(links linkout.Links, detail string) func(graphql.PlasmidDetail) PlasmidView {
	return F.Pipe3(
		plasmidViews,
		R.Lookup[plasmidViewFunc](detail),
		O.GetOrElse(F.Constant(F.Curry2(summaryPlasmidView))),
		// Apply the selected view to links. The result takes the plasmid record.
		I.Flap[func(graphql.PlasmidDetail) PlasmidView](links),
	)
}

func errNoSequence(d graphql.PlasmidDetail) error {
	return fmt.Errorf("plasmid %s has no sequence on record", plasmidDetailIDLens.Get(d))
}

// compactSequence removes all whitespace; single-line FASTA is valid and simplest for LLM clients.

func toPlasmidSequence(links linkout.Links, d graphql.PlasmidDetail) PlasmidSequence {
	id := plasmidDetailIDLens.Get(d)
	name := plasmidDetailNameLens.Get(d)
	seq := F.Pipe3(d, plasmidDetailSequenceLens.Get, strings.Fields, M.ConcatAll(S.Monoid))

	return F.Pipe6(
		PlasmidSequence{},
		plasmidSequenceIDLens.Set(id),
		plasmidSequenceNameLens.Set(name),
		plasmidSequenceLengthLens.Set(S.Size(seq)),
		plasmidSequenceFormatLens.Set(fastaFormat),
		plasmidSequenceFASTALens.Set(fmt.Sprintf(">%s %s\n%s", id, name, seq)),
		plasmidSequenceURLLens.Set(links.PlasmidURL(id)),
	)
}

func toSearchHitView(links linkout.Links, h graphql.StockSearchHit) SearchHitView {
	return F.Pipe2(
		SearchHitView{},
		searchHitViewRecordLens.Set(h),
		searchHitViewURLLens.Set(searchHitURL(links, h)),
	)
}

// searchHitURL returns the stock page of h, or "" for an unknown entity.
func searchHitURL(links linkout.Links, h graphql.StockSearchHit) string {
	return StockURL(links, searchHitEntityLens.Get(h), searchHitIDLens.Get(h))
}

func toSearchResultView(links linkout.Links, r graphql.SearchResult) SearchResultView {
	return F.Pipe3(
		SearchResultView{},
		searchResultViewResultsLens.Set(F.Pipe3(
			r,
			searchResultResultsLens.Get,
			A.Map(F.Bind1st(toSearchHitView, links)),
			notNilSlice[SearchHitView],
		)),
		searchResultViewLimitLens.Set(searchResultLimitLens.Get(r)),
		searchResultViewTotalLens.Set(searchResultTotalCountLens.Get(r)),
	)
}

func toPhenotypeStrainView(links linkout.Links, item graphql.PhenotypeStrainItem) PhenotypeStrainView {
	return F.Pipe2(
		PhenotypeStrainView{},
		phenotypeStrainViewRecordLens.Set(item),
		phenotypeStrainViewURLLens.Set(
			F.Pipe2(item, phenotypeStrainItemIDLens.Get, links.StrainURL),
		),
	)
}

func hitMatchKey(h graphql.StockSearchHit) matchKey {
	return F.Pipe2(
		matchKey{},
		matchKeyFieldLens.Set(searchHitFieldLens.Get(h)),
		matchKeyDisplayTextLens.Set(searchHitDisplayTextLens.Get(h)),
	)
}

// countMatch is the reducer of toMatchedTexts. It writes only to the map that
// toMatchedTexts creates for one call.
func countMatch(counts map[matchKey]int, h graphql.StockSearchHit) map[matchKey]int {
	counts[hitMatchKey(h)]++
	return counts
}

func matchedTextView(counts map[matchKey]int, h graphql.StockSearchHit) MatchedTextView {
	return F.Pipe3(
		MatchedTextView{},
		matchedTextViewFieldLens.Set(searchHitFieldLens.Get(h)),
		matchedTextViewDisplayTextLens.Set(searchHitDisplayTextLens.Get(h)),
		matchedTextViewStrainCountLens.Set(counts[hitMatchKey(h)]),
	)
}

// toMatchedTexts groups ranked rows by matched text in first-seen order, so
// the most relevant text comes first.
func toMatchedTexts(hits []graphql.StockSearchHit) []MatchedTextView {
	return F.Pipe3(
		hits,
		A.Uniq(hitMatchKey),
		// Count every row per matched text, then attach the count to each unique text.
		A.Map(F.Bind1st(matchedTextView, F.Pipe1(hits, A.Reduce(countMatch, make(map[matchKey]int))))),
		notNilSlice[MatchedTextView],
	)
}

func toPhenotypeSearchView(links linkout.Links, r graphql.SearchResult) PhenotypeSearchView {
	hits := searchResultResultsLens.Get(r)

	return F.Pipe4(
		PhenotypeSearchView{},
		phenotypeSearchViewMatchedLens.Set(toMatchedTexts(hits)),
		phenotypeSearchViewResultsLens.Set(F.Pipe2(
			hits,
			A.Map(F.Bind1st(toSearchHitView, links)),
			notNilSlice[SearchHitView],
		)),
		phenotypeSearchViewLimitLens.Set(searchResultLimitLens.Get(r)),
		phenotypeSearchViewCountLens.Set(A.Size(hits)),
	)
}

func phenotypeStrainListView(
	links linkout.Links,
	term string,
	l graphql.PhenotypeStrainList,
) PhenotypeStrainListView {
	strains := phenotypeStrainListStrainsLens.Get(l)
	cursor := phenotypeStrainListCursorLens.Get(l)

	return F.Pipe5(
		PhenotypeStrainListView{},
		phenotypeListViewTermLens.Set(term),
		phenotypeListViewStrainsLens.Set(F.Pipe2(
			strains,
			A.Map(F.Bind1st(toPhenotypeStrainView, links)),
			notNilSlice[PhenotypeStrainView],
		)),
		phenotypeListViewCursorLens.Set(cursor),
		phenotypeListViewHasMoreLens.Set(hasNextPage(cursor)),
		phenotypeListViewPageCountLens.Set(A.Size(strains)),
	)
}
