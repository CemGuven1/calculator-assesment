# Prompts

The prompts used to build this project, in order, quoted as sent. The tool was
Claude Code (Claude Opus 5.5) in the Claude desktop app. The assistant's replies
are not included; the commit history shows what each step produced. Short
housekeeping messages, such as approvals to commit and push, are left out.

## 1. Kick-off

> I'm doing a 2–4 hour technical assessment: a full-stack calculator with a Go REST
> backend and a React + TypeScript frontend. Requirements are below.
> Priorities: correctness, clarity, maintainability over extra features.
> Go: standard library unless there's a strong reason. Frontend: minimal dependencies.
> Work in small steps. Before each phase, say what you'll do and wait for my approval.

Followed by the full assignment brief, pasted unchanged.

## 2. Design

> Before writing any code, propose:
> 1. Repo layout (backend/, frontend/, root-level docs and Docker files).
> 2. API contract: one POST /api/v1/calculate vs. one endpoint per operation.
>    Recommend one. Define the request, response and error JSON, including how
>    unary operations (sqrt) work.
> 3. Edge cases: division by zero, sqrt of a negative, missing or invalid operands,
>    unknown operation, NaN/Inf results, malformed JSON, oversized body.
>    Also define what "percentage" means.
> 4. Testing strategy for each layer, and the UI approach.
> Keep it to one page. I'll save it as docs/DESIGN.md.

## 3. Scaffolding

> Scaffold the repo from the approved design: a Go module with cmd/server,
> internal/calculator and internal/api; a Vite React-TS app with Vitest and React
> Testing Library; a root Makefile (run, test, coverage), .gitignore and .editorconfig.
> No feature code yet. Just confirm `go test ./...` and `npm test` both pass.

## 4. Backend: calculator package

> Implement internal/calculator as pure functions with no HTTP code: add, subtract,
> multiply, divide, power, sqrt, percentage. Return typed errors (ErrDivisionByZero,
> ErrInvalidOperand, ...) and reject NaN/Inf results. Write table-driven tests first.
> Cover normal cases, negatives, decimals and every edge case in DESIGN.md.

## 5. Backend: HTTP API

> Add internal/api using net/http (Go 1.22+ routing). Decode JSON strictly
> (DisallowUnknownFields, MaxBytesReader) and validate the input. Map the calculator
> errors to the right status codes, all with the same error body. Add GET /health,
> CORS for the dev origin, slog request logging, and graceful shutdown in main.
> Test the handlers with httptest, including every bad-input case.

## 6. Frontend: API client and state

> Create a typed API client in src/api/ that matches the backend contract. Read the
> base URL from an env variable and add a Vite dev proxy. Turn backend error codes
> into user-friendly messages. Keep calculator state (input, result, loading, error)
> in a pure reducer behind a useCalculator hook. Unit-test the reducer and the client
> with a mocked fetch.

## 7. Frontend: UI

> Build the calculator UI agreed in DESIGN.md. All math goes through the backend,
> none happens in the browser. Validate input on the client (empty, non-numeric,
> second decimal point). Show loading and error states, support the keyboard, use
> accessible labels, and make it responsive down to 360px. Plain CSS, no UI library.
> Use React Testing Library to test: a successful calculation, validation errors,
> an API error shown to the user, and unary operations.

## 8. Docker and smoke test

> Write a multi-stage Dockerfile that builds the frontend and the Go binary into one
> image. Go serves both the static files and /api. Add docker-compose.yml and
> .dockerignore. Then give me a smoke-test checklist (curl commands plus manual UI
> steps) to check everything end to end.

## 9. Documentation, coverage report and CI

> go ahead, include the GitHub Actions workflow too.

This approved the proposed final phase: a full README, a coverage report, a last pass over
this file, and a GitHub Actions workflow that runs both test suites on every push.
