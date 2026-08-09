# Type Pipeline

VisionStudio maintains type synchronization between Go backend and TypeScript frontend through a **Go-first type pipeline**.

## Pipeline Overview

```
Go structs (pkg/apitypes/types.go)
    │
    │ go generate ./pkg/apitypes
    ▼
JSON Schema (pkg/apitypes/schema/*.schema.json)
    │
    │ npm run generate:types
    ▼
Zod schemas (web/src/api/schemas.gen.ts)
    │
    │ z.infer<>
    ▼
TypeScript types (web/src/api/types.gen.ts)
```

## Why Go-First?

1. **Single source of truth** — Go types define the API contract
2. **Compile-time safety** — Go catches type errors before runtime
3. **Auto-sync** — Frontend types update automatically when backend changes
4. **No drift** — Generated code can't get out of sync

## Regenerating Types

After modifying `pkg/apitypes/types.go`:

```bash
# Generate JSON schemas from Go types
go generate ./pkg/apitypes

# Generate Zod/TypeScript from JSON schemas
cd web && npm run generate:types
```

## JSON Naming Convention

API types use **camelCase** for JSON serialization to match JavaScript conventions:

```go
// pkg/apitypes/types.go
type JudgeResult struct {
    ID           string    `json:"id"`
    InitiativeID string    `json:"initiativeId"`    // camelCase
    SpecPath     string    `json:"specPath"`        // camelCase
    EvaluatedAt  time.Time `json:"evaluatedAt"`     // camelCase
}
```

Store types use **snake_case** for database compatibility:

```go
// pkg/store/store.go
type JudgeResult struct {
    ID           string    `json:"id"`
    InitiativeID string    `json:"initiative_id"`   // snake_case
    SpecPath     string    `json:"spec_path"`       // snake_case
}
```

## Type Conversion

API handlers convert between store and API types:

```go
// cmd/vistudio/api.go
func storeJudgeResultToAPI(r *store.JudgeResult) apitypes.JudgeResult {
    return apitypes.JudgeResult{
        ID:           r.ID,
        InitiativeID: r.InitiativeID,  // Same value, different JSON tag
        SpecPath:     r.SpecPath,
        // ...
    }
}
```

## Compat Layer

Generated TypeScript types have all fields optional (JSON Schema limitation). The compat layer normalizes these:

```typescript
// web/src/api/compat.ts

// Generated type (all optional)
type GenJudgeResult = {
  id?: string
  initiativeId?: string
  // ...
}

// Normalized type (required with defaults)
interface JudgeResult {
  id: string
  initiativeId: string
  // ...
}

// Converter
function toJudgeResult(gen: GenJudgeResult): JudgeResult {
  return {
    id: gen.id ?? '',
    initiativeId: gen.initiativeId ?? '',
    // ...
  }
}
```

Use converters in `web/src/api/client.ts`:

```typescript
export async function getSpecs(): Promise<SpecsResponse> {
  const raw = await fetchJSON<GenSpecsResponse>('/specs')
  return toSpecsResponse(raw)
}
```

## File Structure

```
pkg/apitypes/
├── types.go           # Go types (source of truth)
├── gen/
│   └── main.go        # Schema generator
└── schema/
    ├── JudgeResult.schema.json
    ├── SpecsResponse.schema.json
    └── ...

web/
├── scripts/
│   └── generate-types.mjs   # JSON Schema → Zod/TS
└── src/api/
    ├── schemas.gen.ts       # Generated Zod schemas
    ├── types.gen.ts         # Generated TypeScript types
    ├── compat.ts            # Optional→required normalization
    └── client.ts            # API client with converters
```

## Adding New Types

1. Add Go struct to `pkg/apitypes/types.go` with camelCase JSON tags
2. Add type name to the generator list in `pkg/apitypes/gen/main.go`
3. Run `go generate ./pkg/apitypes`
4. Add type name to `schemaNames` array in `web/scripts/generate-types.mjs`
5. Run `cd web && npm run generate:types`
6. If needed, add compat converter in `web/src/api/compat.ts`
7. Use converter in `web/src/api/client.ts`

## Tools

| Tool | Purpose |
|------|---------|
| `invopop/jsonschema` | Go → JSON Schema |
| `json-schema-to-zod` | JSON Schema → Zod |
| `zod` | Runtime validation + type inference |
