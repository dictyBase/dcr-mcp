# dcr-mcp

`dcr-mcp` connects your coding assistant to the dictyBase Dicty Stock Center. Ask for *Dictyostelium* strains and plasmids in plain language.

## Table of Contents

- [What dcr-mcp Does](#what-dcr-mcp-does)
- [Quick Start](#quick-start)
  - [Step 1: Build the Program](#step-1-build-the-program)
  - [Step 2: Start the Server](#step-2-start-the-server)
  - [Step 3: Confirm That the Server Runs](#step-3-confirm-that-the-server-runs)
- [Connect Your Coding Agent](#connect-your-coding-agent)
  - [Claude Code](#claude-code)
  - [OpenAI Codex CLI](#openai-codex-cli)
  - [Cursor](#cursor)
  - [Windsurf](#windsurf)
  - [VS Code with GitHub Copilot](#vs-code-with-github-copilot)
  - [Any Other MCP Client](#any-other-mcp-client)
  - [Confirm the Connection](#confirm-the-connection)
- [Ask Questions in Plain Language](#ask-questions-in-plain-language)
  - [Find Strains or Plasmids by Gene or Keyword](#find-strains-or-plasmids-by-gene-or-keyword)
  - [Look Up a Name or an ID Fast](#look-up-a-name-or-an-id-fast)
  - [Check If a Stock Is Available](#check-if-a-stock-is-available)
  - [Get the Full Record for a Strain](#get-the-full-record-for-a-strain)
  - [Get the Full Record for a Plasmid](#get-the-full-record-for-a-plasmid)
  - [Find Every Strain with One Phenotype](#find-every-strain-with-one-phenotype)
  - [Read More Pages of Phenotype Results](#read-more-pages-of-phenotype-results)
  - [Run a Multi-Step Research Workflow](#run-a-multi-step-research-workflow)
- [Tips for Better Answers](#tips-for-better-answers)
- [What Each Answer Contains](#what-each-answer-contains)
- [Server Options](#server-options)
- [Troubleshooting](#troubleshooting)

## What dcr-mcp Does

The Dicty Stock Center keeps thousands of *Dictyostelium* strains and plasmids. Each one has an ID, a genotype, phenotypes, a depositor, and linked publications. `dcr-mcp` makes that catalog available to any coding assistant that speaks the Model Context Protocol (MCP).

After you connect `dcr-mcp`, you can type a question in your normal chat window. Your assistant picks the correct lookup, reads the answer, and replies in prose. You do not write a query language, and you do not open a browser.

`dcr-mcp` gives your assistant five abilities:

| Ability | Tool name your assistant calls | Use it for |
| :--- | :--- | :--- |
| Short name suggestions | `autocomplete_stocks` | Fast guesses from 3 or more letters. |
| Full catalog search | `search_stocks` | Broad discovery across free text such as summaries and phenotypes. |
| One strain record | `get_strain` | Every detail for one `DBS...` ID. |
| One plasmid record | `get_plasmid` | Every detail for one `DBP...` ID. |
| Strains for one phenotype | `list_strains_with_phenotype` | A paged list of strains with one phenotype term. |

You do not call these tool names yourself. This guide names them only so that you can recognize them in the tool log of your assistant.

The data is read only. `dcr-mcp` cannot change, add, or delete a stock center record.

## Quick Start

You need Go 1.26 or later to build the program. The three steps below take about one minute.

### Step 1: Build the Program

```bash
go build -o dcr-mcp .
# Result: the file ./dcr-mcp appears in the current directory
```

### Step 2: Start the Server

```bash
./dcr-mcp start-server
# Result: the server listens on 0.0.0.0:8080
```

The default settings are correct for normal use. The server reads the public dictyBase catalog at `https://graphql.dictybase.dev/graphql`. Keep this command active in its own terminal window. The commands `serve` and `start` do the same thing as `start-server`.

To stop the server, press `Ctrl+C`. The server finishes open requests and then exits.

### Step 3: Confirm That the Server Runs

Open a second terminal window and run this check:

```bash
curl http://localhost:8080/healthz
# Result: {"status":"ok"}
```

If you see `{"status":"ok"}`, the server is ready. Your assistant connects at the address `http://localhost:8080/mcp`.

## Connect Your Coding Agent

Pick the section for your tool. Every example uses the local address `http://localhost:8080/mcp`. If your team hosts `dcr-mcp` on a shared server, replace that address with the hosted one and keep the `/mcp` path at the end.

### Claude Code

Register the server one time with the command line:

```bash
claude mcp add --transport http --scope user dcr-mcp http://localhost:8080/mcp
```

To share the server with everybody in a repository, create `.mcp.json` in the project root instead:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### OpenAI Codex CLI

Add this block to `~/.codex/config.toml`:

```toml
[mcp_servers.dcr-mcp]
url = "http://localhost:8080/mcp"
```

You can also register the server from the command line:

```bash
codex mcp add dcr-mcp --url http://localhost:8080/mcp
```

### Cursor

Create `.cursor/mcp.json` in your project, or `~/.cursor/mcp.json` for all projects:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Windsurf

Add the server to `~/.codeium/windsurf/mcp_config.json`. Windsurf uses the key `serverUrl`:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "serverUrl": "http://localhost:8080/mcp"
    }
  }
}
```

### VS Code with GitHub Copilot

Create `.vscode/mcp.json` in your workspace. VS Code uses the key `servers`:

```json
{
  "servers": {
    "dcr-mcp": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Any Other MCP Client

Most other clients read a file named `mcp.json` with the same shape. Add this entry and restart the client:

```json
{
  "mcpServers": {
    "dcr-mcp": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Confirm the Connection

After you restart your client, type this in the chat:

> List the tools you have from dcr-mcp.

The assistant will name five tools: `autocomplete_stocks`, `search_stocks`, `get_strain`, `get_plasmid`, and `list_strains_with_phenotype`. If it names fewer tools, or it names none, go to [Troubleshooting](#troubleshooting).

## Ask Questions in Plain Language

This is the main way to use `dcr-mcp`. You write a normal question. The assistant chooses the lookup, reads the result, and answers you.

Each part below starts with a goal. It gives example prompts, then explains what comes back, then names the tool that runs behind the scenes.

### Find Strains or Plasmids by Gene or Keyword

Use this when you know a gene, a research topic, or a word from a strain description, but you do not know any ID.

**Type one of these:**

> Which Dicty strains are related to the gene mekA?

> Search the stock center for plasmids with GFP in the description.

> Find strains that came from a restriction enzyme mediated integration experiment.

> Show me stocks from the Knecht lab.

**What you get back:** A list of matches. Each match has an ID, the text that matched, a short summary, the strain label, and a yes or no answer for availability. Your assistant normally rewrites this as a short table or a bulleted list.

**Behind the scenes:** `search_stocks`. It covers more of the record than the quick suggestion lookup. It reads free text such as summaries, phenotype notes, assays, and depositor names. The dictyBase catalog owns the exact field coverage, and that coverage can change. `search_stocks` returns up to 15 matches by default, and up to 100 if you ask for more.

**Narrow the search by kind:**

> Search for mekA, but only plasmids.

The words "only strains" or "only plasmids" set the filter for you. Without them, the search covers both kinds.

### Look Up a Name or an ID Fast

Use this when you have part of a name and you want candidate IDs quickly, not a full report.

**Type one of these:**

> Autocomplete Dicty stocks that start with act15.

> Give me 10 quick suggestions for pDGB.

> What stock IDs match axeA?

**What you get back:** A short list of suggestions. Each suggestion has the ID, the kind of stock, the field that matched, and the matched text. There is no summary and no availability flag in this answer. This lookup is for speed.

**Behind the scenes:** `autocomplete_stocks`. It returns 5 suggestions by default, and up to 50 if you ask for more. It needs at least 3 letters.

If your search text has only 2 letters, ask for a full search instead. See [Tips for Better Answers](#tips-for-better-answers).

### Check If a Stock Is Available

Use this before you request material from the stock center.

**Type one of these:**

> Is strain DBS0350840 in stock?

> Is the plasmid pDGB_O2_0006 available to order?

> Of the mekA strains, which ones are in stock right now?

**What you get back:** A clear yes or no for each stock. For the last question, the assistant runs one search and then reports only the available matches.

**Behind the scenes:** `get_strain` or `get_plasmid` when you give an exact ID. `search_stocks` when you describe a group, because the search result carries the availability flag for every match.

Availability comes from the stock center inventory at the moment of the query. Treat it as current, not as a reservation.

### Get the Full Record for a Strain

Use this when you have a strain ID. Strain IDs start with `DBS`.

**Type one of these:**

> Give me the full record for DBS0350840.

> What is the genotype of DBS0350840?

> List the phenotypes and assays for strain DBS0349965.

> Who deposited DBS0350840, and what is its parent strain?

**What you get back:** One complete record. It contains the ID, the label, a summary, the species, the systematic name, and the availability flag. It also contains these parts:

- The genotypes, the names, and the synonyms
- The characteristics, the mutagenesis method, and the genetic modification
- The parent strain and the depositor
- The linked genes and the publications
- The phenotype annotations

Each phenotype annotation carries its own note, assay, environment, and source publication.

**Behind the scenes:** `get_strain`.

Ask for only the part you need. A request such as "just the genotype" produces a one-line answer from the same record.

### Get the Full Record for a Plasmid

Use this when you have a plasmid ID. Plasmid IDs start with `DBP`.

**Type one of these:**

> Show me everything about plasmid DBP0391806.

> Which genes are linked to plasmid DBP0391557, and who deposited it?

> Does DBP0000242 have a sequence and a map image?

> Which publications cite plasmid DBP0391806?

**What you get back:** One complete record. It contains the ID, the plasmid name, a summary, the keywords, the GenBank accession, the availability flag, the depositor, the linked genes, and the publications. It also contains the DNA sequence and the map image reference when the stock center has them. Some plasmids have an empty sequence field, and that is normal.

**Behind the scenes:** `get_plasmid`.

### Find Every Strain with One Phenotype

Use this when you start from a phenotype instead of a gene.

**Type one of these:**

> List Dicty strains with the phenotype increased growth rate.

> Which strains show aberrant cytokinesis? Show 25 of them.

> Find strains annotated with delayed aggregation and tell me which genes they hit.

**What you get back:** A page of strains. Each strain has its ID, label, summary, availability flag, linked genes, and publications. The answer also reports a total count, so you learn how large the full set is.

**Behind the scenes:** `list_strains_with_phenotype`. It returns 10 strains per page by default, and up to 100 per page if you ask for more.

This lookup matches the phenotype term that the stock center uses. A close paraphrase can return nothing. If you get an empty list, ask the assistant to run a full search for the same words. The full search reads phenotype text more loosely and will often show you the exact term to use.

### Read More Pages of Phenotype Results

Phenotype results arrive one page at a time. The answer includes a marker for the next page.

**Type one of these:**

> Show me the next page of those phenotype results.

> Keep paging until you have collected every strain, then summarize the list.

> Get the first 3 pages for increased growth rate and group the strains by gene.

**What you get back:** More strains from the same phenotype set. The assistant carries the page marker forward for you, so you do not need to repeat the phenotype term.

**Behind the scenes:** `list_strains_with_phenotype` again, with the page marker from the previous answer. The marker is a large number. The set ends when the marker comes back as `0`, so ask the assistant to stop at that point rather than after a short page.

### Run a Multi-Step Research Workflow

The strongest use of `dcr-mcp` chains the lookups together. Describe the end result, not the steps. The assistant plans the calls.

**Type one of these:**

> Find the Dicty strain for the gene triA, then summarize its genotype and its references.

> I study chemotaxis. Find candidate strains, keep only the ones in stock, and give me a table with the ID, the label, and the genotype.

> Compare DBS0350840 and DBS0349965 side by side. Include genotype, phenotypes, and depositor.

> Find every strain with the phenotype increased growth rate, then list the publications that cover more than one of them.

> Which plasmids match act15, and which of them are in stock?

**What you get back:** A written answer that joins several lookups. Take the first prompt as an example. The assistant searches for `triA` and selects the best `DBS...` match. It then reads the full strain record and writes a short report with the genotype and the publications.

**Behind the scenes:** A search tool first, then a record tool. The usual pattern is `search_stocks` to turn your words into IDs, and then `get_strain` or `get_plasmid` on the ID that matters.

**Make the result more useful:**

- Name the output format. "Give me a markdown table" or "Write it as a bulleted list".
- Set a budget. "Check the top 5 matches only" keeps a broad question fast.
- Ask for the evidence. "Show the stock ID next to every claim" lets you check the work.
- Ask the assistant to save the result. A coding agent can write the table into a file in your repository.

## Tips for Better Answers

| Situation | What to do |
| :--- | :--- |
| Your gene name has 2 letters, such as `cs` | Ask for a full search. Quick suggestions need 3 or more letters, and a 2-letter word returns the message `query must be at least 3 characters`. |
| You have an ID | Use it directly. `DBS...` means a strain, and `DBP...` means a plasmid. An exact ID skips the search step. |
| You want more results | State the number. "Show me 100 results" raises the limit. The ceiling is 50 for quick suggestions and 100 for a full search or a phenotype page. |
| You want one kind of stock | Add "only strains" or "only plasmids" to the question. |
| A phenotype search returns nothing | Run a full search for the same words. Then repeat the phenotype search with the exact term from the result. |
| You need to read a long list | Ask for the next page, and ask the assistant to continue until the page marker is `0`. |
| The answer looks stale | Ask the assistant to check the record again. `dcr-mcp` holds no cache, so a fresh call reads the live catalog. |
| You want proof | Ask for IDs, publication titles, and DOIs. The strain and plasmid records carry all three. |

## What Each Answer Contains

This table helps you choose how to phrase a request. It lists what the stock center returns for each lookup.

| Lookup | Returns for each item | Also returns | Size limits |
| :--- | :--- | :--- | :--- |
| Quick suggestions | ID, kind, matched field, matched text, match score | The applied limit and the total count | Needs 3 or more letters. 5 by default, 50 at most. |
| Full search | ID, kind, matched field, matched text, match score, summary, strain label, availability | The applied limit and the total count | Needs 2 or more letters. 15 by default, 100 at most. |
| One strain | ID, label, summary, species, systematic name, availability, genotypes, names, characteristics, mutagenesis method, genetic modification, parent strain, depositor, genes, publications, phenotypes | Each phenotype carries a note, an assay, an environment, and a publication | One record for one ID |
| One plasmid | ID, name, summary, keywords, GenBank accession, availability, sequence, map image, depositor, genes, publications | Cross references to other databases | One record for one ID |
| Strains for one phenotype | ID, label, summary, availability, genes, publications | The total count and the marker for the next page | 10 per page by default, 100 at most |

Publications include the title, the journal, the volume, the pages, the publication date, and the author last names. A strain or plasmid record also includes the DOI when the stock center has one.

## Server Options

The defaults work for one person on one machine. Change an option only when you have a reason. Every option accepts a command flag or an environment variable.

| Flag | Environment variable | Default | What it controls |
| :--- | :--- | :--- | :--- |
| `--graphql-endpoint`, `--endpoint` | `GRAPHQL_ENDPOINT` | `https://graphql.dictybase.dev/graphql` | The dictyBase catalog that the server reads. Point this at a test catalog to try changes. |
| `--host` | `MCP_HOST` | `0.0.0.0` | The network address that the server listens on. Use `127.0.0.1` to accept connections from your own machine only. |
| `--port` | `MCP_PORT` | `8080` | The port number. Change it if another program already uses `8080`. The valid range is `0` to `65535`, and `0` picks a free port. |
| `--stateless` | `MCP_STATELESS` | `true` | Whether the server keeps a session between requests. Keep the default. Set it to `false` only if your client reports that it needs a session. |
| `--allowed-origins` | `ALLOWED_ORIGINS` | `*` | Which web pages can reach the server from a browser. Give a comma separated list to restrict access. |

An example with several options:

```bash
./dcr-mcp start-server --host 127.0.0.1 --port 9090 --allowed-origins "https://lab.example.org"
# Result: the server listens on 127.0.0.1:9090 and accepts browser requests from one site
```

The server also answers a health check at `/healthz`. Use it in a monitor or a container probe.

## Troubleshooting

| Problem | Cause and fix |
| :--- | :--- |
| Your assistant does not list the five tools | The server is not running, or the address is wrong. Run `curl http://localhost:8080/healthz` and confirm the reply `{"status":"ok"}`. Then check the `url` value in your configuration file, and confirm that it ends with `/mcp`. |
| Your client connects but shows no tools after a configuration change | Most clients read the configuration file one time at startup. Restart the client. |
| The answer says `query must be at least 3 characters` | The quick suggestion lookup ran with too short a word. Ask for a full search instead. |
| The answer says `query must be at least 2 characters` | The search text is a single character. Add more letters, or use the stock ID. |
| The answer says `limit must be between 1 and 50` or `limit must be between 1 and 100` | You asked for too many results. Lower the number to the ceiling in [What Each Answer Contains](#what-each-answer-contains). |
| The answer says `invalid entity` | The kind filter was not understood. Say "only strains", "only plasmids", or ask for both. |
| The answer says `strain id is required` or `plasmid id is required` | No ID reached the lookup. Give the full ID, such as `DBS0350840` or `DBP0391806`. |
| A phenotype list is empty | The phenotype term does not match a stock center annotation. Run a full search for the same words to find the exact term. |
| The answer reports that a query failed | The server reached the dictyBase catalog and got an error back, or it could not reach the catalog at all. Check your network connection and the `--graphql-endpoint` value. |
| Port `8080` is already in use | Start the server on another port with `--port 9090`, and update the port in your client configuration. |
