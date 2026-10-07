# Development tasks. Requires Go 1.25+, Node 22+ and GNU make.
# Each recipe is a plain `cd <dir> && <command>`, so it also documents the
# command to run by hand where make is unavailable.

.PHONY: install run run-backend run-frontend test test-backend test-frontend \
	coverage coverage-backend coverage-frontend

install:
	cd frontend && npm ci

# Backend on :8080 and the Vite dev server on :5173 (proxies /api to :8080).
run:
	$(MAKE) -j2 run-backend run-frontend

run-backend:
	cd backend && go run ./cmd/server

run-frontend:
	cd frontend && npm run dev

test: test-backend test-frontend

test-backend:
	cd backend && go vet ./... && go test ./...

test-frontend:
	cd frontend && npm test

# Writes backend/coverage.html and frontend/coverage/index.html.
coverage: coverage-backend coverage-frontend

coverage-backend:
	cd backend && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out && go tool cover -html=coverage.out -o coverage.html

coverage-frontend:
	cd frontend && npm run coverage
