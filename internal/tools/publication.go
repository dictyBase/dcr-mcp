package tools

import (
	"context"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	M "github.com/IBM/fp-go/v2/monoid"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
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
	// fullAuthorName is "LastName Initials". The space monoid, lifted to
	// functions with M.FunctionMonoid, joins the two field projections. An
	// empty part drops out with its space.
	fullAuthorName = M.FunctionMonoid[graphql.FullAuthor](spaced).Concat(
		F.Flow2(fullAuthorLastName, oneLine),
		F.Flow2(fullAuthorInitials, oneLine),
	)

	fullAuthorNames = F.Pipe1(fullAuthorName, A.FoldMap[graphql.FullAuthor](commaSep))

	renderPublicationMarkdown = F.Flow2(publicationDoc, renderDoc)

	toPublicationState = F.Curry3(
		func(
			ctx context.Context,
			client *graphql.Client,
			in GetPublicationInput,
		) publicationState {
			return publicationState{
				ctx:    ctx,
				client: client,
				input:  in,
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
	return joinSpaced([]string{
		F.Pipe2(v.Journal, oneLine, md.Italic),
		colonTight.Concat(oneLine(v.Volume), oneLine(v.Pages)),
		F.Pipe3(v.PubDate, year, parens, periods),
		dotted.Concat(md.Link("PubMed", v.PubMedURL), doiLink(v.DOI)),
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

func executeGetPublication(s publicationState) IOE.IOEither[error, graphql.PublicationDetail] {
	return s.client.GetPublication(graphql.IDParams{
		Ctx: s.ctx,
		ID:  s.input.ID,
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
		buildState := F.Pipe1(client, toPublicationState(ctx))
		out := F.Pipe6(
			IOE.Of[error](input),
			IOE.ChainEitherK(ValidatePublicationInput),
			IOE.Map[error](buildState),
			IOE.Chain(executeGetPublication),
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
