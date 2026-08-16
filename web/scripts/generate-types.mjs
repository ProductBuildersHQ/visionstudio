#!/usr/bin/env node
/**
 * Generate Zod schemas and TypeScript types from JSON Schema.
 *
 * Pipeline: Go structs → JSON Schema → Zod → TypeScript
 *
 * Usage:
 *   1. Run `go generate ./pkg/apitypes` to update JSON schemas
 *   2. Run `npm run generate:types` to regenerate Zod/TS
 *
 * Output:
 *   - src/api/schemas.gen.ts (Zod schemas)
 *   - src/api/types.gen.ts (TypeScript types via z.infer<>)
 */

import { readFileSync, writeFileSync, existsSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'
import { jsonSchemaToZod } from 'json-schema-to-zod'

const __dirname = dirname(fileURLToPath(import.meta.url))
const schemaDir = join(__dirname, '../../pkg/apitypes/schema')
const outDir = join(__dirname, '../src/api')

function resolveRefs(schema, defs, visited = new Set()) {
  if (!schema || typeof schema !== 'object') return schema

  if (schema.$ref) {
    const refPath = schema.$ref.replace('#/$defs/', '').replace('#/definitions/', '')
    if (visited.has(refPath)) {
      return { type: 'object', additionalProperties: true }
    }
    if (defs[refPath]) {
      visited.add(refPath)
      const resolved = resolveRefs(JSON.parse(JSON.stringify(defs[refPath])), defs, new Set(visited))
      return resolved
    }
    return schema
  }

  if (Array.isArray(schema)) {
    return schema.map(v => resolveRefs(v, defs, visited))
  }

  const result = {}
  for (const [key, value] of Object.entries(schema)) {
    if (key === '$defs' || key === 'definitions') continue
    if (key === '$schema' || key === '$id' || key === '$ref') continue
    if (Array.isArray(value)) {
      result[key] = value.map(v => resolveRefs(v, defs, visited))
    } else if (typeof value === 'object' && value !== null) {
      result[key] = resolveRefs(value, defs, visited)
    } else {
      result[key] = value
    }
  }
  return result
}

// Types with individual schema files
const schemaNames = [
  'JudgeResult',
  'SpecWorkflow',
  'SpecFile',
  'SpecFilesResponse',
  'SpecsResponse',
  'ContextSpec',
  'Phase',
  'Program',
  'Initiative',
  'RoadmapItem',
  'APIRMI',
  'APIRepository',
  'APIRMIDependency',
  'APIInitiativeDependency',
  'APIStatusCount',
  'APIPhase',
  'APIInitiative',
  'APIProgram',
  'CreateInitiativeRequest',
  'CreateInitiativeResponse',
  'WorkflowSpecDetail',
  'ExecutionResponse',
]

// Types from rubric package embedded in JudgeResult.schema.json
// These are extracted from $defs and generated separately
const rubricTypes = [
  'Rubric',
  'JudgeMetadata',
  'CategoryResult',
  'Finding',
  'ActionItem',
  'NextSteps',
  'Decision',
  'ReportMetadata',
  'PassCriteria',
  'ReferenceData',
  'ChecklistResults',
]

const zodSchemas = []
const typeExports = []
const processedNames = []

// Load JudgeResult schema which contains all rubric type definitions
const judgeResultPath = join(schemaDir, 'JudgeResult.schema.json')
const judgeResultSchema = JSON.parse(readFileSync(judgeResultPath, 'utf8'))
const rubricDefs = judgeResultSchema.$defs || judgeResultSchema.definitions || {}

// First, generate rubric types from embedded definitions
for (const name of rubricTypes) {
  const schema = rubricDefs[name]
  if (!schema) {
    console.warn(`Warning: ${name} not found in JudgeResult.$defs, skipping`)
    continue
  }

  const resolved = resolveRefs(JSON.parse(JSON.stringify(schema)), rubricDefs)

  try {
    const zodCode = jsonSchemaToZod(resolved, {
      module: 'none',
      noImport: true,
    })
    zodSchemas.push(`export const ${name}Schema = ${zodCode}`)
    typeExports.push(`export type ${name} = z.infer<typeof ${name}Schema>`)
    processedNames.push(name)
  } catch (err) {
    console.error(`Error converting ${name}: ${err.message}`)
    zodSchemas.push(`// TODO: fix schema conversion for ${name}`)
    zodSchemas.push(`export const ${name}Schema = z.any()`)
    typeExports.push(`export type ${name} = z.infer<typeof ${name}Schema>`)
    processedNames.push(name)
  }
}

// Then generate types from individual schema files
for (const name of schemaNames) {
  const schemaPath = join(schemaDir, `${name}.schema.json`)
  if (!existsSync(schemaPath)) {
    console.warn(`Warning: ${schemaPath} not found, skipping`)
    continue
  }

  const schemaFile = JSON.parse(readFileSync(schemaPath, 'utf8'))
  const defs = schemaFile.$defs || schemaFile.definitions || {}

  let mainSchema
  if (schemaFile.$ref) {
    const refName = schemaFile.$ref.replace('#/$defs/', '').replace('#/definitions/', '')
    mainSchema = defs[refName]
  } else {
    mainSchema = schemaFile
  }

  if (!mainSchema || mainSchema === true) {
    console.warn(`Warning: ${name} has no valid schema, skipping`)
    continue
  }

  const resolved = resolveRefs(JSON.parse(JSON.stringify(mainSchema)), defs)

  try {
    const zodCode = jsonSchemaToZod(resolved, {
      module: 'none',
      noImport: true,
    })
    zodSchemas.push(`export const ${name}Schema = ${zodCode}`)
    typeExports.push(`export type ${name} = z.infer<typeof ${name}Schema>`)
    processedNames.push(name)
  } catch (err) {
    console.error(`Error converting ${name}: ${err.message}`)
    zodSchemas.push(`// TODO: fix schema conversion for ${name}`)
    zodSchemas.push(`export const ${name}Schema = z.any()`)
    typeExports.push(`export type ${name} = z.infer<typeof ${name}Schema>`)
    processedNames.push(name)
  }
}

const schemasContent = `// AUTO-GENERATED - DO NOT EDIT
// Generated from JSON Schema via: npm run generate:types
// Source: Go structs in pkg/apitypes/types.go

import { z } from 'zod'

${zodSchemas.join('\n\n')}
`

const typesContent = `// AUTO-GENERATED - DO NOT EDIT
// Generated from JSON Schema via: npm run generate:types
// Source: Go structs in pkg/apitypes/types.go

import { z } from 'zod'
import {
${processedNames.map(n => `  ${n}Schema,`).join('\n')}
} from './schemas.gen'

${typeExports.join('\n')}

// Re-export schemas for runtime validation
export {
${processedNames.map(n => `  ${n}Schema,`).join('\n')}
}
`

writeFileSync(join(outDir, 'schemas.gen.ts'), schemasContent)
console.log('Generated src/api/schemas.gen.ts')

writeFileSync(join(outDir, 'types.gen.ts'), typesContent)
console.log('Generated src/api/types.gen.ts')

console.log(`Done! Generated ${processedNames.length} types.`)
