# Dicty Community Resource Plugin Privacy Policy

This policy covers the Dicty Community Resource ChatGPT plugin and its hosted DCR MCP service at `mcp.dictycr.org`.

## Catalog requests

When you use the plugin, ChatGPT sends the hosted service the tool name and the arguments needed to answer your request. Arguments can include a search phrase, a stock ID, a gene name, or a phenotype term. The service uses those arguments to retrieve matching records from the Dicty Stock Center catalog and returns the results to ChatGPT.

## Operational logs

The service records the MCP method, tool name, request duration, and error status or details for operation and troubleshooting. The tool-call logger does not record argument or result payloads as fields. The production ingress and proxy do not retain IP addresses or lookup arguments in their logs.

We keep retained service data in the `dictycr.org` Kubernetes cluster for 30 days. We do not share those retained records with third parties.

## ChatGPT and support

OpenAI handles your ChatGPT conversation and its use of the plugin under the policies and settings that apply to your ChatGPT account. [OpenAI Privacy Policy](https://openai.com/policies/privacy-policy/)

For privacy questions or requests involving your data, email [siddhartha-basu@northwestern.edu](mailto:siddhartha-basu@northwestern.edu).

For general technical support, use the project's GitHub Issues page. Issues are public. Keep personal information and request logs out of issue reports. [Dicty Community Resource support](https://github.com/dictybase/dcr-mcp/issues)
