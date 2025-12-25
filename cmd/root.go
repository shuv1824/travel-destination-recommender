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

	"github.com/gorilla/mux"
	"github.com/shuv1824/recommender/internal/handlers"
	"github.com/shuv1824/recommender/internal/services/weather"
	"github.com/shuv1824/recommender/internal/utils/geodata"
)

func Run() error {
	logger := setupLogger()
	slog.SetDefault(logger)

	if err := geodata.Load("data/districts.json"); err != nil {
		return fmt.Errorf("failed to load geodata: %w", err)
	}

	districts := geodata.Districts()
	slog.Info("Loaded districts", "count", len(districts))

	weatherService := weather.NewCachedWeatherService(districts, 5*time.Minute)
	weatherHandler := handlers.NewWeatherHandler(weatherService)

	// Warm cache on startup (fetch data before serving requests)
	slog.Info("Warming weather cache...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := weatherService.WarmCache(ctx); err != nil {
		slog.Error("Warning: failed to warm cache: ", "error", err)
	} else {
		slog.Info("Cache warmed successfully")
	}
	cancel()

	// Start background cache refresh
	weatherService.StartBackgroundRefresh(context.Background())

	// Initialize router
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	// API v1 subrouter
	api := r.PathPrefix("/api/v1").Subrouter()

	// Weather/Destination routes
	api.HandleFunc("/destinations/top", weatherHandler.GetTopDestinations).Methods(http.MethodGet)

	slog.Info("starting backend server")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
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
			_ = server.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		slog.Info("server stopped gracefully")
	}

	return nil
}
