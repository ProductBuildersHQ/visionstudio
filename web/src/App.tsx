import { useState, useEffect } from 'react'
import { getExecution, getSpecs, getMaturity, getSpend } from './api/client'
import type { ExecutionResponse, SpecsResponse, MaturityResponse, SpendResponse } from './api/types'
import { Sidebar } from './components/Sidebar'
import { InitiativesOverview } from './panels/InitiativesOverview'
import { InitiativeDetail } from './panels/InitiativeDetail'
import { MaturityPanel } from './panels/MaturityPanel'
import { SpendPanel } from './panels/SpendPanel'
import { LoadingState, ErrorState } from './components'

export type NavSection = 'initiatives' | 'maturity' | 'spend'
export type NavTarget =
  | { section: 'initiatives'; view: 'all' }
  | { section: 'initiatives'; view: 'program'; programId: string }
  | { section: 'initiatives'; view: 'standalone' }
  | { section: 'initiatives'; view: 'initiative'; initiativeId: string }
  | { section: 'maturity' }
  | { section: 'spend' }

export default function App() {
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [specs, setSpecs] = useState<SpecsResponse | null>(null)
  const [maturity, setMaturity] = useState<MaturityResponse | null>(null)
  const [spend, setSpend] = useState<SpendResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [navTarget, setNavTarget] = useState<NavTarget>({ section: 'initiatives', view: 'all' })

  const reload = () => {
    setError(null)
    Promise.all([getExecution(), getSpecs(), getMaturity(), getSpend()])
      .then(([e, s, m, sp]) => {
        setExecution(e)
        setSpecs(s)
        setMaturity(m)
        setSpend(sp)
      })
      .catch((err: Error) => setError(err.message))
  }

  useEffect(() => {
    reload()
  }, [])

  const apiStatus = execution ? 'connected' : error ? 'error' : 'loading'

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 text-gray-100 flex items-center justify-center">
        <ErrorState message={error} onRetry={reload} />
      </div>
    )
  }

  if (!execution || !specs || !maturity || !spend) {
    return (
      <div className="min-h-screen bg-gray-900 text-gray-100 flex items-center justify-center">
        <LoadingState message="Loading data..." />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 flex">
      {/* Sidebar */}
      <Sidebar
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        execution={execution}
        navTarget={navTarget}
        onNavigate={setNavTarget}
        apiStatus={apiStatus}
      />

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        <div className="p-6">
          <MainContent
            navTarget={navTarget}
            execution={execution}
            specs={specs}
            maturity={maturity}
            spend={spend}
            onNavigate={setNavTarget}
          />
        </div>
      </main>
    </div>
  )
}

function MainContent({
  navTarget,
  execution,
  specs,
  onNavigate,
}: {
  navTarget: NavTarget
  execution: ExecutionResponse
  specs: SpecsResponse
  maturity: MaturityResponse
  spend: SpendResponse
  onNavigate: (target: NavTarget) => void
}) {
  if (navTarget.section === 'maturity') {
    return <MaturityPanel />
  }

  if (navTarget.section === 'spend') {
    return <SpendPanel />
  }

  // Initiatives section
  if (navTarget.view === 'initiative') {
    const initiative = execution.initiatives.find((i) => i.id === navTarget.initiativeId)
    if (initiative) {
      return (
        <InitiativeDetail
          initiative={initiative}
          execution={execution}
          specs={specs}
          onBack={() => onNavigate({ section: 'initiatives', view: 'all' })}
        />
      )
    }
  }

  // Filter initiatives based on view
  let filteredInitiatives = execution.initiatives
  let title = 'All Initiatives'

  if (navTarget.view === 'program') {
    const program = execution.programs.find((p) => p.id === navTarget.programId)
    filteredInitiatives = execution.initiatives.filter((i) => i.programId === navTarget.programId)
    title = program?.name ?? navTarget.programId
  } else if (navTarget.view === 'standalone') {
    filteredInitiatives = execution.initiatives.filter((i) => !i.programId)
    title = 'Standalone Initiatives'
  }

  return (
    <InitiativesOverview
      title={title}
      initiatives={filteredInitiatives}
      programs={execution.programs}
      rmis={execution.rmis}
      onInitiativeClick={(id) => onNavigate({ section: 'initiatives', view: 'initiative', initiativeId: id })}
      showProgramGroups={navTarget.view === 'all'}
    />
  )
}
