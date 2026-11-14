package tools

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	// AvailabilityItem is one row of check_availability; Error is set when the lookup failed.
	AvailabilityItem struct {
		ID      string `json:"id"`
		Entity  string `json:"entity,omitempty"`
		Label   string `json:"label,omitempty"`
		InStock bool   `json:"in_stock"`
		URL     string `json:"url,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	// AvailabilityReport is the check_availability output.
	AvailabilityReport struct {
		Items []AvailabilityItem `json:"items"`
	}

	stockLookup struct {
		links  linkout.Links
		ctx    context.Context
		client *graphql.Client
		id     string
	}

	availabilityRoute = func(stockLookup) IOE.IOEither[error, AvailabilityItem]
)

const (
	strainEntityName  = "strain"
	plasmidEntityName = "plasmid"
)

var (
	stockLookupCtxLens = L.MakeLens(
		func(l stockLookup) context.Context { return l.ctx },
		func(l stockLookup, ctx context.Context) stockLookup {
			l.ctx = ctx
			return l

		},
	)
	stockLookupClientLens = L.MakeLens(
		func(l stockLookup) *graphql.Client { return l.client },
		func(l stockLookup, client *graphql.Client) stockLookup {
			l.client = client
			return l
		},
	)
	stockLookupIDLens = L.MakeLens(
		func(l stockLookup) string { return l.id },
		func(l stockLookup, id string) stockLookup {
			l.id = id
			return l
		},
	)
	stockLookupLinksLens = L.MakeLens(
		func(l stockLookup) linkout.Links { return l.links },
		func(l stockLookup, links linkout.Links) stockLookup {
			l.links = links
			return l
		},
	)
	availabilityItemIDLens = L.MakeLens(
		func(item AvailabilityItem) string { return item.ID },
		func(item AvailabilityItem, id string) AvailabilityItem {
			item.ID = id
			return item
		},
	)
	availabilityItemEntityLens = L.MakeLens(
		func(item AvailabilityItem) string { return item.Entity },
		func(item AvailabilityItem, entity string) AvailabilityItem {
			item.Entity = entity
			return item
		},
	)
	availabilityItemLabelLens = L.MakeLens(
		func(item AvailabilityItem) string { return item.Label },
		func(item AvailabilityItem, label string) AvailabilityItem {
			item.Label = label
			return item
		},
	)
	availabilityItemInStockLens = L.MakeLens(
		func(item AvailabilityItem) bool { return item.InStock },
		func(item AvailabilityItem, inStock bool) AvailabilityItem {
			item.InStock = inStock
			return item
		},
	)
	availabilityItemURLLens = L.MakeLens(
		func(item AvailabilityItem) string { return item.URL },
		func(item AvailabilityItem, url string) AvailabilityItem {
			item.URL = url
			return item
		},
	)
	availabilityItemErrorLens = L.MakeLens(
		func(item AvailabilityItem) string { return item.Error },
		func(item AvailabilityItem, message string) AvailabilityItem {
			item.Error = message
			return item
		},
	)
	availabilityReportItemsLens = L.MakeLens(
		func(report AvailabilityReport) []AvailabilityItem { return report.Items },
		func(report AvailabilityReport, items []AvailabilityItem) AvailabilityReport {
			report.Items = items
			return report
		},
	)
	availabilityRoutes = map[string]availabilityRoute{
		strainIDPrefix:  strainAvailability,
		plasmidIDPrefix: plasmidAvailability,
	}
)

func toStrainAvailabilityItem(links linkout.Links, d graphql.StrainDetail) AvailabilityItem {
	id := strainDetailIDLens.Get(d)

	return F.Pipe5(
		AvailabilityItem{},
		availabilityItemIDLens.Set(id),
		availabilityItemEntityLens.Set(strainEntityName),
		availabilityItemLabelLens.Set(strainDetailLabelLens.Get(d)),
		availabilityItemInStockLens.Set(strainDetailInStockLens.Get(d)),
		availabilityItemURLLens.Set(links.StrainURL(id)),
	)
}

func toPlasmidAvailabilityItem(links linkout.Links, d graphql.PlasmidDetail) AvailabilityItem {
	id := plasmidDetailIDLens.Get(d)

	return F.Pipe5(
		AvailabilityItem{},
		availabilityItemIDLens.Set(id),
		availabilityItemEntityLens.Set(plasmidEntityName),
		availabilityItemLabelLens.Set(plasmidDetailNameLens.Get(d)),
		availabilityItemInStockLens.Set(plasmidDetailInStockLens.Get(d)),
		availabilityItemURLLens.Set(links.PlasmidURL(id)),
	)
}

func strainAvailability(l stockLookup) IOE.IOEither[error, AvailabilityItem] {
	return F.Pipe4(
		graphql.IDParams{},
		idParamsCtxLens.Set(stockLookupCtxLens.Get(l)),
		idParamsIDLens.Set(stockLookupIDLens.Get(l)),
		stockLookupClientLens.Get(l).GetStrain,
		IOE.Map[error](F.Bind1st(toStrainAvailabilityItem, stockLookupLinksLens.Get(l))),
	)
}

func plasmidAvailability(l stockLookup) IOE.IOEither[error, AvailabilityItem] {
	return F.Pipe4(
		graphql.IDParams{},
		idParamsCtxLens.Set(stockLookupCtxLens.Get(l)),
		idParamsIDLens.Set(stockLookupIDLens.Get(l)),
		stockLookupClientLens.Get(l).GetPlasmid,
		IOE.Map[error](F.Bind1st(toPlasmidAvailabilityItem, stockLookupLinksLens.Get(l))),
	)
}

// lookupAvailability never fails: lookup errors become an item with Error set.
func lookupAvailability(l stockLookup) IOE.IOEither[error, AvailabilityItem] {
	return F.Pipe4(
		l,
		IOE.Of[error, stockLookup],
		IOE.ChainEitherK(F.Bind1st(
			F.Pipe1(availabilityRoutes, F.Bind1of3(routeEffect[stockLookup, AvailabilityItem])),
			stockLookupIDLens.Get,
		)),
		IOE.Flatten[error, AvailabilityItem],
		IOE.OrElse(F.Flow4(
			error.Error,
			availabilityItemErrorLens.Set,
			F.Bind1st(
				F.Pipe1[func(AvailabilityItem) AvailabilityItem, AvailabilityItem, AvailabilityItem],
				F.Pipe1(AvailabilityItem{}, availabilityItemIDLens.Set(stockLookupIDLens.Get(l))),
			),
			IOE.Of[error, AvailabilityItem],
		)),
	)
}

func newCheckAvailabilityHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[CheckAvailabilityInput, AvailabilityReport] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input CheckAvailabilityInput,
	) (*mcp.CallToolResult, AvailabilityReport, error) {
		out := F.Pipe7(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateCheckAvailabilityInput),
			IOE.Map[error](availabilityIDsLens.Get),
			IOE.Map[error](A.Map(F.Flow4(
				F.Pipe1(stockLookup{}, F.Flip(stockLookupIDLens.Set)),
				stockLookupCtxLens.Set(ctx),
				stockLookupClientLens.Set(client),
				stockLookupLinksLens.Set(links),
			))),
			IOE.Chain(IOE.TraverseArrayPar(lookupAvailability)),
			IOE.Map[error](F.Pipe1(AvailabilityReport{}, F.Flip(availabilityReportItemsLens.Set))),
			ioeutils.ToEither[error, AvailabilityReport],
			E.Fold(
				toOutcomeError[AvailabilityReport],
				toOutcomeSuccess[AvailabilityReport],
			),
		)
		return out.Result, out.Output, out.Err
	}
}
