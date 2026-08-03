export const meta = {
  name: 'spec-author',
  description: 'Author specs through visionstudio workflow with evaluation gates',
  whenToUse: 'Use when starting a new initiative or adding specs to an existing one',
  phases: [
    { title: 'Setup', detail: 'Select workflow and check status' },
    { title: 'Author', detail: 'Generate or create specs' },
    { title: 'Evaluate', detail: 'LLM-as-judge evaluation' },
    { title: 'Summary', detail: 'Report workflow status' },
  ],
}

phase('Setup')
const workflowId = args.workflow || 'pbhq-lite'
const initiativeId = args.initiative

log(`Setting up ${workflowId} workflow for ${initiativeId}`)

const setup = await agent(
  `Select workflow and check status:
  1. Call workflow_select with initiative_id="${initiativeId}" and workflow_id="${workflowId}"
  2. Call workflow_status to see required specs and current status
  3. Return JSON with {selected: true/false, status: workflow_status_result}`,
  { label: 'setup', schema: { type: 'object', properties: { selected: { type: 'boolean' }, status: { type: 'object' } } } }
)

if (!setup?.status) {
  return { error: 'Failed to setup workflow', setup }
}

phase('Author')
const missingSpecs = (setup.status.specs || [])
  .filter(s => s.required && s.status === 'missing')
  .map(s => s.type)

if (missingSpecs.length === 0) {
  log('All required specs present')
} else {
  log(`Missing specs: ${missingSpecs.join(', ')}`)

  const authored = await pipeline(
    missingSpecs,
    async (specType) => {
      const result = await agent(
        `Author ${specType} spec for initiative ${initiativeId}:
        1. Check workflow_status for upstream specs that can be sources
        2. If sources exist, use spec_synthesize with dry_run=true first to preview
        3. Review the preview and adjust if needed
        4. Create the spec with spec_create or spec_synthesize (dry_run=false)
        5. Return {spec_type: "${specType}", created: true/false, method: "synthesize"|"manual"}`,
        { label: `author:${specType}`, phase: 'Author', schema: { type: 'object' } }
      )
      return result
    }
  )
}

phase('Evaluate')
const toEvaluate = (setup.status.specs || [])
  .filter(s => s.status === 'draft' || s.status === 'created')
  .map(s => s.type)

const evaluated = await pipeline(
  toEvaluate,
  async (specType) => {
    const result = await agent(
      `Evaluate ${specType} spec for initiative ${initiativeId}:
      1. Call spec_evaluate with spec_type="${specType}"
      2. If score < 85, identify specific improvements needed
      3. Return {spec_type: "${specType}", score: number, verdict: string, improvements: []}`,
      { label: `eval:${specType}`, phase: 'Evaluate', schema: { type: 'object' } }
    )
    return result
  }
)

phase('Summary')
const finalStatus = await agent(
  `Get final workflow status for ${initiativeId}:
  1. Call workflow_status
  2. Return the full status including gates_passed and blockers`,
  { label: 'final-status', schema: { type: 'object' } }
)

return {
  initiative: initiativeId,
  workflow: workflowId,
  status: finalStatus,
  evaluated: evaluated.filter(Boolean),
}
