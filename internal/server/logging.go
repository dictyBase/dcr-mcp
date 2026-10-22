package server

import (
	"context"
	"log/slog"
	"time"

	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func asCallToolRequest(req mcp.Request) (*mcp.CallToolRequest, bool) {
	r, ok := req.(*mcp.CallToolRequest)
	return r, ok
}

func asCallToolResult(res mcp.Result) (*mcp.CallToolResult, bool) {
	r, ok := res.(*mcp.CallToolResult)
	return r, ok
}

func toolName(req mcp.Request) string {
	return F.Pipe2(
		req,
		O.FromValidation(asCallToolRequest),
		O.Fold(
			F.Constant(""),
			func(r *mcp.CallToolRequest) string { return r.Params.Name },
		),
	)
}

func toolIsError(res mcp.Result) bool {
	return F.Pipe2(
		res,
		O.FromValidation(asCallToolResult),
		O.Fold(
			F.Constant(false),
			func(r *mcp.CallToolResult) bool { return r.IsError },
		),
	)
}

// ToolCallLogger logs method, tool name, duration, and error status for each MCP request.
// It never logs arguments or results.
func ToolCallLogger(logger *slog.Logger) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			res, err := next(ctx, method, req)
			logger.InfoContext(ctx, "mcp request",
				slog.String("method", method),
				slog.String("tool", toolName(req)),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.Bool("tool_error", toolIsError(res)),
				slog.Any("error", err),
			)
			return res, err
		}
	}
}
