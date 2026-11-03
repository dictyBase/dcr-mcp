package mdfp_test

import (
	"testing"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/dictybase/dcr-mcp/internal/mdfp"
	md "github.com/nao1215/markdown"
	"github.com/stretchr/testify/require"
)

const wantDocument = "## Title\n\nSome text\n\n- one\n- two\n\nclosing"

func document() []mdfp.Block {
	return A.From(
		mdfp.H2("Title"),
		mdfp.Paragraph("Some text"),
		mdfp.BulletList([]string{"one", "two"}),
		mdfp.Paragraph("closing"),
	)
}

func TestRenderBuildsDocument(t *testing.T) {
	got, err := E.Unwrap(mdfp.Render(document()))
	require.NoError(t, err)
	require.Equal(t, wantDocument, got)
}

func TestRenderEmpty(t *testing.T) {
	got, err := E.Unwrap(mdfp.Render([]mdfp.Block{}))
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestHeadingLevels(t *testing.T) {
	got, err := E.Unwrap(mdfp.Render(A.From(
		mdfp.H1("a"), mdfp.H2("b"), mdfp.H3("c"),
	)))
	require.NoError(t, err)
	require.Equal(t, "# a\n\n## b\n\n### c", got)
}

func TestEmptyBulletListAddsNothing(t *testing.T) {
	got, err := E.Unwrap(mdfp.Render(A.From(
		mdfp.Paragraph("before"),
		mdfp.BulletList([]string{}),
		mdfp.Paragraph("after"),
	)))
	require.NoError(t, err)
	require.Equal(t, "before\n\nafter", got)
}

// Render makes a fresh builder on every call, so rendering the same blocks
// again gives the same document and never accumulates earlier output.
func TestRenderIsRepeatable(t *testing.T) {
	blocks := document()
	first := mdfp.Render(blocks)
	second := mdfp.Render(blocks)
	require.Equal(t, first, second)
	require.Equal(t, E.Of[error](wantDocument), second)
}

// A block is a value: the same block can sit in two documents and neither
// document sees the other.
func TestSharedBlockDoesNotLeak(t *testing.T) {
	shared := mdfp.H2("shared")
	left := mdfp.Render(A.From(shared, mdfp.Paragraph("left")))
	right := mdfp.Render(A.From(shared, mdfp.Paragraph("right")))
	require.Equal(t, E.Of[error]("## shared\n\nleft"), left)
	require.Equal(t, E.Of[error]("## shared\n\nright"), right)
}

func TestRenderIsConcurrentSafe(t *testing.T) {
	blocks := document()
	results := make(chan E.Either[error, string], 8)
	for range 8 {
		go func() { results <- mdfp.Render(blocks) }()
	}
	for range 8 {
		require.Equal(t, E.Of[error](wantDocument), <-results)
	}
}

func invalidTocRange(m *md.Markdown) *md.Markdown {
	return m.TableOfContentsWithRange(md.TableOfContentsDepthH6, md.TableOfContentsDepthH1)
}

func TestRenderReturnsBuilderError(t *testing.T) {
	_, err := E.Unwrap(mdfp.Render(A.From(mdfp.H2("t"), invalidTocRange)))
	require.Error(t, err)
	require.Contains(t, err.Error(), "minDepth")
}

func TestRenderComposesWithPipe(t *testing.T) {
	got := F.Pipe2(
		A.From("a", "b"),
		A.Map(mdfp.Paragraph),
		mdfp.Render,
	)
	require.Equal(t, E.Of[error]("a\n\nb"), got)
}
