package tools

import (
	"context"
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PublicationView is the get_publication output.
type PublicationView struct {
	graphql.PublicationDetail
	PubMedURL string `json:"pubmed_url"`
	DOIURL    string `json:"doi_url,omitempty"`
}

var toPublicationState = F.Curry3(
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

func toPublicationView(p graphql.PublicationDetail) PublicationView {
	return PublicationView{
		PublicationDetail: p,
		PubMedURL:         PubMedURL(p.ID),
		DOIURL:            DOIURL(p.DOI),
	}
}

func fullAuthorName(a graphql.FullAuthor) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", a.LastName, a.Initials))
}

func renderPublicationMarkdown(v PublicationView) string {
	authors := strings.Join(F.Pipe1(v.Authors, A.Map(fullAuthorName)), ", ")
	return fmt.Sprintf(
		"## %s\n%s\n\n*%s* %s:%s (%s). [PubMed](%s)%s\n\n### Abstract\n%s",
		v.Title, authors, v.Journal, v.Volume, v.Pages, year(v.PubDate), v.PubMedURL, doiSuffix(v.DOI), v.Abstract,
	)
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
