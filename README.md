# dcr-mcp

`dcr-mcp` connects your AI assistant to the dictyBase Dicty Stock Center. Ask for *Dictyostelium* strains, plasmids, genes, phenotypes, and publications in plain language.

## Table of Contents

- [What dcr-mcp Does](#what-dcr-mcp-does)
- [Ask Questions in Plain Language](#ask-questions-in-plain-language)
  - [Find Strains and Plasmids for a Gene](#find-strains-and-plasmids-for-a-gene)
  - [Search the Catalog by Keyword or Topic](#search-the-catalog-by-keyword-or-topic)
  - [Look Up a Name or an ID Fast](#look-up-a-name-or-an-id-fast)
  - [Check If Stocks Are Available to Order](#check-if-stocks-are-available-to-order)
  - [Get the Record for a Strain](#get-the-record-for-a-strain)
  - [Get the Record for a Plasmid](#get-the-record-for-a-plasmid)
  - [Get a Plasmid DNA Sequence as FASTA](#get-a-plasmid-dna-sequence-as-fasta)
  - [Find Strains by Phenotype](#find-strains-by-phenotype)
  - [Read More Pages of Phenotype Results](#read-more-pages-of-phenotype-results)
  - [Read a Cited Publication](#read-a-cited-publication)
  - [Run a Multi-Step Research Workflow](#run-a-multi-step-research-workflow)
- [Ready-Made Research Prompts](#ready-made-research-prompts)
- [Tips for Better Answers](#tips-for-better-answers)
- [What Each Answer Contains](#what-each-answer-contains)
- [Connect Your AI Assistant](#connect-your-ai-assistant)
  - [Claude.ai and Claude Desktop](#claudeai-and-claude-desktop)
  - [ChatGPT](#chatgpt)
  - [Claude Code](#claude-code)
  - [OpenAI Codex CLI](#openai-codex-cli)
  - [Cursor](#cursor)
  - [Windsurf](#windsurf)
  - [VS Code with GitHub Copilot](#vs-code-with-github-copilot)
  - [Any Other MCP Client](#any-other-mcp-client)
  - [Confirm the Connection](#confirm-the-connection)
- [Install and Run the Server](#install-and-run-the-server)
  - [Build from Source](#build-from-source)
  - [Server Options](#server-options)
  - [Deploy for Web Clients](#deploy-for-web-clients)
- [Troubleshooting](#troubleshooting)

## What dcr-mcp Does

The Dicty Stock Center keeps thousands of *Dictyostelium* strains and plasmids. Each record has an ID, a genotype, phenotypes, a depositor, linked publications, and a stock center page URL. `dcr-mcp` makes that catalog available to any AI assistant or coding agent that speaks the Model Context Protocol (MCP).

After you connect `dcr-mcp`, you can type a question in your normal chat window. Your assistant picks the correct lookup, reads the answer, and replies in prose with clickable citation links. You do not write a query language, and you do not open a browser.

`dcr-mcp` gives your assistant 12 tools:

| Ability | Tool name | Use it for |
| :--- | :--- | :--- |
| Short name suggestions | `autocomplete_stocks` | Fast guesses from 3 or more letters. |
| Full catalog search | `search_stocks` | Broad discovery across names, genotypes, summaries, and phenotypes. |
| Stocks for a gene | `find_stocks_by_gene` | All strains and matching plasmids for a gene name or `DDB_G` ID. |
| One strain record | `get_strain` | Summary or full record for one `DBS...` ID, with its stock page URL. |
| One plasmid record | `get_plasmid` | Summary or full record for one `DBP...` ID, with its stock page URL. |
| Plasmid DNA sequence | `get_plasmid_sequence` | DNA sequence as FASTA for one `DBP...` ID. |
| Exact phenotype terms | `suggest_phenotype_terms` | Turn loose wording into exact catalog phenotype terms. |
| Strains for one phenotype | `list_strains_with_phenotype` | A paged list of strains annotated with one exact phenotype term. |
| Availability check | `check_availability` | In-stock status and stock page URLs for up to 25 IDs at once. |
| Publication | `get_publication` | Title, authors, abstract, and PubMed/DOI links for a PubMed ID. |
| Document search | `search` | ChatGPT deep research and other search/fetch clients. |
| Document fetch | `fetch` | Full Markdown document for a `DBS...` or `DBP...` ID returned by `search`. |

You do not call these tool names yourself. This guide names them only so that you can recognize them in the tool log of your assistant.

The data is read-only. Every tool carries MCP read-only annotations (`readOnlyHint: true`), so web clients can run lookups without confirmation prompts. `dcr-mcp` cannot change, add, or delete a stock center record.

## Ask Questions in Plain Language

This is the primary way to use `dcr-mcp`. You write a normal question. Your assistant chooses the lookup, reads the result, and answers you with links to the stock center and PubMed.

### Find Strains and Plasmids for a Gene

Use this when you know a gene name (such as `pkaC` or `acaA`) or a `DDB_G...` gene ID and want to see mutants, knockouts, or constructs for it.

**Type one of these:**

> What strains and plasmids are available for the gene pkaC?

> Find every knockout or mutant strain linked to DDB_G0283907.

> Which stocks exist for acaA, and which ones are currently in stock?

> What are the synonyms and gene product for pkaC, and what strains carry mutations in it?

**What you get back:** The gene ID, product description, and synonyms, followed by every strain curated to that gene (with its stock page URL and availability) and any plasmids whose names match the gene. Plasmid matches are text matches on the plasmid name, so ask your assistant to inspect a plasmid with `get_plasmid` before you order it.

**Behind the scenes:** `find_stocks_by_gene`.

### Search the Catalog by Keyword or Topic

Use this when you know a research topic, a construct name, a depositor, or a phrase from a description, rather than a specific gene or stock ID.

**Type one of these:**

> Search the stock center for plasmids with GFP in the description.

> Find strains that came from a restriction enzyme mediated integration experiment.

> Show me stocks from the Knecht lab.

> Search for act15, but only plasmids.

**What you get back:** A Markdown table and structured list of ranked matches (15 by default, up to 100 on request). Each match includes the stock ID, clickable stock center URL, label, matched field, summary, and in-stock status.

**Behind the scenes:** `search_stocks` (or `search` in ChatGPT deep research mode).

### Look Up a Name or an ID Fast

Use this when you have the beginning of a name or ID and want quick candidate IDs rather than a full report.

**Type one of these:**

> Autocomplete Dicty stocks that start with act15.

> Give me 10 quick suggestions for pDGB.

> What stock IDs match axeA?

**What you get back:** A short list of suggestions (5 by default, up to 50) with the stock ID, stock kind, matched field, and display text.

**Behind the scenes:** `autocomplete_stocks`. It requires at least 3 characters.

### Check If Stocks Are Available to Order

Use this before you place an order on the Dicty Stock Center website, especially when you have a list of candidate IDs.

**Type one of these:**

> Check if DBS0350840, DBS0236783, and DBP0391806 are in stock.

> Is strain DBS0350840 available to order right now?

> Which of these strains are in stock: DBS0350840, DBS0349965, DBS0235559?

**What you get back:** A batch report for up to 25 strain (`DBS...`) and plasmid (`DBP...`) IDs at once, showing each stock's label, availability, and direct stock page URL.

**Behind the scenes:** `check_availability` for lists of IDs, or `get_strain` / `get_plasmid` for a single ID.

### Get the Record for a Strain

Use this when you have a strain ID (`DBS...`).

**Type one of these:**

> Give me the record for DBS0350840.

> What is the genotype of DBS0350840?

> Show me every phenotype and publication for strain DBS0350840 in full detail.

> Who deposited DBS0350840, and what species is it?

**What you get back:** A Markdown summary with citation links plus the structured strain record: ID, label, summary, species, systematic name, availability, genotypes, names, characteristics, mutagenesis method, genetic modification, depositor, linked genes, phenotypes, publications, and the stock page URL. By default (`detail=summary`), phenotypes are capped at 10 and publications at 5 (`phenotype_count` and `publication_count` report the totals). Ask for "full detail" (`detail=full`) when you want every annotation.

**Behind the scenes:** `get_strain` (or `fetch` in ChatGPT deep research mode).

### Get the Record for a Plasmid

Use this when you have a plasmid ID (`DBP...`).

**Type one of these:**

> Show me the record for plasmid DBP0391806.

> Which genes are linked to plasmid DBP0391557, and who deposited it?

> How long is the DNA sequence for DBP0000242, and what is its GenBank accession?

> Which publications cite plasmid DBP0391806?

**What you get back:** A Markdown summary and structured record with the ID, name, summary, keywords, GenBank accession, availability, depositor, linked genes, publications, `sequence_length` in base pairs, and the stock page URL. By default (`detail=summary`), the raw DNA sequence is omitted so chat responses stay compact.

**Behind the scenes:** `get_plasmid` (or `fetch` in ChatGPT deep research mode).

### Get a Plasmid DNA Sequence as FASTA

Use this when you need the actual nucleotide sequence for cloning, primer design, or alignment.

**Type one of these:**

> Give me the FASTA sequence for plasmid DBP0000242.

> Fetch the DNA sequence of DBP0391806 so I can design primers.

**What you get back:** The plasmid ID, name, length in base pairs, stock page URL, and single-line FASTA sequence (`>DBP... name`). If the catalog has no sequence on record for that plasmid, the assistant tells you clearly.

**Behind the scenes:** `get_plasmid_sequence`.

### Find Strains by Phenotype

Phenotype annotations in the catalog use controlled terms (for example, `increased growth rate`). You can describe a phenotype in everyday words and let your assistant find the exact catalog term first.

**Type one of these:**

> What phenotype terms in the catalog match "fast growth" or "growth rate"?

> Find Dicty strains with a defect in chemotaxis.

> List Dicty strains with the phenotype increased growth rate.

> Which strains show aberrant cytokinesis? Show 25 of them.

**What you get back:** First, candidate exact terms from the catalog. Then, a paged list of matching strains (10 per page by default, up to 100) with each strain's ID, stock page URL, label, summary, availability, genes, publications (including DOIs), and `total_count`.

**Behind the scenes:** `suggest_phenotype_terms` followed by `list_strains_with_phenotype`.

### Read More Pages of Phenotype Results

Phenotype strain lists arrive one page at a time and include `next_cursor`.

**Type one of these:**

> Show me the next page of those phenotype results.

> Keep paging through the strains for increased growth rate until you have 50 strains, then group them by gene.

**What you get back:** Additional pages from the same phenotype query until `next_cursor` is `0`.

**Behind the scenes:** `list_strains_with_phenotype` with `cursor` set to the previous `next_cursor`.

### Read a Cited Publication

Strain and plasmid records list publications by their PubMed ID. Use this when you want to know what a cited paper reports.

**Type one of these:**

> What does publication 17177856 report?

> Read the abstract of the paper cited on strain DBS0350840.

> Give me the full citation, abstract, and DOI link for PubMed ID 21760904.

**What you get back:** A Markdown summary and structured record with the title, full author list, journal, volume, pages, publication year, abstract, PubMed URL, and DOI URL.

**Behind the scenes:** `get_publication`.

### Run a Multi-Step Research Workflow

You can combine several questions into one instruction and let your assistant chain the tools automatically.

**Type one of these:**

> Find every stock for the gene pkaC, check which strains are in stock, and give me a Markdown table with the ID, label, genotype, key phenotypes, and stock center link.

> Compare DBS0350840 and DBS0349965 side by side: genotype, mutagenesis method, shared phenotypes, unique phenotypes, and availability.

> Check availability for DBS0350840, DBS0236783, and DBP0391806. For any stock that is unavailable, suggest up to two in-stock alternatives.

> Find strains for triA, read the abstract of the primary publication, and summarize how the mutant was constructed.

## Ready-Made Research Prompts

Clients such as Claude.ai, Claude Desktop, and coding agents expose four built-in MCP prompt templates (available in the `+` menu or slash-command picker):

| Prompt name | Required argument | What it does |
| :--- | :--- | :--- |
| `find_stocks_for_gene` | `gene` (e.g. `pkaC`) | Finds every strain and plasmid for a gene, summarizes in-stock strains in a table, and recommends which to order first. |
| `compare_strains` | `ids` (e.g. `DBS0350840, DBS0236783`) | Fetches full records and builds a side-by-side comparison table plus shared and unique phenotypes. |
| `phenotype_survey` | `phenotype` (e.g. `fast growth`) | Resolves the exact phenotype term, pages through matching strains, and groups them by gene. |
| `order_checklist` | `ids` (comma-separated IDs) | Checks batch availability and suggests in-stock alternatives for unavailable items. |

## Tips for Better Answers

| Situation | What to do |
| :--- | :--- |
| You know a gene symbol or `DDB_G` ID | Name the gene directly. Your assistant will call `find_stocks_by_gene` to resolve the gene and list its strains and plasmids. |
| Your search word has 2 letters (such as `cs`) | Ask for a full search (`search_stocks`) or gene lookup (`find_stocks_by_gene`). Quick autocomplete requires at least 3 letters. |
| You have `DBS...` or `DBP...` IDs | Give the IDs directly. Use `check_availability` for a list of IDs, or `get_strain` / `get_plasmid` for one record. |
| You only need the plasmid overview, or you need the DNA sequence | Ask for the plasmid record first (sequence omitted by default). Ask for the FASTA sequence only when you need the bases. |
| A phenotype phrase is approximate | Ask your assistant to suggest exact phenotype terms first (`suggest_phenotype_terms`), then list strains for the chosen term. |
| You want clickable sources | Ask the assistant to include the stock page URL, PubMed link, and DOI link from each tool response. |

## What Each Answer Contains

| Tool | Key fields returned | Notes and limits |
| :--- | :--- | :--- |
| `autocomplete_stocks` | `suggestions` (`id`, `entity`, `field`, `display_text`, `score`), `limit`, `total_count` | Min 3 chars. Default 5, max 50. |
| `search_stocks` | Markdown table + `results` (`id`, `url`, `entity`, `field`, `display_text`, `summary`, `strain_label`, `in_stock`), `limit`, `total_count` | Min 2 chars. Default 15, max 100. |
| `find_stocks_by_gene` | `gene` (`id`, `gene_product`, `synonyms`, `name_description`), `strains` (with `url` and `in_stock`), `plasmids` (with `url`), `plasmid_match_note` | Min 2 chars. Resolves gene name or `DDB_G` ID. |
| `get_strain` | Markdown summary + wire record, `url`, `detail`, `phenotype_count`, `publication_count` | `detail=summary` (default: up to 10 phenotypes, 5 publications) or `detail=full`. |
| `get_plasmid` | Markdown summary + wire record, `url`, `detail`, `sequence_length`, `publication_count` | `detail=summary` (default: omits DNA sequence, up to 5 publications) or `detail=full`. |
| `get_plasmid_sequence` | `id`, `name`, `length`, `format` (`fasta`), `fasta`, `url` | Returns an error if the plasmid has no sequence on record. |
| `suggest_phenotype_terms` | `query`, `terms`, `note` | Min 3 chars. Default 10 terms, max 25. |
| `list_strains_with_phenotype` | `strains` (with `url`, `in_stock`, `genes`, `publications` including `doi`), `next_cursor`, `total_count` | Exact phenotype term. Default 10 per page, max 100. |
| `check_availability` | `items` (`id`, `entity`, `label`, `in_stock`, `url`, or `error`) | 1 to 25 IDs per call, deduplicated and checked concurrently. |
| `get_publication` | Markdown summary + `id`, `title`, `authors`, `journal`, `volume`, `pages`, `pub_date`, `abstract`, `pubmed_url`, `doi_url` | Takes a PubMed ID. |
| `search` / `fetch` | `search`: `results` (`id`, `title`, `url`). `fetch`: `id`, `title`, `text` (Markdown), `url`, `metadata` | ChatGPT deep-research document contract. |

## Connect Your AI Assistant

Point your client at the MCP endpoint `<SERVER_URL>/mcp` (for example, your hosted HTTPS endpoint for web clients, or a local address when running on your own machine).

*Note: Menu paths for web clients reflect the Claude.ai and ChatGPT interfaces as of October 2026. Check [Anthropic's MCP connector guide](https://support.anthropic.com) and [OpenAI's MCP documentation](https://platform.openai.com/docs/mcp) if a menu label changes.*

### Claude.ai and Claude Desktop

1. Open **Settings → Connectors → Add custom connector**.
2. Set the name to `Dicty Stock Center` and the URL to `<SERVER_URL>/mcp`. No authentication is required.
3. Enable the connector in a chat from the tools menu. Built-in prompts appear in the **+** menu.

### ChatGPT

1. Open **Settings → Apps & Connectors → Advanced** and enable **Developer mode**.
2. Click **Create**, set the MCP server URL to `<SERVER_URL>/mcp`, and set authentication to **None**.
3. `dcr-mcp` exposes both the domain tools and the `search` / `fetch` pair required by ChatGPT deep research.

### Claude Code

```bash
claude mcp add --transport http --scope user dcr-mcp <SERVER_URL>/mcp
```

Or add `.mcp.json` in a project root:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "type": "http",
      "url": "<SERVER_URL>/mcp"
    }
  }
}
```

### OpenAI Codex CLI

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.dcr-mcp]
url = "<SERVER_URL>/mcp"
```

Or run:

```bash
codex mcp add dcr-mcp --url <SERVER_URL>/mcp
```

### Cursor

Add to `.cursor/mcp.json` or `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "url": "<SERVER_URL>/mcp"
    }
  }
}
```

### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "serverUrl": "<SERVER_URL>/mcp"
    }
  }
}
```

### VS Code with GitHub Copilot

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "dcr-mcp": {
      "type": "http",
      "url": "<SERVER_URL>/mcp"
    }
  }
}
```

### Any Other MCP Client

Any client that supports remote MCP over Streamable HTTP can connect to `<SERVER_URL>/mcp`:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "type": "http",
      "url": "<SERVER_URL>/mcp"
    }
  }
}
```

### Confirm the Connection

After connecting, ask your assistant:

> List the tools you have from dcr-mcp.

Your assistant will list the 12 tools shown in [What dcr-mcp Does](#what-dcr-mcp-does).

## Install and Run the Server

If you are connecting to a shared or public `dcr-mcp` deployment, you do not need this section. Follow these steps only when running or hosting the server yourself.

### Build from Source

Requires Go 1.26 or later:

```bash
go build -o dcr-mcp .
./dcr-mcp start-server
```

The `serve` and `start` subcommands are aliases for `start-server`. Check `/healthz` to confirm the server is healthy (`{"status":"ok"}`).

### Server Options

| Flag | Environment variable | Default | What it controls |
| :--- | :--- | :--- | :--- |
| `--graphql-endpoint`, `--endpoint` | `GRAPHQL_ENDPOINT` | `https://graphql.dictybase.dev/graphql` | Upstream dictyBase GraphQL endpoint. |
| `--host` | `MCP_HOST` | `0.0.0.0` | Bind address for the HTTP server. |
| `--port` | `MCP_PORT` | `8080` | TCP port (`0`–`65535`). |
| `--stateless` | `MCP_STATELESS` | `true` | Stateless Streamable HTTP mode so any replica can serve any request. |
| `--allowed-origins` | `ALLOWED_ORIGINS` | `*` | Allowed CORS origins for browser-based MCP clients. |
| `--rate-limit` | `MCP_RATE_LIMIT` | `120` | Maximum `/mcp` requests per client IP per minute (`>= 1`). |

### Deploy for Web Clients

1. **Public HTTPS is required for cloud chat clients.** Claude.ai and ChatGPT connect from their own cloud servers and cannot reach a private machine directly. Place `dcr-mcp` behind a TLS-terminating reverse proxy, ingress, or managed container platform and expose `<SERVER_URL>/mcp` over HTTPS.
2. **Container image:** Build the distroless image with `docker build -f build/package/Dockerfile -t dcr-mcp .` and run `docker run -p 8080:8080 dcr-mcp start-server`.
3. **Horizontal scaling:** Keep `--stateless` (`true` by default) so any replica can handle any request without sticky sessions.
4. **Rate limiting:** `--rate-limit` / `MCP_RATE_LIMIT` defaults to 120 requests per client IP per minute on `/mcp` (`/healthz` is exempt). The middleware resolves `X-Forwarded-For` from trusted private/loopback proxy hops and falls back to the peer address when reached directly.
5. **CORS:** Only browser-based MCP clients require CORS. Server-to-server connectors (Claude.ai, ChatGPT, coding CLIs) ignore CORS headers. Restrict `--allowed-origins` to known web origins when needed.
6. **In-memory caching:** Single-record `get_strain` and `get_plasmid` lookups are cached in memory for 10 minutes (up to 1,024 entries each) per replica. Failed lookups are never cached.
7. **Structured logging:** The server writes JSON log lines to `stderr` with `method`, `tool`, `duration_ms`, `tool_error`, and `error`. Tool arguments and results are never logged.

## Troubleshooting

| Problem | Cause and fix |
| :--- | :--- |
| Your assistant does not list the tools | Confirm `<SERVER_URL>/healthz` returns `{"status":"ok"}` and verify that the configured connector URL ends with `/mcp`. Restart desktop/CLI clients after editing their config file. |
| `query must be at least 3 characters` | `autocomplete_stocks` or `suggest_phenotype_terms` received fewer than 3 characters. Use `search_stocks` or `find_stocks_by_gene` for 2-character queries. |
| `gene "..." not found in dictyBase` | The gene symbol or `DDB_G` ID is not in `geneGeneralInformation`. Try `search_stocks` with a keyword or synonym. |
| `plasmid ... has no sequence on record` | The catalog entry for that plasmid does not include a DNA sequence. Check `sequence_length` on `get_plasmid` or inspect the GenBank accession if present. |
| A phenotype list is empty | `list_strains_with_phenotype` requires an exact annotation term. Call `suggest_phenotype_terms` first to discover exact terms. |
| HTTP `429 Too Many Requests` | The client exceeded `--rate-limit` requests per minute on `/mcp`. Wait for the 1-minute window to reset or raise `MCP_RATE_LIMIT`. |

## CI

Workflows: test on push to develop, lint and coverage on pull requests, dependabot, and Kodiak automerge for dependabot upgrades.
