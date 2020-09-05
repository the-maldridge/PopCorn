package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/popcorn/pkg/stats"
	_ "github.com/the-maldridge/popcorn/pkg/stats/fs"
	_ "github.com/the-maldridge/popcorn/pkg/stats/memory"
)

var ()

func main() {
	llevel := os.Getenv("LOG_LEVEL")
	if llevel == "" {
		llevel = "INFO"
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "popcornd",
		Level: hclog.LevelFromString(llevel),
	})

	stats.SetStoreParentLogger(appLogger)
	stats.DoCallbacks()

	store, err := stats.NewStore(os.Getenv("STORE"))
	if err != nil {
		appLogger.Error("Error initializing storage", "error", err)
		os.Exit(1)
	}

	sr := stats.New(appLogger, store)

	bind := os.Getenv("BIND")
	if bind == "" {
		bind = ":8080"
	}

	go func() {
		if err := sr.Serve(bind); err != nil && err.Error() != "http: Server closed" {
			appLogger.Error("Error initializing server", "error", err)
			os.Exit(2)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	appLogger.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	sr.Shutdown(ctx)
	appLogger.Info("Goodbye!")
}
