import { useMemo, useState } from 'react'
import type { APIProgram, APIInitiative, APIRMI, SpecWorkflow } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { ProgressBar } from '../components/ProgressBar'
import { PieChart } from '../components/charts'
import { CreateInitiativeModal } from '../components/CreateInitiativeModal'
import { visibleInitiatives } from '../lib/visibility'

interface InitiativesOverviewProps {
  title: string
  initiatives: APIInitiative[]
  programs: APIProgram[]
  rmis: APIRMI[]
  workflows: SpecWorkflow[]
  onInitiativeClick: (id: string) => void
  onInitiativeCreated: (id: string) => void
  showProgramGroups: boolean
  defaultProgramId?: string
}

export function InitiativesOverview({
  title,
  initiatives: allInitiatives,
  programs,
  rmis,
  workflows,
  onInitiativeClick,
  onInitiativeCreated,
  showProgramGroups,
  defaultProgramId,
}: InitiativesOverviewProps) {
  const [showCreate, setShowCreate] = useState(false)
  const initiatives = useMemo(
    () => visibleInitiatives(allInitiatives, programs),
    [allInitiatives, programs]
  )

  const createModal = showCreate && (
    <CreateInitiativeModal
      workflows={workflows}
      programs={programs}
      defaultProgramId={defaultProgramId}
      onClose={() => setShowCreate(false)}
      onCreated={(id) => {
        setShowCreate(false)
        onInitiativeCreated(id)
      }}
    />
  )

  const newInitiativeButton = (
    <button
      onClick={() => setShowCreate(true)}
      className="px-3 py-1.5 text-sm font-medium bg-purple-600 text-white rounded hover:bg-purple-500 transition-colors"
    >
      + New Initiative
    </button>
  )

  const statusDist = useMemo(() => {
    const initiativeRmis = rmis.filter((r) =>
      initiatives.some((i) => i.id === r.initiativeId)
    )
    const counts: Record<string, number> = {}
    for (const r of initiativeRmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [rmis, initiatives])

  const totalRmis = useMemo(() => {
    return rmis.filter((r) => initiatives.some((i) => i.id === r.initiativeId)).length
  }, [rmis, initiatives])

  if (initiatives.length === 0) {
    return (
      <div className="text-center text-gray-400 py-12 space-y-4">
        <p>No initiatives found</p>
        {newInitiativeButton}
        {createModal}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{title}</h1>
          <p className="text-gray-400 text-sm mt-1">
            {initiatives.length} initiative{initiatives.length !== 1 ? 's' : ''}, {totalRmis} RMIs
          </p>
        </div>
        <div className="flex items-center gap-4">
        {newInitiativeButton}
        {statusDist.length > 0 && (
          <div className="flex items-center gap-4">
            <div className="w-16 h-16">
              <PieChart data={statusDist} size={64} showLegend={false} innerRadius={20} />
            </div>
            <div className="text-sm space-y-1">
              {statusDist.slice(0, 4).map((s) => (
                <div key={s.name} className="flex items-center gap-2">
                  <span className="text-gray-400">{s.name}:</span>
                  <span className="text-gray-200">{s.value}</span>
                </div>
              ))}
            </div>
          </div>
        )}
        </div>
      </div>

      {createModal}

      {/* Initiative Tiles */}
      {showProgramGroups ? (
        <GroupedInitiatives
          initiatives={initiatives}
          programs={programs}
          onInitiativeClick={onInitiativeClick}
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {initiatives.map((init) => (
            <InitiativeTile
              key={init.id}
              initiative={init}
              onClick={() => onInitiativeClick(init.id)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function GroupedInitiatives({
  initiatives,
  programs,
  onInitiativeClick,
}: {
  initiatives: APIInitiative[]
  programs: APIProgram[]
  onInitiativeClick: (id: string) => void
}) {
  const visiblePrograms = programs.filter((p) => !p.hidden)
  const standalone = initiatives.filter((i) => !i.programId)

  return (
    <div className="space-y-8">
      {visiblePrograms.map((program) => {
        const programInits = initiatives.filter((i) => i.programId === program.id)
        if (programInits.length === 0) return null

        const totalProgress =
          programInits.reduce((sum, i) => sum + i.progress, 0) / programInits.length

        return (
          <div key={program.id} className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-medium text-gray-200">{program.name}</h2>
                {program.description && (
                  <p className="text-sm text-gray-500">{program.description}</p>
                )}
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-gray-400">{programInits.length} initiatives</span>
                <div className="w-24">
                  <ProgressBar progress={totalProgress} size="sm" />
                </div>
                <span className="text-sm text-gray-300">{Math.round(totalProgress * 100)}%</span>
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {programInits.map((init) => (
                <InitiativeTile
                  key={init.id}
                  initiative={init}
                  onClick={() => onInitiativeClick(init.id)}
                />
              ))}
            </div>
          </div>
        )
      })}

      {standalone.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-lg font-medium text-gray-300">Standalone</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {standalone.map((init) => (
              <InitiativeTile
                key={init.id}
                initiative={init}
                onClick={() => onInitiativeClick(init.id)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function InitiativeTile({
  initiative,
  onClick,
}: {
  initiative: APIInitiative
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className="w-full text-left p-4 rounded-lg border border-gray-700 bg-gray-800 hover:border-gray-600 hover:bg-gray-750 transition-colors"
    >
      <div className="flex items-center justify-between mb-2">
        <span className="font-mono text-xs text-gray-500">{initiative.id}</span>
        <StatusBadge status={initiative.status} size="sm" />
      </div>
      <h3 className="font-medium text-gray-100 mb-2 line-clamp-2">{initiative.title}</h3>
      <div className="flex items-center justify-between">
        <ProgressBar progress={initiative.progress} className="flex-1 mr-3" size="sm" />
        <span className="text-sm text-gray-400">{Math.round(initiative.progress * 100)}%</span>
      </div>
    </button>
  )
}
