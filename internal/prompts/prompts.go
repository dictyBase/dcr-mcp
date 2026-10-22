// Package prompts registers MCP prompt templates for common Dicty Stock Center research workflows.
package prompts

import (
	"context"
	"fmt"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type spec struct {
	name           string
	title          string
	description    string
	argName        string
	argDescription string
	template       string
}

var (
	findStocksForGene = spec{
		name:           "find_stocks_for_gene",
		title:          "Find strains and plasmids for a gene",
		description:    "Find every stock for a gene and recommend which to order.",
		argName:        "gene",
		argDescription: "Gene name or DDB_G id, for example pkaC",
		template: "Find every Dicty Stock Center strain and plasmid for the gene {{gene}}.\n" +
			"1. Call find_stocks_by_gene with gene={{gene}}.\n" +
			"2. For each strain that is in stock, call get_strain (detail=summary).\n" +
			"3. Reply with a Markdown table of strains: linked ID, label, genotype, key phenotypes, availability.\n" +
			"4. List plasmids separately and note that plasmid matches are name matches to verify.\n" +
			"5. Finish with which stocks to order first and why. Cite each stock with its url.",
	}

	compareStrains = spec{
		name:           "compare_strains",
		title:          "Compare strains side by side",
		description:    "Compare several strains' genotypes, phenotypes, and availability.",
		argName:        "ids",
		argDescription: "Comma-separated strain ids, for example DBS0350840, DBS0236783",
		template: "Compare these Dictyostelium strains: {{ids}}.\n" +
			"1. Call get_strain with detail=full for each id.\n" +
			"2. Build a Markdown table with one column per strain and rows for label, genotype, " +
			"genetic modification, mutagenesis method, characteristics, depositor, and availability.\n" +
			"3. Below the table, list phenotypes unique to each strain and phenotypes they share.\n" +
			"4. Cite each strain with its url and each phenotype with its PubMed link.",
	}

	phenotypeSurvey = spec{
		name:           "phenotype_survey",
		title:          "Survey strains with a phenotype",
		description:    "Find the exact phenotype term, page through matching strains, and summarize by gene.",
		argName:        "phenotype",
		argDescription: "A phenotype in plain words, for example 'fast growth' or 'aggregation defect'",
		template: "Survey strains with the phenotype: {{phenotype}}.\n" +
			"1. Call suggest_phenotype_terms with query={{phenotype}} and pick the best matching exact term; " +
			"if several fit, tell me which you chose.\n" +
			"2. Call list_strains_with_phenotype with that term, following next_cursor until there are no more " +
			"pages or you have 50 strains.\n" +
			"3. Group the strains by gene and give counts, then a table of in-stock strains with linked IDs.\n" +
			"4. Say how many strains matched in total (total_count).",
	}

	orderChecklist = spec{
		name:           "order_checklist",
		title:          "Prepare an order checklist",
		description:    "Check availability for a list of stocks and suggest alternatives for unavailable ones.",
		argName:        "ids",
		argDescription: "Comma-separated strain and/or plasmid ids",
		template: "Prepare an order checklist for these stocks: {{ids}}.\n" +
			"1. Call check_availability with all ids.\n" +
			"2. For available stocks, list linked ID, label, and stock page url as a checklist.\n" +
			"3. For unavailable or unknown ids, explain the problem and use search_stocks or find_stocks_by_gene " +
			"to suggest up to two in-stock alternatives each.\n" +
			"4. Remind me that orders are placed on the Dicty Stock Center website.",
	}

	renderPrompt = F.Curry2(func(sp spec, arg string) *mcp.GetPromptResult {
		return &mcp.GetPromptResult{
			Description: sp.description,
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: strings.ReplaceAll(sp.template, "{{"+sp.argName+"}}", arg)},
			}},
		}
	})
)

func newPrompt(sp spec) *mcp.Prompt {
	return &mcp.Prompt{
		Name:        sp.name,
		Title:       sp.title,
		Description: sp.description,
		Arguments: []*mcp.PromptArgument{{
			Name:        sp.argName,
			Description: sp.argDescription,
			Required:    true,
		}},
	}
}

func templateHandler(sp spec) mcp.PromptHandler {
	missing := F.Constant(fmt.Errorf("argument %q is required", sp.argName))
	return func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return E.Unwrap(F.Pipe5(
			req.Params.Arguments,
			R.Lookup[string](sp.argName),
			O.Map(strings.TrimSpace),
			O.Filter(S.IsNonEmpty),
			E.FromOption[string](missing),
			E.Map[error](renderPrompt(sp)),
		))
	}
}

// RegisterAll registers all research-workflow prompts on srv.
func RegisterAll(srv *mcp.Server) {
	srv.AddPrompt(newPrompt(findStocksForGene), templateHandler(findStocksForGene))
	srv.AddPrompt(newPrompt(compareStrains), templateHandler(compareStrains))
	srv.AddPrompt(newPrompt(phenotypeSurvey), templateHandler(phenotypeSurvey))
	srv.AddPrompt(newPrompt(orderChecklist), templateHandler(orderChecklist))
}
