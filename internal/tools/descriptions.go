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

	toolSearchPhenotypes  = "search_phenotypes"
	titleSearchPhenotypes = "Search strains by phenotype statement"
	descSearchPhenotypes  = "Find Dictyostelium strains by phenotype statement from free words. " +
		"A phenotype statement has a phenotype, an environment, and an assay part (for example " +
		"'abolished vegetative growth', 'in suspension', 'motility assay'). The query matches the label, " +
		"synonyms, definition, and comments of any part. Each result row is one strain. matched_texts " +
		"groups the rows by the matched text. Returns at most 50 rows and has no further pages. " +
		"A matched text can be an environment, an assay, or a definition, so it is not always a phenotype " +
		"term: call get_strain to read the phenotype field before you call list_strains_with_phenotype."

	toolGetStrain  = "get_strain"
	titleGetStrain = "Get a strain record"
	descGetStrain  = "Fetch one Dictyostelium strain record by its DBS ID (for example DBS0350840): " +
		"genotype, phenotypes, depositor, publications, availability, and the stock page url. " +
		"Defaults to detail=summary (at most 10 phenotypes and 5 publications; phenotype_count and " +
		"publication_count give the totals). Pass detail=full only when the user needs every annotation. " +
		"The phenotype field of each phenotype statement is the valid input for list_strains_with_phenotype."

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
	titleListPhenotype = "List strains with an exact phenotype term"
	descListPhenotype  = "List strains annotated with one exact phenotype term. phenotype_term must be the " +
		"exact, case-sensitive label of a Dicty Phenotypes ontology term, copied from the phenotype field " +
		"of get_strain. Environment terms, assay terms, definitions, and free words are rejected with an " +
		"error. To find strains from free words, or by environment or assay, use search_phenotypes. " +
		"Pass next_cursor back as cursor. has_more=false (next_cursor 0) means the last page. " +
		"page_count is the number of strains on this page, not a catalog total."

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
	descCheckAvailability  = "Check whether up to 25 strains (DBS ids) and/or plasmids (DBP ids) are in stock, " +
		"each with a stock page url. Use before recommending stocks to order."

	toolGetPublication  = "get_publication"
	titleGetPublication = "Get a publication"
	descGetPublication  = "Fetch one publication by PubMed id (as listed on strain and plasmid records): title, " +
		"authors, journal, abstract, and PubMed/DOI links. Use when the user asks what a paper " +
		"behind a stock reports."
)
