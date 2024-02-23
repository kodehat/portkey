package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/routes"
	"github.com/kodehat/portkey/internal/utils"
)

const LOGGER_PREFIX string = "[portkey] "

//go:embed static
var static embed.FS

func main() {
	ctx := context.Background()
	if err := run(ctx, config.C, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c config.Config, stdin io.Reader, stdout, stderr io.Writer) error {
	logger := log.New(stdout, LOGGER_PREFIX, log.Lmsgprefix|log.Ldate|log.Ltime)
	mux := http.NewServeMux()
	routes.AddRoutes(mux, static)
	addr := utils.BuildAddr(config.C.Host, config.C.Port)
	logger.Printf("Listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}
