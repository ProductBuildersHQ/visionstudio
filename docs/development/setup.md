# Development Setup

## Prerequisites

- Go 1.23+
- Node.js 20+
- npm

## Clone and Setup

```bash
git clone https://github.com/ProductBuildersHQ/visionstudio.git
cd visionstudio
```

## Build Go Daemon

```bash
go mod tidy
go build -o bin/daemon ./cmd/daemon/
```

## Install Frontend Dependencies

```bash
cd desktop
npm install
```

## Development Workflow

### Terminal 1: Go Daemon

```bash
./bin/daemon
```

### Terminal 2: Vite Dev Server

```bash
cd desktop
npm run dev:renderer
```

### Terminal 3: Electron

```bash
cd desktop
npm run dev:main
```

Or use the combined command:

```bash
cd desktop
npm run dev
```

## Hot Reload

- **Frontend**: Vite provides instant hot reload
- **Go Daemon**: Restart manually after changes

## Building for Production

```bash
# Build Go daemon
go build -o bin/daemon ./cmd/daemon/

# Build frontend
cd desktop
npm run build

# Package Electron app (TODO)
```

## Running Tests

```bash
# Go tests
go test ./...

# Frontend tests (TODO)
cd desktop && npm test
```
