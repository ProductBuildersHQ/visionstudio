import { useState, useEffect } from 'react'
import { SummaryCard, SeverityBadge, LoadingState, ErrorState, EmptyState } from '../../../components/toolkit'
import type { ExtensionViewProps } from '../../../types/extension'
import type { StyleProfile, StyleRule } from './types'

export function RulesView({ context }: ExtensionViewProps) {
  const [profile, setProfile] = useState<StyleProfile | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [expandedRules, setExpandedRules] = useState<Set<string>>(new Set())

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const data = await context.api.getProjectData<StyleProfile>('profile')
      setProfile(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load profile')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [context.projectName])

  if (isLoading) return <LoadingState message="Loading style profile..." />
  if (error) return <ErrorState message={error} onRetry={load} />
  if (!profile || profile.rules.length === 0) {
    return (
      <EmptyState
        icon="📖"
        title="No Rules Loaded"
        description="No style profile has been configured for this project."
      />
    )
  }

  const categories = Array.from(new Set(profile.rules.map(r => r.category))).sort()
  const filtered = selectedCategory
    ? profile.rules.filter(r => r.category === selectedCategory)
    : profile.rules

  const toggleRule = (id: string) => {
    const next = new Set(expandedRules)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    setExpandedRules(next)
  }

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-xl font-semibold text-va-text">{profile.name}</h1>
          {profile.description && (
            <p className="text-sm text-va-text-muted mt-0.5">{profile.description}</p>
          )}
        </div>

        {/* Summary */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          <SummaryCard label="Total Rules" value={profile.rules.length} />
          <SummaryCard label="Categories" value={categories.length} />
          <SummaryCard
            label="Version"
            value={profile.version ?? '—'}
          />
        </div>

        {/* Category filter */}
        <div className="flex flex-wrap gap-2 mb-6">
          <button
            onClick={() => setSelectedCategory(null)}
            className={`px-2.5 py-1 text-xs rounded ${!selectedCategory ? 'bg-va-accent text-white' : 'bg-va-panel text-va-text-muted hover:text-va-text border border-va-border'}`}
          >
            All ({profile.rules.length})
          </button>
          {categories.map(cat => {
            const count = profile.rules.filter(r => r.category === cat).length
            return (
              <button
                key={cat}
                onClick={() => setSelectedCategory(cat)}
                className={`px-2.5 py-1 text-xs rounded capitalize ${selectedCategory === cat ? 'bg-va-accent text-white' : 'bg-va-panel text-va-text-muted hover:text-va-text border border-va-border'}`}
              >
                {cat} ({count})
              </button>
            )
          })}
        </div>

        {/* Rules list */}
        <div className="space-y-2">
          {filtered.map(rule => (
            <div key={rule.id} className="bg-va-panel rounded-lg border border-va-border overflow-hidden">
              <button
                onClick={() => toggleRule(rule.id)}
                className="w-full px-4 py-3 flex items-center gap-3 hover:bg-va-border/30 transition-colors text-left"
              >
                <span className="text-xs text-va-text-muted">{expandedRules.has(rule.id) ? '▼' : '►'}</span>
                <SeverityBadge severity={rule.severity} />
                <span className="flex-1 text-sm font-medium text-va-text">{rule.title}</span>
                <span className="text-[10px] font-mono text-va-text-muted">{rule.id}</span>
              </button>
              {expandedRules.has(rule.id) && (
                <RuleDetail rule={rule} />
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function RuleDetail({ rule }: { rule: StyleRule }) {
  return (
    <div className="border-t border-va-border px-4 py-3 space-y-3">
      {rule.rationale && (
        <div>
          <span className="text-xs font-semibold text-va-text-muted uppercase">Rationale</span>
          <p className="text-sm text-va-text mt-1">{rule.rationale}</p>
        </div>
      )}
      {rule.examples?.good && rule.examples.good.length > 0 && (
        <div>
          <span className="text-xs font-semibold text-va-success uppercase">Good Examples</span>
          {rule.examples.good.map((ex, idx) => (
            <pre key={idx} className="mt-1 text-xs bg-va-bg p-2 rounded border border-va-border overflow-x-auto">
              {ex}
            </pre>
          ))}
        </div>
      )}
      {rule.examples?.bad && rule.examples.bad.length > 0 && (
        <div>
          <span className="text-xs font-semibold text-va-error uppercase">Bad Examples</span>
          {rule.examples.bad.map((ex, idx) => (
            <pre key={idx} className="mt-1 text-xs bg-va-bg p-2 rounded border border-va-border overflow-x-auto">
              {ex}
            </pre>
          ))}
        </div>
      )}
    </div>
  )
}
