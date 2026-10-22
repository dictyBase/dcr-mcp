package tools

import (
	"bytes"
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
)

const (
	yearLen         = 4
	phenotypeHeader = "Phenotypes"
	paperHeader     = "Publications"
	noneRecorded    = "None recorded."
)

var (
	markdownRenderer = goldmark.New(goldmark.WithExtensions(markdown.NewExtension()))

	availabilityLabel = map[bool]string{
		true:  "✅ in stock",
		false: "❌ not in stock",
	}

	// markdownEscaper protects user-supplied catalog text from changing the
	// document structure. goldmark emits text nodes verbatim, so this is the
	// only place that escapes.
	markdownEscaper = strings.NewReplacer(
		`\`, `\\`,
		"`", "\\`",
		"*", `\*`,
		"_", `\_`,
		"[", `\[`,
		"]", `\]`,
		"<", `\<`,
		">", `\>`,
		"|", `\|`,
	)
)

// oneLine escapes text and collapses whitespace so a record field cannot break
// the surrounding block.
func oneLine(s string) string {
	return markdownEscaper.Replace(strings.Join(strings.Fields(s), " "))
}

func heading(level int, text string) ast.Node {
	h := ast.NewHeading(level)
	h.AppendChild(h, ast.NewString([]byte(text)))
	return h
}

func paragraph(parts ...ast.Node) ast.Node {
	p := ast.NewParagraph()
	for _, part := range parts {
		p.AppendChild(p, part)
	}
	return p
}

func text(s string) ast.Node {
	return ast.NewString([]byte(s))
}

func link(label, dest string) ast.Node {
	l := ast.NewLink()
	l.Destination = []byte(dest)
	l.AppendChild(l, ast.NewString([]byte(label)))
	return l
}

func listItem(parts ...ast.Node) ast.Node {
	item := ast.NewListItem(0)
	block := ast.NewTextBlock()
	for _, part := range parts {
		block.AppendChild(block, part)
	}
	item.AppendChild(item, block)
	return item
}

func bulletList(items []ast.Node) ast.Node {
	list := ast.NewList('-')
	for _, item := range items {
		list.AppendChild(list, item)
	}
	return list
}

func renderNodes(nodes []ast.Node) string {
	doc := ast.NewDocument()
	for _, node := range nodes {
		doc.AppendChild(doc, node)
	}
	var buf bytes.Buffer
	if err := markdownRenderer.Renderer().Render(&buf, nil, doc); err != nil {
		return ""
	}
	return buf.String()
}

// section renders a heading plus its items, or a "None recorded." paragraph.
func section(title string, items []ast.Node) []ast.Node {
	if len(items) == 0 {
		return []ast.Node{heading(3, title), paragraph(text(noneRecorded))}
	}
	return []ast.Node{heading(3, title), bulletList(items)}
}

func year(pubDate string) string {
	return pubDate[:min(len(pubDate), yearLen)]
}

func authorNames(authors []graphql.AuthorRef) string {
	names := A.Map(func(a graphql.AuthorRef) string { return oneLine(a.LastName) })(authors)
	return strings.Join(names, ", ")
}

func doiSuffix(doi string) []ast.Node {
	return F.Pipe2(
		doi,
		O.FromPredicate(S.IsNonEmpty),
		O.Fold(
			F.Constant([]ast.Node{}),
			func(d string) []ast.Node {
				return []ast.Node{text(" · "), link("DOI", DOIURL(d))}
			},
		),
	)
}

func publicationItem(p graphql.PublicationRef) ast.Node {
	parts := []ast.Node{
		text("- "),
		text(authorNames(p.Authors)),
		text(fmt.Sprintf(". %s %s (%s). ", oneLine(p.Title), oneLine(p.Journal), year(p.PubDate))),
		link("PubMed", PubMedURL(p.ID)),
	}
	return listItem(append(parts, doiSuffix(p.DOI)...)...)
}

func phenotypeItem(p graphql.PhenotypeItem) ast.Node {
	return listItem(
		text(oneLine(p.Phenotype)),
		text(fmt.Sprintf(" (assay: %s; environment: %s) — ", oneLine(p.Assay), oneLine(p.Environment))),
		link("PubMed "+oneLine(p.Publication.ID), PubMedURL(p.Publication.ID)),
	)
}

func publicationItems(publications []graphql.PublicationRef) []ast.Node {
	return A.Map(publicationItem)(publications)
}

func strainSummaryItems(v StrainView) []ast.Node {
	return []ast.Node{
		listItem(text("Availability: "), text(availabilityLabel[v.InStock])),
		listItem(text("Species: "), text(oneLine(v.Species))),
		listItem(text("Genotype: "), text(oneLine(strings.Join(v.Genotypes, "; ")))),
		listItem(
			text("Depositor: "),
			text(oneLine(v.Depositor.FirstName+" "+v.Depositor.LastName)),
		),
		listItem(text("Stock page: "), link(v.ID, v.URL)),
	}
}

func renderStrainMarkdown(v StrainView) string {
	intro := []ast.Node{
		heading(2, oneLine(v.Label)+" ("+oneLine(v.ID)+")"),
		paragraph(text(oneLine(v.Summary))),
		bulletList(strainSummaryItems(v)),
	}
	footer := paragraph(text(fmt.Sprintf(
		"Showing %d of %d phenotypes and %d of %d publications (detail=%s).",
		len(v.Phenotypes), v.PhenotypeCount, len(v.Publications), v.PublicationCount, v.Detail,
	)))
	nodes := append(intro, section(phenotypeHeader, A.Map(phenotypeItem)(v.Phenotypes))...)
	nodes = append(nodes, section(paperHeader, publicationItems(v.Publications))...)
	return renderNodes(append(nodes, footer))
}

func renderPlasmidMarkdown(v PlasmidView) string {
	intro := []ast.Node{
		heading(2, oneLine(v.Name)+" ("+oneLine(v.ID)+")"),
		paragraph(text(oneLine(v.Summary))),
		bulletList([]ast.Node{
			listItem(text("Availability: "), text(availabilityLabel[v.InStock])),
			listItem(text("Keywords: "), text(oneLine(strings.Join(v.Keywords, ", ")))),
			listItem(text("GenBank: "), text(oneLine(v.GenbankAccession))),
			listItem(text(fmt.Sprintf("Sequence: %d bp (call get_plasmid_sequence)", v.SequenceLength))),
			listItem(
				text("Depositor: "),
				text(oneLine(v.Depositor.FirstName+" "+v.Depositor.LastName)),
			),
			listItem(text("Stock page: "), link(v.ID, v.URL)),
		}),
	}
	return renderNodes(append(intro, section(paperHeader, publicationItems(v.Publications))...))
}

func searchItem(h SearchHitView) ast.Node {
	return listItem(
		link(h.ID, h.URL),
		text(fmt.Sprintf(" — %s · %s: %s · %s",
			oneLine(h.StrainLabel), oneLine(h.Field), oneLine(h.DisplayText), availabilityLabel[h.InStock],
		)),
	)
}

func renderSearchMarkdown(v SearchResultView) string {
	intro := paragraph(text(fmt.Sprintf("Showing %d of %d matches.", len(v.Results), v.TotalCount)))
	if len(v.Results) == 0 {
		return renderNodes([]ast.Node{intro, paragraph(text(noneRecorded))})
	}
	return renderNodes([]ast.Node{intro, bulletList(A.Map(searchItem)(v.Results))})
}
