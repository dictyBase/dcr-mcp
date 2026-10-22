package tools

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
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
