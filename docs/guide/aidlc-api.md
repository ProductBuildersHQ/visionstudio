# AIDLC API Reference

The VisionStudio daemon exposes REST API endpoints for managing AIDLC workflows.

## Base URL

```
http://127.0.0.1:8765/api
```

## Endpoints

### Get AIDLC State

Returns the current workflow state for a project.

```
GET /projects/{project}/aidlc/state
```

**Response:**

```json
{
  "state": {
    "current_phase": "inception",
    "current_document": "vision_document",
    "completed_docs": ["vision_document"],
    "pending_docs": ["requirements_spec", "technical_spec"],
    "in_progress_docs": [],
    "document_scores": {},
    "phase_progress": {
      "inception": 0.33,
      "construction": 0.0,
      "operations": 0.0
    },
    "overall_progress": 0.11,
    "last_updated": "2024-07-16T10:30:00Z"
  }
}
```

### Get AIDLC Workflow

Returns the workflow DAG structure.

```
GET /projects/{project}/aidlc/workflow
```

**Response:**

```json
{
  "workflow": {
    "name": "aidlc-workflow",
    "phases": [...],
    "nodes": {...},
    "edges": [...],
    "progress": {
      "completed": 3,
      "total": 12,
      "percent": 25.0
    }
  },
  "mermaid": "graph TD\n  A[Vision Document] --> B[Requirements]\n  ..."
}
```

### List Documents

Lists all AIDLC documents in a project.

```
GET /projects/{project}/aidlc/documents
```

**Response:**

```json
{
  "documents": [
    {
      "type": "vision_document",
      "phase": "inception",
      "path": "/path/to/aidlc-docs/inception/vision_document.md",
      "title": "Vision Document",
      "status": "completed",
      "updated_at": "2024-07-16T10:30:00Z"
    }
  ]
}
```

### Get Document

Returns a specific AIDLC document with content.

```
GET /projects/{project}/aidlc/documents/{docId}
```

**Response:**

```json
{
  "document": {
    "type": "vision_document",
    "phase": "inception",
    "path": "/path/to/file.md",
    "title": "Vision Document",
    "status": "draft",
    "content": "# Vision Document\n...",
    "score": {
      "rating": "GOOD",
      "score": 0.78,
      "issues": [],
      "dimensions": {
        "clarity": 0.85,
        "completeness": 0.72
      }
    },
    "updated_at": "2024-07-16T10:30:00Z"
  }
}
```

### Get Phase Requirements

Returns requirements for all phases.

```
GET /projects/{project}/aidlc/phase/requirements
```

**Response:**

```json
{
  "current_phase": "inception",
  "requirements": [
    {
      "phase": "inception",
      "required_docs": ["vision_document", "requirements_spec", "technical_spec"],
      "completed_docs": ["vision_document"],
      "pending_docs": ["requirements_spec", "technical_spec"],
      "progress_percent": 33.33,
      "can_advance": false
    }
  ]
}
```

### Transition Phase

Attempts to transition to a new phase.

```
POST /projects/{project}/aidlc/phase/transition
```

**Request:**

```json
{
  "target_phase": "construction",
  "force": false,
  "approved_by": "team-lead",
  "notes": "All inception requirements met"
}
```

**Response:**

```json
{
  "success": true,
  "from_phase": "inception",
  "to_phase": "construction"
}
```

Or if blocked:

```json
{
  "success": false,
  "from_phase": "inception",
  "to_phase": "construction",
  "blocking_docs": ["requirements_spec"],
  "blocking_issues": ["requirements_spec not completed"]
}
```

### Get Sync Diff

Returns differences between VisionSpec and AIDLC directories.

```
GET /projects/{project}/aidlc/sync/diff
```

**Response:**

```json
{
  "diff": {
    "visionspec_dir": ".visionspec",
    "aidlc_docs_dir": "aidlc-docs",
    "actions": [
      {
        "direction": "to_aidlc",
        "doc_type": "vision_document",
        "source_path": ".visionspec/vision_document.md",
        "dest_path": "aidlc-docs/inception/vision_document.md",
        "action": "update",
        "reason": "Source newer"
      }
    ],
    "conflicts": [],
    "computed_at": "2024-07-16T10:30:00Z"
  }
}
```

### Trigger Sync

Synchronizes documents between directories.

```
POST /projects/{project}/aidlc/sync
```

**Request:**

```json
{
  "direction": "bidirectional",
  "dry_run": false
}
```

**Response:**

```json
{
  "result": {
    "direction": "bidirectional",
    "created": ["vision_document.md"],
    "updated": [],
    "skipped": [],
    "errors": [],
    "completed_at": "2024-07-16T10:30:00Z"
  }
}
```

### List Templates

Returns all available document templates.

```
GET /projects/{project}/aidlc/templates
```

**Response:**

```json
{
  "templates": [
    {
      "doc_type": "vision_document",
      "name": "Vision Document",
      "description": "Project vision and goals",
      "phase": "inception",
      "sections": [
        {
          "id": "vision",
          "title": "Vision Statement",
          "required": true
        }
      ]
    }
  ]
}
```

### Get Template

Returns a specific template with content.

```
GET /projects/{project}/aidlc/templates/{docType}
```

**Response:**

```json
{
  "template": {
    "doc_type": "vision_document",
    "name": "Vision Document",
    "description": "Project vision and goals",
    "phase": "inception",
    "sections": [...],
    "content": "---\ntitle: {{.Title}}\n---\n# Vision Document\n..."
  }
}
```

### Create Document

Creates a new document from a template.

```
POST /projects/{project}/aidlc/documents/create
```

**Request:**

```json
{
  "doc_type": "vision_document",
  "title": "My Vision Document",
  "author": "John Doe",
  "description": "Vision for the new feature",
  "overwrite": false
}
```

**Response:**

```json
{
  "document": {
    "type": "vision_document",
    "phase": "inception",
    "path": "/path/to/aidlc-docs/inception/vision_document.md",
    "title": "My Vision Document",
    "content": "---\ntitle: My Vision Document\n---\n...",
    "status": "draft",
    "updated_at": "2024-07-16T10:30:00Z"
  }
}
```

## Error Responses

All endpoints may return error responses:

```json
{
  "error": "Error message description"
}
```

HTTP status codes:

- `200` - Success
- `400` - Bad request (invalid parameters)
- `404` - Resource not found
- `409` - Conflict (e.g., document exists)
- `500` - Internal server error
