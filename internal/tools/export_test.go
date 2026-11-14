package tools

import F "github.com/IBM/fp-go/v2/function"

// Test-only exports of unexported functions for package tools_test. The
// curried forms live here, not in the production files, because only the
// external tests apply the configured value and the detail level in two
// steps.
var (
	StrainViewFor                     = F.Curry2(strainViewFor)
	PlasmidViewFor                    = F.Curry2(plasmidViewFor)
	ToSearchResultView                = F.Curry2(toSearchResultView)
	ToSearchDocuments                 = F.Curry2(toSearchDocuments)
	LabelWithID                       = labelWithID
	ToPhenotypeSearchView             = F.Curry2(toPhenotypeSearchView)
	ToPhenotypeStrainListView         = F.Curry3(phenotypeStrainListView)
	RenderStrainMarkdown              = renderStrainMarkdown
	RenderPlasmidMarkdown             = renderPlasmidMarkdown
	RenderSearchMarkdown              = renderSearchMarkdown
	RenderPublicationMarkdown         = renderPublicationMarkdown
	RenderPhenotypeSearchMarkdown     = renderPhenotypeSearchMarkdown
	RenderPhenotypeStrainListMarkdown = renderPhenotypeStrainListMarkdown
)
