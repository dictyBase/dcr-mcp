package tools

// Test-only exports of unexported functions for package tools_test.
var (
	StrainViewFor                 = strainViewFor
	PlasmidViewFor                = plasmidViewFor
	ToSearchResultView            = toSearchResultView
	ToPhenotypeSearchView         = toPhenotypeSearchView
	ToPhenotypeStrainListView     = toPhenotypeStrainListView
	RenderStrainMarkdown          = renderStrainMarkdown
	RenderPlasmidMarkdown         = renderPlasmidMarkdown
	RenderSearchMarkdown          = renderSearchMarkdown
	RenderPhenotypeSearchMarkdown = renderPhenotypeSearchMarkdown
)
