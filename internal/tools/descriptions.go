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
		"genotype, phenotypes, depositor, publications, availability, and the stock page url. " +
		"Defaults to detail=summary (at most 10 phenotypes and 5 publications; phenotype_count and " +
		"publication_count give the totals). Pass detail=full only when the user needs every annotation."

	toolGetPlasmid  = "get_plasmid"
	titleGetPlasmid = "Get a plasmid record"
	descGetPlasmid  = "Fetch one plasmid record by its DBP ID (for example DBP0391806): summary, keywords, " +
		"genes, GenBank accession, depositor, publications, availability, and the stock page url. " +
		"Defaults to detail=summary, which omits the DNA sequence (sequence_length gives its size). " +
		"Use get_plasmid_sequence when the user needs the sequence."

	toolGetPlasmidSequence  = "get_plasmid_sequence"
	titleGetPlasmidSequence = "Get a plasmid DNA sequence"
	descGetPlasmidSequence  = "Return the DNA sequence of one plasmid (DBP ID) as FASTA with its length. " +
		"Use only when the user explicitly needs the sequence (primer design, cloning, alignment); " +
		"sequences can be several kilobases."

	toolListPhenotype  = "list_strains_with_phenotype"
	titleListPhenotype = "List strains with a phenotype"
	descListPhenotype  = "List strains annotated with one exact phenotype term (for example " +
		"'increased growth rate'), paged with cursor and next_cursor. The term must match exactly; " +
		"pass next_cursor back as cursor to read the next page."

	toolChatGPTSearch  = "search"
	titleChatGPTSearch = "Search stocks (document search)"
	descChatGPTSearch  = "Search the Dicty Stock Center and return a list of matching strain and plasmid documents " +
		"(id, title, url). Pass a returned id to fetch to read the full record. " +
		"Designed for ChatGPT deep research and other clients that use the search/fetch convention."

	toolChatGPTFetch  = "fetch"
	titleChatGPTFetch = "Fetch a stock document"
	descChatGPTFetch  = "Fetch the full record for one strain (DBS id) or plasmid (DBP id) returned by search, " +
		"as a document with title, Markdown text, url, and metadata. " +
		"Designed for ChatGPT deep research and other clients that use the search/fetch convention."

	toolFindByGene  = "find_stocks_by_gene"
	titleFindByGene = "Find stocks for a gene"
	descFindByGene  = "Given a gene name (for example pkaC) or DDB_G id, return the gene's product and synonyms, " +
		"every strain linked to that gene, and plasmids whose names match it. " +
		"Use this first whenever the user asks about mutants, knockouts, or constructs for a gene."

	toolSuggestPhenoTerms  = "suggest_phenotype_terms"
	titleSuggestPhenoTerms = "Suggest exact phenotype terms"
	descSuggestPhenoTerms  = "Turn a loose phenotype description (for example 'grows faster' or 'aggregation') into " +
		"the exact phenotype terms used in the catalog. Call this before list_strains_with_phenotype, " +
		"which only accepts exact terms."

	toolCheckAvailability  = "check_availability"
	titleCheckAvailability = "Check stock availability"
	descCheckAvailability  = "Check whether up to 25 strains (DBS ids) and/or plasmids (DBP ids) are currently in stock, " +
		"with a stock page url for each. Use before recommending stocks to order."

	toolGetPublication  = "get_publication"
	titleGetPublication = "Get a publication"
	descGetPublication  = "Fetch one publication by PubMed id (as listed on strain and plasmid records): title, authors, " +
		"journal, abstract, and PubMed/DOI links. Use when the user asks what a paper behind a stock reports."
)
