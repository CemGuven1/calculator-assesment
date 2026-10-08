# Test coverage

Final numbers, recorded on 2026-10-09 after the review fixes. CI regenerates both reports on
every push: open a run under the repository's **Actions** tab and download the
`backend-coverage` and `frontend-coverage` artifacts. Each includes an HTML report.

## Summary

| Layer | Tests | Coverage |
|---|---|---|
| Backend (Go) | 284 tests and subtests | 97.1% of statements; 100% in `calculator` and `api` |
| Frontend (React + TypeScript) | 153 tests | 99.31% of statements, 99.04% of branches, 100% of functions |

The Go tests also run with the race detector (`go test -race`) in the Docker build and in CI.
On top of the unit tests, `scripts/smoke-test.sh` checks every operation and error case against
a running server, and CI runs it against the Docker image.

To check that the tests catch real bugs, I broke important checks one at a time and confirmed
that tests failed:

| Check broken on purpose | Failing tests |
|---|---|
| The division-by-zero check | 10, in both the math tests and the HTTP tests |
| The guard against infinite results | 19, covering every operation's overflow case |
| Rejecting a `null` operand (it would silently become 0) | 2 at the time, from a single test case; there are now 5 cases |
| The frontend validation check before a request is sent | 10, in the reducer, hook and form tests |

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
npm run coverage                                    # summary, plus an HTML report in coverage/index.html
```

## Backend by package

| Package | Statements | Notes |
|---|---|---|
| `internal/calculator` | 100% | |
| `internal/api` | 100% | |
| `cmd/server` | 86.4% | Only `main()` is uncovered. It connects OS signals to `run()`, which is tested, including a graceful shutdown with a request still in flight. |

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
| `calculator/useCalculator.ts` | 94.73% | 83.33% | 100% | 94.44% |
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
  every error kind; operand counts and `Arity`; NaN and ±Inf in every operand position; `-0`
  normalization; and the exact wording of every error message.
- `api`: every edge case from [DESIGN.md](DESIGN.md) through `httptest`, checking the status,
  headers and exact JSON body. That includes:
  - exact field names, with wrong-case and duplicate fields rejected
  - each operand problem, by position: `null`, a string, a boolean, an object, an array, too
    large, too small
  - the size limit at exactly 1024 and 1025 bytes
  - 404 and 405 routing, including unclean paths such as `/api//v1/calculate`
  - CORS, request logging and panic recovery
  - a logged 500 for an unmapped error
- `cmd/server`: configuration parsing; startup errors; the real static file server over a
  directory laid out like the built frontend, with no directory listings; and graceful shutdown
  with a request in flight.

**Frontend**

- Table tests for input parsing (including separators and numbers too large or too small),
  result formatting (including whole numbers beyond 15 digits) and error messages.
- API client, with a mocked `fetch`:
  - the request it sends and the base URL setting
  - cancelling by the caller, and the 10-second timeout
  - every kind of response, including network failures and responses that aren't JSON
- Reducer: every state transition.
- `useCalculator`: the request lifecycle, and cancelling the request on unmount.
- `Calculator` component (React Testing Library):
  - a successful calculation, including a 16-digit result
  - validation errors
  - API and network errors shown to the user
  - unary operations
  - the loading state
  - keyboard-only use
