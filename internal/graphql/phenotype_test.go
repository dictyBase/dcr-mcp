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
	stopwordBody = `{"errors":[{"message":"expect a search query with at least one non-stopword",
		"path":["searchPhenotypes"],"extensions":{"code":"InvalidArgument"}}],"data":null}`
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
			SearchPhenotypes(graphql.PhenotypeSearchParams{Ctx: ctx, Query: "growth", Limit: 2})))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"query": "growth", "limit": float64(2)}, vars)
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
			SearchPhenotypes(graphql.PhenotypeSearchParams{Ctx: ctx, Query: "growth", Limit: 2})))
		require.Error(t, err)
		require.False(t, errors.Is(err, graphql.ErrInvalidPhenotypeQuery))
		require.Contains(t, err.Error(), "searchPhenotypes query failed")
	})
}
