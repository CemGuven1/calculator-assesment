# Calculator

[![CI](https://github.com/CemGuven1/calculator-assesment/actions/workflows/ci.yml/badge.svg)](https://github.com/CemGuven1/calculator-assesment/actions/workflows/ci.yml)

A full-stack calculator. A Go REST API does all the math, and a React + TypeScript frontend
sends it the numbers and shows the results. It supports addition, subtraction,
multiplication, division, exponentiation, square root and percentage.

- **Backend:** Go 1.25, standard library only.
- **Frontend:** React 19, TypeScript and Vite. Its only runtime dependencies are `react` and
  `react-dom`, and it uses plain CSS.
- **Tests:** Go `testing` with `httptest`; Vitest with React Testing Library. Coverage is
  96.8% for the backend and 99.3% for the frontend (see [docs/COVERAGE.md](docs/COVERAGE.md)).
- **Deployment:** one 16 MB Docker image in which Go serves both the API and the frontend.

## Quick start with Docker

You only need Docker.

```bash
docker compose up --build -d
```

Open <http://localhost:8080>. Building the image runs both test suites, so a failing test
stops the build. To stop the app:

```bash
docker compose down
```

[docs/SMOKE_TEST.md](docs/SMOKE_TEST.md) lists `curl` commands and manual UI steps to check
everything end to end.

## Running locally

**Prerequisites:** Go 1.25 or newer, and Node.js 22 or newer with npm. GNU `make` is optional;
every target is a single command you can also run by hand.

1. Install the frontend dependencies:

   ```bash
   cd frontend
   npm ci
   ```

2. Start the backend in one terminal. It listens on <http://127.0.0.1:8080>.

   ```bash
   cd backend
   go run ./cmd/server
   ```

3. Start the frontend in another terminal, then open <http://localhost:5173>. Vite forwards
   requests for `/api` to the backend.

   ```bash
   cd frontend
   npm run dev
   ```

With `make`, from the repository root:

| Command | What it does |
|---|---|
| `make install` | Installs the frontend dependencies |
| `make run` | Starts the backend and the frontend together |
| `make test` | Runs `go vet` and all Go tests, then all frontend tests |
| `make coverage` | Writes `backend/coverage.html` and `frontend/coverage/index.html` |
| `make docker-up` / `make docker-down` | Builds and starts, or stops, the Docker container |

### Configuration

The backend reads environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `HOST` | `127.0.0.1` | Interface to listen on. The Docker image sets `0.0.0.0`. |
| `PORT` | `8080` | Port to listen on. |
| `STATIC_DIR` | not set | Directory of the built frontend to serve at `/`. The Docker image sets it. |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated browser origins allowed to call the API directly. An empty value turns CORS off, as in the Docker image. |

The frontend reads `VITE_API_BASE_URL`. When it's empty (the default), requests go to the same
origin: through the Vite proxy in development, and to the Go server in Docker. Set it, for
example to `http://localhost:8080`, to call the backend directly. See
[frontend/.env.example](frontend/.env.example).

## API

Requests and responses are JSON. On Windows, run the `curl` examples in Git Bash.

### `POST /api/v1/calculate`

```json
{ "operation": "divide", "operands": [10, 4] }
```

| `operation` | Operands | Result |
|---|---|---|
| `add` | 2 | a + b |
| `subtract` | 2 | a − b |
| `multiply` | 2 | a × b |
| `divide` | 2 | a ÷ b |
| `power` | 2 | a raised to the power b |
| `sqrt` | 1 | √a |
| `percentage` | 2 | a % of b, that is a × b / 100 |

A success returns `200` and the result:

```json
{ "result": 2.5 }
```

Every API error has the same shape. `code` is stable for programs; `message` is for
people.

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "division by zero" } }
```

| Status | `code` | When |
|---|---|---|
| 400 | `INVALID_REQUEST` | Broken JSON, a wrong type, an unknown field, extra data after the object, or a number beyond the float64 range |
| 400 | `UNKNOWN_OPERATION` | `operation` is missing or not supported |
| 400 | `INVALID_OPERANDS` | Wrong number of operands, or a `null` operand |
| 422 | `DIVISION_BY_ZERO` | Dividing by zero, or zero raised to a negative power |
| 422 | `DOMAIN_ERROR` | No real result: square root of a negative number, or a negative base with a fractional exponent |
| 422 | `OVERFLOW` | The result is too large for a float64 |
| 413 | `PAYLOAD_TOO_LARGE` | The body is larger than 1 KiB |
| 415 | `UNSUPPORTED_MEDIA_TYPE` | `Content-Type` is not `application/json` |
| 405 | `METHOD_NOT_ALLOWED` | Wrong HTTP method (the `Allow` header lists the right one) |
| 404 | `NOT_FOUND` | Unknown path under `/api` |
| 500 | `INTERNAL` | Unexpected server error; the details go only to the server log |

400 means the request itself is wrong; 422 means it is valid but the math has no real,
finite answer.

**Examples**

```bash
curl -s -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":[2,3]}' http://localhost:8080/api/v1/calculate
# {"result":5}

curl -s -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","operands":[16]}' http://localhost:8080/api/v1/calculate
# {"result":4}

curl -s -H 'Content-Type: application/json' \
  -d '{"operation":"percentage","operands":[15,200]}' http://localhost:8080/api/v1/calculate
# {"result":30}

curl -s -w ' %{http_code}\n' -H 'Content-Type: application/json' \
  -d '{"operation":"divide","operands":[1,0]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"DIVISION_BY_ZERO","message":"division by zero"}} 422

curl -s -w ' %{http_code}\n' -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","operands":[4,9]}' http://localhost:8080/api/v1/calculate
# {"error":{"code":"INVALID_OPERANDS","message":"invalid operand: sqrt expects 1 operand, got 2"}} 400
```

### `GET /health`

Returns `{"status":"ok"}`. It sits outside the versioned API so that health checks don't
depend on the API version.

## Tests and coverage

```bash
make test        # or: cd backend && go vet ./... && go test ./...
                 #     cd frontend && npm test
make coverage    # HTML reports for both sides
```

[GitHub Actions](.github/workflows/ci.yml) runs three jobs on every push and pull request:

- **Backend:** formatting check, `go vet`, and the tests with `-race` and coverage.
- **Frontend:** lint, the tests with coverage, and the type-checked build.
- **Docker:** builds the image and smoke-tests the running container.

Both coverage reports are attached to each run as artifacts. Current numbers are in
[docs/COVERAGE.md](docs/COVERAGE.md).

## Project structure

```
backend/
  cmd/server/            entry point: configuration, wiring, graceful shutdown
  internal/calculator/   the math and its errors, with no HTTP code
  internal/api/          routes, strict JSON decoding, error-to-status mapping, middleware
frontend/src/
  api/                   typed API client, contract types, user-facing error messages
  calculator/            operations, input parsing, state reducer, useCalculator hook
  components/            Calculator, OperationPicker, NumberField, ResultPanel
docs/                    DESIGN.md, SMOKE_TEST.md, COVERAGE.md
Dockerfile, docker-compose.yml, Makefile, PROMPTS.md
```

## Design decisions

The full design, including every edge case, is in [docs/DESIGN.md](docs/DESIGN.md). The main
decisions:

- **One endpoint, with the operation in the body.** An operation is data, not a resource.
  One handler and one table of operations mean that adding an operation is a single entry.
  Operands are always an array, so `sqrt` simply takes one.
- **All math happens in Go.** The frontend only checks that each input is a number and turns
  the backend's error codes into friendly messages, so no rule exists in two places.
- **Layers that can be tested alone.** The `calculator` package is pure functions with no
  HTTP code, so it is tested on its own. The HTTP layer maps its errors to status codes through
  one table. On the frontend, a pure reducer holds all state transitions, and a thin
  `useCalculator` hook only sends the request.
- **Strict input handling.** The backend rejects unknown fields, data after the JSON object,
  `null` operands, non-JSON `Content-Type` headers and bodies over 1 KiB. Go would otherwise
  silently read a `null` operand as `0`.
- **Nothing non-finite is ever returned.** Specific checks give clear messages, and a final
  check on every result turns NaN into `DOMAIN_ERROR` and ±Infinity into `OVERFLOW`. `-0` is
  returned as `0`.
- **Few dependencies.** The Go code uses only the standard library. The frontend uses only
  React at runtime, with plain CSS and no UI library.
- **Accessibility.** Native radio buttons for the operations, labels that follow the
  operation, error messages linked to their fields, a live region that announces results, and
  focus that moves to the first invalid field. The UI was checked at 360px in light and dark
  mode.
- **One small, hardened image.** A distroless image running as a non-root user, about 16 MB.
  Go serves the page and the API from one origin, so no CORS is needed in production. Run
  locally, the server listens on loopback only, and it shuts down gracefully on SIGINT or
  SIGTERM.

## Assumptions

- **Percentage** means "a % of b", that is a × b / 100, so `15` and `200` give `30`. The
  multiplication comes first, which keeps common cases exact: 7% of 300 is `21` rather than
  `21.000000000000004`.
- **Numbers are float64** (IEEE 754 doubles), with about 15–17 significant digits. The API
  returns the exact float64 result; the UI rounds it to 15 significant digits for display
  only, so `0.1 + 0.2` shows as `0.3`.
- **Power:** `0^0` is `1`, and zero raised to a negative power counts as division by zero. A
  negative base needs a whole-number exponent. Every non-integer float64 is a fraction with an
  even denominator, so a negative number raised to it never has a real result.
- **Input format:** a dot for decimals, with optional exponent notation such as `1e3`. A
  decimal comma such as `1,5` is rejected with a hint.
- **One operation per request.** There is no expression parsing, no history and no memory.

## Known limitations

- The frontend's request has no timeout of its own. The Go server always answers within 10
  seconds, but behind a proxy that stalls, the page would show "Calculating…" until the
  browser gives up.
- Float64 limits apply. Raw results can show binary rounding, as in `0.30000000000000004`.
  Results too small to represent become `0`. For `percentage`, a × b can overflow even when
  a × b / 100 would fit.
- Real roots of negative numbers through fractional exponents, such as the cube root of `-8`,
  aren't supported, for the reason given under Assumptions.
- There's no authentication or rate limiting; both are out of scope for this exercise.

## AI usage

This project was built with Claude Code. Every prompt is listed in [PROMPTS.md](PROMPTS.md).
