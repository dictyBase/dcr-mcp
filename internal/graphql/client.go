package graphql

import (
	"fmt"
	"net/http"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	hasuragql "github.com/hasura/go-graphql-client"
)

const (
	varQuery      = "query"
	varLimit      = "limit"
	varEntity     = "entity"
	varID         = "id"
	varCursor     = "cursor"
	varType       = "type"
	varAnnotation = "annotation"
	phenotypeType = "phenotype"
)

// Client wraps a hasura GraphQL client with functional IOEither query methods.
type Client struct {
	gql *hasuragql.Client
}

func resolveHTTPClient(httpClient *http.Client) *http.Client {
	return F.Pipe2(
		httpClient,
		O.FromPredicate(P.IsNonZero[*http.Client]()),
		O.GetOrElse(F.Constant(http.DefaultClient)),
	)
}

// NewClient constructs a GraphQL Client targeting endpoint.
func NewClient(endpoint string, httpClient *http.Client) *Client {
	resolved := resolveHTTPClient(httpClient)
	return &Client{
		gql: hasuragql.NewClient(endpoint, resolved),
	}
}

// ListStockSuggestions executes the listStockSuggestions GraphQL query.
func (c *Client) ListStockSuggestions(
	params AutocompleteParams,
) IOE.IOEither[error, AutocompleteResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*listStockSuggestionsQuery, error) {
			query := new(listStockSuggestionsQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{
				varQuery:  params.Query,
				varLimit:  params.Limit,
				varEntity: params.Entity,
			})
		}),
		IOE.MapLeft[*listStockSuggestionsQuery](func(err error) error {
			return fmt.Errorf("listStockSuggestions query failed: %w", err)
		}),
		IOE.Map[error](toAutocompleteResult),
	)
}

// SearchStocks executes the searchStocks GraphQL query.
func (c *Client) SearchStocks(
	params SearchParams,
) IOE.IOEither[error, SearchResult] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*searchStocksQuery, error) {
			query := new(searchStocksQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{
				varQuery:  params.Query,
				varLimit:  params.Limit,
				varEntity: params.Entity,
			})
		}),
		IOE.MapLeft[*searchStocksQuery](func(err error) error {
			return fmt.Errorf("searchStocks query failed: %w", err)
		}),
		IOE.Map[error](toSearchResult),
	)
}

// GetStrain executes the strain(id: $id) GraphQL query.
func (c *Client) GetStrain(
	params IDParams,
) IOE.IOEither[error, StrainDetail] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*getStrainQuery, error) {
			query := new(getStrainQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{
				varID: hasuragql.ID(params.ID),
			})
		}),
		IOE.MapLeft[*getStrainQuery](func(err error) error {
			return fmt.Errorf("strain query for %q failed: %w", params.ID, err)
		}),
		IOE.Map[error](toStrainDetail),
	)
}

// GetPlasmid executes the plasmid(id: $id) GraphQL query.
func (c *Client) GetPlasmid(
	params IDParams,
) IOE.IOEither[error, PlasmidDetail] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*getPlasmidQuery, error) {
			query := new(getPlasmidQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{
				varID: hasuragql.ID(params.ID),
			})
		}),
		IOE.MapLeft[*getPlasmidQuery](func(err error) error {
			return fmt.Errorf("plasmid query for %q failed: %w", params.ID, err)
		}),
		IOE.Map[error](toPlasmidDetail),
	)
}

// ListStrainsWithPhenotype executes listStrainsWithAnnotation with type "phenotype".
func (c *Client) ListStrainsWithPhenotype(
	params PhenotypeListParams,
) IOE.IOEither[error, PhenotypeStrainList] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*listStrainsWithPhenotypeQuery, error) {
			query := new(listStrainsWithPhenotypeQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{
				varCursor:     params.Cursor,
				varLimit:      params.Limit,
				varType:       phenotypeType,
				varAnnotation: params.Phenotype,
			})
		}),
		IOE.MapLeft[*listStrainsWithPhenotypeQuery](func(err error) error {
			return fmt.Errorf(
				"listStrainsWithAnnotation query for phenotype %q failed: %w",
				params.Phenotype,
				err,
			)
		}),
		IOE.Map[error](toPhenotypeStrainList),
	)
}
