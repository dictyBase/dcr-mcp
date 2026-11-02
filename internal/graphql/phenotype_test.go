package graphql_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/stretchr/testify/require"
)

const (
	phenoSearchBody = `{"data":{"searchPhenotypes":{"limit":2,"totalCount":1,"results":[
		{"id":"DBS0235587","entity":"STRAIN","field":"TERM_SYNONYM",
		 "display_text":"abolished vegetative growth","score":503.4,"summary":"chcA null",
		 "stock_label":"chcA-","in_stock":true,"stock_type":"gwdi"}]}}}`
	phenoGrowthWord = "growth"
	// #nosec G101 -- test fixture text; "stopword" matches the credential pattern
	stopwordBody = `{"errors":[{"message":"expect a search query with at least one non-stopword",
		"path":["searchPhenotypes"],"extensions":{"code":"InvalidArgument"}}],"data":null}`
	notFoundBody = `{"errors":[{"message":"annotation list not found",
		"path":["listStrainsWithAnnotation"],"extensions":{"code":"NotFound"}}],
		"data":{"listStrainsWithAnnotation":null}}`
	internalBody = `{"errors":[{"message":"syntax error, unexpected RETURN",
		"path":["listStrainsWithAnnotation"],"extensions":{"code":"Internal"}}],
		"data":{"listStrainsWithAnnotation":null}}`
	duplicatePageBody = `{"data":{"listStrainsWithAnnotation":{"nextCursor":0,"strains":[
		{"id":"DBS1","label":"a-","summary":"","in_stock":true,"genes":[]},
		{"id":"DBS2","label":"b-","summary":"","in_stock":true,"genes":[]},
		{"id":"DBS1","label":"a-","summary":"","in_stock":true,"genes":[]}]}}}`
	emptyPublicationsBody = `{"data":{"strain":{"publications":[]}}}`
)

// newFixedServer answers every request with body and records the variables
// of the first request.
func newFixedServer(t *testing.T, body string, vars *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		if vars != nil && *vars == nil {
			*vars = req.Variables
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func TestSearchPhenotypes(t *testing.T) {
	ctx := context.Background()

	t.Run("sends query and limit only and decodes rows", func(t *testing.T) {
		var vars map[string]any
		srv := newFixedServer(t, phenoSearchBody, &vars)
		defer srv.Close()
		res, err := E.Unwrap(ioeutils.ToEither(graphql.NewClient(srv.URL, srv.Client()).
			SearchPhenotypes(graphql.PhenotypeSearchParams{Ctx: ctx, Query: phenoGrowthWord, Limit: 2})))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"query": phenoGrowthWord, "limit": float64(2)}, vars)
		require.Len(t, res.Results, 1)
		require.Equal(t, "chcA-", res.Results[0].StockLabel)
		require.Equal(t, "gwdi", res.Results[0].StockType)
	})

	t.Run("maps InvalidArgument to ErrInvalidPhenotypeQuery", func(t *testing.T) {
		srv := newFixedServer(t, stopwordBody, nil)
		defer srv.Close()
		_, err := E.Unwrap(ioeutils.ToEither(graphql.NewClient(srv.URL, srv.Client()).
			SearchPhenotypes(graphql.PhenotypeSearchParams{Ctx: ctx, Query: "the", Limit: 2})))
		require.True(t, errors.Is(err, graphql.ErrInvalidPhenotypeQuery))
		require.Contains(t, err.Error(), "non-stopword")
	})

	t.Run("wraps a transport error", func(t *testing.T) {
		srv := newFixedServer(t, phenoSearchBody, nil)
		srv.Close()
		_, err := E.Unwrap(ioeutils.ToEither(graphql.NewClient(srv.URL, srv.Client()).
			SearchPhenotypes(graphql.PhenotypeSearchParams{Ctx: ctx, Query: phenoGrowthWord, Limit: 2})))
		require.Error(t, err)
		require.False(t, errors.Is(err, graphql.ErrInvalidPhenotypeQuery))
		require.Contains(t, err.Error(), "searchPhenotypes query failed")
	})
}

func TestListStrainsWithPhenotypeTermContract(t *testing.T) {
	ctx := context.Background()
	params := graphql.PhenotypeListParams{Ctx: ctx, PhenotypeTerm: "in suspension", Limit: 10}

	for name, body := range map[string]string{"NotFound": notFoundBody, "Internal": internalBody} {
		t.Run(name+" maps to ErrNotPhenotypeTerm", func(t *testing.T) {
			srv := newFixedServer(t, body, nil)
			defer srv.Close()
			_, err := E.Unwrap(ioeutils.ToEither(
				graphql.NewClient(srv.URL, srv.Client()).ListStrainsWithPhenotype(params)))
			require.True(t, errors.Is(err, graphql.ErrNotPhenotypeTerm))
			require.Contains(t, err.Error(), `"in suspension" is not an exact phenotype term`)
		})
	}

	t.Run("removes duplicate strains and keeps order", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req gqlRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			w.Header().Set("Content-Type", "application/json")
			body := map[bool]string{true: duplicatePageBody, false: emptyPublicationsBody}
			_, _ = w.Write([]byte(body[req.Variables["annotation"] != nil]))
		}))
		defer srv.Close()
		res, err := E.Unwrap(ioeutils.ToEither(
			graphql.NewClient(srv.URL, srv.Client()).ListStrainsWithPhenotype(
				graphql.PhenotypeListParams{Ctx: ctx, PhenotypeTerm: "abolished vegetative growth", Limit: 10})))
		require.NoError(t, err)
		require.Len(t, res.Strains, 2)
		require.Equal(t, "DBS1", res.Strains[0].ID)
		require.Equal(t, "DBS2", res.Strains[1].ID)
		require.Zero(t, res.NextCursor)
	})
}
