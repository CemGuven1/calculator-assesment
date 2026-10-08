package api

import (
	"cmp"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

// Config configures the handler returned by NewHandler.
type Config struct {
	// Logger receives request logs and recovered panics. Defaults to
	// slog.Default().
	Logger *slog.Logger
	// AllowedOrigins lists the browser origins that may call the API from
	// another origin, such as the Vite dev server.
	AllowedOrigins []string
	// Static, if set, serves every path outside the API, such as the built
	// frontend. Otherwise those paths return 404.
	Static http.Handler
}

// NewHandler returns the application's routes, wrapped in middleware for
// request logging, panic recovery and CORS.
func NewHandler(cfg Config) http.Handler {
	logger := cmp.Or(cfg.Logger, slog.Default())
	h := &handlers{logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/calculate", h.calculate)
	mux.HandleFunc("GET /health", health)

	// These patterns have no method, so they only receive requests that the
	// more specific patterns above don't match. Wrong methods and unknown
	// paths then get a JSON error instead of ServeMux's plain-text one.
	mux.Handle("/api/v1/calculate", methodNotAllowed(http.MethodPost))
	mux.Handle("/health", methodNotAllowed(http.MethodGet, http.MethodHead))
	mux.HandleFunc("/api", notFound) // otherwise ServeMux redirects /api to /api/
	mux.HandleFunc("/api/", notFound)
	if cfg.Static != nil {
		mux.Handle("/", cfg.Static)
	} else {
		mux.HandleFunc("/", notFound)
	}

	return logRequests(logger, recoverPanics(logger, cors(cfg.AllowedOrigins, rejectUncleanAPIPaths(mux))))
}

// handlers holds what the request handlers share.
type handlers struct {
	logger *slog.Logger
}

// rejectUncleanAPIPaths answers API paths such as /api//v1/calculate with a
// JSON 404. ServeMux would redirect them instead, and a client that follows
// the redirect turns a POST into a GET.
func rejectUncleanAPIPaths(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := r.URL.Path; strings.HasPrefix(p, "/api/") && path.Clean(p) != p {
			notFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type healthResponse struct {
	Status string `json:"status"`
}

// health handles GET /health.
func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, &apiError{http.StatusNotFound, codeNotFound, "not found"})
}

// methodNotAllowed answers requests to an existing path with a method other
// than the allowed ones.
func methodNotAllowed(allowed ...string) http.Handler {
	allow := strings.Join(allowed, ", ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		message := fmt.Sprintf("method %s is not allowed; use %s", r.Method, allow)
		writeError(w, &apiError{http.StatusMethodNotAllowed, codeMethodNotAllowed, message})
	})
}
