# Remote Mode: CLI and MCP against VisionStudio Cloud

By default every `visionstudio` command reads and writes the local Dolt
database. **Remote mode** points the same commands, and the same MCP server,
at a VisionStudio Cloud tenant instead. It is not a second implementation. The
CLI and `pkg/mcpserver` talk to `pkg/service`, which talks to the `store.Store`
interface. Remote mode swaps the local Dolt store for `pkg/remote`, an HTTP
implementation of that interface that calls the cloud tenant API at
`/t/<tenant>/api/v1/...`.

Remote mode and local mode are separate systems. Nothing syncs between them,
and remote mode is opt-in per invocation.

## Quick start

```bash
# 1. Store an API key (create one in VisionStudio Cloud). The key is
#    verified against the tenant before it is saved.
visionstudio cloud login --remote https://cloud.example.com --remote-tenant acme

# 2. Check what remote mode will use.
visionstudio cloud status --remote https://cloud.example.com

# 3. Run commands against the tenant.
visionstudio initiative list --remote https://cloud.example.com
visionstudio initiative create --remote https://cloud.example.com \
    --id INIT-ACME-001 --title "Launch" --workflow pbhq-lite
```

## Settings and precedence

| Setting | Resolved from (first match wins) |
|---------|----------------------------------|
| Cloud URL (turns remote mode on) | `--remote <url>` → `$VISIONSTUDIO_REMOTE_URL` |
| Tenant | `--remote-tenant` → `$VISIONSTUDIO_REMOTE_TENANT` → URL ending in `/t/<tenant>` → tenant saved by `cloud login` |
| Credential | `$VISIONSTUDIO_REMOTE_TOKEN` → key saved by `cloud login` for that URL |

Notes:

- `--remote https://cloud.example.com/t/acme` is equivalent to
  `--remote https://cloud.example.com --remote-tenant acme`.
- The tenant flag is `--remote-tenant`, not `--tenant`, because `sync` and
  `pull` already use a local `--tenant` flag with a different meaning.
- `--remote` cannot be combined with `--dsn` or `--data-dir`.
- `cloud login` remembers the URL for `cloud status` and `cloud logout`. It
  never switches other commands to remote mode. Those still need `--remote` or
  `$VISIONSTUDIO_REMOTE_URL`.
- Remote mode applies to commands that open the store through the shared
  connection path, which covers the entity commands (`initiative`, `rmi`,
  `phase`, `program`, `work`, ...) and `visionstudio mcp`. Commands that manage
  the local database or UI directly (`db`, `app`, `dashboard`, `ui`, `sync`,
  `pull`) ignore `--remote`.

## Credentials

`visionstudio cloud login` reads a VisionStudio Cloud API key (`vsc_live_...`)
or a session token from one of these sources:

- standard input, with `--with-token` (for example `op read ... | visionstudio cloud login --with-token`)
- `$VISIONSTUDIO_REMOTE_TOKEN`
- a hidden interactive prompt

The key is stored per cloud URL in
`~/.productbuildershq/visionstudio/credentials.json`. The file is kept separate
from `config.json` and is always written with owner-only (`0600`) permissions.
`visionstudio cloud logout` deletes the local copy, and `--all` deletes every
stored key. Revoke a key in VisionStudio Cloud to invalidate it on the server
side. `cloud status` shows only a masked form of the key, never the full value.

A browser-based device login flow is planned. Until it ships, paste an API key.

## MCP with Claude Code

Run the MCP server locally in remote mode, so agent sessions read and write
the cloud tenant:

```json
{
  "mcpServers": {
    "visionstudio": {
      "command": "visionstudio",
      "args": ["mcp", "--remote", "https://cloud.example.com/t/acme"]
    }
  }
}
```

Run `visionstudio cloud login` first, or pass the key through the server's
environment:

```json
{
  "mcpServers": {
    "visionstudio": {
      "command": "visionstudio",
      "args": ["mcp"],
      "env": {
        "VISIONSTUDIO_REMOTE_URL": "https://cloud.example.com/t/acme",
        "VISIONSTUDIO_REMOTE_TOKEN": "${VISIONSTUDIO_REMOTE_TOKEN}"
      }
    }
  }
}
```

Or register it from the command line:

```bash
claude mcp add visionstudio -- visionstudio mcp --remote https://cloud.example.com/t/acme
```

## What works today

The cloud tenant API serves the full agent-work surface. Store operations
that have no cloud endpoint return a `... is not supported in remote mode`
error that names the area. They never return empty or invented data.

| Area | Remote support |
|------|----------------|
| Programs | create, get, list, update |
| Initiatives | create, get, list, update (including lifecycle transitions) |
| Phases | create, list, delete (refused while the phase has RMIs) |
| RMIs | create, get, list (by initiative, repository, status, or all), update (status changes, moves, origin, context spec) |
| RMI dependencies | create, list (per RMI, all) |
| Assignments (work leases) | claim, get, get active, list (active, all), renew, hand off, release, complete |
| Delivery evidence | create, list (by RMI, by initiative, all) |
| Releases | get, list (by repository, by initiative, all) — read-only |
| Repositories | get, list (by organization, all) — read-only |
| Spec workflows | get, list, select for an initiative, get an initiative's selection |
| Spec documents | create, get, list (by initiative, by repository, all), update, delete |
| Judge results | create, list by initiative |

MCP tools that work end to end include `initiative_list`, `initiative_get`,
`initiative_create`, `rmi_create`, `program_list`, `program_create`,
`work_ready`, `task_claim`, `task_update` (status, evidence, handoff,
completion), `task_release`, `workflow_list` and `workflow_select`. On the
CLI, `initiative`, `rmi` (including `update`, `move` and `dep`), `phase`,
`program` and `work` commands run against the tenant; `rmi create` validates
`--repo` against the tenant's repository registry.

The server enforces the lease rules a shared database needs. A claim while
another session holds a live lease on the RMI, or an update to an assignment
that is no longer active (released, completed, expired, or taken over), fails
with `409 LEASE_CONFLICT`, which `errors.Is` matches as both
`remote.ErrLeaseConflict` and `remote.ErrConflict`.

Creates that set fields the cloud create endpoint cannot carry (an
initiative's status, home repo, program, specs or lifecycle timestamps; an
RMI's status, `origin`, `context_spec` or `completed_at`) are sent as a create
followed by a full update. The two requests are not atomic: if the update
fails, the entity exists with only its create fields and the error says so.

### Remaining gaps

These areas have no cloud endpoint yet and fail with a not-supported error:

- initiative dependencies
- repository writes (register, update, archive, remove) and repository dependencies
- release writes (`release record` and friends)
- organizations and people
- spec workflow writes (workflows come from the embedded catalog)
- capability models and maturity assessments, DevX period reports, and PRISM
  roadmaps, goals and documents

## Embedding: custom HTTP clients and request signing

`pkg/remote` is a library. A program can embed the MCP server against a cloud
tenant with its own HTTP stack, for example an agent launcher that signs every
request:

```go
rs, err := remote.New("https://cloud.example.com", "acme",
    remote.WithTransport(signingRoundTripper), // or remote.WithHTTPClient(hc)
)
if err != nil {
    return err
}
return mcpserver.Run(ctx, service.New(rs))
```

Bearer auth is optional. `remote.WithToken` adds `Authorization: Bearer <key>`,
and when no token is set the store sends no `Authorization` header, so the
transport can authenticate however it needs to. The stock binary uses Go's
default proxy settings (`HTTPS_PROXY`), so a local signing proxy also works
without code changes.

Errors are typed. `*remote.APIError` carries the HTTP status, the server's
machine-readable `Code` (for example `INVALID`, `INVALID_REFERENCE`,
`NOT_FOUND`, `CONFLICT`, `LEASE_CONFLICT`) and its message. `errors.Is`
matches `remote.ErrUnauthorized` (401), `remote.ErrForbidden` (403, not a
member of the tenant), `remote.ErrInvalid` (400), `remote.ErrInvalidReference`
(400 `INVALID_REFERENCE`: the body names an entity that does not exist),
`remote.ErrNotFound` (404), `remote.ErrConflict` (409),
`remote.ErrLeaseConflict` (409 `LEASE_CONFLICT`) and `remote.ErrNotSupported`.

`Store.Ping` checks the URL, credential and tenant membership with
`GET /t/<tenant>/api/v1/me`. `Store.Me` returns the credential's principal and
every tenant membership (`GET /api/v1/me`), which `visionstudio cloud status`
prints after verifying access.

For tests, `pkg/remote/remotetest` provides an in-process fake of the tenant
API. It mirrors the cloud contract (strict request decoding, `[]` for empty
lists, the `{"assignment": null}` and `{"workflow": null}` wrappers, the error
codes, and both lease-conflict rules) and serves requests through the real
service layer over an in-memory store.
