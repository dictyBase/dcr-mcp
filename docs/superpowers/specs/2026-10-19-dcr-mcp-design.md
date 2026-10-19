# `dcr-mcp` — dictyBase Stock Center MCP Server Design

**Date:** 2026-10-19  
**Module:** `github.com/dictybase/dcr-mcp`

---

## 1. Intent & Why

Provide a Kubernetes-ready Model Context Protocol (MCP) server in Go that exposes dictyBase Stock Center GraphQL operations as typed MCP tools. Web applications and external AI agents outside the cluster connect over Streamable HTTP (via Kubernetes Ingress) to autocomplete stocks, run full-text stock searches, fetch individual strains or plasmids by ID, and list strains annotated with a given phenotype.

---

## 2. Goals, Non-Goals, and Anti-Goals

### Goals
- Expose five stock-center tools via `github.com/modelcontextprotocol/go-sdk/mcp`:
  1. `autocomplete_stocks` (`listStockSuggestions`)
  2. `search_stocks` (`searchStocks`)
  3. `get_strain` (`strain`)
  4. `get_plasmid` (`plasmid`)
  5. `list_strains_with_phenotype` (`listStrainsWithAnnotation` with `type: "phenotype"`)
- Query the dictyBase GraphQL backend (`https://graphql.dictybase.dev/graphql` by default, configurable via CLI flag / environment variable).
- Serve MCP over Streamable HTTP (`mcp.NewStreamableHTTPHandler`) with CORS support and a `/healthz` readiness/liveness endpoint for Kubernetes Ingress deployments.
- Provide a `urfave/cli/v3` command-line interface with a subcommand (`start-server`, alias `serve`).
- Write 100% of the codebase in the house `fp-go/v2` + `fp-go-loom` functional style, verified mechanically by `pipelinecheck`, `loomcheck`, `golangci-lint`, and `gopls`.

### Non-Goals
- Stock mutations (`createStrain`, `updateStrain`, `createOrder`, etc.) — read/search operations only in this phase.
- Direct gRPC calls to `modware-stock` or `modware-annotation` — all data access goes through the GraphQL endpoint.
- Stateful in-memory session affinity across pods (stateless Streamable HTTP is the default for multi-replica Ingress compatibility).

### Anti-Goals
- Imperative Go control flow (`if err != nil` chains, `switch` statements in domain logic, bare `for` loops, nested function calls `f(g(x))`, or combinator direct application `comb(arg)(val)`).
- Any monadic stack outside the house allowlist (`Option`, `Either`, `IO`, `IOEither`).
- Hardcoding the GraphQL server URL so it cannot be overridden at runtime.

---

## 3. Constraints & Technology Decisions

### User-Decided
- **Language & Module:** Go (`github.com/dictybase/dcr-mcp`), using `github.com/IBM/fp-go/v2` and `github.com/dictyBase/fp-go-loom`.
- **MCP SDK:** `github.com/modelcontextprotocol/go-sdk/mcp`.
- **CLI Library:** `github.com/urfave/cli/v3` with server launch as a subcommand (`start-server`, alias `serve`).
- **GraphQL Endpoint:** Default `https://graphql.dictybase.dev/graphql`, configurable via CLI parameter and environment variable.
- **Deployment Target:** Kubernetes cluster behind Ingress, accessible to external web apps and MCP clients via Streamable HTTP.
- **Tool Set:**
  1. Stock autocomplete (`listStockSuggestions`)
  2. Stock full search (`searchStocks`)
  3. Single strain by ID (`strain`)
  4. Single plasmid by ID (`plasmid`)
  5. List strains with phenotype (`listStrainsWithAnnotation`)

### Builder Calls (Defaults Chosen)
- **GraphQL Client:** `github.com/hasura/go-graphql-client` (matching `graphql-client` in the workspace) wrapped in `IOE.TryCatchError` + `IOE.MapLeft` + `IOE.Map`.
- **Subcommand & Flags:**
  - Subcommand: `start-server` (aliases: `serve`, `start`)
  - `--graphql-endpoint` (env: `GRAPHQL_ENDPOINT`, default: `"https://graphql.dictybase.dev/graphql"`)
  - `--host` (env: `MCP_HOST`, default: `"0.0.0.0"`)
  - `--port` (env: `MCP_PORT`, default: `8080`)
  - `--stateless` (env: `MCP_STATELESS`, default: `true` — sets `mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}` so multi-pod Kubernetes Ingress requires no sticky sessions)
  - `--allowed-origins` (env: `ALLOWED_ORIGINS`, default: `[]string{"*"}` — CORS headers for browser web apps, exposing and allowing `Mcp-Session-Id` and `MCP-Protocol-Version`)
- **HTTP Routes:**
  - `/mcp` — Streamable HTTP MCP handler wrapped in CORS middleware
  - `/healthz` — Kubernetes liveness/readiness probe returning `200 OK` (`{"status":"ok"}`)
- **Package Layout:**
  - `main.go` / `cli.go` — root `cli.Command` and `start-server` subcommand definition + `RunServer` point-free pipeline.
  - `internal/graphql/` — GraphQL client wrapper, query structs, and domain models (`StockSuggestionList`, `StockSearchResultList`, `StrainDetail`, `PlasmidDetail`, `PhenotypeStrainList`).
  - `internal/tools/` — Pure validation (`E.Either`), `IOEither` GraphQL execution pipelines, state structs + lenses, and `mcp.AddTool` registrations for the 5 tools.
  - `internal/server/` — HTTP mux, CORS middleware, `/healthz` handler, and graceful shutdown on `context.Context` cancellation.

---

## 4. MCP Tools Contract

All tools are registered via typed `mcp.AddTool[In, Out](server, &mcp.Tool{...}, handler)` so the Go SDK automatically generates JSON Schema for inputs and outputs and populates `StructuredContent` alongside JSON text content.

### 4.1 `autocomplete_stocks`
- **GraphQL Operation:** `listStockSuggestions(query: $query, limit: $limit, entity: $entity)`
- **Input (`AutocompleteStocksInput`):**
  - `query` (`string`, required): partial search text; trimmed length must be ≥ 3 runes.
  - `limit` (`int`, optional): max suggestions to return; defaults to `5`, clamped/validated in `1..50`.
  - `entity` (`string`, optional): `ALL`, `STRAIN`, or `PLASMID` (case-insensitive, defaults to `ALL`).
- **Output (`AutocompleteStocksOutput`):**
  - `suggestions` (`[]StockSuggestion`): `id`, `entity`, `field`, `display_text`, `score`
  - `limit` (`int`)
  - `total_count` (`int`)

### 4.2 `search_stocks`
- **GraphQL Operation:** `searchStocks(query: $query, limit: $limit, entity: $entity)`
- **Input (`SearchStocksInput`):**
  - `query` (`string`, required): search text; trimmed length must be ≥ 2 runes.
  - `limit` (`int`, optional): max results to return; defaults to `50`, validated in `1..100`.
  - `entity` (`string`, optional): `ALL`, `STRAIN`, or `PLASMID` (case-insensitive, defaults to `ALL`).
- **Output (`SearchStocksOutput`):**
  - `results` (`[]StockSearchResult`): `id`, `entity`, `field`, `display_text`, `score`, `summary`, `strain_label`, `in_stock`
  - `limit` (`int`)
  - `total_count` (`int`)

### 4.3 `get_strain`
- **GraphQL Operation:** `strain(id: $id)`
- **Input (`GetStrainInput`):**
  - `id` (`string`, required): strain ID (e.g. `DBS0350840`); trimmed length must be non-empty (`Str.IsNonEmpty`).
- **Output (`StrainDetail`):**
  - `id`, `label`, `summary`, `species`, `systematic_name`, `plasmid`, `in_stock`
  - `dbxrefs` (`[]string`), `genotypes` (`[]string`), `names` (`[]string`), `characteristics` (`[]string`)
  - `mutagenesis_method`, `genetic_modification`
  - `parent` (`*ParentStrain`: `id`, `label`)
  - `depositor` (`Depositor`: `first_name`, `last_name`)
  - `genes` (`[]GeneRef`: `name`)
  - `publications` (`[]PublicationRef`: `id`, `pub_date`, `title`, `journal`, `volume`, `pages`, `doi`, `authors`)
  - `phenotypes` (`[]PhenotypeItem`: `phenotype`, `note`, `assay`, `environment`, `publication`)

### 4.4 `get_plasmid`
- **GraphQL Operation:** `plasmid(id: $id)`
- **Input (`GetPlasmidInput`):**
  - `id` (`string`, required): plasmid ID (e.g. `DBP0391806`); trimmed length must be non-empty (`Str.IsNonEmpty`).
- **Output (`PlasmidDetail`):**
  - `id`, `name`, `summary`, `image_map`, `sequence`, `genbank_accession`, `in_stock`
  - `dbxrefs` (`[]string`), `keywords` (`[]string`)
  - `depositor` (`Depositor`: `first_name`, `last_name`)
  - `genes` (`[]GeneRef`: `name`)
  - `publications` (`[]PublicationRef`: `id`, `pub_date`, `title`, `journal`, `volume`, `pages`, `doi`, `authors`)

### 4.5 `list_strains_with_phenotype`
- **GraphQL Operation:** `listStrainsWithAnnotation(cursor: $cursor, limit: $limit, type: "phenotype", annotation: $annotation)`
- **Input (`ListStrainsWithPhenotypeInput`):**
  - `phenotype` (`string`, required): phenotype term (e.g. `"increased growth rate"`); trimmed length must be non-empty (`Str.IsNonEmpty`).
  - `cursor` (`int`, optional): pagination cursor, defaults to `0` (`>= 0`).
  - `limit` (`int`, optional): page size, defaults to `10`, validated in `1..100`.
- **Output (`PhenotypeStrainList`):**
  - `strains` (`[]PhenotypeStrainItem`: `id`, `label`, `summary`, `in_stock`, `genes`, `publications`)
  - `next_cursor` (`int`)
  - `total_count` (`int`)

---

## 5. Functional Pipeline Architecture (`fp-go/v2` + `fp-go-loom`)

1. **Tool Handler Pipeline:**
   - Each tool handler seeds a boundary state struct (`ToolRequest[In]{Ctx, Client, Input}`) and runs a single flat `F.PipeN`:
     - `IOE.Of[error](seed)`
     - `IOE.ChainEitherK(validateAndNormalizeInput)` (pure `Either[error, State]` using `predord.MinStrLen`, `predord.IntBetweenInclusive`, `Str.IsNonEmpty`, `eithererr.ConstErr`, `MO.First` / `O.FromPredicate` for defaults — zero imperative `if` branching)
     - `IOE.Bind(responseLens.Set, executeGraphQLQuery)` (`IOE.TryCatchError` with raw SDK call only → `IOE.MapLeft` error wrapping → `IOE.Map` domain projection)
     - `ioeutils.ToEither`
     - `E.Fold` into `(*mcp.CallToolResult, Out, error)` at the MCP handler boundary.
2. **CLI Entrypoint (`RunServer`):**
   - Point-free seed struct `ServerActionInput{Context: ctx, Command: cmd}` piped through validation, MCP server construction, HTTP listener execution, `ioeutils.ToEither`, and `E.ToError`.
3. **Quality Gates:**
   - `pipelinecheck_test.go` enforcing `Require`, `RequirePointFreeSeed`, `RequirePointFreeBranching`, `RequireTrivialDelegate`, `RequireNoIfErrInTryCatch`, `RequireNoNonRawTryCatchCallback`, `RequireSafePrintf`, `RequireNoRedundantFold`, `RequireNoHandRolledFileJoinInArrayMap`.
   - `loomcheck_test.go` enforcing `loomcheck.Require`.
   - Unit tests using `httptest.NewServer` for all 5 GraphQL tool pipelines + in-memory/HTTP MCP client-server integration tests + live smoke tests.
   - `golangci-lint fmt`, `golangci-lint run ./...`, `go test ./...`, and `gopls check -severity=hint` on all `.go` files.
