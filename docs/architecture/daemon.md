# Go Daemon

The Go daemon provides the backend API for VisionStudio.

## Location

```
cmd/daemon/main.go    # Entry point
pkg/api/types.go      # API types
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/projects` | List all projects |
| GET | `/api/projects/{project}` | Get project details |
| GET | `/api/projects/{project}/specs/{type}` | Get spec content |
| PUT | `/api/projects/{project}/specs/{type}` | Save spec content |
| POST | `/api/projects/{project}/specs/{type}/evaluate` | Evaluate spec |
| POST | `/api/chat` | LLM chat |

## Configuration

The daemon runs on `127.0.0.1:8765` by default.

```bash
./bin/daemon --port 8765
```

## Dependencies

- `chi` - HTTP router
- `cors` - CORS middleware
- `visionspec` - Spec orchestration (planned)
- `omniagent` - LLM integration (planned)

## Extending

To add a new endpoint:

1. Add handler method to `Server` struct
2. Register route in `Router()` method
3. Add corresponding API types in `pkg/api/types.go`
