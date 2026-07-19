import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import type { SampleSummary } from '../../services/api'

interface SamplePickerProps {
  selectedSampleId: string | null
  onSelectSample: (sample: SampleSummary | null) => void
}

/**
 * SamplePicker displays available samples for project initialization.
 * Allows users to select a sample to use as a template.
 */
export function SamplePicker({ selectedSampleId, onSelectSample }: SamplePickerProps) {
  const [samples, setSamples] = useState<SampleSummary[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadSamples()
  }, [])

  async function loadSamples() {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.listSamples()
      setSamples(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load samples')
    } finally {
      setIsLoading(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-va-text-muted text-sm py-2">
        <div className="animate-spin w-4 h-4 border-2 border-va-accent border-t-transparent rounded-full" />
        <span>Loading samples...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-sm text-va-text-muted py-2">
        No samples available
      </div>
    )
  }

  if (samples.length === 0) {
    return (
      <div className="text-sm text-va-text-muted py-2">
        No samples configured
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {/* None option */}
      <SampleCard
        sample={null}
        isSelected={selectedSampleId === null}
        onSelect={() => onSelectSample(null)}
      />

      {/* Sample cards */}
      {samples.map((sample) => (
        <SampleCard
          key={sample.id}
          sample={sample}
          isSelected={selectedSampleId === sample.id}
          onSelect={() => onSelectSample(sample)}
        />
      ))}
    </div>
  )
}

interface SampleCardProps {
  sample: SampleSummary | null
  isSelected: boolean
  onSelect: () => void
}

function SampleCard({ sample, isSelected, onSelect }: SampleCardProps) {
  if (!sample) {
    return (
      <button
        type="button"
        onClick={onSelect}
        className={`w-full p-3 rounded border text-left transition-colors ${
          isSelected
            ? 'border-va-accent bg-va-accent/10'
            : 'border-va-border bg-va-panel hover:border-va-text-muted'
        }`}
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded bg-va-sidebar flex items-center justify-center text-xl">
            📁
          </div>
          <div>
            <div className="font-medium text-va-text">Start from scratch</div>
            <div className="text-xs text-va-text-muted">
              Initialize with empty spec templates
            </div>
          </div>
          {isSelected && (
            <div className="ml-auto">
              <CheckIcon className="w-5 h-5 text-va-accent" />
            </div>
          )}
        </div>
      </button>
    )
  }

  const complexityBadge = sample.complexity === 'enterprise'
    ? { label: 'Enterprise', color: 'bg-purple-500/20 text-purple-400' }
    : { label: 'Simple', color: 'bg-green-500/20 text-green-400' }

  const totalFiles = Object.values(sample.fileCounts).reduce((a, b) => a + b, 0)

  return (
    <button
      type="button"
      onClick={onSelect}
      className={`w-full p-3 rounded border text-left transition-colors ${
        isSelected
          ? 'border-va-accent bg-va-accent/10'
          : 'border-va-border bg-va-panel hover:border-va-text-muted'
      }`}
    >
      <div className="flex items-start gap-3">
        <div className="w-10 h-10 rounded bg-va-sidebar flex items-center justify-center text-xl flex-shrink-0">
          {sample.complexity === 'enterprise' ? '🏢' : '📦'}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-va-text">{sample.name}</span>
            <span className={`text-xs px-1.5 py-0.5 rounded ${complexityBadge.color}`}>
              {complexityBadge.label}
            </span>
          </div>
          <div className="text-xs text-va-text-muted mt-0.5 line-clamp-2">
            {sample.description}
          </div>
          <div className="flex items-center gap-3 mt-2 text-xs text-va-text-muted">
            <span>{totalFiles} files</span>
            {sample.fileCounts.v2mom && (
              <span>V2MOM: {sample.fileCounts.v2mom}</span>
            )}
            {sample.fileCounts.capability && (
              <span>Capabilities: {sample.fileCounts.capability}</span>
            )}
            {sample.fileCounts.roadmap && (
              <span>Roadmap: {sample.fileCounts.roadmap}</span>
            )}
          </div>
        </div>
        {isSelected && (
          <div className="flex-shrink-0">
            <CheckIcon className="w-5 h-5 text-va-accent" />
          </div>
        )}
      </div>
    </button>
  )
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  )
}

export default SamplePicker
