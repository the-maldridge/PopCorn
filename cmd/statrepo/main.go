package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/popcorn/pkg/stats"
)

var ()

func main() {
	llevel := os.Getenv("LOG_LEVEL")
	if llevel == "" {
		llevel = "INFO"
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "statrepo",
		Level: hclog.LevelFromString(llevel),
	})

	sr := stats.New(appLogger)

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
}
