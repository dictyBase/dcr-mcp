// Package mdfp is a pure, fp-go facing layer over github.com/nao1215/markdown.
//
// nao1215/markdown is a mutable builder: every method appends to the builder
// and returns the same pointer. mdfp hides that. A Block is an immutable
// description of one markdown block, and Render is the only function that
// creates a builder. Render makes a fresh builder on every call, so a Block
// can be shared, reused in many documents, and rendered any number of times
// with the same result.
package mdfp

import (
	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	md "github.com/nao1215/markdown"
)

// Block describes one markdown block as a step on a builder. It is an alias,
// not a defined type, so the data-last constructors below can be built with
// F.Curry2 and F.Flip alone.
type Block = func(*md.Markdown) *md.Markdown

var (
	// H1 builds a level 1 heading block.
	H1 = F.Flip(F.Curry2((*md.Markdown).H1))

	// H2 builds a level 2 heading block.
	H2 = F.Flip(F.Curry2((*md.Markdown).H2))

	// H3 builds a level 3 heading block.
	H3 = F.Flip(F.Curry2((*md.Markdown).H3))

	// Paragraph builds a plain text block.
	Paragraph = F.Flip(F.Curry2((*md.Markdown).PlainText))

	// BulletList builds one bullet list block from its items. An empty
	// slice builds a block that adds nothing.
	BulletList = F.Flip(F.Curry2(bulletItems))
)

// bulletItems spreads items into the variadic BulletList method. A method
// expression cannot be curried while its last parameter is variadic.
func bulletItems(m *md.Markdown, items []string) *md.Markdown {
	return m.BulletList(items...)
}

func applyBlock(m *md.Markdown, b Block) *md.Markdown {
	return b(m)
}

// Render applies the blocks, in order, to a fresh builder and returns the
// document. Blocks are separated by a blank line, except the items of one
// list. The result is Left when the builder recorded an error.
func Render(blocks []Block) E.Either[error, string] {
	return F.Pipe3(
		blocks,
		A.Reduce(applyBlock, md.NewMarkdown(nil, md.WithBlockSpacing())),
		E.FromError((*md.Markdown).Error),
		E.Map[error]((*md.Markdown).String),
	)
}
