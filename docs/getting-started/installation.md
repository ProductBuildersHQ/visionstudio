# Installation

## Prerequisites

- Go 1.26 or later
- Node.js 20 or later (only needed if you're building the web UI from source or doing frontend development — the released binary embeds a prebuilt copy)
- [Dolt](https://github.com/dolthub/dolt) — the MySQL-compatible, Git-like database VisionStudio stores initiatives, RMIs, and evaluations in

## Clone the Repository

```bash
git clone https://github.com/ProductBuildersHQ/visionstudio.git
cd visionstudio
```

## Sibling Repos (building from source)

VisionStudio's `go.work` workspace pulls a few packages from sibling repos checked out next to `visionstudio/`, rather than only from published module versions:

```
../../grokify/prism-maturity
../../grokify/prism-roadmap
../../plexusone/devfolio
../prism-build
```

If you don't have these checked out, `go build`/`go run` still works — Go falls back to the published module versions pinned in `go.mod` — but you won't pick up local changes to those repos. Contributors working across repos should clone them at the paths above.

## Build

```bash
go build -o bin/visionstudio ./cmd/visionstudio
```

## Initialize the Database

```bash
./bin/visionstudio db init --migrate
```

This bootstraps a Dolt database at `~/.productbuildershq/visionstudio` (override with `--data-dir`, or point at an existing MySQL-compatible server with `--dsn`). Re-running this command is always safe — it's additive-only, so it's also the fix if you ever see a `does not have column` error after pulling new code (see [Troubleshooting](troubleshooting.md)).

## Verify Installation

Start everything with one command — this starts the database (if it isn't already running) and serves the web UI + API, opening your browser:

```bash
./bin/visionstudio app start
```

You should see the VisionStudio dashboard open at `http://localhost:9400`. Press `Ctrl-C` to stop; if `app start` started the database itself, it stops that too on exit.

### Running the pieces separately

```bash
./bin/visionstudio db start              # start the database in the background
./bin/visionstudio ui --port 9400        # serve the UI + API, connecting to the running database
./bin/visionstudio db stop               # stop the background database when done
```

Use `./bin/visionstudio db status` to check whether the database is reachable, and `./bin/visionstudio --help` for the full command list.

## Building the Web UI from Source

The released binary embeds a prebuilt `web/dist`. If you're changing frontend code, rebuild it before building the Go binary so the new assets get embedded:

```bash
cd web
npm install
npm run build
cd ..
go build -o bin/visionstudio ./cmd/visionstudio
```

For frontend development with hot reload instead, see [Development Setup](../development/setup.md).
