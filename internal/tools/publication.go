package tools

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/fp-go-loom/strutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/mdfp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	md "github.com/nao1215/markdown"
)

// PublicationView is the get_publication output.
type PublicationView struct {
	graphql.PublicationDetail
	PubMedURL string `json:"pubmed_url"`
	DOIURL    string `json:"doi_url,omitempty"`
}

const abstractHeader = "Abstract"

var (
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

	toPublicationParams = F.Curry2(
		func(ctx context.Context, in GetPublicationInput) graphql.IDParams {
			return graphql.IDParams{
				Ctx: ctx,
				ID:  in.ID,
			}
		},
	)
)

func toPublicationView(p graphql.PublicationDetail) PublicationView {
	return PublicationView{
		PublicationDetail: p,
		PubMedURL:         PubMedURL(p.ID),
		DOIURL:            DOIURL(p.DOI),
	}
}

func fullAuthorLastName(a graphql.FullAuthor) string { return a.LastName }

func fullAuthorInitials(a graphql.FullAuthor) string { return a.Initials }

// citation is "*Journal* volume:pages (year). [PubMed](url) · [DOI](url)".
func citation(v PublicationView) string {
	return strutils.JoinSpace([]string{
		F.Pipe2(v.Journal, oneLine, md.Italic),
		strutils.Colon.Concat(oneLine(v.Volume), oneLine(v.Pages)),
		F.Pipe3(v.PubDate, year, strutils.Parens, strutils.Period),
		strutils.MiddleDot.Concat(md.Link("PubMed", v.PubMedURL), doiLink(v.DOI)),
	})
}

func publicationDoc(v PublicationView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.H2(oneLine(v.Title)),
		mdfp.Paragraph(fullAuthorNames(v.Authors)),
		mdfp.Paragraph(citation(v)),
		mdfp.H3(abstractHeader),
		mdfp.Paragraph(oneLine(v.Abstract)),
	})
}

func newGetPublicationHandler(
	client *graphql.Client,
) mcp.ToolHandlerFor[GetPublicationInput, PublicationView] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input GetPublicationInput,
	) (*mcp.CallToolResult, PublicationView, error) {
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePublicationInput),
			IOE.Map[error](toPublicationParams(ctx)),
			IOE.Chain(client.GetPublication),
			IOE.Map[error](toPublicationView),
			ioeutils.ToEither[error, PublicationView],
			E.Fold(
				toOutcomeError[PublicationView],
				toOutcomeWithMarkdown(renderPublicationMarkdown),
			),
		)
		return out.Result, out.Output, out.Err
	}
}
