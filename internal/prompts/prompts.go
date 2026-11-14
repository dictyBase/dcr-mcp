// Package prompts registers MCP prompt templates for common Dicty Stock Center research workflows.
package prompts

import (
	"context"
	"fmt"
	A "github.com/IBM/fp-go/v2/array"
	L "github.com/IBM/fp-go/v2/optics/lens"
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
	specNameLens = L.MakeLens(
		func(x spec) string { return x.name },
		func(x spec, value string) spec {
			x.name = value
			return x
		},
	)
	specTitleLens = L.MakeLens(
		func(x spec) string { return x.title },
		func(x spec, value string) spec {
			x.title = value
			return x
		},
	)
	specDescriptionLens = L.MakeLens(
		func(x spec) string { return x.description },
		func(x spec, value string) spec {
			x.description = value
			return x
		},
	)
	specArgNameLens = L.MakeLens(
		func(x spec) string { return x.argName },
		func(x spec, value string) spec {
			x.argName = value
			return x
		},
	)
	specArgDescriptionLens = L.MakeLens(
		func(x spec) string { return x.argDescription },
		func(x spec, value string) spec {
			x.argDescription = value
			return x
		},
	)
	specTemplateLens = L.MakeLens(
		func(x spec) string { return x.template },
		func(x spec, value string) spec {
			x.template = value
			return x
		},
	)
	getPromptResultDescriptionLens = L.MakeLens(
		func(x mcp.GetPromptResult) string { return x.Description },
		func(x mcp.GetPromptResult, value string) mcp.GetPromptResult {
			x.Description = value
			return x
		},
	)
	getPromptResultMessagesLens = L.MakeLens(
		func(x mcp.GetPromptResult) []*mcp.PromptMessage { return x.Messages },
		func(x mcp.GetPromptResult, value []*mcp.PromptMessage) mcp.GetPromptResult {
			x.Messages = value
			return x
		},
	)
	promptMessageRoleLens = L.MakeLens(
		func(x mcp.PromptMessage) mcp.Role { return x.Role },
		func(x mcp.PromptMessage, value mcp.Role) mcp.PromptMessage {
			x.Role = value
			return x
		},
	)
	promptMessageContentLens = L.MakeLens(
		func(x mcp.PromptMessage) mcp.Content { return x.Content },
		func(x mcp.PromptMessage, value mcp.Content) mcp.PromptMessage {
			x.Content = value
			return x
		},
	)
	textContentTextLens = L.MakeLens(
		func(x mcp.TextContent) string { return x.Text },
		func(x mcp.TextContent, value string) mcp.TextContent {
			x.Text = value
			return x
		},
	)
	promptNameLens = L.MakeLens(
		func(x mcp.Prompt) string { return x.Name },
		func(x mcp.Prompt, value string) mcp.Prompt {
			x.Name = value
			return x
		},
	)
	promptTitleLens = L.MakeLens(
		func(x mcp.Prompt) string { return x.Title },
		func(x mcp.Prompt, value string) mcp.Prompt {
			x.Title = value
			return x
		},
	)
	promptDescriptionLens = L.MakeLens(
		func(x mcp.Prompt) string { return x.Description },
		func(x mcp.Prompt, value string) mcp.Prompt {
			x.Description = value
			return x
		},
	)
	promptArgumentsLens = L.MakeLens(
		func(x mcp.Prompt) []*mcp.PromptArgument { return x.Arguments },
		func(x mcp.Prompt, value []*mcp.PromptArgument) mcp.Prompt {
			x.Arguments = value
			return x
		},
	)
	promptArgumentNameLens = L.MakeLens(
		func(x mcp.PromptArgument) string { return x.Name },
		func(x mcp.PromptArgument, value string) mcp.PromptArgument {
			x.Name = value
			return x
		},
	)
	promptArgumentDescriptionLens = L.MakeLens(
		func(x mcp.PromptArgument) string { return x.Description },
		func(x mcp.PromptArgument, value string) mcp.PromptArgument {
			x.Description = value
			return x
		},
	)
	promptArgumentRequiredLens = L.MakeLens(
		func(x mcp.PromptArgument) bool { return x.Required },
		func(x mcp.PromptArgument, value bool) mcp.PromptArgument {
			x.Required = value
			return x
		},
	)
	findStocksForGene = F.Pipe6(
		spec{},
		specNameLens.Set("find_stocks_for_gene"),
		specTitleLens.Set("Find strains and plasmids for a gene"),
		specDescriptionLens.Set("Find every stock for a gene and recommend which to order."),
		specArgNameLens.Set("gene"),
		specArgDescriptionLens.Set("Gene name or DDB_G id, for example pkaC"),
		specTemplateLens.Set("Find every Dicty Stock Center strain and plasmid for the gene {{gene}}.\n"+
			"1. Call find_stocks_by_gene with gene={{gene}}.\n"+
			"2. For each strain that is in stock, call get_strain (detail=summary).\n"+
			"3. Reply with a Markdown table of strains: linked ID, label, genotype, key phenotypes, availability.\n"+
			"4. List plasmids separately and note that plasmid matches are name matches to verify.\n"+
			"5. Finish with which stocks to order first and why. Cite each stock with its url."),
	)

	compareStrains = F.Pipe6(
		spec{},
		specNameLens.Set("compare_strains"),
		specTitleLens.Set("Compare strains side by side"),
		specDescriptionLens.Set("Compare several strains' genotypes, phenotypes, and availability."),
		specArgNameLens.Set("ids"),
		specArgDescriptionLens.Set("Comma-separated strain ids, for example DBS0350840, DBS0236783"),
		specTemplateLens.Set("Compare these Dictyostelium strains: {{ids}}.\n"+
			"1. Call get_strain with detail=full for each id.\n"+
			"2. Build a Markdown table with one column per strain and rows for label, genotype, "+
			"genetic modification, mutagenesis method, characteristics, depositor, and availability.\n"+
			"3. Below the table, list phenotypes unique to each strain and phenotypes they share.\n"+
			"4. Cite each strain with its url and each phenotype with its PubMed link."),
	)

	phenotypeSurvey = F.Pipe6(
		spec{},
		specNameLens.Set("phenotype_survey"),
		specTitleLens.Set("Survey strains with a phenotype"),
		specDescriptionLens.Set("Find strains by phenotype statement, confirm exact phenotype terms, and summarize by gene."),
		specArgNameLens.Set("phenotype"),
		specArgDescriptionLens.Set("A phenotype, environment, or assay in plain words, for example 'aggregation defect'"),
		specTemplateLens.Set("Survey strains with the phenotype: {{phenotype}}.\n"+
			"1. Call search_phenotypes with query={{phenotype}}. Show matched_texts and say which ones fit "+
			"the request, and whether each is a phenotype, an environment, or an assay.\n"+
			"2. Call get_strain on the top strains under the fitting texts to read the full phenotype "+
			"statement and its publication.\n"+
			"3. If a phenotype term from step 2 fits (the phenotype field, not the environment or the assay), "+
			"call list_strains_with_phenotype with phenotype_term set to that exact text. Follow next_cursor "+
			"until has_more is false or you have 50 unique strains. If the request is about an environment "+
			"or an assay, skip this step and use the search_phenotypes rows only.\n"+
			"4. Group the strains by gene and give counts, then a table of in-stock strains with linked IDs.\n"+
			"5. Report the number of unique strains you collected. Call it a catalog total only if has_more "+
			"became false."),
	)

	orderChecklist = F.Pipe6(
		spec{},
		specNameLens.Set("order_checklist"),
		specTitleLens.Set("Prepare an order checklist"),
		specDescriptionLens.Set("Check availability for a list of stocks and suggest alternatives for unavailable ones."),
		specArgNameLens.Set("ids"),
		specArgDescriptionLens.Set("Comma-separated strain and/or plasmid ids"),
		specTemplateLens.Set("Prepare an order checklist for these stocks: {{ids}}.\n"+
			"1. Call check_availability with all ids.\n"+
			"2. For available stocks, list linked ID, label, and stock page url as a checklist.\n"+
			"3. For unavailable or unknown ids, explain the problem and use search_stocks or find_stocks_by_gene "+
			"to suggest up to two in-stock alternatives each.\n"+
			"4. Remind me that orders are placed on the Dicty Stock Center website."),
	)
)

// renderPrompt fills the template of sp with arg as one user message.
func renderPrompt(sp spec, arg string) *mcp.GetPromptResult {
	return new(F.Pipe2(
		mcp.GetPromptResult{},
		getPromptResultDescriptionLens.Set(specDescriptionLens.Get(sp)),
		getPromptResultMessagesLens.Set(A.Of(new(F.Pipe2(
			mcp.PromptMessage{},
			promptMessageRoleLens.Set("user"),
			promptMessageContentLens.Set(new(F.Pipe1(
				mcp.TextContent{},
				textContentTextLens.Set(strings.ReplaceAll(
					specTemplateLens.Get(sp),
					"{{"+specArgNameLens.Get(sp)+"}}",
					arg,
				)),
			))),
		)))),
	))
}

func newPrompt(sp spec) *mcp.Prompt {
	return new(F.Pipe4(
		mcp.Prompt{},
		promptNameLens.Set(specNameLens.Get(sp)),
		promptTitleLens.Set(specTitleLens.Get(sp)),
		promptDescriptionLens.Set(specDescriptionLens.Get(sp)),
		promptArgumentsLens.Set(A.Of(new(F.Pipe3(
			mcp.PromptArgument{},
			promptArgumentNameLens.Set(specArgNameLens.Get(sp)),
			promptArgumentDescriptionLens.Set(specArgDescriptionLens.Get(sp)),
			promptArgumentRequiredLens.Set(true),
		)))),
	))
}

func templateHandler(sp spec) mcp.PromptHandler {
	return func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return E.Unwrap(F.Pipe5(
			req.Params.Arguments,
			R.Lookup[string](sp.argName),
			O.Map(strings.TrimSpace),
			O.Filter(S.IsNonEmpty),
			E.FromOption[string](F.Constant(fmt.Errorf("argument %q is required", specArgNameLens.Get(sp)))),
			E.Map[error](F.Bind1st(renderPrompt, sp)),
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
