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

The cloud tenant API's first slice covers initiatives and RMIs. Store
operations that have no cloud endpoint return a
`... is not supported in remote mode` error. They never return empty or
invented data.

| Area | Remote support |
|------|----------------|
| Initiatives | create, get, list |
| RMIs | create, get, list (by initiative, all; by repository and status filtered client-side) |
| Phases | list |
| Programs | list, get |
| Everything else (updates, transitions, claims, evidence, dependencies, repositories, releases, specs, workflows, ...) | not supported yet |

Create calls are also rejected, before any request is sent, if they set
fields that the cloud create endpoint cannot accept. Examples are an
initiative's home repo or program, and an RMI's `origin` or `context_spec`.
Without this check those fields would be silently dropped.

MCP tools that work end to end: `initiative_list`, `initiative_create`
(without `program_id`), `rmi_create` (without `origin`/`context_spec`),
`program_list`, and `workflow_list`, which reads the embedded catalog. The
remaining tools fail with a clear not-supported error. These include
`initiative_get` (needs releases by initiative), `work_ready` (needs
dependencies and assignments), and `task_claim`/`task_update`/`task_release`
(need assignment, evidence and RMI-update endpoints). On the CLI, `rmi create`
currently fails because it validates `--repo` against the repository registry.

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

Errors are typed. `*remote.APIError` carries the HTTP status and the server's
message, and `errors.Is` matches `remote.ErrUnauthorized` (401),
`remote.ErrForbidden` (403, not a member of the tenant), `remote.ErrNotFound`
(404), `remote.ErrConflict` (409) and `remote.ErrNotSupported`.

For tests, `pkg/remote/remotetest` provides an in-process fake of the tenant
API. It serves requests through the real service layer over an in-memory store.
