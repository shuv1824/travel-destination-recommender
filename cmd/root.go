package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() error {
	logger := setupLogger()
	slog.SetDefault(logger)

	slog.Info("starting backend server")

	server := &http.Server{
		Addr: ":8080",
	}
	return startServer(server)
}

func setupLogger() *slog.Logger {
	var handler slog.Handler
	isDevelopment := true

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if isDevelopment {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func startServer(server *http.Server) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverError := make(chan error, 1)

	go func() {
		slog.Info("server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	select {
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		slog.Info("server stopped gracefully")
	}

	return nil
}
