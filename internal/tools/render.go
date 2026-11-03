package tools

import (
	"bytes"
	"fmt"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	EQ "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	predord "github.com/dictyBase/fp-go-loom/predicate/ord"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
)

const (
	yearLen           = 4
	phenotypeHeader   = "Phenotypes"
	paperHeader       = "Publications"
	noneRecorded      = "None recorded."
	regularStockType  = "regular"
	matchedTextHeader = "Matched text"
	strainHeader      = "Strains"
	strainWord        = "strain"
	strainsWord       = "strains"
	noValue           = "—"
)

var (
	markdownRenderer = goldmark.New(goldmark.WithExtensions(markdown.NewExtension()))

	availabilityLabel = map[bool]string{
		true:  "✅ in stock",
		false: "❌ not in stock",
	}

	strainNoun = map[bool]string{
		true:  strainWord,
		false: strainsWord,
	}

	morePagesLabel = map[bool]string{
		true:  "yes (pass next_cursor as cursor)",
		false: "no",
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

	hasNodes = predarrays.IsNonEmpty[ast.Node]()

	// isSpecialStockType selects stock types worth a label: gwdi, goldenbraid.
	isSpecialStockType = F.Pipe1(
		S.IsNonEmpty,
		P.And(F.Pipe2(regularStockType, EQ.Equals(predord.StringEq), P.Not[string])),
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

// withChildren appends kids to parent and returns parent.
//
// fp-go:allow-loop goldmark nodes are mutable and AppendChild is the only API
func withChildren(parent ast.Node, kids []ast.Node) ast.Node {
	for _, kid := range kids {
		parent.AppendChild(parent, kid)
	}
	return parent
}

func paragraph(parts ...ast.Node) ast.Node {
	return withChildren(ast.NewParagraph(), parts)
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
	block := withChildren(ast.NewTextBlock(), parts)
	return withChildren(ast.NewListItem(0), []ast.Node{block})
}

func bulletList(items []ast.Node) ast.Node {
	return withChildren(ast.NewList('-'), items)
}

// renderNodes renders nodes as Markdown. A render error gives an empty string.
func renderNodes(nodes []ast.Node) string {
	doc := withChildren(ast.NewDocument(), nodes)
	buf := new(bytes.Buffer)
	rendered := E.TryCatchError(
		buf,
		markdownRenderer.Renderer().Render(buf, nil, doc),
	)
	return F.Pipe2(
		rendered,
		E.Map[error]((*bytes.Buffer).String),
		E.GetOrElse(F.Constant1[error]("")),
	)
}

// noneRecordedNode builds a fresh paragraph, because goldmark nodes are
// mutable and must not be shared between documents.
func noneRecordedNode() ast.Node {
	return paragraph(text(noneRecorded))
}

// listOrNone renders items as a bullet list, or a "None recorded." paragraph.
func listOrNone(items []ast.Node) ast.Node {
	return F.Pipe3(
		items,
		O.FromPredicate(hasNodes),
		O.Map(bulletList),
		O.GetOrElse(noneRecordedNode),
	)
}

// section renders a heading plus its items, or a "None recorded." paragraph.
func section(title string, items []ast.Node) []ast.Node {
	return []ast.Node{heading(3, title), listOrNone(items)}
}

func year(pubDate string) string {
	return pubDate[:min(len(pubDate), yearLen)]
}

func authorName(a graphql.AuthorRef) string { return oneLine(a.LastName) }

func authorNames(authors []graphql.AuthorRef) string {
	return F.Pipe2(
		authors,
		A.Map(authorName),
		F.Bind2nd(strings.Join, ", "),
	)
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

func valueOrDash(s string) string {
	return F.Pipe3(
		s,
		O.FromPredicate(S.IsNonEmpty),
		O.Map(oneLine),
		O.GetOrElse(F.Constant(noValue)),
	)
}

// phenotypeItem labels each part of a phenotype statement, so a client copies
// the phenotype term (not the environment or the assay) into
// list_strains_with_phenotype.
func phenotypeItem(p graphql.PhenotypeItem) ast.Node {
	return listItem(
		text("Phenotype term: "+oneLine(p.Phenotype)),
		text(fmt.Sprintf(" · environment: %s · assay: %s — ",
			valueOrDash(p.Environment), valueOrDash(p.Assay),
		)),
		link("PubMed "+oneLine(p.Publication.ID), PubMedURL(p.Publication.ID)),
	)
}

func publicationItems(publications []graphql.PublicationRef) []ast.Node {
	return F.Pipe1(publications, A.Map(publicationItem))
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
	nodes := make([]ast.Node, 0, len(intro)+16)
	nodes = append(nodes, intro...)
	nodes = append(nodes, section(phenotypeHeader, F.Pipe1(v.Phenotypes, A.Map(phenotypeItem)))...)
	nodes = append(nodes, section(paperHeader, publicationItems(v.Publications))...)
	nodes = append(nodes, footer)
	return renderNodes(nodes)
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

func stockTypeLabel(stockType string) string {
	return " · " + oneLine(stockType)
}

func searchItem(h SearchHitView) ast.Node {
	suffix := F.Pipe3(
		h.StockType,
		O.FromPredicate(isSpecialStockType),
		O.Map(stockTypeLabel),
		O.GetOrElse(F.Constant("")),
	)
	line := fmt.Sprintf(" — %s · %s: %s · %s%s",
		oneLine(h.StockLabel), oneLine(h.Field), oneLine(h.DisplayText),
		availabilityLabel[h.InStock], suffix,
	)
	return listItem(link(h.ID, h.URL), text(line))
}

func matchedTextItem(m MatchedTextView) ast.Node {
	line := fmt.Sprintf("%s (%s) — %d %s",
		oneLine(m.DisplayText), oneLine(m.Field), m.StrainCount, strainNoun[m.StrainCount == 1],
	)
	return listItem(text(line))
}

func renderPhenotypeSearchMarkdown(v PhenotypeSearchView) string {
	intro := paragraph(text(fmt.Sprintf(
		"Top %d strain matches (maximum %d, no further pages).",
		v.Count, maxPhenotypeSearchLimit,
	)))
	nodes := make([]ast.Node, 0, 8)
	nodes = append(nodes, intro)
	nodes = append(nodes, section(matchedTextHeader, F.Pipe1(v.MatchedTexts, A.Map(matchedTextItem)))...)
	nodes = append(nodes, section(strainHeader, F.Pipe1(v.Results, A.Map(searchItem)))...)
	return renderNodes(nodes)
}

func phenotypeStrainItem(s PhenotypeStrainView) ast.Node {
	return listItem(
		link(s.ID, s.URL),
		text(fmt.Sprintf(" — %s · %s", oneLine(s.Label), availabilityLabel[s.InStock])),
	)
}

func renderPhenotypeStrainListMarkdown(v PhenotypeStrainListView) string {
	intro := paragraph(text(fmt.Sprintf(
		"%d strains with phenotype term \"%s\" on this page. More pages: %s.",
		v.PageCount, oneLine(v.PhenotypeTerm), morePagesLabel[v.HasMore],
	)))
	nodes := make([]ast.Node, 0, 4)
	nodes = append(nodes, intro)
	nodes = append(nodes, section(strainHeader, F.Pipe1(v.Strains, A.Map(phenotypeStrainItem)))...)
	return renderNodes(nodes)
}

func renderSearchMarkdown(v SearchResultView) string {
	intro := paragraph(text(fmt.Sprintf("Showing %d of %d matches.", len(v.Results), v.TotalCount)))
	hits := listOrNone(F.Pipe1(v.Results, A.Map(searchItem)))
	return renderNodes([]ast.Node{intro, hits})
}
