# dcr-mcp

`dcr-mcp` connects your AI assistant to the dictyBase Dicty Stock Center. Ask about *Dictyostelium* strains, plasmids, genes, phenotypes, and publications in plain language.

## Table of Contents

- [What dcr-mcp Does](#what-dcr-mcp-does)
- [Ask Questions in Plain Language](#ask-questions-in-plain-language)
  - [Find Strains and Plasmids for a Gene](#find-strains-and-plasmids-for-a-gene)
  - [Search the Catalog](#search-the-catalog)
  - [Find a Name or an ID](#find-a-name-or-an-id)
  - [Check Stock Availability](#check-stock-availability)
  - [Read a Strain Record](#read-a-strain-record)
  - [Read a Plasmid Record](#read-a-plasmid-record)
  - [Find Strains by Phenotype](#find-strains-by-phenotype)
  - [Read More Pages of Phenotype Results](#read-more-pages-of-phenotype-results)
  - [Read a Cited Publication](#read-a-cited-publication)
  - [Run a Multi-Step Research Workflow](#run-a-multi-step-research-workflow)
- [Ready-Made Research Prompts](#ready-made-research-prompts)
- [Tips for Better Answers](#tips-for-better-answers)
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

## What dcr-mcp Does

The Dicty Stock Center keeps *Dictyostelium* strains and plasmids. Each record has an ID, a genotype, phenotypes, a depositor, publications, and a stock page URL.

`dcr-mcp` makes this catalog available to AI assistants that use the Model Context Protocol (MCP). Type a question in your chat window. The assistant selects the correct lookup and answers with links.

`dcr-mcp` gives your assistant 11 tools:

| Ability | Tool name | Use it for |
| :--- | :--- | :--- |
| Name suggestions | `autocomplete_stocks` | Fast suggestions from 3 or more letters. |
| Catalog search | `search_stocks` | Broad search of names, genotypes, summaries, and phenotypes. |
| Stocks for a gene | `find_stocks_by_gene` | Strains and plasmids for a gene name or `DDB_G` ID. |
| Strain record | `get_strain` | One `DBS...` record. |
| Plasmid record | `get_plasmid` | One `DBP...` record. |
| Phenotype terms | `suggest_phenotype_terms` | Exact catalog terms for a phenotype. |
| Strains by phenotype | `list_strains_with_phenotype` | Strains that have one exact phenotype term. |
| Availability check | `check_availability` | Stock status of up to 25 IDs. |
| Publication | `get_publication` | One publication, selected by PubMed ID. |
| Document search | `search` | Search for ChatGPT deep research. |
| Document fetch | `fetch` | One record, selected by an ID from `search`. |

You do not call these tools yourself. The names help you find them in the tool log of your assistant.

The tools are read-only. `dcr-mcp` cannot change a stock center record.

## Ask Questions in Plain Language

Write a normal question. Your assistant selects the lookup and answers with links to the stock center and PubMed.

### Find Strains and Plasmids for a Gene

Use a gene name (for example `pkaC`) or a `DDB_G...` gene ID.

**Type one of these:**

> What strains and plasmids are available for the gene pkaC?

> Find every strain linked to DDB_G0283907.

> Which stocks exist for pkaC, and which stocks are in stock?

> What are the synonyms and the gene product for pkaC?

**Behind the scenes:** `find_stocks_by_gene`.

Plasmid matches are text matches on the plasmid name. Ask your assistant to read the plasmid record before you order it.

### Search the Catalog

Use a research topic, a construct name, a depositor, or a phrase from a description.

**Type one of these:**

> Search the stock center for strains with GFP in the genotype.

> Search for act15, but only plasmids.

> Show stocks that mention Knecht.

**Behind the scenes:** `search_stocks`, or `search` in ChatGPT deep research.

### Find a Name or an ID

Use the first letters of a name or an ID. Enter at least 3 characters.

**Type one of these:**

> Autocomplete Dicty stocks that start with act15.

> Give me 10 quick suggestions for pDGB.

**Behind the scenes:** `autocomplete_stocks`.

### Check Stock Availability

Check availability before you order from the Dicty Stock Center website. You can check up to 25 IDs at the same time.

**Type one of these:**

> Check if DBS0350840, DBS0236783, and DBP0391806 are in stock.

> Is strain DBS0350840 available to order?

**Behind the scenes:** `check_availability`.

### Read a Strain Record

Use a strain ID (`DBS...`).

**Type one of these:**

> Show the record for DBS0350840.

> What is the genotype of DBS0350840?

> Show every phenotype and publication for strain DBS0350707 in full detail.

> Who deposited DBS0350840?

**Behind the scenes:** `get_strain`, or `fetch` in ChatGPT deep research.

By default, the record shows a summary. Ask for "full detail" to see every annotation.

### Read a Plasmid Record

Use a plasmid ID (`DBP...`).

**Type one of these:**

> Show the record for plasmid DBP0391806.

> Which genes are linked to plasmid DBP0391557?

> Who deposited plasmid DBP0391806?

> Which publications cite plasmid DBP0391806?

**Behind the scenes:** `get_plasmid`, or `fetch` in ChatGPT deep research.

### Find Strains by Phenotype

The catalog uses controlled phenotype terms. Describe the phenotype in your own words. The assistant finds the exact term first.

**Type one of these:**

> Which phenotype terms in the catalog match "chemo"?

> Find Dicty strains with aberrant chemotaxis.

> List Dicty strains with the phenotype aberrant chemotaxis.

**Behind the scenes:** `suggest_phenotype_terms`, then `list_strains_with_phenotype`.

### Read More Pages of Phenotype Results

Phenotype lists arrive in pages. Ask for the next page.

**Type one of these:**

> Show the next page of those phenotype results.

> Read all pages for aberrant chemotaxis, then group the strains by gene.

**Behind the scenes:** `list_strains_with_phenotype` with `cursor`.

### Read a Cited Publication

Strain and plasmid records list publications by PubMed ID. Use the ID to read the paper details.

**Type one of these:**

> What does publication 17177856 report?

> Read the abstract of the paper cited on strain DBS0350840.

**Behind the scenes:** `get_publication`.

### Run a Multi-Step Research Workflow

Combine several questions in one instruction. Your assistant chains the tools automatically.

**Type one of these:**

> Find every stock for the gene pkaC. Check which strains are in stock. Give me a table with the ID, label, and stock center link.

> Compare DBS0350840 and DBS0349965 side by side: genotype, depositor, and availability.

> Check availability for DBS0350840, DBS0236783, and DBP0391806. Suggest in-stock alternatives for each unavailable stock.

> Find strains for triA, read the abstract of the primary publication, and summarize it.

## Ready-Made Research Prompts

Some clients, such as Claude.ai, Claude Desktop, and coding agents, show four built-in prompt templates. Open them from the `+` menu or the slash-command picker.

| Prompt name | Required argument | Purpose |
| :--- | :--- | :--- |
| `find_stocks_for_gene` | `gene` (for example `pkaC`) | Finds the strains and plasmids for a gene and recommends which to order first. |
| `compare_strains` | `ids` (for example `DBS0350840, DBS0236783`) | Compares strains in a side-by-side table. |
| `phenotype_survey` | `phenotype` (for example `aberrant chemotaxis`) | Finds the exact phenotype term and groups the matching strains by gene. |
| `order_checklist` | `ids` (comma-separated IDs) | Checks availability and suggests in-stock alternatives. |

## Tips for Better Answers

| Situation | Action |
| :--- | :--- |
| You know a gene symbol or `DDB_G` ID | Name the gene. Your assistant calls `find_stocks_by_gene`. |
| Your search word has 2 letters | Ask for a full search (`search_stocks`) or a gene lookup (`find_stocks_by_gene`). Autocomplete needs 3 or more letters. |
| You have `DBS...` or `DBP...` IDs | Give the IDs. Use `check_availability` for a list of IDs. Use `get_strain` or `get_plasmid` for one record. |
| A phenotype phrase is approximate | Ask your assistant to suggest exact phenotype terms first. Then list the strains for the selected term. |
| You want clickable sources | Ask your assistant to include the stock page URL, the PubMed link, and the DOI link. |

## Connect Your AI Assistant

Point your client at the MCP endpoint `<SERVER_URL>/mcp`. For web clients, use your hosted HTTPS endpoint. For a local server, use a local address.

*Note: Menu paths for web clients reflect the Claude.ai and ChatGPT interfaces as of October 2026. Check [Anthropic's MCP connector guide](https://support.anthropic.com) and [OpenAI's MCP documentation](https://platform.openai.com/docs/mcp) if a menu label changes.*

### Claude.ai and Claude Desktop

1. Open **Settings → Connectors → Add custom connector**.
2. Set the name to `Dicty Stock Center` and the URL to `<SERVER_URL>/mcp`. Authentication is not necessary.
3. Enable the connector in a chat from the tools menu. The built-in prompts show in the **+** menu.

### ChatGPT

1. Open **Settings → Apps & Connectors → Advanced** and enable **Developer mode**.
2. Click **Create**. Set the MCP server URL to `<SERVER_URL>/mcp` and set authentication to **None**.
3. `dcr-mcp` includes the `search` and `fetch` pair that ChatGPT deep research needs.

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

Add this to `~/.codex/config.toml`:

```toml
[mcp_servers.dcr-mcp]
url = "<SERVER_URL>/mcp"
```

Or run:

```bash
codex mcp add dcr-mcp --url <SERVER_URL>/mcp
```

### Cursor

Add this to `.cursor/mcp.json` or `~/.cursor/mcp.json`:

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

Add this to `~/.codeium/windsurf/mcp_config.json`:

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

Add this to `.vscode/mcp.json`:

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

Use any client that supports remote MCP over Streamable HTTP:

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

Restart the client after you edit its configuration file. Then ask your assistant:

> List the tools you have from dcr-mcp.

Your assistant lists the `dcr-mcp` tools. Confirm that `<SERVER_URL>/healthz` returns `{"status":"ok"}`.

## Install and Run the Server

Skip this section if you use a shared or public `dcr-mcp` server. Use it only to run or host the server yourself.

### Build from Source

Go 1.26 or later is necessary.

```bash
go build -o dcr-mcp .
./dcr-mcp start-server
```

The `serve` and `start` subcommands are aliases for `start-server`. Open `/healthz` to confirm that the server is healthy (`{"status":"ok"}`).

### Server Options

| Flag | Environment variable | Default | Purpose |
| :--- | :--- | :--- | :--- |
| `--graphql-endpoint`, `--endpoint` | `GRAPHQL_ENDPOINT` | `https://graphql.dictybase.dev/graphql` | Upstream dictyBase GraphQL endpoint. |
| `--host` | `MCP_HOST` | `0.0.0.0` | Bind address for the HTTP server. |
| `--port` | `MCP_PORT` | `8080` | TCP port (`0`–`65535`). |
| `--stateless` | `MCP_STATELESS` | `true` | Stateless Streamable HTTP mode. Any replica can serve any request. |
| `--allowed-origins` | `ALLOWED_ORIGINS` | `*` | Allowed CORS origins for browser-based MCP clients. |
| `--rate-limit` | `MCP_RATE_LIMIT` | `120` | Maximum `/mcp` requests for each client IP each minute (`>= 1`). |

### Deploy for Web Clients

1. **Use public HTTPS for cloud chat clients.** Claude.ai and ChatGPT connect from their own cloud servers. They cannot reach a private machine. Put `dcr-mcp` behind a TLS-terminating reverse proxy, an ingress, or a managed container platform. Expose `<SERVER_URL>/mcp` over HTTPS.
2. **Use the container image.** Build the distroless image with `docker build -f build/package/Dockerfile -t dcr-mcp .`. Run it with `docker run -p 8080:8080 dcr-mcp start-server`.
3. **Scale horizontally.** Keep `--stateless` set to `true`. Then any replica can handle any request without sticky sessions.
4. **Limit the request rate.** `--rate-limit` sets the maximum number of `/mcp` requests for each client IP each minute. The default is 120. `/healthz` is exempt. The middleware reads `X-Forwarded-For` from trusted private and loopback proxy hops. It uses the peer address when the client connects directly.
5. **Set CORS only for browsers.** Only browser-based MCP clients need CORS. Server-to-server connectors, such as Claude.ai, ChatGPT, and coding CLIs, ignore CORS headers. Set `--allowed-origins` to known web origins when necessary.
6. **Use the in-memory cache.** The server caches single-record `get_strain` and `get_plasmid` lookups for 10 minutes (up to 1,024 entries each) for each replica. The server does not cache failed lookups.
7. **Read the structured logs.** The server writes JSON log lines to `stderr` with `method`, `tool`, `duration_ms`, `tool_error`, and `error`. The server does not log tool arguments or results.

## CI

Workflows: test on push to develop, lint and coverage on pull requests, dependabot, and Kodiak automerge for dependabot upgrades.
