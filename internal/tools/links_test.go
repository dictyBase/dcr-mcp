package tools_test

import (
	"testing"

	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestLinks(t *testing.T) {
	require.Equal(t, "https://dictycr.org/strains/DBS0350840", tools.StrainURL("DBS0350840"))
	require.Equal(t, "https://dictycr.org/plasmids/DBP0391806", tools.PlasmidURL("DBP0391806"))
	require.Equal(t, "https://pubmed.ncbi.nlm.nih.gov/17177856/", tools.PubMedURL("17177856"))
	require.Equal(t, "https://doi.org/10.1111/j.1432-0436.2006.00086.x", tools.DOIURL("10.1111/j.1432-0436.2006.00086.x"))
	require.Empty(t, tools.DOIURL(""))

	require.Equal(t, "https://dictycr.org/strains/DBS1", tools.StockURL(graphql.StockEntityStrain, "DBS1"))
	require.Equal(t, "https://dictycr.org/plasmids/DBP1", tools.StockURL(graphql.StockEntityPlasmid, "DBP1"))
	require.Empty(t, tools.StockURL(graphql.StockEntityAll, "X"))
}

func TestIDPrefix(t *testing.T) {
	require.Equal(t, "DBS", tools.IDPrefix("dbs0350840"))
	require.Equal(t, "DBP", tools.IDPrefix("DBP0391806"))
	require.Equal(t, "AB", tools.IDPrefix("ab"))
	require.Empty(t, tools.IDPrefix(""))
}
