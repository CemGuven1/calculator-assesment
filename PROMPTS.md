# Prompts

The prompts used to build this project, in order, quoted as sent. The tool was
Claude Code (Claude Opus 5.5) in the Claude desktop app. The assistant's replies
are not included; the commit history shows what each step produced.

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
