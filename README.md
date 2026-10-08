# Calculator

[![CI](https://github.com/CemGuven1/calculator-assesment/actions/workflows/ci.yml/badge.svg)](https://github.com/CemGuven1/calculator-assesment/actions/workflows/ci.yml)

## Overview

A full-stack calculator. A Go REST API does all the math, and a React + TypeScript frontend
sends it the numbers and shows the results. It supports addition, subtraction, multiplication,
division, exponentiation, square root and percentage.

```
browser ── React app ── POST /api/v1/calculate ──▶ Go API ──▶ calculator package (pure math)
```

- **Backend:** Go 1.25, standard library only. The math lives in a package with no HTTP code;
  the API layer decodes requests strictly and maps errors to status codes.
- **Frontend:** React 19, TypeScript and Vite. Its only runtime dependencies are `react` and
  `react-dom`, and it uses plain CSS. The UI validates input, shows loading and error states,
  works with the keyboard alone and fits screens down to 360px.
- **Tests:** Go `testing` with `httptest`, and Vitest with React Testing Library. A smoke-test
  script checks every operation and error case against a running server. GitHub Actions runs
  all of it on every push.
- **Deployment:** one 16 MB Docker image in which Go serves both the API and the frontend.

The design, including every edge case, is in [docs/DESIGN.md](docs/DESIGN.md).

## Prerequisites

| To… | You need |
|---|---|
| Run everything in Docker | Docker with Compose v2 |
| Run the backend | Go 1.25 or newer |
| Run the frontend | Node.js 22.13+ or 24+, with npm |
| Try the API examples and the smoke test | `bash` and `curl`. On Windows, use Git Bash or WSL. |
| Use the `make` shortcuts (optional) | GNU make |

## Running the app

### With Docker

Builds the frontend and the backend into one image and runs it. Building the image also runs
both test suites, so a failing test stops the build.

```bash
docker compose up --build -d
```

Open <http://localhost:8080>. To stop it:

```bash
docker compose down
```

### Backend

```bash
cd backend
go run ./cmd/server
```

It logs `msg=listening addr=127.0.0.1:8080` and serves the API at <http://localhost:8080>.
Stop it with Ctrl+C: it finishes requests in progress first, and a second Ctrl+C stops it at
once. Settings come from environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `HOST` | `127.0.0.1` | Interface to listen on. The Docker image sets `0.0.0.0`. |
| `PORT` | `8080` | Port to listen on. |
| `STATIC_DIR` | not set | Directory of the built frontend to serve at `/`. The Docker image sets it. |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated browser origins allowed to call the API directly. An empty value turns CORS off, as in the Docker image. |

### Frontend

Start the backend first, then in a second terminal:

```bash
cd frontend
npm ci
npm run dev
```

Open <http://localhost:5173>. The Vite dev server forwards `/api` requests to the backend on
port 8080. To call a backend elsewhere instead, set `VITE_API_BASE_URL` (see
[frontend/.env.example](frontend/.env.example)).

### With make

From the repository root:

| Command | What it does |
|---|---|
| `make install` | Installs the frontend dependencies |
| `make run` | Starts the backend and the frontend dev server together |
| `make test` | Runs `go vet` and the Go tests, then the frontend tests |
| `make coverage` | Writes `backend/coverage.html` and `frontend/coverage/index.html` |
| `make docker-up` / `make docker-down` | Builds and starts, or stops, the Docker container |

## Tests and coverage

**Backend**

```bash
cd backend
go vet ./...
go test ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

The race detector (`go test -race ./...`) needs a C compiler, so it runs in the Docker build and
in CI rather than on every machine.

**Frontend**

```bash
cd frontend
npm test
npm run lint
npm run coverage
```

`npm run coverage` prints a summary and writes an HTML report to `frontend/coverage/index.html`.

**Smoke test.** With the app running (Docker, or the backend on its own), check every
operation and error case from the API reference below:

```bash
bash scripts/smoke-test.sh
```

It prints one line per check and exits non-zero if any fails. Pass a URL to test another
address, for example `bash scripts/smoke-test.sh http://localhost:9090`. For a full end-to-end
checklist, including manual UI steps, see [docs/SMOKE_TEST.md](docs/SMOKE_TEST.md).

**CI.** [GitHub Actions](.github/workflows/ci.yml) runs three jobs on every push and pull
request. The backend job runs `gofmt`, `go vet`, and the tests with `-race` and coverage. The
frontend job runs lint, the tests with coverage, and the type-checked build. The Docker job
builds the image and runs the smoke test against the container. Both coverage reports are
attached to each run as artifacts. The current numbers are in [docs/COVERAGE.md](docs/COVERAGE.md).

## API reference

Requests and responses are JSON. Request bodies must be sent with
`Content-Type: application/json` and be at most 1 KiB. Every API error has the same shape:
`code` is stable for programs, and `message` is for people.

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "division by zero" } }
```

### `GET /health`

```bash
curl -s http://localhost:8080/health
# {"status":"ok"}
```

### `POST /api/v1/calculate`

```json
{ "operation": "divide", "operands": [10, 4] }
```

A success returns `200` and `{"result": <number>}`.

| `operation` | Operands | Result |
|---|---|---|
| `add` | 2 | a + b |
| `subtract` | 2 | a − b |
| `multiply` | 2 | a × b |
| `divide` | 2 | a ÷ b |
| `power` | 2 | a raised to the power b |
| `sqrt` | 1 | √a |
| `percentage` | 2 | a % of b, that is a × b / 100 |

**One example per operation:**

```bash
curl -s -H 'Content-Type: application/json' -d '{"operation":"add","operands":[2,3]}' http://localhost:8080/api/v1/calculate
# {"result":5}

curl -s -H 'Content-Type: application/json' -d '{"operation":"subtract","operands":[5,8]}' http://localhost:8080/api/v1/calculate
# {"result":-3}

curl -s -H 'Content-Type: application/json' -d '{"operation":"multiply","operands":[4,2.5]}' http://localhost:8080/api/v1/calculate
# {"result":10}

curl -s -H 'Content-Type: application/json' -d '{"operation":"divide","operands":[10,4]}' http://localhost:8080/api/v1/calculate
# {"result":2.5}

curl -s -H 'Content-Type: application/json' -d '{"operation":"power","operands":[2,10]}' http://localhost:8080/api/v1/calculate
# {"result":1024}

curl -s -H 'Content-Type: application/json' -d '{"operation":"sqrt","operands":[16]}' http://localhost:8080/api/v1/calculate
# {"result":4}

curl -s -H 'Content-Type: application/json' -d '{"operation":"percentage","operands":[15,200]}' http://localhost:8080/api/v1/calculate
# {"result":30}
```

### Errors

| Status | `code` | When |
|---|---|---|
| 422 | `DIVISION_BY_ZERO` | Dividing by zero, or raising zero to a negative power |
| 422 | `DOMAIN_ERROR` | No real result: the square root of a negative number, or a negative base with a fractional exponent |
| 422 | `OVERFLOW` | The result is too large for a float64 |
| 400 | `UNKNOWN_OPERATION` | `operation` is missing, not a string, or not supported |
| 400 | `INVALID_OPERANDS` | The wrong number of operands, or an operand that is `null`, not a number, or beyond float64 |
| 400 | `INVALID_REQUEST` | The body is not JSON or not a single object, or has a field that is unknown, in the wrong case or repeated |
| 413 | `PAYLOAD_TOO_LARGE` | The body is larger than 1 KiB |
| 415 | `UNSUPPORTED_MEDIA_TYPE` | `Content-Type` is not `application/json` |
| 405 | `METHOD_NOT_ALLOWED` | The wrong HTTP method; the `Allow` header names the right one |
| 404 | `NOT_FOUND` | An unknown path under `/api` |
| 500 | `INTERNAL` | An unexpected server error; the details go only to the server log |

422 means the request is valid but the math has no real, finite answer. A 400 says which part of
the request to fix, checked in this order: the body as a whole, then the operation, then the
operands.

**One example per error case.** `-w 'HTTP %{http_code}\n'` prints the status after the body.

```bash
# 422 DIVISION_BY_ZERO
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"divide","operands":[1,0]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"DIVISION_BY_ZERO","message":"division by zero"}}
# HTTP 422

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"power","operands":[0,-1]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"DIVISION_BY_ZERO","message":"division by zero: zero raised to a negative power"}}
# HTTP 422

# 422 DOMAIN_ERROR
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"sqrt","operands":[-4]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"DOMAIN_ERROR","message":"result is not a real number: square root of a negative number"}}
# HTTP 422

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"power","operands":[-8,0.5]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"DOMAIN_ERROR","message":"result is not a real number: negative base with a fractional exponent"}}
# HTTP 422

# 422 OVERFLOW
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"multiply","operands":[1e308,10]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"OVERFLOW","message":"result is out of range"}}
# HTTP 422

# 400 UNKNOWN_OPERATION
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"modulo","operands":[1,2]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"UNKNOWN_OPERATION","message":"unknown operation \"modulo\"; supported operations: add, divide, multiply, percentage, power, sqrt, subtract"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operands":[1,2]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"UNKNOWN_OPERATION","message":"operation is required"}}
# HTTP 400

# 400 INVALID_OPERANDS
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"sqrt","operands":[4,9]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_OPERANDS","message":"invalid operand: sqrt expects 1 operand, got 2"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"add","operands":[1,null]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_OPERANDS","message":"operands[1] must be a number, got null"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"add","operands":["2",3]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_OPERANDS","message":"operands[0] must be a number, got a string"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"add","operands":[1e400,1]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_OPERANDS","message":"operands[0] is too large: 1e400"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"divide","operands":[1,1e-400]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_OPERANDS","message":"operands[1] is too small: 1e-400"}}
# HTTP 400

# 400 INVALID_REQUEST
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_REQUEST","message":"request body is not valid JSON"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '[2,3]' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_REQUEST","message":"request body must be a JSON object"}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"add","operands":[1,2],"precision":2}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_REQUEST","message":"unknown field \"precision\""}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"Operation":"add","operands":[1,2]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_REQUEST","message":"unknown field \"Operation\""}}
# HTTP 400

curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d '{"operation":"add","operands":[1,2],"operands":[3,4]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_REQUEST","message":"duplicate field \"operands\""}}
# HTTP 400

# 413 PAYLOAD_TOO_LARGE (a body of about 2 KB)
curl -s -w 'HTTP %{http_code}\n' -H 'Content-Type: application/json' -d "{\"operation\":\"add\",$(printf '%*s' 2000 '')\"operands\":[1,2]}" http://localhost:8080/api/v1/calculate
# {"error":{"code":"PAYLOAD_TOO_LARGE","message":"request body must not exceed 1024 bytes"}}
# HTTP 413

# 415 UNSUPPORTED_MEDIA_TYPE (curl -d sends a form Content-Type by default)
curl -s -w 'HTTP %{http_code}\n' -d '{"operation":"add","operands":[1,2]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}}
# HTTP 415

# 405 METHOD_NOT_ALLOWED
curl -s -w 'HTTP %{http_code}\n' http://localhost:8080/api/v1/calculate
# {"error":{"code":"METHOD_NOT_ALLOWED","message":"method GET is not allowed; use POST"}}
# HTTP 405

# 404 NOT_FOUND
curl -s -w 'HTTP %{http_code}\n' http://localhost:8080/api/v1/nope
# {"error":{"code":"NOT_FOUND","message":"not found"}}
# HTTP 404
```

`500 INTERNAL` has no example: no request can trigger it. It is the safety net for a panic or an
error the API doesn't recognize, and both are covered by tests.

## Design decisions & assumptions

The full design is in [docs/DESIGN.md](docs/DESIGN.md). Each decision below lists the reason
and what it costs.

### Decisions and trade-offs

- **One endpoint, with the operation in the body.**
  - *Why:* an operation is data, not a resource. One handler and one table of operations mean
    that adding an operation is a single entry, and the frontend needs one client function.
    Operands are always an array, so `sqrt` simply takes one.
  - *Trade-off:* an unknown operation is a 400 rather than a routing 404, and per-operation
    caching or metrics by URL aren't possible.
- **All math happens in Go.** The frontend only checks that each input is a number, then turns
  the backend's error codes into friendly messages.
  - *Why:* no rule exists in two places, so the frontend and the backend can't disagree.
  - *Trade-off:* every calculation is a network round trip, and nothing works offline.
- **Numbers are float64 throughout.**
  - *Why:* it is the standard library's number type and maps directly onto JSON numbers.
  - *Trade-off:* binary rounding (`0.1 + 0.2` is `0.30000000000000004`) and about 15–17
    significant digits. An exact decimal type would need a dependency or much more code.
- **Strict input handling.** Field names must match exactly and appear once, and each operand
  is checked on its own, so every problem gets a precise message.
  - *Why:* the API never computes with a value the client didn't send. Go's decoder alone would
    read a `null` operand, or one too small for float64, as `0`, and would accept `OPERATION`
    or a repeated field.
  - *Trade-off:* a small first pass over the JSON tokens instead of a one-line decoder setting.
- **Error codes say what to fix.** 422 means "valid request, but no real, finite answer";
  400 codes name the part of the request that is wrong.
  - *Why:* a client can tell a bad request from impossible math, and the codes are stable even
    if messages change.
  - *Trade-off:* more codes to document and test.
- **Nothing non-finite is ever returned.** Specific checks give clear messages, and a final
  check on every result turns NaN into `DOMAIN_ERROR` and ±Infinity into `OVERFLOW`. `-0` is
  returned as `0`.
- **Layers that can be tested alone.**
  - *Why:* the `calculator` package is pure functions with no HTTP code. The HTTP layer maps
    its errors through one table. On the frontend, a pure reducer holds every state transition
    and a thin `useCalculator` hook only sends the request.
  - *Trade-off:* more, smaller files than a single handler or component.
- **A form, not a keypad.**
  - *Why:* a form maps one-to-one onto the API, makes per-field validation natural and is
    accessible by default. The operation buttons are native radio buttons, labels follow the
    operation, errors are linked to their fields, results are announced to screen readers, and
    focus moves to the first invalid field.
  - *Trade-off:* it looks less like a physical calculator, and calculations can't be chained.
- **Display formatting only changes how a number looks.** Whole numbers up to 2^53 keep every
  digit, because float64 holds them exactly. Other results are rounded to 15 significant digits,
  so `0.1 + 0.2` shows `0.3`, and from 10^15 up they use exponent form, so rounding never
  produces trailing zeros that look exact.
- **Few dependencies.**
  - *Why:* the Go code uses only the standard library. The frontend uses only React at runtime,
    with plain CSS.
  - *Trade-off:* the form's styling is hand-written rather than taken from a component library.
- **One small, hardened image.** A distroless image running as a non-root user, about 16 MB, in
  which Go serves the page and the API from one origin, so no CORS is needed. Building it runs
  every test.
  - *Trade-off:* the frontend and the backend scale together, and image builds are slower
    because they run the tests.
- **Safe local defaults.** The server listens on loopback only unless `HOST` is set, shuts down
  gracefully, and the frontend gives up on a request after 10 seconds.

### Assumptions

- **Percentage** means "a % of b", that is a × b / 100, so `15` and `200` give `30`. Multiplying
  first keeps common cases exact: 7% of 300 is `21`, not `21.000000000000004`.
- **Power:** `0^0` is `1`, and zero raised to a negative power counts as division by zero. A
  negative base needs a whole-number exponent. Every non-integer float64 is a fraction with an
  even denominator, so a negative number raised to it never has a real result.
- **Input format:** a dot for decimals, optional exponent notation such as `1e3`, and no
  thousands separators. `1,5`, `1,000` and `1 000` are rejected with a hint.
- **One operation per request,** with no expression parsing, history or memory.
- **No authentication or rate limiting:** both are out of scope for this exercise.

### Known limitations

- Float64 limits apply. Raw API results can show binary rounding, as in `0.30000000000000004`.
  For `percentage`, a × b can overflow even when a × b / 100 would fit.
- Real roots of negative numbers through fractional exponents, such as the cube root of `-8`,
  aren't supported, for the reason given under Assumptions.

### What I'd do with more time

- Exact decimal arithmetic, so that `0.1 + 0.2` is exactly `0.3`, and an nth-root operation that
  gives real roots of negative numbers.
- Browser end-to-end tests (for example Playwright) against the Docker image in CI, with an
  automated accessibility check.
- An OpenAPI description of the API, with the frontend's types generated from it, so the
  contract can't drift between the two sides.
- Calculation history, chaining (using a result as the next first number), and keyboard
  shortcuts for the operations.
- Production hardening: request IDs and metrics, rate limiting, security and cache headers for
  the static files, and a Docker health check.
- The small code-review items I chose to leave as they are:
  - Go: request parsing returns a custom `*apiError` where `error` is the usual idiom.
  - React: element IDs are hard-coded rather than generated with `useId`.
  - The request is sent from an effect keyed on the status object, and a cancelled request is
    reported as a network error.
  - A handler that panicked after it started writing a response would get the JSON error
    appended to that response. No current handler can do this.

## Project structure

```
backend/
  cmd/server/            entry point: configuration, static files, graceful shutdown
  internal/calculator/   the math and its errors, with no HTTP code
  internal/api/          routes, strict request decoding, error-to-status mapping, middleware
frontend/src/
  api/                   typed API client, contract types, user-facing error messages
  calculator/            operations, input parsing, display formatting, reducer, useCalculator
  components/            Calculator, OperationPicker, NumberField, ResultPanel
scripts/smoke-test.sh    checks every operation and error case against a running server
docs/                    DESIGN.md, SMOKE_TEST.md, COVERAGE.md
Dockerfile, docker-compose.yml, Makefile, PROMPTS.md
```

## AI usage

This project was built with Claude Code. Every prompt is listed in [PROMPTS.md](PROMPTS.md).
