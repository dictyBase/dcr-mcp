package tools

import (
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

	hasItems = predarrays.IsNonEmpty[string]()

	// isSpecialStockType selects stock types worth a label: gwdi, goldenbraid.
	isSpecialStockType = F.Pipe1(
		S.IsNonEmpty,
		P.And(F.Pipe2(regularStockType, EQ.Equals(predord.StringEq), P.Not[string])),
	)

	noneBlock = mdfp.Paragraph(noneRecorded)

	// listOrNone renders items as a bullet list block, or "None recorded.".
	listOrNone = F.Flow3(
		O.FromPredicate(hasItems),
		O.Map(mdfp.BulletList),
		O.GetOrElse(F.Constant(noneBlock)),
	)

	// renderDoc renders blocks as Markdown. A render error gives an empty
	// string. The fixed block set cannot produce one.
	renderDoc = F.Flow2(
		mdfp.Render,
		E.GetOrElse(F.Constant1[error]("")),
	)

	authorNames = F.Flow2(
		A.Map(authorName),
		F.Bind2nd(strings.Join, ", "),
	)
	phenotypeItems = A.Map(phenotypeItem)
	paperItems     = A.Map(publicationItem)
	hitItems       = A.Map(searchItem)
	matchedItems   = A.Map(matchedTextItem)
	strainItems    = A.Map(phenotypeStrainItem)

	renderStrainMarkdown  = F.Flow2(strainBlocks, renderDoc)
	renderPlasmidMarkdown = F.Flow2(plasmidBlocks, renderDoc)
	renderSearchMarkdown  = F.Flow2(searchBlocks, renderDoc)

	renderPhenotypeSearchMarkdown     = F.Flow2(phenotypeSearchBlocks, renderDoc)
	renderPhenotypeStrainListMarkdown = F.Flow2(phenotypeStrainListBlocks, renderDoc)
)

// oneLine escapes text and collapses whitespace so a record field cannot break
// the surrounding block.
func oneLine(s string) string {
	return markdownEscaper.Replace(strings.Join(strings.Fields(s), " "))
}

// section is a heading plus its items, or a "None recorded." paragraph.
func section(title string, items []string) []mdfp.Block {
	return []mdfp.Block{mdfp.H3(title), listOrNone(items)}
}

func year(pubDate string) string {
	return pubDate[:min(len(pubDate), yearLen)]
}

func authorName(a graphql.AuthorRef) string { return oneLine(a.LastName) }

func fullName(first, last string) string {
	return valueOrDash(strings.TrimSpace(first + " " + last))
}

func doiSuffix(doi string) string {
	return F.Pipe3(
		doi,
		O.FromPredicate(S.IsNonEmpty),
		O.Map(F.Flow2(DOIURL, F.Bind1st(md.Link, "DOI"))),
		O.Fold(F.Constant(""), S.Prepend(" · ")),
	)
}

func publicationItem(p graphql.PublicationRef) string {
	return fmt.Sprintf("%s. %s %s (%s). %s%s",
		authorNames(p.Authors), oneLine(p.Title), oneLine(p.Journal), year(p.PubDate),
		md.Link("PubMed", PubMedURL(p.ID)), doiSuffix(p.DOI),
	)
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
func phenotypeItem(p graphql.PhenotypeItem) string {
	return fmt.Sprintf("Phenotype term: %s · environment: %s · assay: %s — %s",
		oneLine(p.Phenotype), valueOrDash(p.Environment), valueOrDash(p.Assay),
		md.Link("PubMed "+oneLine(p.Publication.ID), PubMedURL(p.Publication.ID)),
	)
}

func strainSummaryItems(v StrainView) []string {
	return []string{
		"Availability: " + availabilityLabel[v.InStock],
		"Species: " + valueOrDash(v.Species),
		"Genotype: " + valueOrDash(strings.Join(v.Genotypes, "; ")),
		"Depositor: " + fullName(v.Depositor.FirstName, v.Depositor.LastName),
		"Stock page: " + md.Link(v.ID, v.URL),
	}
}

func strainBlocks(v StrainView) []mdfp.Block {
	return A.Flatten([][]mdfp.Block{
		{
			mdfp.H2(fmt.Sprintf("%s (%s)", oneLine(v.Label), oneLine(v.ID))),
			mdfp.Paragraph(oneLine(v.Summary)),
			mdfp.BulletList(strainSummaryItems(v)),
		},
		section(phenotypeHeader, phenotypeItems(v.Phenotypes)),
		section(paperHeader, paperItems(v.Publications)),
		{mdfp.Paragraph(fmt.Sprintf(
			"Showing %d of %d phenotypes and %d of %d publications (detail=%s).",
			len(v.Phenotypes), v.PhenotypeCount, len(v.Publications), v.PublicationCount, v.Detail,
		))},
	})
}

func plasmidSummaryItems(v PlasmidView) []string {
	return []string{
		"Availability: " + availabilityLabel[v.InStock],
		"Keywords: " + valueOrDash(strings.Join(v.Keywords, ", ")),
		"GenBank: " + valueOrDash(v.GenbankAccession),
		fmt.Sprintf("Sequence: %d bp (call get_plasmid_sequence)", v.SequenceLength),
		"Depositor: " + fullName(v.Depositor.FirstName, v.Depositor.LastName),
		"Stock page: " + md.Link(v.ID, v.URL),
	}
}

func plasmidBlocks(v PlasmidView) []mdfp.Block {
	return A.Flatten([][]mdfp.Block{
		{
			mdfp.H2(fmt.Sprintf("%s (%s)", oneLine(v.Name), oneLine(v.ID))),
			mdfp.Paragraph(oneLine(v.Summary)),
			mdfp.BulletList(plasmidSummaryItems(v)),
		},
		section(paperHeader, paperItems(v.Publications)),
	})
}

func stockTypeLabel(stockType string) string {
	return " · " + oneLine(stockType)
}

func searchItem(h SearchHitView) string {
	suffix := F.Pipe3(
		h.StockType,
		O.FromPredicate(isSpecialStockType),
		O.Map(stockTypeLabel),
		O.GetOrElse(F.Constant("")),
	)
	return fmt.Sprintf("%s — %s · %s: %s · %s%s",
		md.Link(h.ID, h.URL), oneLine(h.StockLabel), oneLine(h.Field),
		oneLine(h.DisplayText), availabilityLabel[h.InStock], suffix,
	)
}

func matchedTextItem(m MatchedTextView) string {
	return fmt.Sprintf("%s (%s) — %d %s",
		oneLine(m.DisplayText), oneLine(m.Field), m.StrainCount, strainNoun[m.StrainCount == 1],
	)
}

func searchBlocks(v SearchResultView) []mdfp.Block {
	return []mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf("Showing %d of %d matches.", len(v.Results), v.TotalCount)),
		listOrNone(hitItems(v.Results)),
	}
}

func phenotypeSearchBlocks(v PhenotypeSearchView) []mdfp.Block {
	return A.Flatten([][]mdfp.Block{
		{mdfp.Paragraph(fmt.Sprintf(
			"Top %d strain matches (maximum %d, no further pages).",
			v.Count, maxPhenotypeSearchLimit,
		))},
		section(matchedTextHeader, matchedItems(v.MatchedTexts)),
		section(strainHeader, hitItems(v.Results)),
	})
}

func phenotypeStrainItem(s PhenotypeStrainView) string {
	return fmt.Sprintf("%s — %s · %s",
		md.Link(s.ID, s.URL), oneLine(s.Label), availabilityLabel[s.InStock],
	)
}

func phenotypeStrainListBlocks(v PhenotypeStrainListView) []mdfp.Block {
	return A.Flatten([][]mdfp.Block{
		{mdfp.Paragraph(fmt.Sprintf(
			"%d strains with phenotype term \"%s\" on this page. More pages: %s.",
			v.PageCount, oneLine(v.PhenotypeTerm), morePagesLabel[v.HasMore],
		))},
		section(strainHeader, strainItems(v.Strains)),
	})
}
