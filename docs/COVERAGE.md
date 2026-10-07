# Test coverage

Recorded on 2026-10-08. CI regenerates both reports on every push: open a run under the
repository's **Actions** tab and download the `backend-coverage` and `frontend-coverage`
artifacts (each includes an HTML report).

## Summary

| Layer | Tests | Coverage |
|---|---|---|
| Backend (Go) | 245 tests and subtests | 96.8% of statements; 100% in `calculator` and `api` |
| Frontend (React + TypeScript) | 132 tests | 99.3% of statements, 98.9% of branches, 100% of functions |

The Go tests also run with the race detector (`go test -race`) in the Docker build and in CI.

## Regenerate the reports

```bash
make coverage
```

Without `make`:

```bash
cd backend
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out                    # summary per function
go tool cover -html=coverage.out -o coverage.html   # HTML report

cd ../frontend
npm run coverage                                    # HTML report in coverage/index.html
```

## Backend by package

| Package | Statements | Notes |
|---|---|---|
| `internal/calculator` | 100% | |
| `internal/api` | 100% | |
| `cmd/server` | 85.1% | Only `main()` is uncovered. It connects OS signals to `run()`, which is tested, including a graceful shutdown with a request still in flight. |

## Frontend by file

Type-only files (`api/types.ts`, `env.d.ts`) have no statements and are left out.

| File | Statements | Branches | Functions | Lines |
|---|---|---|---|---|
| `api/client.ts` | 100% | 100% | 100% | 100% |
| `api/errorMessages.ts` | 100% | 100% | 100% | 100% |
| `calculator/format.ts` | 100% | 100% | 100% | 100% |
| `calculator/operations.ts` | 100% | 100% | 100% | 100% |
| `calculator/parseOperand.ts` | 100% | 100% | 100% | 100% |
| `calculator/reducer.ts` | 100% | 100% | 100% | 100% |
| `calculator/useCalculator.ts` | 94.7% | 83.3% | 100% | 94.4% |
| `components/Calculator.tsx` | 100% | 100% | 100% | 100% |
| `components/NumberField.tsx` | 100% | 100% | 100% | 100% |
| `components/OperationPicker.tsx` | 100% | 100% | 100% | 100% |
| `components/ResultPanel.tsx` | 100% | 100% | 100% | 100% |
| `App.tsx` | 100% | 100% | 100% | 100% |

The one uncovered line in `useCalculator.ts` ignores a response that arrives after its request
was cancelled. Today that only happens after the component unmounts, when React drops state
updates anyway, so no test can observe the difference. The check is kept as a safeguard.

## What the tests cover

**Backend**

- `calculator`: table-driven tests for each operation (normal cases, negatives, decimals);
  every error kind; operand counts; NaN and ±Inf in every operand position; `-0`
  normalization; and the exact wording of every error message.
- `api`: every edge case from [DESIGN.md](DESIGN.md) through `httptest`, checking the status,
  headers and exact JSON body. Also covered:
  - `Content-Type` handling
  - the size limit at exactly 1024 and 1025 bytes
  - 404 and 405 routing
  - static files
  - CORS
  - request logging
  - panic recovery
- `cmd/server`: configuration parsing, startup errors, and graceful shutdown with a request in
  flight.

**Frontend**

- Input parsing, result formatting and error messages: table tests.
- API client, with a mocked `fetch`: the request it sends, the base URL setting, and every
  kind of response, including network failures and responses that aren't JSON.
- Reducer: every state transition.
- `useCalculator`: the request lifecycle, and cancelling the request on unmount.
- `Calculator` component (React Testing Library):
  - a successful calculation
  - validation errors
  - API and network errors shown to the user
  - unary operations
  - the loading state
  - keyboard-only use
