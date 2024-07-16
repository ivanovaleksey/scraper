package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scraper/internal/scraper"
)

func main() {
	port := flag.Int("port", 8000, "web server port")
	withWeb := flag.Bool("web", false, "start web server")
	flag.Parse()

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	service := scraper.NewService(logger)
	dataDir, err := service.Run()
	if err != nil {
		logger.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}

	if *withWeb {
		logger.Info("starting web server", slog.Int("port", *port), slog.String("dir", dataDir))

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: http.FileServer(http.Dir(dataDir)),
		}

		done := make(chan struct{}, 1)
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			defer close(done)
			err := server.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("failed to run web server", slog.Any("error", err))
			}
		}()

		<-ch
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			logger.Error("failed to shutdown web server", slog.Any("error", err))
		}
		<-done
	}
}
