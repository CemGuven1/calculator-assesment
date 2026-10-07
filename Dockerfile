# syntax=docker/dockerfile:1

# Builds the frontend and the Go server into one small image. The Go binary
# serves both the API (/api, /health) and the built frontend from the same
# origin, so the browser needs no CORS.
#
#   docker build -t calculator .
#   docker run --rm -p 8080:8080 calculator      # then open http://localhost:8080
#
# Both test suites run during the build: if a test fails, no image is produced.

# ---- Frontend: test, then build the static files ------------------------------
FROM node:22-alpine AS frontend
WORKDIR /src/frontend

# Install dependencies in their own layer, so code changes don't reinstall them.
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund

COPY frontend/ ./
RUN npm test && npm run build

# ---- Backend: vet, test with the race detector, build a static binary ---------
# The full (Debian) image includes the C toolchain that -race needs.
FROM golang:1.25 AS backend
WORKDIR /src/backend

COPY backend/ ./
RUN go vet ./... && go test -race ./...
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/server ./cmd/server

# ---- Runtime: only the binary and the static files ----------------------------
# Distroless has no shell or package manager, and :nonroot runs as an
# unprivileged user.
FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app

COPY --from=backend /out/server ./server
COPY --from=frontend /src/frontend/dist ./public

# Listen on all interfaces inside the container, and serve the frontend.
# CORS is off: the page and the API share one origin.
ENV HOST=0.0.0.0 \
    PORT=8080 \
    STATIC_DIR=/app/public \
    CORS_ORIGINS=""

EXPOSE 8080
USER nonroot
ENTRYPOINT ["/app/server"]
