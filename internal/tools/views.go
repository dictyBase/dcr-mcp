package tools

import (
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictybase/dcr-mcp/internal/graphql"
)

type (
	// StrainView is the get_strain output: the GraphQL record plus presentation fields.
	StrainView struct {
		graphql.StrainDetail
		URL              string `json:"url"`
		Detail           string `json:"detail"`
		PhenotypeCount   int    `json:"phenotype_count"`
		PublicationCount int    `json:"publication_count"`
	}

	// PlasmidView is the get_plasmid output: the GraphQL record plus presentation fields.
	PlasmidView struct {
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

var (
	strainViews = map[string]func(graphql.StrainDetail) StrainView{
		detailSummary: summaryStrainView,
		detailFull:    fullStrainView,
	}

	plasmidViews = map[string]func(graphql.PlasmidDetail) PlasmidView{
		detailSummary: summaryPlasmidView,
		detailFull:    fullPlasmidView,
	}

	hasSequence = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(func(d graphql.PlasmidDetail) string { return d.Sequence }),
	)

	toMatchedTextView = F.Curry2(matchedTextView)

	hasNextPage = P.IsNonZero[int64]()

	toPhenotypeStrainListView = F.Curry2(phenotypeStrainListView)
)

// takeAtMost keeps the first n items and never returns nil, so JSON emits [] not null.
func takeAtMost[T any](n int) func([]T) []T {
	hasElem := predarrays.IsNonEmpty[T]()
	return F.Flow3(
		A.Slice[T](0, n),
		O.FromPredicate(hasElem),
		O.GetOrElse(func() []T { return make([]T, 0) }),
	)
}

func emptySlice[T any]() []T { return make([]T, 0) }

// notNilSlice keeps JSON output [] instead of null, because fp-go A.Map and
// A.Uniq return nil for an empty input.
func notNilSlice[T any](items []T) []T {
	hasElem := predarrays.IsNonEmpty[T]()
	return F.Pipe2(
		items,
		O.FromPredicate(hasElem),
		O.GetOrElse(emptySlice[T]),
	)
}

func fullStrainView(d graphql.StrainDetail) StrainView {
	return StrainView{
		StrainDetail:     d,
		URL:              StrainURL(d.ID),
		Detail:           detailFull,
		PhenotypeCount:   A.Size(d.Phenotypes),
		PublicationCount: A.Size(d.Publications),
	}
}

func summaryStrainView(d graphql.StrainDetail) StrainView {
	view := fullStrainView(d)
	view.Detail = detailSummary
	view.Phenotypes = F.Pipe1(d.Phenotypes, takeAtMost[graphql.PhenotypeItem](summaryPhenotypeCap))
	view.Publications = F.Pipe1(d.Publications, takeAtMost[graphql.PublicationRef](summaryPublicationCap))
	return view
}

func strainViewFor(detail string) func(graphql.StrainDetail) StrainView {
	return F.Pipe2(
		strainViews,
		R.Lookup[func(graphql.StrainDetail) StrainView](detail),
		O.GetOrElse(F.Constant(summaryStrainView)),
	)
}

func fullPlasmidView(d graphql.PlasmidDetail) PlasmidView {
	return PlasmidView{
		PlasmidDetail:    d,
		URL:              PlasmidURL(d.ID),
		Detail:           detailFull,
		SequenceLength:   len(d.Sequence),
		PublicationCount: A.Size(d.Publications),
	}
}

func summaryPlasmidView(d graphql.PlasmidDetail) PlasmidView {
	view := fullPlasmidView(d)
	view.Detail = detailSummary
	view.Sequence = ""
	view.Publications = F.Pipe1(d.Publications, takeAtMost[graphql.PublicationRef](summaryPublicationCap))
	return view
}

func plasmidViewFor(detail string) func(graphql.PlasmidDetail) PlasmidView {
	return F.Pipe2(
		plasmidViews,
		R.Lookup[func(graphql.PlasmidDetail) PlasmidView](detail),
		O.GetOrElse(F.Constant(summaryPlasmidView)),
	)
}

func errNoSequence(d graphql.PlasmidDetail) error {
	return fmt.Errorf("plasmid %s has no sequence on record", d.ID)
}

// compactSequence removes all whitespace; single-line FASTA is valid and simplest for LLM clients.
func compactSequence(seq string) string {
	return strings.Join(strings.Fields(seq), "")
}

func toPlasmidSequence(d graphql.PlasmidDetail) PlasmidSequence {
	seq := compactSequence(d.Sequence)
	return PlasmidSequence{
		ID:     d.ID,
		Name:   d.Name,
		Length: len(seq),
		Format: "fasta",
		FASTA:  fmt.Sprintf(">%s %s\n%s", d.ID, d.Name, seq),
		URL:    PlasmidURL(d.ID),
	}
}

func toSearchHitView(h graphql.StockSearchHit) SearchHitView {
	return SearchHitView{StockSearchHit: h, URL: StockURL(h.Entity, h.ID)}
}

func toSearchResultView(r graphql.SearchResult) SearchResultView {
	return SearchResultView{
		Results:    F.Pipe1(r.Results, A.Map(toSearchHitView)),
		Limit:      r.Limit,
		TotalCount: r.TotalCount,
	}
}

func toPhenotypeStrainView(item graphql.PhenotypeStrainItem) PhenotypeStrainView {
	return PhenotypeStrainView{PhenotypeStrainItem: item, URL: StrainURL(item.ID)}
}

func hitMatchKey(h graphql.StockSearchHit) matchKey {
	return matchKey{field: h.Field, displayText: h.DisplayText}
}

// countMatch is the reducer of toMatchedTexts. It writes only to the map that
// toMatchedTexts creates for one call.
func countMatch(counts map[matchKey]int, h graphql.StockSearchHit) map[matchKey]int {
	counts[hitMatchKey(h)]++
	return counts
}

func matchedTextView(counts map[matchKey]int, h graphql.StockSearchHit) MatchedTextView {
	return MatchedTextView{
		Field:       h.Field,
		DisplayText: h.DisplayText,
		StrainCount: counts[hitMatchKey(h)],
	}
}

// toMatchedTexts groups ranked rows by matched text in first-seen order, so
// the most relevant text comes first.
func toMatchedTexts(hits []graphql.StockSearchHit) []MatchedTextView {
	counts := F.Pipe1(hits, A.Reduce(countMatch, make(map[matchKey]int)))
	return F.Pipe3(
		hits,
		A.Uniq(hitMatchKey),
		A.Map(F.Pipe1(counts, toMatchedTextView)),
		notNilSlice[MatchedTextView],
	)
}

func toPhenotypeSearchView(r graphql.SearchResult) PhenotypeSearchView {
	return PhenotypeSearchView{
		MatchedTexts: toMatchedTexts(r.Results),
		Results:      F.Pipe2(r.Results, A.Map(toSearchHitView), notNilSlice[SearchHitView]),
		Limit:        r.Limit,
		Count:        A.Size(r.Results),
	}
}

func phenotypeStrainListView(
	term string,
	l graphql.PhenotypeStrainList,
) PhenotypeStrainListView {
	return PhenotypeStrainListView{
		PhenotypeTerm: term,
		Strains:       F.Pipe2(l.Strains, A.Map(toPhenotypeStrainView), notNilSlice[PhenotypeStrainView]),
		NextCursor:    l.NextCursor,
		HasMore:       hasNextPage(l.NextCursor),
		PageCount:     A.Size(l.Strains),
	}
}
