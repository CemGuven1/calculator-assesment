package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.DiscardHandler)

func TestLoadConfig(t *testing.T) {
	defaultOrigins := []string{"http://localhost:5173"}
	tests := []struct {
		name string
		env  map[string]string
		want config
	}{
		{
			name: "defaults",
			want: config{addr: "127.0.0.1:8080", corsOrigins: defaultOrigins},
		},
		{
			name: "all set",
			env: map[string]string{
				"HOST": "0.0.0.0", "PORT": "9090", "STATIC_DIR": "/srv/www",
				"CORS_ORIGINS": "http://a.example, http://b.example",
			},
			want: config{addr: "0.0.0.0:9090", staticDir: "/srv/www", corsOrigins: []string{"http://a.example", "http://b.example"}},
		},
		{
			name: "empty values fall back to defaults",
			env:  map[string]string{"HOST": "", "PORT": "", "STATIC_DIR": ""},
			want: config{addr: "127.0.0.1:8080", corsOrigins: defaultOrigins},
		},
		{
			name: "empty CORS_ORIGINS disables CORS",
			env:  map[string]string{"CORS_ORIGINS": ""},
			want: config{addr: "127.0.0.1:8080"},
		},
		{
			name: "blank CORS_ORIGINS entries are dropped",
			env:  map[string]string{"CORS_ORIGINS": " , http://a.example,, "},
			want: config{addr: "127.0.0.1:8080", corsOrigins: []string{"http://a.example"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookup := func(key string) (string, bool) {
				v, ok := tc.env[key]
				return v, ok
			}
			if got := loadConfig(lookup); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	file := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(file, []byte("<html></html>"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		cfg     config
		wantErr bool
	}{
		{name: "valid config", cfg: config{addr: "127.0.0.1:0", staticDir: t.TempDir()}},
		{name: "missing static dir", cfg: config{addr: "127.0.0.1:0", staticDir: filepath.Join(t.TempDir(), "missing")}, wantErr: true},
		{name: "static dir is a file", cfg: config{addr: "127.0.0.1:0", staticDir: file}, wantErr: true},
		{name: "invalid address", cfg: config{addr: "127.0.0.1:-1"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// A cancelled context makes a successful run shut down straight away.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := run(ctx, tc.cfg, discardLogger)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("run() error = %v, want error: %t", err, tc.wantErr)
			}
		})
	}
}

// Uses a directory laid out like the built frontend, and the real file server.
func TestStaticFiles(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"index.html":         "<title>Calculator</title>",
		"assets/app.js":      "console.log('app')",
		"assets/icons/a.svg": "<svg/>",
	} {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	handler, err := newHandler(config{staticDir: dir}, discardLogger)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", 200, "<title>Calculator</title>"},
		{"/assets/app.js", 200, "console.log('app')"},
		{"/assets/icons/a.svg", 200, "<svg/>"},
		// Directories without an index.html are not listed.
		{"/assets/", 404, "404 page not found\n"},
		{"/assets/icons/", 404, "404 page not found\n"},
		{"/missing.js", 404, "404 page not found\n"},
		// The API keeps its JSON errors.
		{"/api/nope", 404, `{"error":{"code":"NOT_FOUND","message":"not found"}}` + "\n"},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != tc.wantStatus || rec.Body.String() != tc.wantBody {
				t.Errorf("GET %s = %d %q, want %d %q", tc.path, rec.Code, rec.Body.String(), tc.wantStatus, tc.wantBody)
			}
		})
	}
}

func TestServeReturnsListenerErrors(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()

	if err := serve(context.Background(), ln, http.NotFoundHandler(), discardLogger); err == nil {
		t.Error("serve() on a closed listener = nil, want an error")
	}
}

func TestServeFinishesInFlightRequestsOnShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	slow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "done")
	})

	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, ln, slow, discardLogger) }()

	type response struct {
		body string
		err  error
	}
	responses := make(chan response, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			responses <- response{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		responses <- response{string(body), err}
	}()

	<-started
	cancel() // shut down while the request is in flight

	select {
	case err := <-served:
		t.Fatalf("serve returned %v before the in-flight request finished", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	if r := <-responses; r.err != nil || r.body != "done" {
		t.Errorf("in-flight request got (%q, %v), want (\"done\", nil)", r.body, r.err)
	}
	if err := <-served; err != nil {
		t.Errorf("serve() = %v, want nil after a graceful shutdown", err)
	}
}
