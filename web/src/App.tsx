import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, useNavigate, useParams } from 'react-router-dom'
import { getExecution, getSpecs, getMaturity, getSpend } from './api/client'
import type { ExecutionResponse, SpecsResponse, MaturityResponse, SpendResponse } from './api/types'
import { Sidebar } from './components/Sidebar'
import { InitiativesOverview } from './panels/InitiativesOverview'
import { InitiativeDetail } from './panels/InitiativeDetail'
import { MaturityPanel } from './panels/MaturityPanel'
import { PerformancePanel } from './panels/PerformancePanel'
import { SpecViewer } from './panels/SpecViewer'
import { RepositoriesPanel } from './panels/RepositoriesPanel'
import { RepositoryDetail } from './panels/RepositoryDetail'
import { LoadingState, ErrorState } from './components'

export type NavSection = 'initiatives' | 'repositories' | 'maturity' | 'performance'
export type NavTarget =
  | { section: 'initiatives'; view: 'all' }
  | { section: 'initiatives'; view: 'program'; programId: string }
  | { section: 'initiatives'; view: 'standalone' }
  | { section: 'initiatives'; view: 'initiative'; initiativeId: string }
  | { section: 'repositories'; view?: 'all' }
  | { section: 'repositories'; view: 'repository'; repositoryId: string }
  | { section: 'maturity' }
  | { section: 'performance' }

export default function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  )
}

function AppContent() {
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [specs, setSpecs] = useState<SpecsResponse | null>(null)
  const [maturity, setMaturity] = useState<MaturityResponse | null>(null)
  const [spend, setSpend] = useState<SpendResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const navigate = useNavigate()

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

  const handleNavigate = (target: NavTarget) => {
    if (target.section === 'maturity') {
      navigate('/maturity')
    } else if (target.section === 'performance') {
      navigate('/performance')
    } else if (target.section === 'repositories') {
      if (target.view === 'repository') {
        navigate(`/repository/${target.repositoryId}`)
      } else {
        navigate('/repositories')
      }
    } else if (target.view === 'all') {
      navigate('/')
    } else if (target.view === 'program') {
      navigate(`/program/${target.programId}`)
    } else if (target.view === 'standalone') {
      navigate('/standalone')
    } else if (target.view === 'initiative') {
      navigate(`/initiative/${target.initiativeId}`)
    }
  }

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
      <Sidebar
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        execution={execution}
        onNavigate={handleNavigate}
        apiStatus={apiStatus}
      />

      <main className="flex-1 overflow-auto">
        <div className="p-6">
          <Routes>
            <Route
              path="/"
              element={
                <InitiativesOverview
                  title="All Initiatives"
                  initiatives={execution.initiatives}
                  programs={execution.programs}
                  rmis={execution.rmis}
                  onInitiativeClick={(id) => navigate(`/initiative/${id}`)}
                  showProgramGroups={true}
                />
              }
            />
            <Route
              path="/program/:programId"
              element={
                <ProgramView
                  execution={execution}
                  onInitiativeClick={(id) => navigate(`/initiative/${id}`)}
                />
              }
            />
            <Route
              path="/standalone"
              element={
                <InitiativesOverview
                  title="Standalone Initiatives"
                  initiatives={execution.initiatives.filter((i) => !i.programId)}
                  programs={execution.programs}
                  rmis={execution.rmis}
                  onInitiativeClick={(id) => navigate(`/initiative/${id}`)}
                  showProgramGroups={false}
                />
              }
            />
            <Route
              path="/initiative/:initiativeId"
              element={
                <InitiativeView
                  execution={execution}
                  specs={specs}
                  onBack={() => navigate('/')}
                />
              }
            />
            <Route path="/initiative/:initiativeId/spec/:specType" element={<SpecViewer />} />
            <Route path="/initiative/:initiativeId/spec" element={<SpecViewer />} />
            <Route
              path="/repositories"
              element={
                <RepositoriesPanel
                  execution={execution}
                  onRepositoryClick={(id) => navigate(`/repository/${id}`)}
                />
              }
            />
            <Route
              path="/repository/*"
              element={
                <RepositoryView
                  execution={execution}
                  onBack={() => navigate('/repositories')}
                  onInitiativeClick={(id) => navigate(`/initiative/${id}`)}
                />
              }
            />
            <Route path="/maturity" element={<MaturityPanel />} />
            <Route path="/performance" element={<PerformancePanel />} />
          </Routes>
        </div>
      </main>
    </div>
  )
}

function ProgramView({
  execution,
  onInitiativeClick,
}: {
  execution: ExecutionResponse
  onInitiativeClick: (id: string) => void
}) {
  const { programId } = useParams<{ programId: string }>()
  const program = execution.programs.find((p) => p.id === programId)
  const initiatives = execution.initiatives.filter((i) => i.programId === programId)

  return (
    <InitiativesOverview
      title={program?.name ?? programId ?? 'Program'}
      initiatives={initiatives}
      programs={execution.programs}
      rmis={execution.rmis}
      onInitiativeClick={onInitiativeClick}
      showProgramGroups={false}
    />
  )
}

function InitiativeView({
  execution,
  specs,
  onBack,
}: {
  execution: ExecutionResponse
  specs: SpecsResponse
  onBack: () => void
}) {
  const { initiativeId } = useParams<{ initiativeId: string }>()
  const initiative = execution.initiatives.find((i) => i.id === initiativeId)

  if (!initiative) {
    return (
      <div className="text-center text-gray-400 py-12">
        <p>Initiative not found: {initiativeId}</p>
      </div>
    )
  }

  return (
    <InitiativeDetail
      initiative={initiative}
      execution={execution}
      specs={specs}
      onBack={onBack}
    />
  )
}

function RepositoryView({
  execution,
  onBack,
  onInitiativeClick,
}: {
  execution: ExecutionResponse
  onBack: () => void
  onInitiativeClick: (id: string) => void
}) {
  const { '*': repositoryId } = useParams()
  const repository = execution.repositories.find((r) => r.id === repositoryId)

  if (!repository) {
    return (
      <div className="text-center text-gray-400 py-12">
        <p>Repository not found: {repositoryId}</p>
      </div>
    )
  }

  return (
    <RepositoryDetail
      repository={repository}
      execution={execution}
      onBack={onBack}
      onInitiativeClick={onInitiativeClick}
    />
  )
}
