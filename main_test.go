package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
	"github.com/kodehat/portkey/internal/utils"
)

var CONFIG_EMPTY config.Config = config.Config{
	Host:    "localhost",
	Port:    "3000",
	Portals: []models.Portal{},
	Pages:   []models.Page{},
}

func setup(t *testing.T, config config.Config) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go run(ctx, config, os.Stdin, os.Stdout, os.Stderr)
	return utils.WaitForReady(ctx, time.Duration(5)*time.Second, "http://localhost:3000/healthz")
}

func TestStart(t *testing.T) {
	err := setup(t, CONFIG_EMPTY)
	if err != nil {
		t.Fatalf("server was unable to start %v", err)
	}
}
