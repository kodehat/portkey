package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/server"
)

//go:embed static
var static embed.FS

func main() {
	ctx := context.Background()
	build.LoadBuildDetails(getCssResourceHash())
	config.Load()
	if err := run(ctx, config.C, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, config config.Config, stdin io.Reader, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	logger := slog.New(config.GetLogHandler(stdout))
	slog.SetDefault(logger)
	srv := server.NewServer(
		logger,
		&config,
		static,
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: srv,
	}
	go func() {
		logger.Info("server is now accepting connections", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		// make a new context for the Shutdown (thanks Alessandro Rosetti)
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func getCssResourceHash() string {
	cssFile, _ := static.ReadFile("static/css/main.css")
	hasher := sha256.New()
	hasher.Write(cssFile)
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha[:8]
}
