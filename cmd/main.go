package main

import (
	"log/slog"
	"os"

	"scraper/internal/scraper"
)

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	logger.Debug("start")

	service := scraper.NewService(logger)
	err := service.Run()
	if err != nil {
		logger.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Debug("done")
}
