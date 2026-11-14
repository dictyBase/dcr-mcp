package tools

import (
	"fmt"
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
)

type (
	// stockRoute takes the configured website first and the raw record ID second.
	stockRoute = func(linkout.Links) func(string) string

	// stockEffect builds a lazy stock operation from its state.
	stockEffect[S, A any] = func(S) IOE.IOEither[error, A]
)

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

// routeEffect selects a route, rejects unknown IDs, and applies the state.
// The result is a lazy effect. This function does not execute it.
func routeEffect[S, A any](
	routes map[string]stockEffect[S, A],
	idOf func(S) string,
	state S,
) IOE.IOEither[error, A] {
	return F.Pipe6(
		state,    // Start with the state that contains the ID and request dependencies.
		idOf,     // Extract the stock ID from the state.
		IDPrefix, // Convert the first three characters to an uppercase route key.
		// Bind routes as the first lookup argument. The prefix supplies the second.
		// The lookup returns Some(route) for a known prefix, or None otherwise.
		F.Bind1st(R.MonadLookup[stockEffect[S, A], string], routes),
		// Convert Some(route) to a lazy Right(route), or None to a lazy Left(error).
		IOE.FromOption[stockEffect[S, A]](
			// Constant supplies this error when FromOption encounters None.
			F.Constant(fmt.Errorf(
				"unrecognized stock id %q: expected a DBS (strain) or DBP (plasmid) id",
				idOf(state), // Include the complete ID, not just its prefix.
			)),
		),
		// Apply the selected route to state. A Left passes through unchanged.
		IOE.Flap[error, IOE.IOEither[error, A]](state),
		// Merge the route effect into the outer effect: IOEither[IOEither[A]] -> IOEither[A].
		IOE.Flatten[error, A],
	)
}
