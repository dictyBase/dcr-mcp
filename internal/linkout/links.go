// Package linkout builds public dictyBase website links for catalog records.
package linkout

import (
	"net/url"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	eitherparse "github.com/dictyBase/fp-go-loom/either/parse"
	eithererr "github.com/dictyBase/fp-go-loom/eithererr"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	predstrings "github.com/dictyBase/fp-go-loom/predicate/strings"
	"github.com/dictyBase/fp-go-loom/strutils"
)

// Links holds one checked website origin. The zero value uses
// DefaultBaseURL, so a direct library caller needs no parse step.
type Links struct {
	baseURL string
}

const (
	// DefaultBaseURL is the website that serves the four record pages.
	DefaultBaseURL = "https://dictybase.dev"

	genePath        = "/gene/"
	strainPath      = "/stockcenter/strains/"
	plasmidPath     = "/stockcenter/plasmids/"
	publicationPath = "/publication/"

	schemeHTTP  = "http"
	schemeHTTPS = "https"
	rootSlash   = "/"
	colon       = ":"
	markerChars = "?#"

	minTCPPort = 1
	maxTCPPort = 65535
)

var (
	baseURLLens = L.MakeLens(
		func(links Links) string { return links.baseURL },
		func(links Links, base string) Links {
			links.baseURL = base
			return links
		},
	)

	// The five lenses below hold every direct field selector of a
	// net/url record. Each setter copies the record, so no predicate and
	// no pipeline step changes the parsed value in place. Hostname and
	// Port stay method expressions, because net/url exposes them as
	// methods instead of fields.
	urlSchemeLens = L.MakeLens(
		func(target *url.URL) string { return target.Scheme },
		func(target *url.URL, scheme string) *url.URL {
			clone := *target
			clone.Scheme = scheme
			return &clone
		},
	)
	urlHostLens = L.MakeLens(
		func(target *url.URL) string { return target.Host },
		func(target *url.URL, host string) *url.URL {
			clone := *target
			clone.Host = host
			return &clone
		},
	)
	urlPathLens = L.MakeLens(
		func(target *url.URL) string { return target.Path },
		func(target *url.URL, path string) *url.URL {
			clone := *target
			clone.Path = path
			return &clone
		},
	)
	urlRawPathLens = L.MakeLens(
		func(target *url.URL) string { return target.RawPath },
		func(target *url.URL, rawPath string) *url.URL {
			clone := *target
			clone.RawPath = rawPath
			return &clone
		},
	)
	urlUserLens = L.MakeLens(
		func(target *url.URL) *url.Userinfo { return target.User },
		func(target *url.URL, user *url.Userinfo) *url.URL {
			clone := *target
			clone.User = user
			return &clone
		},
	)

	// hasURLMarker finds a query marker or a fragment marker in the trimmed
	// option text. This text check runs before the parse step, so an empty
	// query and an empty fragment also fail. It covers ForceQuery, RawQuery,
	// Fragment, and RawFragment with one rule.
	hasURLMarker  = F.Bind2nd(strings.ContainsAny, markerChars)
	isBaseURLText = F.Pipe1(S.IsNonEmpty, P.And(P.Not(hasURLMarker)))

	isWebScheme    = F.Pipe1(predord.StrEq(schemeHTTPS), P.Or(predord.StrEq(schemeHTTP)))
	isRootOnlyPath = F.Flow2(F.Bind2nd(strings.Trim, rootSlash), S.IsEmpty)
	// isOptionalPort accepts an absent port and a decimal port from 1
	// through 65535. net/url accepts any digit string as a port, so the
	// range check belongs here.
	isOptionalPort = F.Pipe1(S.IsEmpty, P.Or(F.Flow3(
		eitherparse.ParseInt,
		E.Map[error](predord.IntBetweenInclusive(minTCPPort, maxTCPPort)),
		E.GetOrElse(F.Constant1[error](false)),
	)))
	isAbsentUser = F.Flow2(O.FromNillable2[url.Userinfo], O.IsNone[url.Userinfo])

	// hasHostname also rejects a relative URL, a protocol-relative URL, and
	// an opaque URL, because each of those leaves the authority empty.
	hasWebScheme     = F.Pipe1(isWebScheme, P.ContraMap(urlSchemeLens.Get))
	hasHostname      = F.Pipe1(S.IsNonEmpty, P.ContraMap((*url.URL).Hostname))
	hasNoUserInfo    = F.Pipe1(isAbsentUser, P.ContraMap(urlUserLens.Get))
	hasNoEmptyPort   = F.Pipe1(P.Not(predstrings.HasSuffix(colon)), P.ContraMap(urlHostLens.Get))
	hasAllowedPort   = F.Pipe1(isOptionalPort, P.ContraMap((*url.URL).Port))
	hasRootOnlyPath  = F.Pipe1(isRootOnlyPath, P.ContraMap(urlPathLens.Get))
	hasNoEscapedPath = F.Pipe1(S.IsEmpty, P.ContraMap(urlRawPathLens.Get))

	isWebOrigin = F.Pipe6(
		hasWebScheme,
		P.And(hasHostname),
		P.And(hasNoUserInfo),
		P.And(hasNoEmptyPort),
		P.And(hasAllowedPort),
		P.And(hasRootOnlyPath),
		P.And(hasNoEscapedPath),
	)

	errBaseURLText = eithererr.ConstErr[string](
		"linkout-base-url must be a nonempty URL without a query or a fragment",
	)
	// errUnparsableURL drops the net/url message on purpose. That message
	// repeats the raw option value, which can hold user information.
	errUnparsableURL = eithererr.ConstErr[error](
		"linkout-base-url must be an absolute http or https URL",
	)
	errNotWebOrigin = eithererr.ConstErr[*url.URL](
		"linkout-base-url must be an http or https origin with a host, " +
			"an optional port from 1 through 65535, no user information, and no path",
	)

	parseBaseURL = F.Flow2(
		E.Eitherize1(url.Parse),
		E.MapLeft[*url.URL](errUnparsableURL),
	)
)

// Parse checks one website origin for the --linkout-base-url option. The
// last step clears the root-only path and prints the checked record, which
// yields "scheme://host" without a join helper. F.Flip then turns the lens
// setter into a one-argument constructor, so the file needs no seed helper.
func Parse(raw string) E.Either[error, Links] {
	return F.Pipe5(
		raw,
		strings.TrimSpace,
		E.FromPredicate(isBaseURLText, errBaseURLText),
		E.Chain(parseBaseURL),
		E.Chain(E.FromPredicate(isWebOrigin, errNotWebOrigin)),
		E.Map[error](F.Flow3(
			urlPathLens.Set(S.Monoid.Empty()),
			(*url.URL).String,
			F.Pipe1(Links{}, F.Flip(baseURLLens.Set)),
		)),
	)
}

// recordURL joins the origin, one fixed route, and one escaped ID. The zero
// value falls back to DefaultBaseURL inside the prefix step. An empty ID
// gives an empty URL, so a caller never emits a link to a record list.
// The join skips path.Join and url.JoinPath, which would clean the path and
// escape the ID a second time.
func (links Links) recordURL(route, id string) string {
	return F.Pipe1(
		id,
		strutils.MapNonEmpty(F.Flow2(
			url.PathEscape,
			S.Prepend(F.Pipe4(
				links,
				baseURLLens.Get,
				O.FromPredicate(S.IsNonEmpty),
				O.GetOrElse(F.Constant(DefaultBaseURL)),
				S.Append(route),
			)),
		)),
	)
}

// GeneURL returns the public gene page for id.
func (links Links) GeneURL(id string) string {
	return links.recordURL(genePath, id)
}

// StrainURL returns the public strain page for id.
func (links Links) StrainURL(id string) string {
	return links.recordURL(strainPath, id)
}

// PlasmidURL returns the public plasmid page for id.
func (links Links) PlasmidURL(id string) string {
	return links.recordURL(plasmidPath, id)
}

// PublicationURL returns the public publication page for id.
func (links Links) PublicationURL(id string) string {
	return links.recordURL(publicationPath, id)
}
