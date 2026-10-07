package api_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/CemGuven1/calculator-assesment/backend/internal/api"
)

const calculatePath = "/api/v1/calculate"

var discardLogger = slog.New(slog.DiscardHandler)

func newHandler() http.Handler {
	return api.NewHandler(api.Config{Logger: discardLogger})
}

func serve(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// postCalculate builds a POST /api/v1/calculate request with a JSON body.
func postCalculate(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, calculatePath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// errorJSON is the expected body of an error response.
func errorJSON(code, message string) string {
	b, err := json.Marshal(map[string]map[string]string{"error": {"code": code, "message": message}})
	if err != nil {
		panic(err)
	}
	return string(b)
}

// assertJSON checks the status, the JSON headers and the exact body.
func assertJSON(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantBody string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d", rec.Code, wantStatus)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := strings.TrimSuffix(rec.Body.String(), "\n"); got != wantBody {
		t.Errorf("body = %s\n  want %s", got, wantBody)
	}
}

func TestCalculate(t *testing.T) {
	const supported = "supported operations: add, divide, multiply, percentage, power, sqrt, subtract"

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		// Success.
		{"add", `{"operation":"add","operands":[2,3]}`, 200, `{"result":5}`},
		{"subtract to a negative", `{"operation":"subtract","operands":[3,5]}`, 200, `{"result":-2}`},
		{"multiply decimals", `{"operation":"multiply","operands":[4,2.5]}`, 200, `{"result":10}`},
		{"divide", `{"operation":"divide","operands":[10,4]}`, 200, `{"result":2.5}`},
		{"power", `{"operation":"power","operands":[2,10]}`, 200, `{"result":1024}`},
		{"sqrt takes one operand", `{"operation":"sqrt","operands":[16]}`, 200, `{"result":4}`},
		{"percentage", `{"operation":"percentage","operands":[15,200]}`, 200, `{"result":30}`},
		{"results keep float64 precision", `{"operation":"add","operands":[0.1,0.2]}`, 200, `{"result":0.30000000000000004}`},
		{"exponent notation in operands", `{"operation":"multiply","operands":[1e3,-2E-1]}`, 200, `{"result":-200}`},
		{"large results use exponent notation", `{"operation":"multiply","operands":[2e300,0.5]}`, 200, `{"result":1e+300}`},
		{"negative zero is sent as 0", `{"operation":"multiply","operands":[-1,0]}`, 200, `{"result":0}`},
		{"field order and whitespace don't matter", "\n { \"operands\": [1, 2], \"operation\": \"add\" }\n", 200, `{"result":3}`},

		// Valid request, but no real, finite answer: 422.
		{"divide by zero", `{"operation":"divide","operands":[1,0]}`, 422, errorJSON("DIVISION_BY_ZERO", "division by zero")},
		{"zero to a negative power", `{"operation":"power","operands":[0,-1]}`, 422, errorJSON("DIVISION_BY_ZERO", "division by zero: zero raised to a negative power")},
		{"sqrt of a negative", `{"operation":"sqrt","operands":[-4]}`, 422, errorJSON("DOMAIN_ERROR", "result is not a real number: square root of a negative number")},
		{"negative base, fractional exponent", `{"operation":"power","operands":[-8,0.5]}`, 422, errorJSON("DOMAIN_ERROR", "result is not a real number: negative base with a fractional exponent")},
		{"overflow", `{"operation":"multiply","operands":[1e308,10]}`, 422, errorJSON("OVERFLOW", "result is out of range")},

		// Invalid operation: 400.
		{"unknown operation", `{"operation":"modulo","operands":[1,2]}`, 400, errorJSON("UNKNOWN_OPERATION", `unknown operation "modulo"; `+supported)},
		{"operation names are case-sensitive", `{"operation":"ADD","operands":[1,2]}`, 400, errorJSON("UNKNOWN_OPERATION", `unknown operation "ADD"; `+supported)},
		{"missing operation", `{"operands":[1,2]}`, 400, errorJSON("UNKNOWN_OPERATION", "operation is required")},
		{"empty operation", `{"operation":"","operands":[1,2]}`, 400, errorJSON("UNKNOWN_OPERATION", "operation is required")},
		{"null operation", `{"operation":null,"operands":[1,2]}`, 400, errorJSON("UNKNOWN_OPERATION", "operation is required")},

		// Invalid operands: 400.
		{"missing operands", `{"operation":"add"}`, 400, errorJSON("INVALID_OPERANDS", "invalid operand: add expects 2 operands, got 0")},
		{"null operands", `{"operation":"add","operands":null}`, 400, errorJSON("INVALID_OPERANDS", "invalid operand: add expects 2 operands, got 0")},
		{"empty operands", `{"operation":"add","operands":[]}`, 400, errorJSON("INVALID_OPERANDS", "invalid operand: add expects 2 operands, got 0")},
		{"too few operands", `{"operation":"add","operands":[1]}`, 400, errorJSON("INVALID_OPERANDS", "invalid operand: add expects 2 operands, got 1")},
		{"too many operands", `{"operation":"add","operands":[1,2,3]}`, 400, errorJSON("INVALID_OPERANDS", "invalid operand: add expects 2 operands, got 3")},
		{"two operands for sqrt", `{"operation":"sqrt","operands":[4,9]}`, 400, errorJSON("INVALID_OPERANDS", "invalid operand: sqrt expects 1 operand, got 2")},
		{"null operand", `{"operation":"add","operands":[1,null]}`, 400, errorJSON("INVALID_OPERANDS", "operands[1] must be a number, got null")},

		// Malformed request: 400.
		{"empty body", ``, 400, errorJSON("INVALID_REQUEST", "request body is empty")},
		{"whitespace-only body", " \n ", 400, errorJSON("INVALID_REQUEST", "request body is empty")},
		{"not JSON", `hello`, 400, errorJSON("INVALID_REQUEST", "request body is not valid JSON")},
		{"truncated JSON", `{"operation":"add","operands":[1,`, 400, errorJSON("INVALID_REQUEST", "request body is not valid JSON")},
		{"trailing comma", `{"operation":"add","operands":[1,2],}`, 400, errorJSON("INVALID_REQUEST", "request body is not valid JSON")},
		{"NaN is not JSON", `{"operation":"add","operands":[NaN,1]}`, 400, errorJSON("INVALID_REQUEST", "request body is not valid JSON")},
		{"Infinity is not JSON", `{"operation":"add","operands":[Infinity,1]}`, 400, errorJSON("INVALID_REQUEST", "request body is not valid JSON")},
		{"array instead of object", `[1,2]`, 400, errorJSON("INVALID_REQUEST", "request body must be a JSON object")},
		{"string instead of object", `"add"`, 400, errorJSON("INVALID_REQUEST", "request body must be a JSON object")},
		{"operation is a number", `{"operation":5,"operands":[1,2]}`, 400, errorJSON("INVALID_REQUEST", "operation must be a string")},
		{"operands is a string", `{"operation":"add","operands":"1,2"}`, 400, errorJSON("INVALID_REQUEST", "operands must be an array of numbers")},
		{"operands is an object", `{"operation":"add","operands":{"a":1}}`, 400, errorJSON("INVALID_REQUEST", "operands must be an array of numbers")},
		{"operand is a string", `{"operation":"add","operands":["1",2]}`, 400, errorJSON("INVALID_REQUEST", "operands must be an array of numbers")},
		{"operand is a boolean", `{"operation":"add","operands":[true,2]}`, 400, errorJSON("INVALID_REQUEST", "operands must be an array of numbers")},
		{"operand beyond float64 range", `{"operation":"add","operands":[1e400,1]}`, 400, errorJSON("INVALID_REQUEST", "number 1e400 is out of range")},
		{"unknown field", `{"operation":"add","operands":[1,2],"precision":2}`, 400, errorJSON("INVALID_REQUEST", `unknown field "precision"`)},
		{"misspelled field", `{"operation":"add","operand":[1,2]}`, 400, errorJSON("INVALID_REQUEST", `unknown field "operand"`)},
		{"two objects", `{"operation":"add","operands":[1,2]}{"operation":"add","operands":[3,4]}`, 400, errorJSON("INVALID_REQUEST", "request body must contain a single JSON object")},
		{"trailing data", `{"operation":"add","operands":[1,2]} x`, 400, errorJSON("INVALID_REQUEST", "request body must contain a single JSON object")},
	}
	h := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertJSON(t, serve(h, postCalculate(tc.body)), tc.wantStatus, tc.wantBody)
		})
	}
}

func TestCalculateContentType(t *testing.T) {
	unsupported := errorJSON("UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
	tests := []struct {
		name        string
		contentType string
		wantStatus  int
		wantBody    string
	}{
		{"missing", "", 415, unsupported},
		{"plain text", "text/plain", 415, unsupported},
		{"form", "application/x-www-form-urlencoded", 415, unsupported},
		{"malformed", "application/json; charset", 415, unsupported},
		{"with charset", "application/json; charset=utf-8", 200, `{"result":3}`},
		{"media types are case-insensitive", "Application/JSON", 200, `{"result":3}`},
	}
	h := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := postCalculate(`{"operation":"add","operands":[1,2]}`)
			req.Header.Del("Content-Type")
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			assertJSON(t, serve(h, req), tc.wantStatus, tc.wantBody)
		})
	}
}

func TestCalculateBodySizeLimit(t *testing.T) {
	const valid = `{"operation":"add","operands":[1,2]}`
	tooLarge := errorJSON("PAYLOAD_TOO_LARGE", "request body must not exceed 1024 bytes")
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"exactly at the limit", valid + strings.Repeat(" ", 1024-len(valid)), 200, `{"result":3}`},
		{"one byte over the limit", valid + strings.Repeat(" ", 1025-len(valid)), 413, tooLarge},
		{"object over the limit", `{"operation":"add",` + strings.Repeat(" ", 2000) + `"operands":[1,2]}`, 413, tooLarge},
	}
	h := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertJSON(t, serve(h, postCalculate(tc.body)), tc.wantStatus, tc.wantBody)
		})
	}
}

func TestCalculateBodyReadError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, calculatePath, iotest.ErrReader(errors.New("connection reset")))
	req.Header.Set("Content-Type", "application/json")

	assertJSON(t, serve(newHandler(), req), 400, errorJSON("INVALID_REQUEST", "request body could not be read"))
}

func TestRouting(t *testing.T) {
	notFound := errorJSON("NOT_FOUND", "not found")
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantAllow  string
		wantBody   string
	}{
		{"health", http.MethodGet, "/health", 200, "", `{"status":"ok"}`},
		{"GET calculate", http.MethodGet, calculatePath, 405, "POST", errorJSON("METHOD_NOT_ALLOWED", "method GET is not allowed; use POST")},
		{"PUT calculate", http.MethodPut, calculatePath, 405, "POST", errorJSON("METHOD_NOT_ALLOWED", "method PUT is not allowed; use POST")},
		{"DELETE calculate", http.MethodDelete, calculatePath, 405, "POST", errorJSON("METHOD_NOT_ALLOWED", "method DELETE is not allowed; use POST")},
		{"POST health", http.MethodPost, "/health", 405, "GET, HEAD", errorJSON("METHOD_NOT_ALLOWED", "method POST is not allowed; use GET, HEAD")},
		{"unknown API path", http.MethodGet, "/api/v1/unknown", 404, "", notFound},
		{"unknown API version", http.MethodPost, "/api/v2/calculate", 404, "", notFound},
		{"trailing slash", http.MethodPost, calculatePath + "/", 404, "", notFound},
		{"root without static files", http.MethodGet, "/", 404, "", notFound},
	}
	h := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(h, httptest.NewRequest(tc.method, tc.path, nil))
			assertJSON(t, rec, tc.wantStatus, tc.wantBody)
			if got := rec.Header().Get("Allow"); got != tc.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tc.wantAllow)
			}
		})
	}
}

func TestStaticFiles(t *testing.T) {
	static := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "static "+r.URL.Path)
	})
	h := api.NewHandler(api.Config{Logger: discardLogger, Static: static})

	for _, path := range []string{"/", "/assets/app.js"} {
		rec := serve(h, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != 200 || rec.Body.String() != "static "+path {
			t.Errorf("GET %s = %d %q, want 200 %q", path, rec.Code, rec.Body.String(), "static "+path)
		}
	}

	// The API keeps its own routes and JSON errors.
	assertJSON(t, serve(h, httptest.NewRequest(http.MethodGet, "/health", nil)), 200, `{"status":"ok"}`)
	assertJSON(t, serve(h, httptest.NewRequest(http.MethodGet, "/api/unknown", nil)), 404, errorJSON("NOT_FOUND", "not found"))
	assertJSON(t, serve(h, postCalculate(`{"operation":"add","operands":[1,2]}`)), 200, `{"result":3}`)
}
