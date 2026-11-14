package tools

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/fp-go-loom/strutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/linkout"
	"github.com/dictybase/dcr-mcp/internal/mdfp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	md "github.com/nao1215/markdown"
)

// PublicationView is the get_publication output.
type PublicationView struct {
	graphql.PublicationDetail
	URL       string `json:"url"`
	PubMedURL string `json:"pubmed_url"`
	DOIURL    string `json:"doi_url,omitempty"`
}

const abstractHeader = "Abstract"

var (
	publicationViewRecordLens = L.MakeLens(
		func(v PublicationView) graphql.PublicationDetail { return v.PublicationDetail },
		func(v PublicationView, p graphql.PublicationDetail) PublicationView {
			v.PublicationDetail = p
			return v
		},
	)
	publicationViewURLLens = L.MakeLens(
		func(v PublicationView) string { return v.URL },
		func(v PublicationView, url string) PublicationView {
			v.URL = url
			return v
		},
	)
	publicationViewPubMedLens = L.MakeLens(
		func(v PublicationView) string { return v.PubMedURL },
		func(v PublicationView, url string) PublicationView {
			v.PubMedURL = url
			return v
		},
	)
	publicationViewDOILinkLens = L.MakeLens(
		func(v PublicationView) string { return v.DOIURL },
		func(v PublicationView, url string) PublicationView {
			v.DOIURL = url
			return v
		},
	)
	publicationDetailIDLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.ID },
		func(p graphql.PublicationDetail, id string) graphql.PublicationDetail {
			p.ID = id
			return p
		},
	)
	publicationDetailDOILens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.DOI },
		func(p graphql.PublicationDetail, doi string) graphql.PublicationDetail {
			p.DOI = doi
			return p
		},
	)
	publicationDetailJournalLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.Journal },
		func(p graphql.PublicationDetail, journal string) graphql.PublicationDetail {
			p.Journal = journal
			return p
		},
	)
	publicationDetailVolumeLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.Volume },
		func(p graphql.PublicationDetail, volume string) graphql.PublicationDetail {
			p.Volume = volume
			return p
		},
	)
	publicationDetailPagesLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.Pages },
		func(p graphql.PublicationDetail, pages string) graphql.PublicationDetail {
			p.Pages = pages
			return p
		},
	)
	publicationDetailPubDateLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.PubDate },
		func(p graphql.PublicationDetail, pubDate string) graphql.PublicationDetail {
			p.PubDate = pubDate
			return p
		},
	)
	publicationDetailTitleLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.Title },
		func(p graphql.PublicationDetail, title string) graphql.PublicationDetail {
			p.Title = title
			return p
		},
	)
	publicationDetailAbstractLens = L.MakeLens(
		func(p graphql.PublicationDetail) string { return p.Abstract },
		func(p graphql.PublicationDetail, abstract string) graphql.PublicationDetail {
			p.Abstract = abstract
			return p
		},
	)
	publicationDetailAuthorsLens = L.MakeLens(
		func(p graphql.PublicationDetail) []graphql.FullAuthor { return p.Authors },
		func(p graphql.PublicationDetail, authors []graphql.FullAuthor) graphql.PublicationDetail {
			p.Authors = authors
			return p
		},
	)

	publicationViewIDLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailIDLens),
	)
	publicationViewDOILens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailDOILens),
	)
	publicationViewJournalLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailJournalLens),
	)
	publicationViewVolumeLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailVolumeLens),
	)
	publicationViewPagesLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailPagesLens),
	)
	publicationViewPubDateLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailPubDateLens),
	)
	publicationViewTitleLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailTitleLens),
	)
	publicationViewAbstractLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailAbstractLens),
	)
	publicationViewAuthorsLens = F.Pipe1(
		publicationViewRecordLens,
		L.Compose[PublicationView](publicationDetailAuthorsLens),
	)
	// fullAuthorName is "LastName Initials". An empty part drops out with its
	// space.
	fullAuthorName = F.Pipe1(
		[]func(graphql.FullAuthor) string{
			F.Flow2(fullAuthorLastName, oneLine),
			F.Flow2(fullAuthorInitials, oneLine),
		},
		strutils.ConcatFields[graphql.FullAuthor](strutils.Space),
	)

	fullAuthorNames = F.Pipe1(fullAuthorName, A.FoldMap[graphql.FullAuthor](strutils.Comma))

	renderPublicationMarkdown = F.Flow2(publicationDoc, renderDoc)
)

func fullAuthorLastName(a graphql.FullAuthor) string { return a.LastName }

func fullAuthorInitials(a graphql.FullAuthor) string { return a.Initials }

// citation is "*Journal* volume:pages (year). [PubMed](url) · [DOI](url)".
func citation(v PublicationView) string {
	return strutils.JoinSpace([]string{
		F.Pipe3(v, publicationViewJournalLens.Get, oneLine, md.Italic),
		strutils.Colon.Concat(
			F.Pipe2(v, publicationViewVolumeLens.Get, oneLine),
			F.Pipe2(v, publicationViewPagesLens.Get, oneLine),
		),
		F.Pipe4(v, publicationViewPubDateLens.Get, year, strutils.Parens, strutils.Period),
		strutils.JoinMiddleDot([]string{

			F.Pipe2(v, publicationViewURLLens.Get, publicationLink),
			F.Pipe2(v, publicationViewPubMedLens.Get, F.Bind1st(md.Link, "PubMed")),
			F.Pipe2(v, publicationViewDOILens.Get, doiLink),
		}),
	})
}

func publicationDoc(v PublicationView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		F.Pipe3(v, publicationViewTitleLens.Get, oneLine, mdfp.H2),
		F.Pipe3(v, publicationViewAuthorsLens.Get, fullAuthorNames, mdfp.Paragraph),
		F.Pipe2(v, citation, mdfp.Paragraph),
		mdfp.H3(abstractHeader),
		F.Pipe3(v, publicationViewAbstractLens.Get, oneLine, mdfp.Paragraph),
	})
}

func newGetPublicationHandler(
	client *graphql.Client,
	links linkout.Links,
) mcp.ToolHandlerFor[GetPublicationInput, PublicationView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPublicationInput,
	) (*mcp.CallToolResult, PublicationView, error) {
		out := F.Pipe9(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePublicationInput),
			IOE.Map[error](F.Flow3(
				publicationIDLens.Get,
				F.Pipe1(graphql.IDParams{}, F.Flip(idParamsIDLens.Set)),
				idParamsCtxLens.Set(ctx),
			)),
			IOE.Chain(client.GetPublication),
			IOE.Map[error](F.Pipe1(PublicationView{}, F.Flip(publicationViewRecordLens.Set))),
			IOE.Let[error](
				publicationViewURLLens.Set,
				F.Flow2(publicationViewIDLens.Get, links.PublicationURL),
			),
			IOE.Let[error](
				publicationViewPubMedLens.Set,
				F.Flow2(publicationViewIDLens.Get, PubMedURL),
			),
			IOE.Let[error](
				publicationViewDOILinkLens.Set,
				F.Flow2(publicationViewDOILens.Get, DOIURL),
			),
			ioeutils.ToEither[error, PublicationView],
			E.Fold(
				toOutcomeError[PublicationView],
				F.Bind1st(toOutcomeWithMarkdown[PublicationView], renderPublicationMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}
