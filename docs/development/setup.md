# Development Setup

## Prerequisites

- Go 1.26+
- Node.js 20+
- npm

## Clone and Setup

```bash
git clone https://github.com/ProductBuildersHQ/visionstudio.git
cd visionstudio
```

## Build CLI

```bash
go mod tidy
go build ./cmd/visionstudio
```

## Initialize Database

VisionStudio uses Dolt (MySQL-compatible) for persistent storage:

```bash
# Initialize and migrate database
go run ./cmd/visionstudio db init --migrate
```

## Install Frontend Dependencies

```bash
cd web
npm install
```

## Regenerate Types

After modifying Go API types in `pkg/apitypes/types.go`:

```bash
# Generate JSON schemas from Go types
go generate ./pkg/apitypes

# Generate Zod/TypeScript from JSON schemas
cd web && npm run generate:types
```

See [Type Pipeline](../architecture/types.md) for details.

## Development Workflow

### Unified Dashboard (Recommended)

Run the Go daemon with embedded frontend:

```bash
go run ./cmd/visionstudio dashboard --port 9401 --unified
```

Open http://127.0.0.1:9401 in your browser.

### Frontend Hot Reload

For frontend development with hot reload:

**Terminal 1: Go API Server**

```bash
go run ./cmd/visionstudio dashboard --port 9401
```

**Terminal 2: Vite Dev Server**

```bash
cd web
npm run dev
```

Open http://localhost:5173 (Vite proxies API calls to port 9401).

## Hot Reload

- **Frontend**: Vite provides instant hot reload
- **Go Daemon**: Restart manually after changes

## Building for Production

```bash
# Build CLI
go build -o bin/visionstudio ./cmd/visionstudio

# Build frontend (embedded in unified mode)
cd web
npm run build
```

## Running Tests

```bash
# Go tests
go test ./...

# Lint
golangci-lint run

# Frontend tests
cd web && npm test
```

## Common Tasks

| Task | Command |
|------|---------|
| Run dashboard | `go run ./cmd/visionstudio dashboard --port 9401 --unified` |
| Initialize DB | `go run ./cmd/visionstudio db init --migrate` |
| Regenerate Go→TS types | `go generate ./pkg/apitypes && cd web && npm run generate:types` |
| Lint | `golangci-lint run` |
| Test | `go test ./...` |
