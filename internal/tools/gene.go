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
	// GeneInfoView is the gene record of find_stocks_by_gene plus its page url.
	GeneInfoView struct {
		graphql.GeneInfo
		URL string `json:"url"`
	}
	// GeneStrainView is one strain linked to a gene plus its stock page url.
	GeneStrainView struct {
		graphql.GeneStrainItem
		URL string `json:"url"`
	}

	// GeneStocks is the find_stocks_by_gene output.
	GeneStocks struct {
		Gene             GeneInfoView     `json:"gene"`
		Strains          []GeneStrainView `json:"strains"`
		Plasmids         []SearchHitView  `json:"plasmids"`
		PlasmidMatchNote string           `json:"plasmid_match_note"`
	}
)

const plasmidMatchNote = "Plasmids are text matches on the gene name, not curated gene links; " +
	"confirm with get_plasmid (genes field) before recommending one."

var (
	geneParamsCtxLens = L.MakeLens(
		func(p graphql.GeneParams) context.Context { return p.Ctx },
		func(p graphql.GeneParams, ctx context.Context) graphql.GeneParams {
			p.Ctx = ctx
			return p
		},
	)
	geneParamsGeneLens = L.MakeLens(
		func(p graphql.GeneParams) string { return p.Gene },
		func(p graphql.GeneParams, gene string) graphql.GeneParams {
			p.Gene = gene
			return p
		},
	)
	searchParamsCtxLens = L.MakeLens(
		func(p graphql.SearchParams) context.Context { return p.Ctx },
		func(p graphql.SearchParams, ctx context.Context) graphql.SearchParams {
			p.Ctx = ctx
			return p
		},
	)
	searchParamsQueryLens = L.MakeLens(
		func(p graphql.SearchParams) string { return p.Query },
		func(p graphql.SearchParams, query string) graphql.SearchParams {
			p.Query = query
			return p
		},
	)
	searchParamsLimitLens = L.MakeLens(
		func(p graphql.SearchParams) int { return p.Limit },
		func(p graphql.SearchParams, limit int) graphql.SearchParams {
			p.Limit = limit
			return p
		},
	)
	searchParamsEntityLens = L.MakeLens(
		func(p graphql.SearchParams) graphql.StockEntityType { return p.Entity },
		func(p graphql.SearchParams, entity graphql.StockEntityType) graphql.SearchParams {
			p.Entity = entity
			return p
		},
	)
	geneInfoIDLens = L.MakeLens(
		func(info graphql.GeneInfo) string { return info.ID },
		func(info graphql.GeneInfo, id string) graphql.GeneInfo {
			info.ID = id
			return info
		},
	)
	geneStrainViewRecordLens = L.MakeLens(
		func(v GeneStrainView) graphql.GeneStrainItem { return v.GeneStrainItem },
		func(v GeneStrainView, item graphql.GeneStrainItem) GeneStrainView {
			v.GeneStrainItem = item
			return v
		},
	)
	geneStrainViewURLLens = L.MakeLens(
		func(v GeneStrainView) string { return v.URL },
		func(v GeneStrainView, url string) GeneStrainView {
			v.URL = url
			return v
		},
	)
	geneStocksGeneViewLens = L.MakeLens(
		func(g GeneStocks) GeneInfoView { return g.Gene },
		func(g GeneStocks, view GeneInfoView) GeneStocks {
			g.Gene = view
			return g
		},
	)
	geneInfoViewRecordLens = L.MakeLens(
		func(v GeneInfoView) graphql.GeneInfo { return v.GeneInfo },
		func(v GeneInfoView, info graphql.GeneInfo) GeneInfoView {
			v.GeneInfo = info
			return v
		},
	)
	geneInfoViewURLLens = L.MakeLens(
		func(v GeneInfoView) string { return v.URL },
		func(v GeneInfoView, url string) GeneInfoView {
			v.URL = url
			return v
		},
	)
	geneStocksStrainViewsLens = L.MakeLens(
		func(g GeneStocks) []GeneStrainView { return g.Strains },
		func(g GeneStocks, items []GeneStrainView) GeneStocks {
			g.Strains = items
			return g
		},
	)
	geneStocksPlasmidViewsLens = L.MakeLens(
		func(g GeneStocks) []SearchHitView { return g.Plasmids },
		func(g GeneStocks, items []SearchHitView) GeneStocks {
			g.Plasmids = items
			return g
		},
	)
	geneStocksNoteLens = L.MakeLens(
		func(g GeneStocks) string { return g.PlasmidMatchNote },
		func(g GeneStocks, note string) GeneStocks {
			g.PlasmidMatchNote = note
			return g
		},
	)

	geneStocksInputGeneLens = F.Pipe1(
		geneStocksInputLens,
		L.Compose[geneStocksState](geneNameLens),
	)
	geneStocksGeneIDLens = F.Pipe1(
		geneStocksGeneLens,
		L.Compose[geneStocksState](geneInfoIDLens),
	)
)

func toGeneStrainView(links linkout.Links, item graphql.GeneStrainItem) GeneStrainView {
	return F.Pipe2(
		GeneStrainView{},
		geneStrainViewRecordLens.Set(item),
		geneStrainViewURLLens.Set(
			F.Pipe2(item, geneStrainItemIDLens.Get, links.StrainURL),
		),
	)
}

func lookupGene(s geneStocksState) IOE.IOEither[error, graphql.GeneInfo] {
	return F.Pipe3(
		graphql.GeneParams{},
		geneParamsCtxLens.Set(geneStocksCtxLens.Get(s)),
		geneParamsGeneLens.Set(geneStocksInputGeneLens.Get(s)),
		geneStocksClientLens.Get(s).GetGeneInfo,
	)
}

func strainsForGene(s geneStocksState) IOE.IOEither[error, []graphql.GeneStrainItem] {
	return F.Pipe3(
		graphql.GeneParams{},
		geneParamsCtxLens.Set(geneStocksCtxLens.Get(s)),
		geneParamsGeneLens.Set(geneStocksGeneIDLens.Get(s)),
		geneStocksClientLens.Get(s).ListStrainsWithGene,
	)
}

func plasmidsForGene(s geneStocksState) IOE.IOEither[error, graphql.SearchResult] {
	return F.Pipe5(
		graphql.SearchParams{},
		searchParamsCtxLens.Set(geneStocksCtxLens.Get(s)),
		searchParamsQueryLens.Set(geneStocksInputGeneLens.Get(s)),
		searchParamsLimitLens.Set(genePlasmidLimit),
		searchParamsEntityLens.Set(graphql.StockEntityPlasmid),
		geneStocksClientLens.Get(s).SearchStocks,
	)
}

func toGeneStocks(s geneStocksState) GeneStocks {
	links := geneStocksLinksLens.Get(s)
	gene := geneStocksGeneLens.Get(s)

	return F.Pipe4(
		GeneStocks{},
		geneStocksGeneViewLens.Set(F.Pipe2(
			GeneInfoView{},
			geneInfoViewRecordLens.Set(gene),
			geneInfoViewURLLens.Set(F.Pipe2(gene, geneInfoIDLens.Get, links.GeneURL)),
		)),
		geneStocksStrainViewsLens.Set(F.Pipe2(
			s,
			geneStocksStrainsLens.Get,
			A.Map(F.Bind1st(toGeneStrainView, links)),
		)),
		geneStocksPlasmidViewsLens.Set(F.Pipe3(
			s,
			geneStocksPlasmidsLens.Get,
			F.Bind1st(toSearchResultView, links),
			searchResultViewResultsLens.Get,
		)),
		geneStocksNoteLens.Set(plasmidMatchNote),
	)
}

func newFindStocksByGeneHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[FindStocksByGeneInput, GeneStocks] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input FindStocksByGeneInput,
	) (*mcp.CallToolResult, GeneStocks, error) {
		out := F.Pipe8(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidateFindStocksByGeneInput),
			IOE.Map[error](F.Flow4(
				F.Pipe1(geneStocksState{}, F.Flip(geneStocksInputLens.Set)),
				geneStocksCtxLens.Set(ctx),
				geneStocksClientLens.Set(client),
				geneStocksLinksLens.Set(links),
			)),
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
