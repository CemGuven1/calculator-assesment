package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// logRequests logs one line per request with its method, path, status and
// duration. Server errors (5xx) are logged at error level.
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		level := slog.LevelInfo
		if rec.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		logger.LogAttrs(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// statusRecorder remembers the status code sent, for the request log.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status, r.wroteHeader = status, true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true // an implicit 200, which status already holds
	return r.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying ResponseWriter.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// recoverPanics turns a panic in a handler into a JSON 500 response, and logs
// the panic with its stack trace. Without it, net/http would drop the
// connection and the client would get no response at all.
func recoverPanics(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if v == http.ErrAbortHandler {
				panic(v) // a deliberate abort, which net/http handles quietly
			}
			logger.ErrorContext(r.Context(), "panic serving request",
				"panic", v, "stack", string(debug.Stack()))
			writeError(w, internalError())
		}()
		next.ServeHTTP(w, r)
	})
}

// cors lets browsers on the allowed origins, such as the Vite dev server, call
// the API directly. Same-origin requests don't need it: those through the Vite
// proxy, and those from the frontend when this server serves it.
func cors(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Add("Vary", "Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Answer preflight requests here, since the router rejects OPTIONS.
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
