# Installation

## Prerequisites

- Go 1.23 or later
- Node.js 20 or later
- npm

## Clone the Repository

```bash
git clone https://github.com/ProductBuildersHQ/visionstudio.git
cd visionstudio
```

## Build the Go Daemon

```bash
go build -o bin/daemon ./cmd/daemon/
```

## Install Frontend Dependencies

```bash
cd desktop
npm install
```

## Verify Installation

Start the daemon:

```bash
./bin/daemon
```

You should see:

```
level=INFO msg="Starting VisionStudio daemon" addr=127.0.0.1:8765
```

Start the frontend:

```bash
cd desktop
npm run dev
```

The Electron app should open with the VisionStudio interface.
