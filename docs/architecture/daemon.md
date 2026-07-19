# Go Daemon

The Go daemon provides the backend API for VisionStudio.

## Location

```
cmd/daemon/
├── main.go           # Entry point and core routes
├── aidlc.go          # AIDLC workflow handlers
├── v2mom.go          # V2MOM cascade handlers
├── capability.go     # Capability stack handlers
├── roadmap.go        # Roadmap handlers
├── organization.go   # Organization handlers
├── methodologies.go  # Methodology selection handlers
└── samples.go        # Sample projects handlers

pkg/
├── api/types.go      # API request/response types
└── config/
    ├── projects.go   # Project configuration
    └── organization.go # Organization configuration
```

## API Endpoints

### Core

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/profiles` | List available profiles |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects` | List all projects |
| POST | `/api/projects` | Create project |
| GET | `/api/projects/{project}` | Get project details |
| DELETE | `/api/projects/{project}` | Delete project |

### Specs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{project}/specs/{type}` | Get spec content |
| PUT | `/api/projects/{project}/specs/{type}` | Save spec content |
| POST | `/api/projects/{project}/specs/{type}/evaluate` | Evaluate spec |

### Methodology

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/methodologies/requirements` | List requirements methodologies |
| GET | `/api/methodologies/implementation` | List implementation methodologies |
| GET | `/api/projects/{project}/methodology` | Get project methodology |
| PUT | `/api/projects/{project}/methodology` | Update methodology selection |

### AIDLC Workflow

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{project}/aidlc/workflow` | Get workflow status |
| GET | `/api/projects/{project}/aidlc/phases` | List phases and deliverables |
| POST | `/api/projects/{project}/aidlc/transition` | Transition to next phase |
| GET | `/api/projects/{project}/aidlc/sync` | Get sync status |
| GET | `/api/projects/{project}/aidlc/documents/{docType}` | Get AIDLC document |
| POST | `/api/projects/{project}/aidlc/documents/{docType}` | Create/generate document |
| PUT | `/api/projects/{project}/aidlc/documents/{docType}` | Update document |

### V2MOM

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{project}/v2moms` | List project V2MOMs |
| GET | `/api/projects/{project}/v2moms/{id}` | Get specific V2MOM |
| PUT | `/api/projects/{project}/v2moms/{id}` | Update V2MOM |
| GET | `/api/v2mom/cascade` | Get full V2MOM cascade |

### Capability Stack

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{project}/capabilities` | List capabilities |
| GET | `/api/projects/{project}/capabilities/{id}` | Get capability |
| POST | `/api/projects/{project}/capabilities` | Create capability |
| PUT | `/api/projects/{project}/capabilities/{id}` | Update capability |
| DELETE | `/api/projects/{project}/capabilities/{id}` | Delete capability |

### Roadmap

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{project}/roadmap` | Get roadmap |
| POST | `/api/projects/{project}/roadmap/initiatives` | Create initiative |
| PUT | `/api/projects/{project}/roadmap/initiatives/{id}` | Update initiative |
| DELETE | `/api/projects/{project}/roadmap/initiatives/{id}` | Delete initiative |

### Maturity Model

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{project}/maturity` | List maturity models |
| GET | `/api/projects/{project}/maturity/{id}` | Get maturity model |
| GET | `/api/projects/{project}/maturity/{id}/dashboard` | Get HTML dashboard |

### Organization

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/organization` | Get organization details |
| PUT | `/api/organization` | Update organization |
| GET | `/api/organization/v2mom` | Get organization V2MOM |
| PUT | `/api/organization/v2mom` | Update organization V2MOM |
| GET | `/api/organization/teams` | List teams |
| POST | `/api/organization/teams` | Create team |

### Samples

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/samples` | List available samples |
| GET | `/api/samples/{id}` | Get sample details |
| POST | `/api/samples/{id}/import` | Import sample to workspace |

### Chat

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/chat` | LLM chat |

## Configuration

The daemon runs on `127.0.0.1:8765` by default.

```bash
./bin/daemon --port 8765
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `chi` | HTTP router |
| `cors` | CORS middleware |
| `visionspec` | Spec orchestration and AIDLC |
| `structured-evaluation` | LLM-as-Judge evaluation |

## Handler Structure

Each handler file follows a consistent pattern:

```go
// Handler methods
func (s *Server) handleListX(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetX(w http.ResponseWriter, r *http.Request)
func (s *Server) handleCreateX(w http.ResponseWriter, r *http.Request)
func (s *Server) handleUpdateX(w http.ResponseWriter, r *http.Request)
func (s *Server) handleDeleteX(w http.ResponseWriter, r *http.Request)

// Helper methods
func (s *Server) discoverX(projectPath string) []api.XSummary
func (s *Server) loadX(projectPath, id string) (api.X, error)
```

## Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "Error message here"
}
```

HTTP status codes:

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad request (validation error) |
| 404 | Not found |
| 409 | Conflict (already exists) |
| 500 | Internal server error |

## Extending

To add a new endpoint:

1. Create handler file in `cmd/daemon/` (e.g., `myfeature.go`)
2. Add handler methods to `Server` struct
3. Add API types in `pkg/api/types.go`
4. Register routes in `main.go` Router() method
5. Update this documentation
