import { useState, useEffect } from 'react'
import { getExecution } from '../api/client'
import type { ExecutionResponse, APIInitiative, APIPhase, APIRMI } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { ProgressBar } from '../components/ProgressBar'

export function ExecutionPanel() {
  const [data, setData] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedInitiative, setSelectedInitiative] = useState<string | null>(null)

  useEffect(() => {
    getExecution()
      .then(setData)
      .catch((err: Error) => setError(err.message))
  }, [])

  if (error) {
    return <div className="text-red-400">Error: {error}</div>
  }

  if (!data) {
    return <div className="text-gray-400">Loading...</div>
  }

  const selectedInit = data.initiatives.find((i) => i.id === selectedInitiative)
  const phases = selectedInit
    ? data.phases.filter((p) => p.initiativeId === selectedInit.id)
    : []
  const rmis = selectedInit
    ? data.rmis.filter((r) => r.initiativeId === selectedInit.id)
    : []

  return (
    <div className="grid grid-cols-3 gap-6">
      {/* Programs & Initiatives List */}
      <div className="col-span-1 space-y-4">
        <h2 className="text-lg font-semibold">Initiatives</h2>
        <div className="space-y-2">
          {data.programs.map((prog) => (
            <div key={prog.id} className="space-y-1">
              <h3 className="text-sm font-medium text-gray-400">{prog.name}</h3>
              {data.initiatives
                .filter((i) => i.programId === prog.id)
                .map((init) => (
                  <InitiativeCard
                    key={init.id}
                    initiative={init}
                    selected={selectedInitiative === init.id}
                    onClick={() => setSelectedInitiative(init.id)}
                  />
                ))}
            </div>
          ))}
          {/* Standalone initiatives */}
          {data.initiatives.filter((i) => !i.programId).length > 0 && (
            <div className="space-y-1">
              <h3 className="text-sm font-medium text-gray-400">Standalone</h3>
              {data.initiatives
                .filter((i) => !i.programId)
                .map((init) => (
                  <InitiativeCard
                    key={init.id}
                    initiative={init}
                    selected={selectedInitiative === init.id}
                    onClick={() => setSelectedInitiative(init.id)}
                  />
                ))}
            </div>
          )}
        </div>
      </div>

      {/* Detail Panel */}
      <div className="col-span-2">
        {selectedInit ? (
          <InitiativeDetail
            initiative={selectedInit}
            phases={phases}
            rmis={rmis}
          />
        ) : (
          <div className="text-gray-400">Select an initiative to view details</div>
        )}
      </div>
    </div>
  )
}

function InitiativeCard({
  initiative,
  selected,
  onClick,
}: {
  initiative: APIInitiative
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-3 rounded-lg border transition-colors ${
        selected
          ? 'border-blue-500 bg-blue-500/10'
          : 'border-gray-700 bg-gray-800 hover:border-gray-600'
      }`}
    >
      <div className="flex items-center justify-between">
        <span className="font-medium text-sm">{initiative.id}</span>
        <StatusBadge status={initiative.status} />
      </div>
      <div className="text-gray-400 text-xs mt-1 truncate">{initiative.title}</div>
      <ProgressBar progress={initiative.progress} className="mt-2" />
    </button>
  )
}

function InitiativeDetail({
  initiative,
  phases,
  rmis,
}: {
  initiative: APIInitiative
  phases: APIPhase[]
  rmis: APIRMI[]
}) {
  const sortedPhases = [...phases].sort((a, b) => a.sequenceNumber - b.sequenceNumber)

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold">{initiative.id}</h2>
        <p className="text-gray-400 mt-1">{initiative.title}</p>
        {initiative.description && (
          <p className="text-gray-500 text-sm mt-2">{initiative.description}</p>
        )}
        <div className="flex items-center gap-4 mt-3">
          <StatusBadge status={initiative.status} />
          <span className="text-sm text-gray-400">
            {Math.round(initiative.progress * 100)}% complete
          </span>
        </div>
      </div>

      <div className="space-y-4">
        {sortedPhases.map((phase) => {
          const phaseRmis = rmis
            .filter((r) => r.phaseId === phase.id)
            .sort((a, b) => a.sequenceNumber - b.sequenceNumber)

          return (
            <div key={phase.id} className="bg-gray-800 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-medium">{phase.title}</h3>
                <span className="text-sm text-gray-400">
                  {Math.round(phase.progress * 100)}%
                </span>
              </div>
              <ProgressBar progress={phase.progress} className="mb-3" />
              <div className="space-y-2">
                {phaseRmis.map((rmi) => (
                  <div
                    key={rmi.id}
                    className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-xs font-mono text-gray-500">{rmi.id}</span>
                      <span className="text-sm">{rmi.title}</span>
                    </div>
                    <StatusBadge status={rmi.status} size="sm" />
                  </div>
                ))}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
