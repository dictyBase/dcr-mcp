package tools

import (
	"fmt"
	"strconv"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	M "github.com/IBM/fp-go/v2/monoid"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
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

	// String monoids and semigroups for every separator in the output. An
	// intersperse monoid skips an empty operand, so an optional part (a
	// missing DOI, a regular stock type) drops out together with its
	// separator, with no branch.
	spaced     = S.IntersperseMonoid(" ")
	dotted     = S.IntersperseMonoid(" · ")
	commaSep   = S.IntersperseMonoid(", ")
	semicolons = S.IntersperseMonoid("; ")
	colonTight = S.IntersperseMonoid(":")
	labeled    = S.IntersperseSemigroup(": ")
	dashed     = S.IntersperseSemigroup(" — ")

	joinSpaced = M.ConcatAll(spaced)
	joinDotted = M.ConcatAll(dotted)
	joinComma  = M.ConcatAll(commaSep)
	joinSemi   = M.ConcatAll(semicolons)

	// oneLine escapes text and collapses whitespace so a record field cannot
	// break the surrounding block.
	oneLine = F.Flow3(strings.Fields, joinSpaced, markdownEscaper.Replace)

	// valueOrDash renders a field, or a dash when the field is empty.
	valueOrDash = F.Flow3(
		O.FromPredicate(S.IsNonEmpty),
		O.Map(oneLine),
		O.GetOrElse(F.Constant(noValue)),
	)

	parens  = F.Flow2(S.Prepend("("), S.Append(")"))
	periods = S.Append(".")

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
	doiLink = F.Flow3(
		O.FromPredicate(S.IsNonEmpty),
		O.Map(F.Flow2(DOIURL, F.Bind1st(md.Link, "DOI"))),
		O.GetOrElse(F.Constant(S.Monoid.Empty())),
	)

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

	authorNames    = F.Pipe1(F.Flow2(authorLastName, oneLine), A.FoldMap[graphql.AuthorRef](commaSep))
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
	return mdfp.H2(spaced.Concat(oneLine(label), F.Pipe2(id, oneLine, parens)))
}

func fullName(d graphql.Depositor) string {
	return F.Pipe1(spaced.Concat(d.FirstName, d.LastName), valueOrDash)
}

func publicationItem(p graphql.PublicationRef) string {
	return joinSpaced([]string{
		F.Pipe2(p.Authors, authorNames, periods),
		oneLine(p.Title),
		oneLine(p.Journal),
		F.Pipe3(p.PubDate, year, parens, periods),
		dotted.Concat(md.Link("PubMed", PubMedURL(p.ID)), doiLink(p.DOI)),
	})
}

// phenotypeItem labels each part of a phenotype statement, so a client copies
// the phenotype term (not the environment or the assay) into
// list_strains_with_phenotype.
func phenotypeItem(p graphql.PhenotypeItem) string {
	parts := joinDotted([]string{
		labeled.Concat("Phenotype term", valueOrDash(p.Phenotype)),
		labeled.Concat("environment", valueOrDash(p.Environment)),
		labeled.Concat("assay", valueOrDash(p.Assay)),
	})
	pubmed := md.Link(spaced.Concat("PubMed", oneLine(p.Publication.ID)), PubMedURL(p.Publication.ID))
	return dashed.Concat(parts, pubmed)
}

func strainSummaryItems(v StrainView) []string {
	return []string{
		labeled.Concat("Availability", availabilityLabel[v.InStock]),
		labeled.Concat("Species", valueOrDash(v.Species)),
		labeled.Concat("Genotype", F.Pipe2(v.Genotypes, joinSemi, valueOrDash)),
		labeled.Concat("Depositor", fullName(v.Depositor)),
		labeled.Concat("Stock page", md.Link(v.ID, v.URL)),
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
		labeled.Concat("Availability", availabilityLabel[v.InStock]),
		labeled.Concat("Keywords", F.Pipe2(v.Keywords, joinComma, valueOrDash)),
		labeled.Concat("GenBank", valueOrDash(v.GenbankAccession)),
		labeled.Concat("Sequence", sequenceText(v.SequenceLength)),
		labeled.Concat("Depositor", fullName(v.Depositor)),
		labeled.Concat("Stock page", md.Link(v.ID, v.URL)),
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
	return dashed.Concat(md.Link(h.ID, h.URL), joinDotted([]string{
		oneLine(h.StockLabel),
		labeled.Concat(oneLine(h.Field), oneLine(h.DisplayText)),
		availabilityLabel[h.InStock],
		stockTypeTag(h.StockType),
	}))
}

func matchedTextItem(m MatchedTextView) string {
	return dashed.Concat(
		spaced.Concat(oneLine(m.DisplayText), F.Pipe2(m.Field, oneLine, parens)),
		spaced.Concat(strconv.Itoa(m.StrainCount), strainNoun[m.StrainCount == 1]),
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
	return dashed.Concat(
		md.Link(s.ID, s.URL),
		dotted.Concat(oneLine(s.Label), availabilityLabel[s.InStock]),
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
