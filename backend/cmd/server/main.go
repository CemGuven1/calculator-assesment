// Command server runs the calculator HTTP API. When STATIC_DIR is set it also
// serves the built frontend, so a single process can run the whole app.
//
// It is configured with environment variables:
//
//	HOST          interface to listen on (default 127.0.0.1; use 0.0.0.0 in a container)
//	PORT          port to listen on (default 8080)
//	STATIC_DIR    directory of the built frontend to serve at / (optional)
//	CORS_ORIGINS  comma-separated browser origins allowed to call the API
//	              (default http://localhost:5173, the Vite dev server; set it
//	              to an empty string to disable CORS)
//
// On SIGINT or SIGTERM it stops accepting connections and lets in-flight
// requests finish before exiting.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/CemGuven1/calculator-assesment/backend/internal/api"
)

// shutdownTimeout is how long in-flight requests get to finish on shutdown.
const shutdownTimeout = 10 * time.Second

type config struct {
	addr        string
	staticDir   string
	corsOrigins []string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// After the first signal, restore the default behavior, so a second
	// Ctrl+C exits at once instead of waiting for the graceful shutdown.
	context.AfterFunc(ctx, stop)
	err := run(ctx, loadConfig(os.LookupEnv), logger)
	stop()
	if err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

// loadConfig reads the configuration through lookup, which is os.LookupEnv
// outside of tests.
func loadConfig(lookup func(key string) (string, bool)) config {
	get := func(key, fallback string) string {
		if v, ok := lookup(key); ok && v != "" {
			return v
		}
		return fallback
	}
	cfg := config{
		addr:        net.JoinHostPort(get("HOST", "127.0.0.1"), get("PORT", "8080")),
		staticDir:   get("STATIC_DIR", ""),
		corsOrigins: []string{"http://localhost:5173"},
	}
	// Unlike the other settings, an empty CORS_ORIGINS means something:
	// it disables CORS.
	if v, ok := lookup("CORS_ORIGINS"); ok {
		cfg.corsOrigins = splitList(v)
	}
	return cfg
}

// splitList splits a comma-separated list, dropping blank entries.
func splitList(s string) []string {
	var items []string
	for item := range strings.SplitSeq(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// run serves the app on cfg.addr until ctx is cancelled.
func run(ctx context.Context, cfg config, logger *slog.Logger) error {
	handler, err := newHandler(cfg, logger)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return err
	}
	logger.Info("listening", "addr", ln.Addr().String(), "static_dir", cfg.staticDir, "cors_origins", cfg.corsOrigins)
	return serve(ctx, ln, handler, logger)
}

// newHandler builds the API handler, which also serves cfg.staticDir if set.
func newHandler(cfg config, logger *slog.Logger) (http.Handler, error) {
	var static http.Handler
	if cfg.staticDir != "" {
		info, err := os.Stat(cfg.staticDir)
		if err != nil {
			return nil, fmt.Errorf("STATIC_DIR: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("STATIC_DIR: %s is not a directory", cfg.staticDir)
		}
		static = staticFiles(cfg.staticDir)
	}
	return api.NewHandler(api.Config{Logger: logger, AllowedOrigins: cfg.corsOrigins, Static: static}), nil
}

// serve serves HTTP on ln until ctx is cancelled, then shuts down gracefully:
// it stops accepting connections and waits up to shutdownTimeout for
// in-flight requests to finish.
func serve(ctx context.Context, ln net.Listener, handler http.Handler, logger *slog.Logger) error {
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		return err // Serve failed before any shutdown was requested.
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
