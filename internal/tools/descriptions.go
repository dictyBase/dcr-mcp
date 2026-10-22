package tools

// Tool names, human-readable titles, and model-facing descriptions.
// Descriptions say *when* to pick each tool, because web chat clients choose
// tools from these strings alone.
const (
	toolAutocomplete  = "autocomplete_stocks"
	titleAutocomplete = "Autocomplete stock names and IDs"
	descAutocomplete  = "Fast type-ahead lookup of Dicty Stock Center strain and plasmid names or IDs " +
		"from a partial string of 3 or more characters. Use when the user gives a partial name or ID. " +
		"For free-text questions use search_stocks instead."

	toolSearch  = "search_stocks"
	titleSearch = "Search the Dicty Stock Center"
	descSearch  = "Full-text search of the Dicty Stock Center catalog (Dictyostelium strains with DBS IDs " +
		"and plasmids with DBP IDs) across names, genotypes, summaries, and phenotypes. " +
		"Use for discovery questions. Returns ranked hits with availability."

	toolGetStrain  = "get_strain"
	titleGetStrain = "Get a strain record"
	descGetStrain  = "Fetch one Dictyostelium strain record by its DBS ID (for example DBS0350840): " +
		"genotype, phenotypes, depositor, publications, and whether it is in stock."

	toolGetPlasmid  = "get_plasmid"
	titleGetPlasmid = "Get a plasmid record"
	descGetPlasmid  = "Fetch one plasmid record by its DBP ID (for example DBP0391806): " +
		"summary, keywords, genes, GenBank accession, depositor, publications, and availability."

	toolListPhenotype  = "list_strains_with_phenotype"
	titleListPhenotype = "List strains with a phenotype"
	descListPhenotype  = "List strains annotated with one exact phenotype term (for example " +
		"'increased growth rate'), paged with cursor and next_cursor. The term must match exactly; " +
		"pass next_cursor back as cursor to read the next page."
)
