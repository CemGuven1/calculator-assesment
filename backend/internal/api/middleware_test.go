package api_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CemGuven1/calculator-assesment/backend/internal/api"
)

func TestCORS(t *testing.T) {
	const devOrigin = "http://localhost:5173"
	const otherOrigin = "http://evil.example"
	h := api.NewHandler(api.Config{Logger: discardLogger, AllowedOrigins: []string{devOrigin}})

	tests := []struct {
		name       string
		method     string
		origin     string
		preflight  bool
		wantStatus int
		// Expected header values; "" means the header must be absent.
		wantHeaders map[string]string
	}{
		{
			name: "request from the allowed origin", method: http.MethodPost, origin: devOrigin, wantStatus: 200,
			wantHeaders: map[string]string{"Access-Control-Allow-Origin": devOrigin, "Vary": "Origin"},
		},
		{
			name: "request from another origin", method: http.MethodPost, origin: otherOrigin, wantStatus: 200,
			wantHeaders: map[string]string{"Access-Control-Allow-Origin": ""},
		},
		{
			name: "same-origin request", method: http.MethodPost, wantStatus: 200,
			wantHeaders: map[string]string{"Access-Control-Allow-Origin": ""},
		},
		{
			name: "preflight from the allowed origin", method: http.MethodOptions, origin: devOrigin, preflight: true, wantStatus: 204,
			wantHeaders: map[string]string{
				"Access-Control-Allow-Origin":  devOrigin,
				"Access-Control-Allow-Methods": "GET, POST",
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Max-Age":       "600",
			},
		},
		{
			name: "preflight from another origin", method: http.MethodOptions, origin: otherOrigin, preflight: true, wantStatus: 204,
			wantHeaders: map[string]string{"Access-Control-Allow-Origin": "", "Access-Control-Allow-Methods": ""},
		},
		{
			name: "OPTIONS that is not a preflight goes to the router", method: http.MethodOptions, origin: devOrigin, wantStatus: 405,
			wantHeaders: map[string]string{"Access-Control-Allow-Origin": devOrigin, "Allow": "POST"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, calculatePath, nil)
			if tc.method == http.MethodPost {
				req = postCalculate(`{"operation":"add","operands":[1,2]}`)
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.preflight {
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "content-type")
			}

			rec := serve(h, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			for name, want := range tc.wantHeaders {
				if got := rec.Header().Get(name); got != want {
					t.Errorf("%s = %q, want %q", name, got, want)
				}
			}
		})
	}
}

// logEntries parses the JSON log lines written to buf.
func logEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	dec := json.NewDecoder(buf)
	for dec.More() {
		var entry map[string]any
		if err := dec.Decode(&entry); err != nil {
			t.Fatalf("decoding log line: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func TestRequestLogging(t *testing.T) {
	var buf bytes.Buffer
	h := api.NewHandler(api.Config{Logger: slog.New(slog.NewJSONHandler(&buf, nil))})

	serve(h, postCalculate(`{"operation":"divide","operands":[1,0]}`))

	entries := logEntries(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("got %d log lines, want 1: %v", len(entries), entries)
	}
	entry := entries[0]
	want := map[string]any{"level": "INFO", "msg": "request", "method": "POST", "path": calculatePath, "status": 422.0}
	for key, value := range want {
		if entry[key] != value {
			t.Errorf("log %s = %v, want %v", key, entry[key], value)
		}
	}
	if _, ok := entry["duration"].(float64); !ok {
		t.Errorf("log duration = %v, want a number", entry["duration"])
	}
}

func TestPanicRecovery(t *testing.T) {
	var buf bytes.Buffer
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
	h := api.NewHandler(api.Config{Logger: slog.New(slog.NewJSONHandler(&buf, nil)), Static: panicking})

	rec := serve(h, httptest.NewRequest(http.MethodGet, "/", nil))

	// The client gets the standard error body, without panic details.
	assertJSON(t, rec, 500, errorJSON("INTERNAL", "internal server error"))

	// The panic and its stack go to the log, and the request is logged as an error.
	entries := logEntries(t, &buf)
	if len(entries) != 2 {
		t.Fatalf("got %d log lines, want 2: %v", len(entries), entries)
	}
	if p := entries[0]; p["level"] != "ERROR" || p["msg"] != "panic serving request" || p["panic"] != "boom" || p["stack"] == "" {
		t.Errorf("panic log = %v", p)
	}
	if r := entries[1]; r["level"] != "ERROR" || r["msg"] != "request" || r["status"] != 500.0 {
		t.Errorf("request log = %v", r)
	}
}

// http.ErrAbortHandler is how a handler deliberately aborts a response, so it
// must reach net/http instead of becoming a 500.
func TestAbortHandlerPanicIsNotRecovered(t *testing.T) {
	aborting := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) })
	h := api.NewHandler(api.Config{Logger: discardLogger, Static: aborting})

	defer func() {
		if got := recover(); got != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler", got)
		}
	}()
	serve(h, httptest.NewRequest(http.MethodGet, "/", nil))
}

// The logging middleware wraps the ResponseWriter; handlers must still reach
// optional features such as flushing through http.ResponseController.
func TestResponseControllerWorksThroughMiddleware(t *testing.T) {
	var flushErr error
	flushing := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flushErr = http.NewResponseController(w).Flush()
	})
	h := api.NewHandler(api.Config{Logger: discardLogger, Static: flushing})

	serve(h, httptest.NewRequest(http.MethodGet, "/", nil))

	if flushErr != nil {
		t.Errorf("Flush() = %v, want nil", flushErr)
	}
}

func TestNilLoggerUsesDefault(t *testing.T) {
	h := api.NewHandler(api.Config{})

	assertJSON(t, serve(h, httptest.NewRequest(http.MethodGet, "/health", nil)), 200, `{"status":"ok"}`)
}
