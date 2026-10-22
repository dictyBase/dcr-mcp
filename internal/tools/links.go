package tools

import (
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictybase/dcr-mcp/internal/graphql"
)

const (
	stockCenterBaseURL = "https://dictycr.org"
	pubmedBaseURL      = "https://pubmed.ncbi.nlm.nih.gov"
	doiBaseURL         = "https://doi.org"

	stockIDPrefixLen = 3
)

var (
	// StrainURL returns the public Dicty Stock Center page for a strain ID.
	StrainURL = S.Prepend(stockCenterBaseURL + "/strains/")
	// PlasmidURL returns the public Dicty Stock Center page for a plasmid ID.
	PlasmidURL = S.Prepend(stockCenterBaseURL + "/plasmids/")
	// PubMedURL returns the PubMed page for a dictyBase publication ID (a PubMed ID).
	PubMedURL = F.Flow2(S.Prepend(pubmedBaseURL+"/"), S.Append("/"))

	stockURLByEntity = map[graphql.StockEntityType]func(string) string{
		graphql.StockEntityStrain:  StrainURL,
		graphql.StockEntityPlasmid: PlasmidURL,
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
func StockURL(entity graphql.StockEntityType, id string) string {
	return F.Pipe3(
		stockURLByEntity,
		R.Lookup[func(string) string](entity),
		O.Map(func(toURL func(string) string) string { return toURL(id) }),
		O.GetOrElse(F.Constant("")),
	)
}

// IDPrefix returns the upper-cased first three characters of id ("DBS", "DBP", ...).
func IDPrefix(id string) string {
	return strings.ToUpper(id[:min(len(id), stockIDPrefixLen)])
}
