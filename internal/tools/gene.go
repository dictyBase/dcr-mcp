package tools

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const plasmidMatchNote = "Plasmids are text matches on the gene name, not curated gene links; " +
	"confirm with get_plasmid (genes field) before recommending one."

// GeneStrainView is one strain linked to a gene plus its stock page url.
type GeneStrainView struct {
	graphql.GeneStrainItem
	URL string `json:"url"`
}

// GeneStocks is the find_stocks_by_gene output.
type GeneStocks struct {
	Gene             graphql.GeneInfo `json:"gene"`
	Strains          []GeneStrainView `json:"strains"`
	Plasmids         []SearchHitView  `json:"plasmids"`
	PlasmidMatchNote string           `json:"plasmid_match_note"`
}

var toGeneStocksState = F.Curry3(
	func(
		ctx context.Context,
		client *graphql.Client,
		in FindStocksByGeneInput,
	) geneStocksState {
		return geneStocksState{
			ctx:    ctx,
			client: client,
			input:  in,
		}
	},
)

func toGeneStrainView(item graphql.GeneStrainItem) GeneStrainView {
	return GeneStrainView{GeneStrainItem: item, URL: StrainURL(item.ID)}
}

func lookupGene(s geneStocksState) IOE.IOEither[error, graphql.GeneInfo] {
	return s.client.GetGeneInfo(graphql.GeneParams{Ctx: s.ctx, Gene: s.input.Gene})
}

func strainsForGene(s geneStocksState) IOE.IOEither[error, []graphql.GeneStrainItem] {
	return s.client.ListStrainsWithGene(graphql.GeneParams{Ctx: s.ctx, Gene: s.gene.ID})
}

func plasmidsForGene(s geneStocksState) IOE.IOEither[error, graphql.SearchResult] {
	return s.client.SearchStocks(graphql.SearchParams{
		Ctx:    s.ctx,
		Query:  s.input.Gene,
		Limit:  genePlasmidLimit,
		Entity: graphql.StockEntityPlasmid,
	})
}

func toGeneStocks(s geneStocksState) GeneStocks {
	return GeneStocks{
		Gene:             s.gene,
		Strains:          F.Pipe1(s.strains, A.Map(toGeneStrainView)),
		Plasmids:         toSearchResultView(s.plasmids).Results,
		PlasmidMatchNote: plasmidMatchNote,
	}
}

func newFindStocksByGeneHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[FindStocksByGeneInput, GeneStocks] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input FindStocksByGeneInput,
	) (*mcp.CallToolResult, GeneStocks, error) {
		buildState := F.Pipe1(client, toGeneStocksState(ctx))
		out := F.Pipe8(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateFindStocksByGeneInput),
			IOE.Map[error](buildState),
			IOE.Bind(geneStocksGeneLens.Set, lookupGene),
			IOE.Bind(geneStocksStrainsLens.Set, strainsForGene),
			IOE.Bind(geneStocksPlasmidsLens.Set, plasmidsForGene),
			IOE.Map[error](toGeneStocks),
			ioeutils.ToEither[error, GeneStocks],
			E.Fold(
				toOutcomeError[GeneStocks],
				toOutcomeSuccess[GeneStocks],
			),
		)
		return out.Result, out.Output, out.Err
	}
}
