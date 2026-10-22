// Command dcr-mcp provides a Streamable HTTP MCP server for dictyBase stock center queries.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
)

func invokeCLI(ctx context.Context, args []string) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		return struct{}{}, NewApp().Run(ctx, args)
	})
}

func handleExit(err error) int {
	_, _ = fmt.Fprintln(os.Stderr, err)
	return 1
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := F.Pipe2(
		invokeCLI(ctx, os.Args),
		ioeutils.ToEither[error, struct{}],
		E.Fold(
			handleExit,
			F.Constant1[struct{}](0),
		),
	)
	stop()
	os.Exit(code)
}
