package graphql

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	EQ "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/hashicorp/golang-lru/v2/expirable"
	hasuragql "github.com/hasura/go-graphql-client"
)

// Client wraps a hasura GraphQL client with functional IOEither query methods
// and a short-lived cache for single-record lookups.
type Client struct {
	gql          *hasuragql.Client
	strainCache  *expirable.LRU[string, StrainDetail]
	plasmidCache *expirable.LRU[string, PlasmidDetail]
}

const (
	varQuery         = "query"
	varLimit         = "limit"
	varEntity        = "entity"
	varID            = "id"
	varCursor        = "cursor"
	varType          = "type"
	varAnnotation    = "annotation"
	phenotypeType    = "phenotype"
	varGene          = "gene"
	notFoundCode     = "NotFound"
	defaultCacheSize = 1024
	defaultCacheTTL  = 10 * time.Minute
)

var (
	hasNotFoundCode = F.Pipe2(
		notFoundCode,
		EQ.Equals(predord.StringEq),
		P.ContraMap(gqlErrorCode),
	)

	// isNotFound reports whether err carries extensions.code == "NotFound".
	// dictyBase uses it for "no strains for this gene", an empty result, not a failure.
	isNotFound = F.Flow2(gqlErrors, A.Any(hasNotFoundCode))

	// hasCode matches one GraphQL error whose extensions.code equals code.
	hasCode = func(code string) P.Predicate[hasuragql.Error] {
		return F.Pipe2(code, EQ.Equals(predord.StringEq), P.ContraMap(gqlErrorCode))
	}

	// isClientErrorEntry reports whether one GraphQL error was minted by the
	// GraphQL client itself (request, encode, or decode failure) instead of the
	// upstream resolver.
	isClientErrorEntry = F.Pipe5(
		hasCode(hasuragql.ErrRequestError),
		P.Or(hasCode(hasuragql.ErrJsonEncode)),
		P.Or(hasCode(hasuragql.ErrJsonDecode)),
		P.Or(hasCode(hasuragql.ErrGraphQLEncode)),
		P.Or(hasCode(hasuragql.ErrGraphQLDecode)),
		P.Or(hasCode(hasuragql.ErrGraphQLExtensionsDecode)),
	)

	// hasDataError reports whether err carries an upstream GraphQL data error,
	// the class one bad PMID produces. Only that class degrades to an empty
	// publication list; client-side transport and decode errors propagate, never
	// degrade silently.
	hasDataError = F.Flow2(gqlErrors, A.Any(P.Not(isClientErrorEntry)))

	geneFound = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(func(g GeneInfo) string { return g.ID }),
	)

	publicationFound = F.Pipe1(
		S.IsNonEmpty,
		P.ContraMap(func(p PublicationDetail) string { return p.ID }),
	)
)

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
		gql:          hasuragql.NewClient(endpoint, resolved),
		strainCache:  expirable.NewLRU[string, StrainDetail](defaultCacheSize, nil, defaultCacheTTL),
		plasmidCache: expirable.NewLRU[string, PlasmidDetail](defaultCacheSize, nil, defaultCacheTTL),
	}
}

func storeIn[V any](cache *expirable.LRU[string, V], key string) func(V) IO.IO[struct{}] {
	return func(v V) IO.IO[struct{}] {
		return func() struct{} {
			cache.Add(key, v)
			return struct{}{}
		}
	}
}

// cached serves key from cache, or runs load and stores a successful result.
// IOE.Defer postpones the cache lookup until the IOEither runs.
func cached[V any](
	cache *expirable.LRU[string, V],
	key string,
	load IOE.IOEither[error, V],
) IOE.IOEither[error, V] {
	return IOE.Defer(func() IOE.IOEither[error, V] {
		return F.Pipe2(
			key,
			O.FromValidation(cache.Get),
			O.Fold(
				func() IOE.IOEither[error, V] {
					return F.Pipe1(load, IOE.ChainFirstIOK[error](storeIn(cache, key)))
				},
				IOE.Of[error, V],
			),
		)
	})
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

func (c *Client) getStrainUncached(
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

// GetStrain executes the strain(id: $id) GraphQL query, served from cache when fresh.
func (c *Client) GetStrain(params IDParams) IOE.IOEither[error, StrainDetail] {
	return cached(c.strainCache, params.ID, c.getStrainUncached(params))
}

func (c *Client) getPlasmidUncached(
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

// GetPlasmid executes the plasmid(id: $id) GraphQL query, served from cache when fresh.
func (c *Client) GetPlasmid(params IDParams) IOE.IOEither[error, PlasmidDetail] {
	return cached(c.plasmidCache, params.ID, c.getPlasmidUncached(params))
}

// ListStrainsWithPhenotype executes listStrainsWithAnnotation with type "phenotype".
// The list query carries no publications field: one legacy non-numeric PMID
// would fail the whole list. The pipeline fetches publications per strain in
// parallel and degrades only upstream GraphQL data errors (the bad-PMID class)
// to an empty publication list; transport and decode errors propagate.
func (c *Client) ListStrainsWithPhenotype(
	params PhenotypeListParams,
) IOE.IOEither[error, PhenotypeStrainList] {
	return F.Pipe3(
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
		IOE.Chain(func(list PhenotypeStrainList) IOE.IOEither[error, PhenotypeStrainList] {
			return F.Pipe2(
				list.Strains,
				IOE.TraverseArrayPar(
					func(item PhenotypeStrainItem) IOE.IOEither[error, []PhenotypePublicationRef] {
						return F.Pipe3(
							IOE.TryCatchError(func() (*StrainPublicationQuery, error) {
								query := new(StrainPublicationQuery)
								return query, c.gql.Query(params.Ctx, query, map[string]any{
									varID: hasuragql.ID(item.ID),
								})
							}),
							IOE.Map[error](toStrainPublications),
							IOE.OrElse(func(err error) IOE.IOEither[error, []PhenotypePublicationRef] {
								return F.Pipe2(
									err,
									O.FromPredicate(hasDataError),
									O.Fold(
										func() IOE.IOEither[error, []PhenotypePublicationRef] {
											return IOE.Left[[]PhenotypePublicationRef](err)
										},
										func(error) IOE.IOEither[error, []PhenotypePublicationRef] {
											return F.Pipe1(
												emptySlice[PhenotypePublicationRef](),
												IOE.Of[error, []PhenotypePublicationRef],
											)
										},
									),
								)
							}),
							IOE.MapLeft[[]PhenotypePublicationRef](func(err error) error {
								return fmt.Errorf("publication query for strain %q failed: %w", item.ID, err)
							}),
						)
					},
				),
				IOE.Map[error](func(pubLists [][]PhenotypePublicationRef) PhenotypeStrainList {
					return PhenotypeStrainList{
						Strains: A.ZipWith(
							list.Strains,
							pubLists,
							func(item PhenotypeStrainItem, pubs []PhenotypePublicationRef) PhenotypeStrainItem {
								item.Publications = pubs
								return item
							},
						),
						NextCursor: list.NextCursor,
						TotalCount: list.TotalCount,
					}
				}),
			)
		}),
	)
}

func gqlErrorCode(e hasuragql.Error) string {
	code, _ := e.Extensions["code"].(string)
	return code
}

func gqlErrors(err error) []hasuragql.Error {
	var errs hasuragql.Errors
	_ = errors.As(err, &errs)
	return errs
}

func recoverNotFound[T any](empty func() T) func(error) IOE.IOEither[error, T] {
	return func(err error) IOE.IOEither[error, T] {
		return F.Pipe2(
			err,
			O.FromPredicate(isNotFound),
			O.Fold(
				func() IOE.IOEither[error, T] { return IOE.Left[T](err) },
				func(error) IOE.IOEither[error, T] { return IOE.Of[error](empty()) },
			),
		)
	}
}

// GetGeneInfo resolves a gene name or DDB_G id via geneGeneralInformation.
func (c *Client) GetGeneInfo(params GeneParams) IOE.IOEither[error, GeneInfo] {
	notFound := func(GeneInfo) error {
		return fmt.Errorf("gene %q not found in dictyBase", params.Gene)
	}
	return F.Pipe3(
		IOE.TryCatchError(func() (*geneInfoQuery, error) {
			query := new(geneInfoQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{varGene: params.Gene})
		}),
		IOE.MapLeft[*geneInfoQuery](func(err error) error {
			return fmt.Errorf("geneGeneralInformation query for %q failed: %w", params.Gene, err)
		}),
		IOE.Map[error](toGeneInfo),
		IOE.ChainEitherK(E.FromPredicate(geneFound, notFound)),
	)
}

// ListStrainsWithGene lists strains for a DDB_G gene id; NotFound yields an empty list.
func (c *Client) ListStrainsWithGene(params GeneParams) IOE.IOEither[error, []GeneStrainItem] {
	return F.Pipe3(
		IOE.TryCatchError(func() (*strainsWithGeneQuery, error) {
			query := new(strainsWithGeneQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{varGene: params.Gene})
		}),
		IOE.Map[error](toGeneStrainItems),
		IOE.OrElse(recoverNotFound(emptySlice[GeneStrainItem])),
		IOE.MapLeft[[]GeneStrainItem](func(err error) error {
			return fmt.Errorf("listStrainsWithGene query for %q failed: %w", params.Gene, err)
		}),
	)
}

// GetPublication fetches one publication by its PubMed id.
func (c *Client) GetPublication(params IDParams) IOE.IOEither[error, PublicationDetail] {
	notFound := func(PublicationDetail) error {
		return fmt.Errorf("publication %q not found", params.ID)
	}
	return F.Pipe3(
		IOE.TryCatchError(func() (*publicationQuery, error) {
			query := new(publicationQuery)
			return query, c.gql.Query(params.Ctx, query, map[string]any{
				varID: hasuragql.ID(params.ID),
			})
		}),
		IOE.MapLeft[*publicationQuery](func(err error) error {
			return fmt.Errorf("publication query for %q failed: %w", params.ID, err)
		}),
		IOE.Map[error](toPublicationDetail),
		IOE.ChainEitherK(E.FromPredicate(publicationFound, notFound)),
	)
}
