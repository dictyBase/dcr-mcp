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

func document() mdfp.Block {
	return mdfp.Concat(A.From(
		mdfp.H2("Title"),
		mdfp.Paragraph("Some text"),
		mdfp.BulletList(A.From("one", "two")),
		mdfp.Paragraph("closing"),
	))
}

func render(t *testing.T, b mdfp.Block) string {
	t.Helper()
	got, err := E.Unwrap(mdfp.Render(b))
	require.NoError(t, err)
	return got
}

func TestRenderBuildsDocumentInOrder(t *testing.T) {
	require.Equal(t, wantDocument, render(t, document()))
}

func TestRenderEmpty(t *testing.T) {
	require.Empty(t, render(t, mdfp.Empty))
	require.Empty(t, render(t, mdfp.Concat([]mdfp.Block{})))
}

func TestHeadingLevels(t *testing.T) {
	got := render(t, mdfp.Concat(A.From(mdfp.H1("a"), mdfp.H2("b"), mdfp.H3("c"))))
	require.Equal(t, "# a\n\n## b\n\n### c", got)
}

func TestEmptyBulletListAddsNothing(t *testing.T) {
	got := render(t, mdfp.Concat(A.From(
		mdfp.Paragraph("before"),
		mdfp.BulletList([]string{}),
		mdfp.Paragraph("after"),
	)))
	require.Equal(t, "before\n\nafter", got)
}

// The monoid laws make sections composable: a section is one block, and a
// document built from sections renders the same as one flat list of blocks.
func TestMonoidLaws(t *testing.T) {
	a, b, c := mdfp.H2("a"), mdfp.Paragraph("b"), mdfp.Paragraph("c")
	concat := mdfp.Monoid.Concat

	t.Run("left identity", func(t *testing.T) {
		require.Equal(t, render(t, a), render(t, concat(mdfp.Empty, a)))
	})
	t.Run("right identity", func(t *testing.T) {
		require.Equal(t, render(t, a), render(t, concat(a, mdfp.Empty)))
	})
	t.Run("associativity", func(t *testing.T) {
		require.Equal(t,
			render(t, concat(concat(a, b), c)),
			render(t, concat(a, concat(b, c))),
		)
	})
	t.Run("document order", func(t *testing.T) {
		require.Equal(t, "## a\n\nb", render(t, concat(a, b)))
	})
	t.Run("nested sections equal flat blocks", func(t *testing.T) {
		nested := mdfp.Concat(A.From(a, mdfp.Concat(A.From(b, c))))
		flat := mdfp.Concat(A.From(a, b, c))
		require.Equal(t, render(t, flat), render(t, nested))
	})
}

// Render makes a fresh builder on every call, so rendering the same block
// again gives the same document and never accumulates earlier output.
func TestRenderIsRepeatable(t *testing.T) {
	doc := document()
	require.Equal(t, mdfp.Render(doc), mdfp.Render(doc))
	require.Equal(t, E.Of[error](wantDocument), mdfp.Render(doc))
}

// A block is a value: the same block can sit in two documents and neither
// document sees the other.
func TestSharedBlockDoesNotLeak(t *testing.T) {
	shared := mdfp.H2("shared")
	left := mdfp.Render(mdfp.Monoid.Concat(shared, mdfp.Paragraph("left")))
	right := mdfp.Render(mdfp.Monoid.Concat(shared, mdfp.Paragraph("right")))
	require.Equal(t, E.Of[error]("## shared\n\nleft"), left)
	require.Equal(t, E.Of[error]("## shared\n\nright"), right)
}

func TestRenderIsConcurrentSafe(t *testing.T) {
	doc := document()
	results := make(chan E.Either[error, string], 8)
	for range 8 {
		go func() { results <- mdfp.Render(doc) }()
	}
	for range 8 {
		require.Equal(t, E.Of[error](wantDocument), <-results)
	}
}

func invalidTocRange(m *md.Markdown) *md.Markdown {
	return m.TableOfContentsWithRange(md.TableOfContentsDepthH6, md.TableOfContentsDepthH1)
}

func TestRenderReturnsBuilderError(t *testing.T) {
	_, err := E.Unwrap(mdfp.Render(mdfp.Monoid.Concat(mdfp.H2("t"), invalidTocRange)))
	require.Error(t, err)
	require.Contains(t, err.Error(), "minDepth")
}

func TestRenderComposesWithPipe(t *testing.T) {
	got := F.Pipe3(
		A.From("a", "b"),
		A.Map(mdfp.Paragraph),
		mdfp.Concat,
		mdfp.Render,
	)
	require.Equal(t, E.Of[error]("a\n\nb"), got)
}
