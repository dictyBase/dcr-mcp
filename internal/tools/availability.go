package tools

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AvailabilityItem is one row of check_availability; Error is set when the lookup failed.
type AvailabilityItem struct {
	ID      string `json:"id"`
	Entity  string `json:"entity,omitempty"`
	Label   string `json:"label,omitempty"`
	InStock bool   `json:"in_stock"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AvailabilityReport is the check_availability output.
type AvailabilityReport struct {
	Items []AvailabilityItem `json:"items"`
}

type stockLookup struct {
	ctx    context.Context
	client *graphql.Client
	id     string
}

type availabilityRoute func(stockLookup) IOE.IOEither[error, AvailabilityItem]

var (
	availabilityRoutes = map[string]availabilityRoute{
		strainIDPrefix:  strainAvailability,
		plasmidIDPrefix: plasmidAvailability,
	}
	toAvailabilityState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in CheckAvailabilityInput,
		) availabilityState {
			return availabilityState{
				ctx:    ctx,
				client: client,
				input:  in,
			}
		},
	)
)

func toStrainAvailabilityItem(d graphql.StrainDetail) AvailabilityItem {
	return AvailabilityItem{
		ID:      d.ID,
		Entity:  "strain",
		Label:   d.Label,
		InStock: d.InStock,
		URL:     StrainURL(d.ID),
	}
}

func toPlasmidAvailabilityItem(d graphql.PlasmidDetail) AvailabilityItem {
	return AvailabilityItem{
		ID:      d.ID,
		Entity:  "plasmid",
		Label:   d.Name,
		InStock: d.InStock,
		URL:     PlasmidURL(d.ID),
	}
}

func strainAvailability(l stockLookup) IOE.IOEither[error, AvailabilityItem] {
	return F.Pipe2(
		graphql.IDParams{Ctx: l.ctx, ID: l.id},
		l.client.GetStrain,
		IOE.Map[error](toStrainAvailabilityItem),
	)
}

func plasmidAvailability(l stockLookup) IOE.IOEither[error, AvailabilityItem] {
	return F.Pipe2(
		graphql.IDParams{Ctx: l.ctx, ID: l.id},
		l.client.GetPlasmid,
		IOE.Map[error](toPlasmidAvailabilityItem),
	)
}

// lookupAvailability never fails: lookup errors become an item with Error set.
func lookupAvailability(l stockLookup) IOE.IOEither[error, AvailabilityItem] {
	return F.Pipe3(
		availabilityRoutes,
		R.Lookup[availabilityRoute](IDPrefix(l.id)),
		O.Fold(
			func() IOE.IOEither[error, AvailabilityItem] {
				return IOE.Left[AvailabilityItem](unknownIDError(l.id))
			},
			func(route availabilityRoute) IOE.IOEither[error, AvailabilityItem] { return route(l) },
		),
		IOE.OrElse(func(err error) IOE.IOEither[error, AvailabilityItem] {
			return IOE.Of[error](AvailabilityItem{ID: l.id, Error: err.Error()})
		}),
	)
}

func toAvailabilityReport(items []AvailabilityItem) AvailabilityReport {
	return AvailabilityReport{Items: items}
}

func checkAvailability(s availabilityState) IOE.IOEither[error, AvailabilityReport] {
	return F.Pipe2(
		s.input.IDs,
		IOE.TraverseArrayPar(func(id string) IOE.IOEither[error, AvailabilityItem] {
			return lookupAvailability(stockLookup{ctx: s.ctx, client: s.client, id: id})
		}),
		IOE.Map[error](toAvailabilityReport),
	)
}

func newCheckAvailabilityHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[CheckAvailabilityInput, AvailabilityReport] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input CheckAvailabilityInput,
	) (*mcp.CallToolResult, AvailabilityReport, error) {
		buildState := F.Pipe1(client, toAvailabilityState(ctx))
		out := F.Pipe5(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateCheckAvailabilityInput),
			IOE.Map[error](buildState),
			IOE.Chain(checkAvailability),
			ioeutils.ToEither[error, AvailabilityReport],
			E.Fold(
				toOutcomeError[AvailabilityReport],
				toOutcomeSuccess[AvailabilityReport],
			),
		)
		return out.Result, out.Output, out.Err
	}
}
