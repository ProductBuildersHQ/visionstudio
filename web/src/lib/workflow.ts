import type { APIInitiative, SpecWorkflow } from '../api/types'

/**
 * Default workflow ID for an initiative type. Mirrors the backend's
 * specworkflow.DefaultWorkflowForType.
 */
export function defaultWorkflowForType(initType?: string): string {
  switch (initType) {
    case 'maintenance':
    case 'refactor':
    case 'migration':
      return 'quick-fix'
    default:
      return 'pbhq-lite'
  }
}

/**
 * Resolve the workflow definition that applies to an initiative: its explicit
 * workflowId if set, otherwise the default for its type. Returns undefined if
 * the workflow isn't in the provided catalog list.
 */
export function resolveWorkflow(
  initiative: Pick<APIInitiative, 'workflowId' | 'type'>,
  workflows: SpecWorkflow[]
): SpecWorkflow | undefined {
  const id = initiative.workflowId || defaultWorkflowForType(initiative.type)
  return workflows.find((w) => w.id === id)
}

/** Filename (e.g. "PRD.md") → uppercase spec type (e.g. "PRD"). */
function fileToSpecType(filename: string): string {
  return filename.replace(/\.md$/i, '').toUpperCase()
}

/**
 * The workflow's required spec types (uppercase, matching SpecFile.specType),
 * in flow order. Prefers the explicit sequence, filtered to required specs;
 * falls back to the specsRequired filename list.
 */
export function requiredSpecTypes(workflow: SpecWorkflow): string[] {
  const required = (workflow.specsRequired ?? []).map(fileToSpecType)
  const seq = workflow.sequence ?? []
  if (seq.length === 0) return required
  const requiredSet = new Set(required)
  const inSeq = seq.filter((t) => requiredSet.has(t))
  const trailing = required.filter((t) => !seq.includes(t))
  return [...inSeq, ...trailing]
}
