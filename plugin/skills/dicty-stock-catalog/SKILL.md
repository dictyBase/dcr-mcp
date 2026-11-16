---
name: dicty-stock-catalog
description: Use when a user wants to find or compare Dictyostelium strains or plasmids, read Dicty Stock Center records, search phenotype annotations, or check stock availability.
---

# Dicty Stock Center catalog

- Use the connected DCR tools for catalog facts and cite the URL returned with each stock. If the connection fails, report that the live lookup could not be completed.
- Use `get_strain` for a DBS strain ID and `get_plasmid` for a DBP plasmid ID. Start with summary detail. Request full detail when the needed fields are missing from the summary.
- Use `search_stocks` for general discovery and filter by STRAIN or PLASMID when the user specifies the record type. Use `autocomplete_stocks` for partial labels and IDs of at least three characters. If several stocks share a label, include their IDs and available distinguishing details.
- For a named gene or DDB_G ID, use `find_stocks_by_gene`. Treat plasmid results as text matches. Confirm a claimed gene association against the plasmid record and describe uncertainty when the record lacks a curated link.
- Use `search_phenotypes` for free-text phenotype, environment, or assay queries. For an exact phenotype list, read the phenotype field of a matching strain with `get_strain` and pass that label to `list_strains_with_phenotype`. Follow `next_cursor` until the requested page or the end. Interpret `page_count` as a count for that page.
- Use `check_availability` for current stock status, up to 25 IDs per call. Identify missing records and individual lookup errors separately from catalog items marked out of stock.
- Use `get_publication` with the PubMed ID from a stock record when the user asks about a cited paper. In a research surface that uses document search, pass IDs returned by `search` to `fetch`.
- The catalog integration supports read-only lookups. Direct users who ask to place an order, deposit a stock, or change a record to the Stock Center's own workflow. Ask for an ID, name, or search phrase when a requested lookup has no identifiable subject.
