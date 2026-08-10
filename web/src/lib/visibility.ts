import type { APIInitiative, APIProgram } from '../api/types'

// isInitiativeVisible reports whether an initiative should appear in
// initiative-listing UI: not itself hidden, and not attached to a hidden
// program. Hiding a program cascades to hide its initiatives everywhere
// initiatives are listed, not just in program-grouped views.
export function isInitiativeVisible(
  initiative: APIInitiative,
  programsById: Map<string, APIProgram>
): boolean {
  if (initiative.hidden) return false
  if (initiative.programId) {
    const program = programsById.get(initiative.programId)
    if (program?.hidden) return false
  }
  return true
}

// visibleInitiatives filters a list of initiatives down to the ones that
// should appear in the UI, applying isInitiativeVisible against the given
// programs.
export function visibleInitiatives(
  initiatives: APIInitiative[],
  programs: APIProgram[]
): APIInitiative[] {
  const programsById = new Map(programs.map((p) => [p.id, p]))
  return initiatives.filter((i) => isInitiativeVisible(i, programsById))
}

// hiddenInitiativeIds returns the set of initiative IDs that are hidden
// (individually, or via a hidden program), for callers that only have raw
// initiative IDs to filter — e.g. IDs referenced by records that don't
// carry the full initiative object.
export function hiddenInitiativeIds(
  initiatives: APIInitiative[],
  programs: APIProgram[]
): Set<string> {
  const programsById = new Map(programs.map((p) => [p.id, p]))
  return new Set(
    initiatives.filter((i) => !isInitiativeVisible(i, programsById)).map((i) => i.id)
  )
}
