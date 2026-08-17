# Troubleshooting

Common failure modes when running VisionStudio locally, and how to fix them. If you hit something not covered here, check `visionstudio app status` first — it reports whether the database and web UI are ready.

## `Error 1105 (HY000): table "X" does not have column "Y"`

**Symptom:** an API call (or a CLI command that hits the database) fails with a raw SQL error naming a missing column — for example:

```
list judge results for INIT-ACTS-001: Error 1105 (HY000): table "judge_results" does not have column "rubric_id"
```

**Cause:** the local Dolt database is behind the Ent schema baked into the binary you're running. This happens whenever a commit that changes `ent/schema/*.go` lands — via `git pull`, a rebase, or another session's work — and nobody has run a migration against *this particular* database since. The code expects a column (or table) that the database doesn't have yet.

**Fix:**

```bash
visionstudio db init --migrate
```

This is Ent's schema auto-migration, and it's **additive-only** — it adds missing columns and tables but never drops existing ones, even ones no longer referenced by current code. That makes it always safe to run, including as a first troubleshooting step any time an API call throws a raw SQL error you don't immediately recognize. Restart the UI+API process afterward (`visionstudio app restart`) so it picks up a fresh connection.

## Dashboard shows old behavior, or a fix doesn't seem to take effect

**Symptom:** you rebuilt VisionStudio (new code, `go install`, or a frontend rebuild), but the running dashboard still behaves like the old version — a bug you just fixed is still there, or a new feature is missing.

**Cause:** the UI+API process currently listening on your port is still running the *previous* binary. `go build`/`go install` produces a new executable, but any process already running keeps executing the code it loaded at startup — it doesn't hot-reload.

**Fix:**

```bash
visionstudio app restart
```

This finds whatever UI+API server this machine has recorded as running (tracked via `ui.pid`, written on every `ui`/`app start`/`app restart`/`dashboard --unified` invocation) — even one started from a different terminal or a detached/background process — stops it, and starts a fresh one from the current binary. The database is left alone if it's already running.

If `app restart` reports nothing was running to stop, but you can still reach the dashboard in a browser, the process predates `ui.pid` tracking (i.e., it was started before this feature existed) and needs a manual kill once: find it with `lsof -i :<port>` and stop it directly. Any process started after that will be trackable going forward.

Don't forget the rebuild step itself — `app restart` replaces the *running process*, but if you only ran `go run` (which compiles fresh each invocation) rather than `go build`/`go install`, there's no separate installed binary to be stale in the first place. If you *did* change frontend code, remember `web/dist` needs `npm run build` before a Go rebuild will embed it (see "Rebuilding the embedded web UI" in `CLAUDE.md`).

## `cannot reach the VisionStudio database at 127.0.0.1:13306`

**Symptom:** any command fails with a connection-refused error pointing at the Dolt port.

**Cause:** the Dolt SQL server itself isn't running. Most commonly this happens because an earlier `visionstudio app start` was the one that started it — and `app start` stops the database on exit if it's the one that started it, so ending that session (Ctrl-C, terminal closed, process killed) took the database down with it.

**Fix:**

```bash
visionstudio db start      # start Dolt alone, in the background
visionstudio db status     # verify it's reachable
```

Or bring both up together: `visionstudio app start` (foreground, stops the database on exit if it started it) or `visionstudio app restart` (same, but also replaces a stale UI+API process — see above).
