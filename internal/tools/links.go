package tools

import (
	"fmt"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
)

// stockRoute is a curried record-page builder: it takes the configured
// website first and the raw record ID second.
type stockRoute = func(linkout.Links) func(string) string

const (
	pubmedBaseURL = "https://pubmed.ncbi.nlm.nih.gov"
	doiBaseURL    = "https://doi.org"

	strainIDPrefix   = "DBS"
	plasmidIDPrefix  = "DBP"
	stockIDPrefixLen = 3
)

var (
	// PubMedURL returns the PubMed page for a dictyBase publication ID (a PubMed ID).
	PubMedURL = F.Flow2(S.Prepend(pubmedBaseURL+"/"), S.Append("/"))

	stockURLByEntity = map[graphql.StockEntityType]stockRoute{
		graphql.StockEntityStrain:  F.Curry2(linkout.Links.StrainURL),
		graphql.StockEntityPlasmid: F.Curry2(linkout.Links.PlasmidURL),
	}
)

// DOIURL returns the doi.org resolver link, or "" when doi is empty.
func DOIURL(doi string) string {
	return F.Pipe3(
		doi,
		O.FromPredicate(S.IsNonEmpty),
		O.Map(S.Prepend(doiBaseURL+"/")),
		O.GetOrElse(F.Constant("")),
	)
}

// StockURL returns the stock page for id given its entity, or "" for unknown entities.
func StockURL(links linkout.Links, entity graphql.StockEntityType, id string) string {
	return F.Pipe4(
		stockURLByEntity,
		R.Lookup[stockRoute](entity),
		O.Ap[func(string) string](O.Of(links)),
		O.Ap[string](O.Of(id)),
		O.GetOrElse(F.Constant(S.Monoid.Empty())),
	)
}

// IDPrefix returns the upper-cased first three characters of id ("DBS", "DBP", ...).
func IDPrefix(id string) string {
	return strings.ToUpper(id[:min(len(id), stockIDPrefixLen)])
}

// routeEffect resolves the ID prefix of state to the lazy effect that
// serves it, or to the unknown-ID error. The step is pure: O.Map applies
// the found route to state, which builds an unexecuted IOEither, so no
// GraphQL request runs here. check_availability and fetch both call it.
func routeEffect[S, A any](
	routes map[string]func(S) IOE.IOEither[error, A],
	idOf func(S) string,
	state S,
) E.Either[error, IOE.IOEither[error, A]] {
	return F.Pipe3(
		routes,
		R.Lookup[func(S) IOE.IOEither[error, A]](F.Pipe2(state, idOf, IDPrefix)),
		O.Map(F.Bind1st(
			F.Pipe1[func(S) IOE.IOEither[error, A], S, IOE.IOEither[error, A]],
			state,
		)),
		E.FromOption[IOE.IOEither[error, A]](
			F.Constant(fmt.Errorf(
				"unrecognized stock id %q: expected a DBS (strain) or DBP (plasmid) id",
				idOf(state),
			)),
		),
	)
}
