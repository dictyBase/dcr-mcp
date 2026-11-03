package tools

import (
	"fmt"
	"strconv"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictyBase/fp-go-loom/strutils"
	"github.com/dictybase/dcr-mcp/internal/graphql"
	"github.com/dictybase/dcr-mcp/internal/mdfp"
	md "github.com/nao1215/markdown"
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
	// document structure. nao1215/markdown writes text verbatim, so this is
	// the only place that escapes.
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

	// oneLine escapes text and collapses whitespace so a record field cannot
	// break the surrounding block. Every separator below is a strutils
	// monoid: an empty operand drops out with its separator, so an optional
	// part (a missing DOI, a regular stock type) needs no branch.
	oneLine = F.Flow2(strutils.CollapseWhitespace, markdownEscaper.Replace)

	// valueOrDash renders a field, or a dash when the field is empty.
	valueOrDash = F.Flow3(
		O.FromPredicate(S.IsNonEmpty),
		O.Map(oneLine),
		O.GetOrElse(F.Constant(noValue)),
	)

	hasItems = predarrays.IsNonEmpty[string]()

	// isSpecialStockType selects stock types worth a label: gwdi, goldenbraid.
	isSpecialStockType = F.Pipe1(
		S.IsNonEmpty,
		P.And(F.Pipe2(regularStockType, S.Equals, P.Not[string])),
	)

	// stockTypeTag is the stock type label, or empty for a regular stock.
	stockTypeTag = F.Flow3(
		O.FromPredicate(isSpecialStockType),
		O.Map(oneLine),
		O.GetOrElse(F.Constant(S.Monoid.Empty())),
	)

	// doiLink is a DOI link, or empty when the publication has no DOI.
	doiLink = strutils.MapNonEmpty(F.Flow2(DOIURL, F.Bind1st(md.Link, "DOI")))

	sequenceText = S.Format[int]("%d bp (call get_plasmid_sequence)")

	noneBlock = mdfp.Paragraph(noneRecorded)

	// listOrNone renders items as a bullet list block, or "None recorded.".
	listOrNone = F.Flow3(
		O.FromPredicate(hasItems),
		O.Map(mdfp.BulletList),
		O.GetOrElse(F.Constant(noneBlock)),
	)

	// renderDoc renders a document block as Markdown. A render error gives an
	// empty string. The fixed block set cannot produce one.
	renderDoc = F.Flow2(
		mdfp.Render,
		E.GetOrElse(F.Constant1[error](S.Monoid.Empty())),
	)

	authorNames    = F.Pipe1(F.Flow2(authorLastName, oneLine), A.FoldMap[graphql.AuthorRef](strutils.Comma))
	phenotypeItems = A.Map(phenotypeItem)
	paperItems     = A.Map(publicationItem)
	hitItems       = A.Map(searchItem)
	matchedItems   = A.Map(matchedTextItem)
	strainItems    = A.Map(phenotypeStrainItem)

	renderStrainMarkdown  = F.Flow2(strainDoc, renderDoc)
	renderPlasmidMarkdown = F.Flow2(plasmidDoc, renderDoc)
	renderSearchMarkdown  = F.Flow2(searchDoc, renderDoc)

	renderPhenotypeSearchMarkdown     = F.Flow2(phenotypeSearchDoc, renderDoc)
	renderPhenotypeStrainListMarkdown = F.Flow2(phenotypeStrainListDoc, renderDoc)
)

// section is a heading plus its items, or a "None recorded." paragraph.
func section(title string, items []string) mdfp.Block {
	return mdfp.Monoid.Concat(mdfp.H3(title), listOrNone(items))
}

func year(pubDate string) string {
	return pubDate[:min(len(pubDate), yearLen)]
}

func authorLastName(a graphql.AuthorRef) string { return a.LastName }

// heading is "label (id)". An empty label drops out with its space.
func heading(label, id string) mdfp.Block {
	return mdfp.H2(strutils.Space.Concat(oneLine(label), F.Pipe2(id, oneLine, strutils.Parens)))
}

func fullName(d graphql.Depositor) string {
	return F.Pipe1(strutils.Space.Concat(d.FirstName, d.LastName), valueOrDash)
}

func publicationItem(p graphql.PublicationRef) string {
	return strutils.JoinSpace([]string{
		F.Pipe2(p.Authors, authorNames, strutils.Period),
		oneLine(p.Title),
		oneLine(p.Journal),
		F.Pipe3(p.PubDate, year, strutils.Parens, strutils.Period),
		strutils.MiddleDot.Concat(md.Link("PubMed", PubMedURL(p.ID)), doiLink(p.DOI)),
	})
}

// phenotypeItem labels each part of a phenotype statement, so a client copies
// the phenotype term (not the environment or the assay) into
// list_strains_with_phenotype.
func phenotypeItem(p graphql.PhenotypeItem) string {
	parts := strutils.JoinMiddleDot([]string{
		strutils.Label.Concat("Phenotype term", valueOrDash(p.Phenotype)),
		strutils.Label.Concat("environment", valueOrDash(p.Environment)),
		strutils.Label.Concat("assay", valueOrDash(p.Assay)),
	})
	pubmed := md.Link(strutils.Space.Concat("PubMed", oneLine(p.Publication.ID)), PubMedURL(p.Publication.ID))
	return strutils.EmDash.Concat(parts, pubmed)
}

func strainSummaryItems(v StrainView) []string {
	return []string{
		strutils.Label.Concat("Availability", availabilityLabel[v.InStock]),
		strutils.Label.Concat("Species", valueOrDash(v.Species)),
		strutils.Label.Concat("Genotype", F.Pipe2(v.Genotypes, strutils.JoinSemicolon, valueOrDash)),
		strutils.Label.Concat("Depositor", fullName(v.Depositor)),
		strutils.Label.Concat("Stock page", md.Link(v.ID, v.URL)),
	}
}

func strainDoc(v StrainView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		heading(v.Label, v.ID),
		mdfp.Paragraph(oneLine(v.Summary)),
		mdfp.BulletList(strainSummaryItems(v)),
		section(phenotypeHeader, phenotypeItems(v.Phenotypes)),
		section(paperHeader, paperItems(v.Publications)),
		mdfp.Paragraph(fmt.Sprintf(
			"Showing %d of %d phenotypes and %d of %d publications (detail=%s).",
			len(v.Phenotypes), v.PhenotypeCount, len(v.Publications), v.PublicationCount, v.Detail,
		)),
	})
}

func plasmidSummaryItems(v PlasmidView) []string {
	return []string{
		strutils.Label.Concat("Availability", availabilityLabel[v.InStock]),
		strutils.Label.Concat("Keywords", F.Pipe2(v.Keywords, strutils.JoinComma, valueOrDash)),
		strutils.Label.Concat("GenBank", valueOrDash(v.GenbankAccession)),
		strutils.Label.Concat("Sequence", sequenceText(v.SequenceLength)),
		strutils.Label.Concat("Depositor", fullName(v.Depositor)),
		strutils.Label.Concat("Stock page", md.Link(v.ID, v.URL)),
	}
}

func plasmidDoc(v PlasmidView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		heading(v.Name, v.ID),
		mdfp.Paragraph(oneLine(v.Summary)),
		mdfp.BulletList(plasmidSummaryItems(v)),
		section(paperHeader, paperItems(v.Publications)),
	})
}

func searchItem(h SearchHitView) string {
	return strutils.EmDash.Concat(md.Link(h.ID, h.URL), strutils.JoinMiddleDot([]string{
		oneLine(h.StockLabel),
		strutils.Label.Concat(oneLine(h.Field), oneLine(h.DisplayText)),
		availabilityLabel[h.InStock],
		stockTypeTag(h.StockType),
	}))
}

func matchedTextItem(m MatchedTextView) string {
	return strutils.EmDash.Concat(
		strutils.Space.Concat(oneLine(m.DisplayText), F.Pipe2(m.Field, oneLine, strutils.Parens)),
		strutils.Space.Concat(strconv.Itoa(m.StrainCount), strainNoun[m.StrainCount == 1]),
	)
}

func searchDoc(v SearchResultView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf("Showing %d of %d matches.", len(v.Results), v.TotalCount)),
		listOrNone(hitItems(v.Results)),
	})
}

func phenotypeSearchDoc(v PhenotypeSearchView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf(
			"Top %d strain matches (maximum %d, no further pages).",
			v.Count, maxPhenotypeSearchLimit,
		)),
		section(matchedTextHeader, matchedItems(v.MatchedTexts)),
		section(strainHeader, hitItems(v.Results)),
	})
}

func phenotypeStrainItem(s PhenotypeStrainView) string {
	return strutils.EmDash.Concat(
		md.Link(s.ID, s.URL),
		strutils.MiddleDot.Concat(oneLine(s.Label), availabilityLabel[s.InStock]),
	)
}

func phenotypeStrainListDoc(v PhenotypeStrainListView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf(
			"%d strains with phenotype term \"%s\" on this page. More pages: %s.",
			v.PageCount, oneLine(v.PhenotypeTerm), morePagesLabel[v.HasMore],
		)),
		section(strainHeader, strainItems(v.Strains)),
	})
}
