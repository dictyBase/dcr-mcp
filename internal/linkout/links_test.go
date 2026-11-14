package linkout_test

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/stretchr/testify/require"
)

const (
	exampleBase       = "https://catalog.example.org"
	customBase        = exampleBase + ":8443"
	defaultBase       = "https://dictybase.dev"
	testGeneID        = "DDB_G0283907"
	testStrainID      = "DBS0350840"
	testPlasmidID     = "DBP0391806"
	testPublicationID = "17177856"
	flagName          = "linkout-base-url"
)

func mustParse(t *testing.T, raw string) linkout.Links {
	t.Helper()
	links, err := E.Unwrap(linkout.Parse(raw))
	require.NoError(t, err)
	return links
}

func TestLinksRoutes(t *testing.T) {
	zero := linkout.Links{}
	require.Equal(t, defaultBase, linkout.DefaultBaseURL)
	require.Equal(t, defaultBase+"/gene/DDB_G0283907", zero.GeneURL(testGeneID))
	require.Equal(t, defaultBase+"/stockcenter/strains/DBS0350840", zero.StrainURL(testStrainID))
	require.Equal(t, defaultBase+"/stockcenter/plasmids/DBP0391806", zero.PlasmidURL(testPlasmidID))
	require.Equal(t, defaultBase+"/publication/17177856", zero.PublicationURL(testPublicationID))

	custom := mustParse(t, customBase)
	require.Equal(t, customBase+"/gene/DDB_G0283907", custom.GeneURL(testGeneID))
	require.Equal(t, customBase+"/stockcenter/strains/DBS0350840", custom.StrainURL(testStrainID))
	require.Equal(t, customBase+"/stockcenter/plasmids/DBP0391806", custom.PlasmidURL(testPlasmidID))
	require.Equal(t, customBase+"/publication/17177856", custom.PublicationURL(testPublicationID))

	alpha := mustParse(t, "https://alpha.example.org")
	beta := mustParse(t, "https://beta.example.org")
	require.Equal(t, "https://alpha.example.org/gene/DDB_G0283907", alpha.GeneURL(testGeneID))
	require.Equal(t, "https://beta.example.org/gene/DDB_G0283907", beta.GeneURL(testGeneID))
	require.Equal(t, defaultBase+"/gene/DDB_G0283907", zero.GeneURL(testGeneID))
}

func TestLinksIDs(t *testing.T) {
	zero := linkout.Links{}
	require.Equal(t, defaultBase+"/gene/a%2Fb%3Fc%23d%25", zero.GeneURL("a/b?c#d%"))
	require.Equal(t, defaultBase+"/stockcenter/strains/a%20b", zero.StrainURL("a b"))
	require.Equal(t, defaultBase+"/publication/dbs0350840", zero.PublicationURL("dbs0350840"))

	require.Empty(t, zero.GeneURL(""))
	require.Empty(t, zero.StrainURL(""))
	require.Empty(t, zero.PlasmidURL(""))
	require.Empty(t, zero.PublicationURL(""))

	custom := mustParse(t, customBase)
	require.Empty(t, custom.GeneURL(""))
	require.Empty(t, custom.StrainURL(""))
	require.Empty(t, custom.PlasmidURL(""))
	require.Empty(t, custom.PublicationURL(""))
}

func TestParseBaseURL(t *testing.T) {
	accepted := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain origin", raw: exampleBase, want: exampleBase},
		{name: "outer space and root slashes", raw: " " + exampleBase + "/// ", want: exampleBase},
		{name: "upper-case scheme and one slash", raw: "HTTP://localhost:8080/", want: "http://localhost:8080"},
		{name: "bracketed IPv6 host", raw: "http://[::1]:8080", want: "http://[::1]:8080"},
		{name: "custom port", raw: customBase, want: customBase},
	}
	for _, tt := range accepted {
		t.Run(tt.name, func(t *testing.T) {
			links := mustParse(t, tt.raw)
			require.Equal(t, tt.want+"/gene/DDB_G0283907", links.GeneURL(testGeneID))
		})
	}

	rejected := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "whitespace", raw: "   "},
		{name: "host only", raw: "catalog.example.org"},
		{name: "protocol relative", raw: "//catalog.example.org"},
		{name: "path only", raw: "/catalog"},
		{name: "ftp scheme", raw: "ftp://catalog.example.org"},
		{name: "javascript scheme", raw: "javascript:alert(1)"},
		{name: "empty authority", raw: "https:///"},
		{name: "path prefix", raw: exampleBase + "/prefix"},
		{name: "query", raw: exampleBase + "?x=1"},
		{name: "empty query", raw: exampleBase + "?"},
		{name: "query with slashes", raw: exampleBase + "?x=///"},
		{name: "fragment", raw: exampleBase + "#part"},
		{name: "empty fragment", raw: exampleBase + "#"},
		{name: "fragment with slashes", raw: exampleBase + "#///"},
		{name: "text port", raw: exampleBase + ":abc"},
		{name: "port zero", raw: exampleBase + ":0"},
		{name: "port over range", raw: exampleBase + ":65536"},
		{name: "terminal colon", raw: exampleBase + ":"},
		{name: "encoded root path", raw: exampleBase + "/%2F"},
		{name: "bad escape", raw: exampleBase + "/%zz"},
	}
	for _, tt := range rejected {
		t.Run(tt.name, func(t *testing.T) {
			_, err := E.Unwrap(linkout.Parse(tt.raw))
			require.Error(t, err)
			require.ErrorContains(t, err, flagName)
		})
	}

	t.Run("user information stays out of the error text", func(t *testing.T) {
		_, err := E.Unwrap(linkout.Parse("https://user:secret@catalog.example.org"))
		require.Error(t, err)
		require.ErrorContains(t, err, flagName)
		require.NotContains(t, err.Error(), "secret")
	})
}
