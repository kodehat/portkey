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

	authpkg "github.com/kodehat/portkey/internal/auth"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/favicon"
	"github.com/kodehat/portkey/internal/metrics"
	"github.com/kodehat/portkey/internal/server"
)

//go:embed static
var static embed.FS

func main() {
	ctx := context.Background()
	build.LoadBuildDetails(getCssResourceHash())
	config.Load()
	logger := slog.New(config.C.GetLogHandler(os.Stdout))
	slog.SetDefault(logger)
	favicon.Init(config.C.Favicon.CacheDir, config.C.Favicon.CacheEnabled, logger)
	if config.C.Auth.Enabled {
		if err := authpkg.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "auth init failed: %s\n", err)
			os.Exit(1)
		}
	}
	metrics.Load()
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
		static,
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Server.Host, config.Server.Port),
		Handler: srv,
	}
	go func() {
		logger.Info("server is now accepting connections", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var metricHttpServer *http.Server
	if config.Metrics.Enabled {
		metricsSrv := server.NewMetricsServer(logger)
		metricHttpServer = &http.Server{
			Addr:    net.JoinHostPort(config.Metrics.Host, config.Metrics.Port),
			Handler: metricsSrv,
		}
		go func() {
			logger.Info("metrics server is now accepting connections", "address", metricHttpServer.Addr)
			if err := metricHttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
			}
		}()
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		// Use a fresh context for shutdown — the parent ctx is already cancelled.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
		if metricHttpServer != nil {
			if err := metricHttpServer.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "error shutting down metrics server: %s\n", err)
			}
		}
	})
	wg.Wait()
	return nil
}

func getCssResourceHash() string {
	cssFile, err := static.ReadFile("static/css/main.css")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not read CSS for hash: %s\n", err)
		return "unknown"
	}
	hasher := sha256.New()
	hasher.Write(cssFile)
	return hex.EncodeToString(hasher.Sum(nil))[:8]
}
