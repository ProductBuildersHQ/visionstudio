export const meta = {
  name: 'spec-synthesize',
  description: 'Synthesize a spec from upstream sources following workflow DAG',
  whenToUse: 'Use when generating a downstream spec from existing upstream specs',
  phases: [
    { title: 'Discover', detail: 'Find upstream sources' },
    { title: 'Preview', detail: 'Dry-run synthesis' },
    { title: 'Generate', detail: 'Create the spec' },
    { title: 'Validate', detail: 'Evaluate generated spec' },
  ],
}

const DAG = {
  trd: ['prd'],
  plan: ['prd', 'trd'],
  roadmap: ['plan'],
  press: ['mrd', 'prd'],
  faq: ['mrd', 'press'],
  tpd: ['prd', 'trd'],
  uxd: ['prd'],
}

phase('Discover')
const initiativeId = args.initiative
const targetType = args.spec_type

if (!targetType) {
  return { error: 'spec_type is required in args' }
}

const requiredSources = DAG[targetType] || []
log(`Synthesizing ${targetType} from sources: ${requiredSources.join(', ') || 'none defined'}`)

const sources = await agent(
  `Find source specs for synthesizing ${targetType}:
  1. Call spec_list for initiative ${initiativeId}
  2. For each source in [${requiredSources.join(', ')}]:
     - Call spec_read to get content
     - Include in sources array
  3. Return array of {type, path, content}`,
  { label: 'discover', schema: { type: 'array', items: { type: 'object' } } }
)

if (!sources || sources.length === 0) {
  return {
    error: 'No sources found',
    target: targetType,
    required: requiredSources,
    suggestion: `Create ${requiredSources[0] || 'prd'} first`,
  }
}

const missingRequired = requiredSources.filter(r => !sources.find(s => s.type === r))
if (missingRequired.length > 0) {
  return {
    error: 'Missing required sources',
    target: targetType,
    missing: missingRequired,
    found: sources.map(s => s.type),
  }
}

phase('Preview')
log('Running dry-run synthesis')

const preview = await agent(
  `Preview synthesis of ${targetType}:
  1. Call spec_synthesize with:
     - target_spec_type: "${targetType}"
     - initiative_id: "${initiativeId}"
     - sources: ${JSON.stringify(sources.map(s => ({ type: s.type, content: s.content })))}
     - dry_run: true
  2. Return the preview result`,
  { label: 'preview', schema: { type: 'object' } }
)

log(`Preview generated (${preview?.content?.length || 0} chars)`)

phase('Generate')
const userApproved = args.auto_approve || false

if (!userApproved) {
  log('Dry-run complete. Set auto_approve=true to generate.')
  return {
    status: 'preview',
    target: targetType,
    sources: sources.map(s => s.type),
    preview: preview?.content?.substring(0, 1000) + '...',
    next: `Run again with auto_approve=true to generate`,
  }
}

log('Generating spec')

const generated = await agent(
  `Generate ${targetType} spec:
  1. Call spec_synthesize with:
     - target_spec_type: "${targetType}"
     - initiative_id: "${initiativeId}"
     - sources: ${JSON.stringify(sources.map(s => ({ type: s.type, content: s.content })))}
     - dry_run: false
  2. Return the synthesis result`,
  { label: 'generate', schema: { type: 'object' } }
)

phase('Validate')
log('Evaluating generated spec')

const evaluation = await agent(
  `Evaluate the generated ${targetType} spec:
  1. Call spec_evaluate with spec_type="${targetType}" for initiative ${initiativeId}
  2. Return the evaluation result`,
  { label: 'validate', schema: { type: 'object' } }
)

const passed = evaluation?.score >= 85 && evaluation?.verdict !== 'fail'

return {
  status: passed ? 'success' : 'needs_improvement',
  target: targetType,
  sources: sources.map(s => s.type),
  score: evaluation?.score,
  verdict: evaluation?.verdict,
  findings: evaluation?.findings || [],
  next: passed ? 'Spec ready for use' : 'Review findings and iterate',
}
