package tools

import (
	"fmt"
	"strconv"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
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
	renderStrainViewGenotypesLens = L.MakeLens(
		func(v StrainView) []string { return v.Genotypes },
		func(v StrainView, value []string) StrainView {
			v.Genotypes = value
			return v
		},
	)
	renderStrainViewDepositorLens = L.MakeLens(
		func(v StrainView) graphql.Depositor { return v.Depositor },
		func(v StrainView, value graphql.Depositor) StrainView {
			v.Depositor = value
			return v
		},
	)
	renderStrainViewSummaryLens = L.MakeLens(
		func(v StrainView) string { return v.Summary },
		func(v StrainView, value string) StrainView {
			v.Summary = value
			return v
		},
	)
	renderPlasmidViewKeywordsLens = L.MakeLens(
		func(v PlasmidView) []string { return v.Keywords },
		func(v PlasmidView, value []string) PlasmidView {
			v.Keywords = value
			return v
		},
	)
	renderPlasmidViewDepositorLens = L.MakeLens(
		func(v PlasmidView) graphql.Depositor { return v.Depositor },
		func(v PlasmidView, value graphql.Depositor) PlasmidView {
			v.Depositor = value
			return v
		},
	)
	renderPlasmidViewSummaryLens = L.MakeLens(
		func(v PlasmidView) string { return v.Summary },
		func(v PlasmidView, value string) PlasmidView {
			v.Summary = value
			return v
		},
	)
	renderSearchHitViewIDLens = L.MakeLens(
		func(v SearchHitView) string { return v.ID },
		func(v SearchHitView, value string) SearchHitView {
			v.ID = value
			return v
		},
	)
	renderSearchHitViewStockLabelLens = L.MakeLens(
		func(v SearchHitView) string { return v.StockLabel },
		func(v SearchHitView, value string) SearchHitView {
			v.StockLabel = value
			return v
		},
	)
	renderSearchHitViewFieldLens = L.MakeLens(
		func(v SearchHitView) string { return v.Field },
		func(v SearchHitView, value string) SearchHitView {
			v.Field = value
			return v
		},
	)
	renderSearchHitViewDisplayTextLens = L.MakeLens(
		func(v SearchHitView) string { return v.DisplayText },
		func(v SearchHitView, value string) SearchHitView {
			v.DisplayText = value
			return v
		},
	)
	renderSearchHitViewInStockLens = L.MakeLens(
		func(v SearchHitView) bool { return v.InStock },
		func(v SearchHitView, value bool) SearchHitView {
			v.InStock = value
			return v
		},
	)
	renderSearchHitViewStockTypeLens = L.MakeLens(
		func(v SearchHitView) string { return v.StockType },
		func(v SearchHitView, value string) SearchHitView {
			v.StockType = value
			return v
		},
	)
	renderPhenotypeStrainViewIDLens = L.MakeLens(
		func(v PhenotypeStrainView) string { return v.ID },
		func(v PhenotypeStrainView, value string) PhenotypeStrainView {
			v.ID = value
			return v
		},
	)
	renderPhenotypeStrainViewLabelLens = L.MakeLens(
		func(v PhenotypeStrainView) string { return v.Label },
		func(v PhenotypeStrainView, value string) PhenotypeStrainView {
			v.Label = value
			return v
		},
	)
	renderPhenotypeStrainViewInStockLens = L.MakeLens(
		func(v PhenotypeStrainView) bool { return v.InStock },
		func(v PhenotypeStrainView, value bool) PhenotypeStrainView {
			v.InStock = value
			return v
		},
	)
	// publicationLink is the local publication link, or empty for a record
	// without an ID.
	publicationLink   = strutils.MapNonEmpty(F.Bind1st(md.Link, "Publication"))
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

func publicationItem(p PublicationRefView) string {
	return strutils.JoinSpace([]string{
		F.Pipe3(p, publicationRefViewAuthorsLens.Get, authorNames, strutils.Period),
		F.Pipe2(p, publicationRefViewTitleLens.Get, oneLine),
		F.Pipe2(p, publicationRefViewJournalLens.Get, oneLine),
		F.Pipe4(p, publicationRefViewPubDateLens.Get, year, strutils.Parens, strutils.Period),
		strutils.JoinMiddleDot([]string{
			F.Pipe2(p, publicationRefViewURLLens.Get, publicationLink),
			F.Pipe3(p, publicationRefViewIDLens.Get, PubMedURL, F.Bind1st(md.Link, "PubMed")),
			F.Pipe2(p, publicationRefViewDOILens.Get, doiLink),
		}),
	})
}

// phenotypeItem labels each part of a phenotype statement, so a client copies
// the phenotype term (not the environment or the assay) into
// list_strains_with_phenotype.
func phenotypeItem(p PhenotypeItemView) string {
	paperID := phenotypeItemViewPaperIDLens.Get(p)

	return strutils.EmDash.Concat(
		strutils.JoinMiddleDot([]string{
			strutils.Label.Concat(
				"Phenotype term",
				F.Pipe2(p, phenotypeItemViewPhenotypeLens.Get, valueOrDash),
			),
			strutils.Label.Concat(
				"environment",
				F.Pipe2(p, phenotypeItemViewEnvironmentLens.Get, valueOrDash),
			),
			strutils.Label.Concat(
				"assay",
				F.Pipe2(p, phenotypeItemViewAssayLens.Get, valueOrDash),
			),
		}),
		strutils.JoinMiddleDot([]string{
			F.Pipe2(p, phenotypeItemViewPaperURLLens.Get, publicationLink),
			md.Link(
				strutils.Space.Concat("PubMed", oneLine(paperID)),
				PubMedURL(paperID),
			),
		}),
	)
}

func strainSummaryItems(v StrainView) []string {
	return []string{
		strutils.Label.Concat("Availability", availabilityLabel[strainViewInStockLens.Get(v)]),
		strutils.Label.Concat("Species", valueOrDash(strainViewSpeciesLens.Get(v))),
		strutils.Label.Concat("Genotype", F.Pipe2(renderStrainViewGenotypesLens.Get(v), strutils.JoinSemicolon, valueOrDash)),
		strutils.Label.Concat("Depositor", fullName(renderStrainViewDepositorLens.Get(v))),
		strutils.Label.Concat("Stock page", md.Link(strainViewIDLens.Get(v), strainViewURLLens.Get(v))),
	}
}

func strainDoc(v StrainView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		heading(strainViewLabelLens.Get(v), strainViewIDLens.Get(v)),
		mdfp.Paragraph(oneLine(renderStrainViewSummaryLens.Get(v))),
		mdfp.BulletList(strainSummaryItems(v)),
		section(phenotypeHeader, phenotypeItems(strainViewPhenotypesLens.Get(v))),
		section(paperHeader, paperItems(strainViewPublicationsLens.Get(v))),
		mdfp.Paragraph(fmt.Sprintf(
			"Showing %d of %d phenotypes and %d of %d publications (detail=%s).",
			len(strainViewPhenotypesLens.Get(v)),
			strainViewPhenotypeCountLens.Get(v),
			len(strainViewPublicationsLens.Get(v)),
			strainViewPublicationCountLens.Get(v),
			strainViewDetailLens.Get(v),
		)),
	})
}

func plasmidSummaryItems(v PlasmidView) []string {
	return []string{
		strutils.Label.Concat("Availability", availabilityLabel[plasmidViewInStockLens.Get(v)]),
		strutils.Label.Concat("Keywords", F.Pipe2(renderPlasmidViewKeywordsLens.Get(v), strutils.JoinComma, valueOrDash)),
		strutils.Label.Concat("GenBank", valueOrDash(plasmidViewGenbankLens.Get(v))),
		strutils.Label.Concat("Sequence", sequenceText(plasmidViewSequenceLengthLens.Get(v))),
		strutils.Label.Concat("Depositor", fullName(renderPlasmidViewDepositorLens.Get(v))),
		strutils.Label.Concat("Stock page", md.Link(plasmidViewIDLens.Get(v), plasmidViewURLLens.Get(v))),
	}
}

func plasmidDoc(v PlasmidView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		heading(plasmidViewNameLens.Get(v), plasmidViewIDLens.Get(v)),
		mdfp.Paragraph(oneLine(renderPlasmidViewSummaryLens.Get(v))),
		mdfp.BulletList(plasmidSummaryItems(v)),
		section(paperHeader, paperItems(plasmidViewPublicationsLens.Get(v))),
	})
}

func searchItem(h SearchHitView) string {
	return strutils.EmDash.Concat(
		md.Link(renderSearchHitViewIDLens.Get(h), searchHitViewURLLens.Get(h)),
		strutils.JoinMiddleDot([]string{
			oneLine(renderSearchHitViewStockLabelLens.Get(h)),
			strutils.Label.Concat(
				oneLine(renderSearchHitViewFieldLens.Get(h)),
				oneLine(renderSearchHitViewDisplayTextLens.Get(h)),
			),
			availabilityLabel[renderSearchHitViewInStockLens.Get(h)],
			stockTypeTag(renderSearchHitViewStockTypeLens.Get(h)),
		}),
	)
}

func matchedTextItem(m MatchedTextView) string {
	return strutils.EmDash.Concat(
		strutils.Space.Concat(oneLine(m.DisplayText), F.Pipe2(m.Field, oneLine, strutils.Parens)),
		strutils.Space.Concat(strconv.Itoa(m.StrainCount), strainNoun[m.StrainCount == 1]),
	)
}

func searchDoc(v SearchResultView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf(
			"Showing %d of %d matches.",
			len(searchResultViewResultsLens.Get(v)),
			searchResultViewTotalLens.Get(v),
		)),
		listOrNone(hitItems(searchResultViewResultsLens.Get(v))),
	})
}

func phenotypeSearchDoc(v PhenotypeSearchView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf(
			"Top %d strain matches (maximum %d, no further pages).",
			phenotypeSearchViewCountLens.Get(v), maxPhenotypeSearchLimit,
		)),
		section(matchedTextHeader, matchedItems(phenotypeSearchViewMatchedLens.Get(v))),
		section(strainHeader, hitItems(phenotypeSearchViewResultsLens.Get(v))),
	})
}

func phenotypeStrainItem(s PhenotypeStrainView) string {
	return strutils.EmDash.Concat(
		md.Link(renderPhenotypeStrainViewIDLens.Get(s), phenotypeStrainViewURLLens.Get(s)),
		strutils.MiddleDot.Concat(
			oneLine(renderPhenotypeStrainViewLabelLens.Get(s)),
			availabilityLabel[renderPhenotypeStrainViewInStockLens.Get(s)],
		),
	)
}

func phenotypeStrainListDoc(v PhenotypeStrainListView) mdfp.Block {
	return mdfp.Concat([]mdfp.Block{
		mdfp.Paragraph(fmt.Sprintf(
			"%d strains with phenotype term \"%s\" on this page. More pages: %s.",
			phenotypeListViewPageCountLens.Get(v),
			oneLine(phenotypeListViewTermLens.Get(v)),
			morePagesLabel[phenotypeListViewHasMoreLens.Get(v)],
		)),
		section(strainHeader, strainItems(phenotypeListViewStrainsLens.Get(v))),
	})
}
