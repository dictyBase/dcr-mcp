// Package mdfp is a pure, fp-go facing layer over github.com/nao1215/markdown.
//
// nao1215/markdown is a mutable builder: every method appends to the builder
// and returns the same pointer. mdfp hides that. A Block is an endomorphism on
// the builder: an immutable description of markdown that runs only when Render
// applies it. Blocks form a monoid (Monoid), so a section or a whole document
// is itself one Block, built with M.ConcatAll.
//
// Render is the only function that creates a builder. It makes a fresh one on
// every call, so a Block can be shared, reused in many documents, and rendered
// any number of times with the same result.
package mdfp

import (
	E "github.com/IBM/fp-go/v2/either"
	EN "github.com/IBM/fp-go/v2/endomorphism"
	F "github.com/IBM/fp-go/v2/function"
	M "github.com/IBM/fp-go/v2/monoid"
	md "github.com/nao1215/markdown"
)

// Block describes markdown as a step on a builder.
type Block = EN.Endomorphism[*md.Markdown]

var (
	// Monoid combines blocks in document order: Concat(a, b) renders a, then
	// b. The endomorphism monoid composes right to left, so M.Reverse flips it.
	// Empty is the block that renders nothing.
	Monoid = M.Reverse(EN.Monoid[*md.Markdown]())

	// Concat combines blocks into one block, in document order.
	Concat = M.ConcatAll(Monoid)

	// Empty is the block that renders nothing.
	Empty = Monoid.Empty()

	// H1 builds a level 1 heading block.
	H1 = F.Curry2(F.Swap((*md.Markdown).H1))

	// H2 builds a level 2 heading block.
	H2 = F.Curry2(F.Swap((*md.Markdown).H2))

	// H3 builds a level 3 heading block.
	H3 = F.Curry2(F.Swap((*md.Markdown).H3))

	// Paragraph builds a plain text block.
	Paragraph = F.Curry2(F.Swap((*md.Markdown).PlainText))

	// BulletList builds one bullet list block from its items. An empty
	// slice builds a block that adds nothing.
	BulletList = F.Curry2(bulletItems)
)

// bulletItems spreads items into the variadic BulletList method. A method
// expression cannot be curried while its last parameter is variadic.
func bulletItems(items []string, m *md.Markdown) *md.Markdown {
	return m.BulletList(items...)
}

// Render runs the block on a fresh builder and returns the document. Blocks
// are separated by a blank line, except the items of one list. The result is
// Left when the builder recorded an error.
func Render(block Block) E.Either[error, string] {
	return F.Pipe3(
		md.NewMarkdown(nil, md.WithBlockSpacing()),
		block,
		E.FromError((*md.Markdown).Error),
		E.Map[error]((*md.Markdown).String),
	)
}
