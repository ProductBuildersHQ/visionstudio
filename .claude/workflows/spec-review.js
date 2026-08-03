export const meta = {
  name: 'spec-review',
  description: 'Review and evaluate all specs for an initiative with multi-lens verification',
  whenToUse: 'Use before major milestones to verify spec quality and gate compliance',
  phases: [
    { title: 'Gather', detail: 'Collect all specs for review' },
    { title: 'Evaluate', detail: 'Run LLM evaluation on each spec' },
    { title: 'Verify', detail: 'Cross-check findings with multiple lenses' },
    { title: 'Report', detail: 'Generate review summary' },
  ],
}

phase('Gather')
const initiativeId = args.initiative

log(`Gathering specs for ${initiativeId}`)

const specs = await agent(
  `Gather all specs for initiative ${initiativeId}:
  1. Call spec_list to get all spec documents
  2. For each spec, call spec_read to get content
  3. Return array of {type, content, file_path, status}`,
  { label: 'gather', schema: { type: 'array', items: { type: 'object' } } }
)

if (!specs || specs.length === 0) {
  return { error: 'No specs found', initiative: initiativeId }
}

log(`Found ${specs.length} specs to review`)

phase('Evaluate')
const evaluations = await parallel(
  specs.map(spec => async () => {
    const result = await agent(
      `Evaluate ${spec.type} spec:
      1. Call spec_evaluate for spec_type="${spec.type}"
      2. Return full evaluation result including score, verdict, findings`,
      { label: `eval:${spec.type}`, phase: 'Evaluate', schema: { type: 'object' } }
    )
    return { ...spec, evaluation: result }
  })
)

phase('Verify')
const LENSES = ['completeness', 'consistency', 'feasibility']

const verified = await pipeline(
  evaluations.filter(e => e?.evaluation),
  async (spec) => {
    if (spec.evaluation.score >= 85 && spec.evaluation.verdict === 'pass') {
      return { ...spec, verified: true, lensResults: [] }
    }

    const lensResults = await parallel(
      LENSES.map(lens => async () => {
        const verdict = await agent(
          `Verify ${spec.type} spec through ${lens} lens:

          Content:
          ${spec.content?.substring(0, 2000)}

          Evaluation findings:
          ${JSON.stringify(spec.evaluation.findings || [])}

          Question: Is this spec ${lens === 'completeness' ? 'complete with all required sections' : lens === 'consistency' ? 'consistent with other specs in the workflow' : 'technically feasible to implement'}?

          Return {lens: "${lens}", passed: boolean, reason: string}`,
          { label: `verify:${spec.type}:${lens}`, phase: 'Verify', schema: { type: 'object' } }
        )
        return verdict
      })
    )

    const passedLenses = lensResults.filter(r => r?.passed).length
    return {
      ...spec,
      verified: passedLenses >= 2,
      lensResults: lensResults.filter(Boolean),
    }
  }
)

phase('Report')
const passing = verified.filter(s => s.verified)
const failing = verified.filter(s => !s.verified)

const summary = {
  initiative: initiativeId,
  total_specs: specs.length,
  passing: passing.length,
  failing: failing.length,
  gate_status: failing.length === 0 ? 'PASSED' : 'BLOCKED',
  specs: verified.map(s => ({
    type: s.type,
    score: s.evaluation?.score,
    verdict: s.evaluation?.verdict,
    verified: s.verified,
    blockers: s.verified ? [] : s.lensResults?.filter(l => !l.passed).map(l => l.reason) || [],
  })),
}

if (failing.length > 0) {
  log(`BLOCKED: ${failing.length} spec(s) need attention`)
  summary.next_steps = failing.map(s => `Improve ${s.type}: ${s.lensResults?.find(l => !l.passed)?.reason || 'score below threshold'}`)
} else {
  log('All gates passed')
}

return summary
