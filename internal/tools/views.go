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

// StrainView is the get_strain output: the GraphQL record plus presentation fields.
type StrainView struct {
	graphql.StrainDetail
	URL              string `json:"url"`
	Detail           string `json:"detail"`
	PhenotypeCount   int    `json:"phenotype_count"`
	PublicationCount int    `json:"publication_count"`
}

// takeAtMost keeps the first n items and never returns nil, so JSON emits [] not null.
func takeAtMost[T any](n int) func([]T) []T {
	hasElem := predarrays.IsNonEmpty[T]()
	return F.Flow3(
		A.Slice[T](0, n),
		O.FromPredicate(hasElem),
		O.GetOrElse(func() []T { return make([]T, 0) }),
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

var strainViews = map[string]func(graphql.StrainDetail) StrainView{
	detailSummary: summaryStrainView,
	detailFull:    fullStrainView,
}

func strainViewFor(detail string) func(graphql.StrainDetail) StrainView {
	return F.Pipe2(
		strainViews,
		R.Lookup[func(graphql.StrainDetail) StrainView](detail),
		O.GetOrElse(F.Constant(summaryStrainView)),
	)
}

// PlasmidView is the get_plasmid output: the GraphQL record plus presentation fields.
type PlasmidView struct {
	graphql.PlasmidDetail
	URL              string `json:"url"`
	Detail           string `json:"detail"`
	SequenceLength   int    `json:"sequence_length"`
	PublicationCount int    `json:"publication_count"`
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

var plasmidViews = map[string]func(graphql.PlasmidDetail) PlasmidView{
	detailSummary: summaryPlasmidView,
	detailFull:    fullPlasmidView,
}

func plasmidViewFor(detail string) func(graphql.PlasmidDetail) PlasmidView {
	return F.Pipe2(
		plasmidViews,
		R.Lookup[func(graphql.PlasmidDetail) PlasmidView](detail),
		O.GetOrElse(F.Constant(summaryPlasmidView)),
	)
}

// PlasmidSequence is the get_plasmid_sequence output.
type PlasmidSequence struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Length int    `json:"length"`
	Format string `json:"format"`
	FASTA  string `json:"fasta"`
	URL    string `json:"url"`
}

var hasSequence = F.Pipe1(
	S.IsNonEmpty,
	P.ContraMap(func(d graphql.PlasmidDetail) string { return d.Sequence }),
)

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

// SearchHitView is one search_stocks hit plus its stock page url.
type SearchHitView struct {
	graphql.StockSearchHit
	URL string `json:"url"`
}

// SearchResultView is the search_stocks output.
type SearchResultView struct {
	Results    []SearchHitView `json:"results"`
	Limit      int             `json:"limit"`
	TotalCount int             `json:"total_count"`
}

// PhenotypeStrainView is one list_strains_with_phenotype item plus its stock page url.
type PhenotypeStrainView struct {
	graphql.PhenotypeStrainItem
	URL string `json:"url"`
}

// PhenotypeStrainListView is the list_strains_with_phenotype output.
type PhenotypeStrainListView struct {
	Strains    []PhenotypeStrainView `json:"strains"`
	NextCursor int64                 `json:"next_cursor"`
	TotalCount int                   `json:"total_count"`
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

func toPhenotypeStrainListView(l graphql.PhenotypeStrainList) PhenotypeStrainListView {
	return PhenotypeStrainListView{
		Strains:    F.Pipe1(l.Strains, A.Map(toPhenotypeStrainView)),
		NextCursor: l.NextCursor,
		TotalCount: l.TotalCount,
	}
}
