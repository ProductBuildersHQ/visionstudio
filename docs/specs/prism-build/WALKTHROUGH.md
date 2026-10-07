# PRISM Control — End-to-End Agent Walkthrough

This walkthrough demonstrates a Claude Code session using the PRISM Control session protocol to plan, claim, execute, and complete work on a roadmap item.

## Prerequisites

1. A running Dolt SQL server with the prismcontrol database:

   ```bash
   prismctl db init
   prismctl db serve
   ```

2. An initiative, phases, and RMIs already created (via `prismctl` or MCP).

3. Either:
   - **CLI mode:** `prismctl` on your PATH
   - **MCP mode:** `.mcp.json` configured:

     ```json
     {
       "mcpServers": {
         "prism-control": {
           "command": "prismctl",
           "args": ["mcp", "--dsn", "root:@tcp(127.0.0.1:3306)/prismcontrol"]
         }
       }
     }
     ```

## Step 1: Plan — Survey Available Work

The agent starts by listing initiatives and finding ready work.

### Via CLI

```bash
# Browse initiatives
prismctl initiative list

# Get detail with phase progress
prismctl initiative get INIT-PRISMCONTROL-001

# Find claimable items
prismctl work ready --initiative INIT-PRISMCONTROL-001
```

### Via MCP

```
tool: initiative_list
args: {}

tool: initiative_get
args: {"id": "INIT-PRISMCONTROL-001"}

tool: work_ready
args: {"initiative_id": "INIT-PRISMCONTROL-001"}
```

**Expected output** — a list of RMIs with `status=ready`, all dependencies completed, and no active assignment. The agent picks one to work on.

## Step 2: Claim — Reserve the Work Item

The agent claims the chosen RMI, establishing a lease and receiving the git trailer.

### Via CLI

```bash
prismctl work claim RMI-PRISMCONTROL-013 \
  --worker "session-$(date +%s)" \
  --workspace "/Users/example/prism-control" \
  --lease-hours 4
```

Output:

```
Claimed RMI-PRISMCONTROL-013 (assignment: assign-RMI-PRISMCONTROL-013-1721700000000)
Worker:  session-1721700000
Lease:   2026-07-23 17:00

Git trailer for commits:
  Refs: RMI-PRISMCONTROL-013
```

### Via MCP

```
tool: task_claim
args: {
  "rmi_id": "RMI-PRISMCONTROL-013",
  "worker": "session-1721700000",
  "workspace": "/Users/example/prism-control",
  "lease_hours": 4
}
```

Response:

```json
{
  "assignment_id": "assign-RMI-PRISMCONTROL-013-1721700000000",
  "rmi_id": "RMI-PRISMCONTROL-013",
  "worker": "session-1721700000",
  "lease_expires": "2026-07-23T17:00:00Z",
  "trailer_line": "Refs: RMI-PRISMCONTROL-013"
}
```

**Key:** The `trailer_line` is the exact text to append to every commit made under this claim.

## Step 3: Execute — Write Code with Trailers

The agent works in the product repository. Every commit carries the `Refs:` trailer in the footer:

```bash
git commit -m "$(cat <<'EOF'
feat(mcpserver): add stdio MCP server with 9 core tools

Implement pkg/mcpserver with initiative_list, initiative_get,
initiative_create, rmi_create, work_ready, task_claim, task_release,
task_update, and report_initiative tools.

Refs: RMI-PRISMCONTROL-013
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

### Lease Management

If the work takes longer than the initial lease:

```bash
# CLI
prismctl work renew assign-RMI-PRISMCONTROL-013-1721700000000 --lease-hours 4

# MCP — no dedicated renew tool; use task_update or release+reclaim
```

### Midpoint Handoff Updates

If the session needs to pause or hand off:

```bash
# CLI
prismctl work update assign-RMI-PRISMCONTROL-013-1721700000000 \
  --handoff '{"completed":["server skeleton","tool registration"],"remaining":["tests"],"decisions":["used raw AddTool API"],"next_action":"write unit tests"}'

# MCP
tool: task_update
args: {
  "assignment_id": "assign-RMI-PRISMCONTROL-013-1721700000000",
  "handoff": {
    "completed": ["server skeleton", "tool registration"],
    "remaining": ["tests"],
    "decisions": ["used raw AddTool API"],
    "next_action": "write unit tests"
  }
}
```

## Step 4: Update — Record Evidence and Complete

After all commits are pushed and tests pass, the agent records evidence and marks the work done.

### Add Evidence

```bash
# CLI
prismctl work update assign-RMI-PRISMCONTROL-013-1721700000000 \
  --handoff '{"completed":["MCP server","CLI command","tests"],"remaining":[],"next_action":"none"}'
```

```
# MCP
tool: task_update
args: {
  "rmi_id": "RMI-PRISMCONTROL-013",
  "evidence": [
    {"type": "commit", "reference": "abc123def"},
    {"type": "commit", "reference": "def456abc"},
    {"type": "pr", "reference": "https://github.com/ProductBuildersHQ/prism-build/pull/7"}
  ]
}
```

### Complete the Assignment

```bash
# CLI
prismctl work complete assign-RMI-PRISMCONTROL-013-1721700000000
```

```
# MCP
tool: task_update
args: {
  "assignment_id": "assign-RMI-PRISMCONTROL-013-1721700000000",
  "complete": true,
  "handoff": {
    "completed": ["MCP server", "CLI command", "unit tests", "walkthrough"],
    "remaining": [],
    "next_action": "none"
  }
}
```

### Or Release for Another Session

If the work isn't finished, release instead of completing:

```bash
# CLI
prismctl work release assign-RMI-PRISMCONTROL-013-1721700000000 \
  --handoff '{"completed":["server"],"remaining":["tests","docs"],"next_action":"write unit tests using in-memory transport"}'
```

```
# MCP
tool: task_release
args: {
  "assignment_id": "assign-RMI-PRISMCONTROL-013-1721700000000",
  "handoff": {
    "completed": ["server"],
    "remaining": ["tests", "docs"],
    "next_action": "write unit tests using in-memory transport"
  }
}
```

The RMI returns to `ready` status and appears in `work_ready` results for the next session.

## Step 5: Report — Review Progress

At any point, generate an initiative report:

```bash
# CLI (future — report command is Phase 4)
# For now, use initiative get:
prismctl initiative get INIT-PRISMCONTROL-001
```

```
# MCP
tool: report_initiative
args: {"id": "INIT-PRISMCONTROL-001"}
```

Response includes per-phase progress and RMI status breakdown:

```json
{
  "initiative_id": "INIT-PRISMCONTROL-001",
  "title": "PRISM Control Platform",
  "status": "executing",
  "phases": [
    {"phase_id": "phase-1", "title": "Foundation", "status": "completed", "rmis_total": 5, "rmis_completed": 5},
    {"phase_id": "phase-2", "title": "Coordination Core", "status": "completed", "rmis_total": 6, "rmis_completed": 6},
    {"phase_id": "phase-3", "title": "Agent Interface", "status": "in_progress", "rmis_total": 3, "rmis_completed": 2}
  ],
  "summary": {
    "total_rmis": 14,
    "completed_rmis": 13,
    "in_progress": 1,
    "ready": 0,
    "blocked": 0
  }
}
```

## Protocol Summary

| Step | Action | CLI Command | MCP Tool |
|------|--------|-------------|----------|
| Plan | Survey work | `initiative list`, `work ready` | `initiative_list`, `work_ready` |
| Claim | Reserve RMI | `work claim` | `task_claim` |
| Execute | Commit with trailer | git commit with `Refs:` | — |
| Update | Evidence + handoff | `work update`, `work complete` | `task_update` |
| Release | Hand off to next session | `work release` | `task_release` |
| Report | Phase progress | `initiative get` | `report_initiative` |

## Product Repo Setup

For repos participating in PRISM-tracked work, add to their `CLAUDE.md`:

```markdown
## PRISM Control

This repo's roadmap items are tracked in [prism-control](https://github.com/ProductBuildersHQ/prism-build). Use `prismctl work ready --repo github.com/ProductBuildersHQ/<this-repo>` to find claimable work, and carry the `Refs: RMI-<REPOSLUG>-<NNN>` trailer on every commit.
```
