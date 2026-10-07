# Calculator — Design

## 1. Repo layout

```
calculator/
├── backend/                     Go 1.25, standard library only
│   ├── cmd/server/main.go       config (env vars), wiring, graceful shutdown
│   └── internal/
│       ├── calculator/          pure domain: operation registry, arity, domain errors
│       └── api/                 routing, JSON decode/validate, error → status, middleware
├── frontend/                    Vite + React + TS (runtime deps: react, react-dom only)
│   └── src/
│       ├── api/                 typed fetch client, contract types, user-facing error messages
│       ├── calculator/          operations, input parsing, state reducer, useCalculator hook
│       └── components/          Calculator UI
├── docs/DESIGN.md
├── Dockerfile                   multi-stage: node build → go build → minimal runtime
├── README.md                    setup, run, API examples, decisions
└── PROMPTS.md                   AI prompts used
```

`calculator` has no HTTP imports, so the math is tested without a server. In dev, Vite proxies
`/api` to Go on `127.0.0.1:8080`. In Docker, the Go binary also serves the built SPA from
`STATIC_DIR`. Either way the browser sees a single origin. CORS is still enabled for the Vite dev
origin (configurable with `CORS_ORIGINS`), so the API can also be called from it directly.

## 2. API contract

**Decision: one `POST /api/v1/calculate`, operation named in the body.** An operation is data,
not a resource. One handler, one validation path and one registry (`name → arity, func`) means
adding an operation is a single table entry. The frontend needs one client function. Trade-off:
an unknown operation returns 400 instead of a routing 404. Also: `GET /health` → `{"status":"ok"}`,
outside the versioned API so infrastructure health checks don't depend on the API version.

```jsonc
// Request: operands is always an array whose length must equal the operation's arity
{ "operation": "divide", "operands": [10, 4] }
{ "operation": "sqrt",   "operands": [16] }      // unary: exactly one operand

// 200
{ "result": 2.5 }

// 4xx/5xx: one shape for every error
{ "error": { "code": "DIVISION_BY_ZERO", "message": "division by zero" } }
```

| `operation` | Arity | Result |
|---|---|---|
| `add`, `subtract`, `multiply`, `divide` | 2 | a + b, a − b, a × b, a ÷ b |
| `power` | 2 | a^b |
| `sqrt` | 1 | √a |
| `percentage` | 2 | **a % of b** = a × b / 100, e.g. `[15, 200]` → `30` |

Operation names are lowercase and case-sensitive. `code` is stable and meant for programs; `message` is for people.

## 3. Edge cases

**400** = the request is invalid. **422** = the request is valid, but the math has no real, finite answer.

| Case | Status | `code` |
|---|---|---|
| Malformed JSON, wrong type (`"5"`), unknown field, trailing data, number beyond float64 (`1e400`) | 400 | `INVALID_REQUEST` |
| `operation` missing, empty or unknown | 400 | `UNKNOWN_OPERATION` |
| `operands` missing, `null`, wrong count, or containing `null`¹ | 400 | `INVALID_OPERANDS` |
| `divide` by 0; `power` with a = 0, b < 0 | 422 | `DIVISION_BY_ZERO` |
| `sqrt` of a negative; `power` with a < 0 and non-integer b | 422 | `DOMAIN_ERROR` |
| Result overflows to ±Inf (`1e308 × 10`) | 422 | `OVERFLOW` |
| Body > 1 KiB | 413 | `PAYLOAD_TOO_LARGE` |
| `Content-Type` not `application/json` | 415 | `UNSUPPORTED_MEDIA_TYPE` |
| Wrong method / unknown `/api` path | 405 / 404 | `METHOD_NOT_ALLOWED` / `NOT_FOUND` |
| Panic | 500 | `INTERNAL` (recovered; details only in logs) |

¹ Go silently decodes a `null` array element as `0`. Operands are decoded as `[]*float64` and any nil is rejected.

- Explicit checks run first so messages are precise.
- A final guard then checks every result: NaN → `DOMAIN_ERROR`, ±Inf → `OVERFLOW`. NaN/Inf can never be serialized.
- `-0` is normalized to `0`.
- `0^0 = 1`, following Go's `math.Pow`.
- `NaN`/`Infinity` inputs can't happen, because they aren't valid JSON.
- Results are raw float64: `0.1 + 0.2` returns `0.30000000000000004`. Rounding for display is the UI's job.

## 4. Testing strategy

**Backend** (`go test -race -cover ./...` plus an HTML coverage report)
- `calculator`: table-driven tests for each operation: happy paths, negatives, decimals,
  every 422 case, arity errors and `-0`. Pure functions, so aim for ~100% coverage.
- `api`: `httptest` against the real router, one table row per edge case above. Each test
  checks status, `Content-Type` and the exact JSON body.

**Frontend** (Vitest + Testing Library + jsdom, `@vitest/coverage-v8`)
- Pure units: input parsing, the state reducer, error messages, and the API client with a
  mocked `fetch`. The client turns each outcome (2xx, 4xx error body, non-JSON body, network
  failure) into a typed result.
- Components, with `fetch` mocked: `sqrt` hides the second input; invalid input shows an
  inline error and sends no request; success shows the result; 422 and network failures
  show a message; the button is disabled while loading.
- No browser end-to-end tests, to fit the time budget. The Docker image is smoke-tested with the README's curl examples.

## 5. UI approach

**A form, not a keypad.** A keypad needs client-side state for button presses and makes per-field
validation awkward. A form maps 1:1 onto the API.
- Layout: operation buttons (native radios styled as a grid; arrow keys change the choice) →
  number fields labelled for the operation (the second is hidden for `sqrt`) → **Calculate**
  (Enter submits) → result panel.
- Inputs are plain `type="text"`. `type="number"` reports partial input like `1e` as empty, so
  validation would be wrong, and `inputMode="decimal"` has no minus key on iOS.
  Our parser accepts `-3.5` and `1e3`. It rejects empty input, `abc`, a second decimal point
  and `Infinity`.
- The frontend validates **input shape only**: required, finite number. The math rules
  (÷0, √−x) live only in Go. The UI turns the server's error `code` into a friendly message.
  No rule exists in two places.
- The result shows the full expression, e.g. `15% of 200 = 30`, rounded to 15 significant digits.
- Errors appear in a `role="alert"` region. Invalid fields get `aria-invalid` and `aria-describedby`,
  and focus moves to the first one after a failed submit. While a request runs, the button is
  `aria-disabled` rather than `disabled`, which would drop keyboard focus.
- Plain CSS with variables, built for mobile first and checked at 360px: fields side by side
  from 480px, 44px touch targets, long results shrink to stay on one line, and dark mode
  follows the system.
