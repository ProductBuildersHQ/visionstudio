import { useState, useEffect, useCallback, useRef, useMemo } from 'react'
import type { ExtensionViewProps } from '../../../types/extension'
import type { Violation, LintSummary, LintResponse, ProfileInfo, FixSuggestion } from './types'

const API_BASE = 'http://127.0.0.1:8765/api/extensions/api-style-spec'

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!response.ok) {
    const body = await response.text().catch(() => '')
    throw new Error(body || `HTTP ${response.status}`)
  }
  return response.json()
}

export function EditorView(_props: ExtensionViewProps) {
  const [specContent, setSpecContent] = useState(SAMPLE_SPEC)
  const [violations, setViolations] = useState<Violation[]>([])
  const [summary, setSummary] = useState<LintSummary | null>(null)
  const [lintStatus, setLintStatus] = useState<'idle' | 'linting' | 'error'>('idle')
  const [lintError, setLintError] = useState<string | null>(null)
  const [profile, setProfile] = useState('default')
  const [profiles, setProfiles] = useState<ProfileInfo[]>([])
  const [selectedViolation, setSelectedViolation] = useState<Violation | null>(null)
  const [fixSuggestions, setFixSuggestions] = useState<FixSuggestion[]>([])
  const [isFixing, setIsFixing] = useState(false)
  const [showFixPanel, setShowFixPanel] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const editorRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    fetchAPI<ProfileInfo[]>('/profiles').then(setProfiles).catch(() => {})
  }, [])

  const runLint = useCallback(async (content: string, profileName: string) => {
    if (!content.trim()) {
      setViolations([])
      setSummary(null)
      setLintStatus('idle')
      return
    }
    setLintStatus('linting')
    setLintError(null)
    try {
      const result = await fetchAPI<LintResponse>('/lint', {
        method: 'POST',
        body: JSON.stringify({ spec: content, profile: profileName }),
      })
      setViolations(result.violations ?? [])
      setSummary(result.summary ?? null)
      setLintStatus('idle')
    } catch (err) {
      setLintError(err instanceof Error ? err.message : 'Lint failed')
      setLintStatus('error')
    }
  }, [])

  // Live linting: debounce on content/profile changes
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => runLint(specContent, profile), 800)
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current) }
  }, [specContent, profile, runLint])

  const handleSuggestFixes = useCallback(async () => {
    setIsFixing(true)
    setShowFixPanel(true)
    try {
      const result = await fetchAPI<{ suggestions?: FixSuggestion[] }>('/suggest-fixes', {
        method: 'POST',
        body: JSON.stringify({ spec: specContent, profile }),
      })
      setFixSuggestions(result.suggestions ?? [])
    } catch {
      setFixSuggestions([])
    } finally {
      setIsFixing(false)
    }
  }, [specContent, profile])

  const handleApplyFix = useCallback((fix: FixSuggestion) => {
    if (!fix.suggestedValue || !fix.path) return
    // Simple replacement: find currentValue in content and replace with suggestedValue
    if (fix.currentValue && specContent.includes(fix.currentValue)) {
      setSpecContent(specContent.replace(fix.currentValue, fix.suggestedValue))
    }
  }, [specContent])

  const goToLine = useCallback((line: number) => {
    const editor = editorRef.current
    if (!editor) return
    const lines = editor.value.split('\n')
    let pos = 0
    for (let i = 0; i < Math.min(line - 1, lines.length); i++) {
      pos += lines[i].length + 1
    }
    editor.focus()
    editor.setSelectionRange(pos, pos + (lines[line - 1]?.length ?? 0))
  }, [])

  // Group violations by line for gutter markers
  const violationsByLine = useMemo(() => {
    const map = new Map<number, Violation[]>()
    for (const v of violations) {
      if (v.line) {
        const list = map.get(v.line) ?? []
        list.push(v)
        map.set(v.line, list)
      }
    }
    return map
  }, [violations])

  const lineCount = specContent.split('\n').length

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 bg-va-sidebar border-b border-va-border shrink-0">
        <div className="flex items-center gap-3">
          <h2 className="text-sm font-semibold text-va-text">API Style Lint Editor</h2>
          <LintStatusBadge status={lintStatus} error={lintError} />
        </div>
        <div className="flex items-center gap-2">
          <select
            value={profile}
            onChange={e => setProfile(e.target.value)}
            className="px-2 py-1 text-xs bg-va-panel border border-va-border rounded text-va-text"
          >
            {profiles.length > 0 ? (
              profiles.map(p => (
                <option key={p.name} value={p.name}>{p.name}</option>
              ))
            ) : (
              <option value="default">default</option>
            )}
          </select>
          <button
            onClick={() => runLint(specContent, profile)}
            disabled={lintStatus === 'linting'}
            className="px-3 py-1 text-xs bg-va-accent text-white rounded hover:bg-va-accent/80 disabled:opacity-50"
          >
            Lint
          </button>
          <button
            onClick={handleSuggestFixes}
            disabled={isFixing || violations.length === 0}
            className="px-3 py-1 text-xs bg-va-panel border border-va-border text-va-text rounded hover:bg-va-accent hover:text-white disabled:opacity-50 transition-colors"
          >
            {isFixing ? 'Generating...' : 'Suggest Fixes'}
          </button>
        </div>
      </div>

      {/* Summary bar */}
      {summary && (
        <div className="flex items-center gap-4 px-4 py-1.5 bg-va-panel border-b border-va-border text-[11px] shrink-0">
          <span className="text-va-text-muted">Violations:</span>
          {summary.errors > 0 && <span className="text-red-400">{summary.errors} errors</span>}
          {summary.warnings > 0 && <span className="text-orange-400">{summary.warnings} warnings</span>}
          {summary.infos > 0 && <span className="text-blue-400">{summary.infos} info</span>}
          {summary.hints > 0 && <span className="text-va-text-muted">{summary.hints} hints</span>}
          <span className="text-va-text-muted ml-auto">{summary.total} total</span>
        </div>
      )}

      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Editor with gutter */}
        <div className="flex-1 flex overflow-hidden">
          {/* Line numbers + markers */}
          <div className="w-14 bg-va-sidebar border-r border-va-border overflow-y-auto text-right select-none shrink-0 gutter-scroll" style={{ overflowX: 'hidden' }}>
            {Array.from({ length: lineCount }, (_, i) => {
              const lineNum = i + 1
              const lineViolations = violationsByLine.get(lineNum)
              const worstSeverity = lineViolations
                ? getWorstSeverity(lineViolations)
                : null
              return (
                <div
                  key={lineNum}
                  className={`px-1 text-[11px] leading-[1.375rem] font-mono cursor-pointer hover:bg-va-panel ${
                    lineViolations ? 'text-va-text' : 'text-va-text-muted'
                  }`}
                  onClick={() => {
                    if (lineViolations) setSelectedViolation(lineViolations[0])
                    goToLine(lineNum)
                  }}
                  title={lineViolations?.map(v => v.message).join('\n')}
                >
                  {worstSeverity && <SeverityMarker severity={worstSeverity} />}
                  {lineNum}
                </div>
              )
            })}
          </div>

          {/* Textarea editor */}
          <textarea
            ref={editorRef}
            value={specContent}
            onChange={e => setSpecContent(e.target.value)}
            className="flex-1 p-2 bg-va-bg text-va-text font-mono text-[13px] leading-[1.375rem] resize-none focus:outline-none overflow-auto"
            spellCheck={false}
            wrap="off"
          />
        </div>

        {/* Violations panel */}
        <div className={`border-l border-va-border overflow-hidden flex flex-col shrink-0 ${
          showFixPanel ? 'w-[420px]' : 'w-80'
        }`}>
          {showFixPanel ? (
            <FixPanel
              fixes={fixSuggestions}
              isLoading={isFixing}
              onApply={handleApplyFix}
              onClose={() => setShowFixPanel(false)}
            />
          ) : (
            <ViolationsPanel
              violations={violations}
              selected={selectedViolation}
              onSelect={v => { setSelectedViolation(v); if (v.line) goToLine(v.line) }}
            />
          )}
        </div>
      </div>
    </div>
  )
}

function LintStatusBadge({ status, error }: { status: string; error: string | null }) {
  if (status === 'linting') {
    return (
      <span className="flex items-center gap-1 text-[10px] text-va-accent">
        <span className="w-2 h-2 rounded-full bg-va-accent animate-pulse" />
        Linting...
      </span>
    )
  }
  if (status === 'error') {
    return (
      <span className="text-[10px] text-va-error" title={error ?? undefined}>
        Lint error
      </span>
    )
  }
  return null
}

function SeverityMarker({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    error: 'bg-red-500',
    warn: 'bg-orange-400',
    info: 'bg-blue-400',
    hint: 'bg-va-text-muted',
  }
  return (
    <span className={`inline-block w-1.5 h-1.5 rounded-full mr-0.5 ${colors[severity] ?? colors.hint}`} />
  )
}

function getWorstSeverity(violations: Violation[]): string {
  const order = ['error', 'warn', 'info', 'hint']
  let worst = 'hint'
  for (const v of violations) {
    const idx = order.indexOf(v.severity)
    if (idx >= 0 && idx < order.indexOf(worst)) worst = v.severity
  }
  return worst
}

function ViolationsPanel({
  violations,
  selected,
  onSelect,
}: {
  violations: Violation[]
  selected: Violation | null
  onSelect: (v: Violation) => void
}) {
  const grouped = useMemo(() => {
    const map = new Map<string, Violation[]>()
    for (const v of violations) {
      const key = v.severity
      const list = map.get(key) ?? []
      list.push(v)
      map.set(key, list)
    }
    return map
  }, [violations])

  const severityOrder = ['error', 'warn', 'info', 'hint']
  const severityLabels: Record<string, string> = {
    error: 'Errors',
    warn: 'Warnings',
    info: 'Info',
    hint: 'Hints',
  }
  const severityColors: Record<string, string> = {
    error: 'text-red-400',
    warn: 'text-orange-400',
    info: 'text-blue-400',
    hint: 'text-va-text-muted',
  }

  return (
    <div className="h-full flex flex-col">
      <div className="px-3 py-2 border-b border-va-border shrink-0">
        <span className="text-xs font-semibold text-va-text-muted uppercase tracking-wider">
          Violations ({violations.length})
        </span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {violations.length === 0 ? (
          <div className="p-4 text-xs text-va-text-muted text-center">
            No violations found
          </div>
        ) : (
          severityOrder.map(sev => {
            const items = grouped.get(sev)
            if (!items || items.length === 0) return null
            return (
              <div key={sev}>
                <div className={`px-3 py-1 text-[10px] font-semibold uppercase tracking-wider ${severityColors[sev]}`}>
                  {severityLabels[sev]} ({items.length})
                </div>
                {items.map((v, i) => (
                  <button
                    key={i}
                    onClick={() => onSelect(v)}
                    className={`w-full text-left px-3 py-2 text-xs border-b border-va-border hover:bg-va-panel transition-colors ${
                      selected === v ? 'bg-va-panel' : ''
                    }`}
                  >
                    <div className="flex items-center gap-1.5 mb-0.5">
                      <SeverityMarker severity={v.severity} />
                      <span className="text-va-text-muted font-mono text-[10px]">{v.ruleId}</span>
                      {v.line && (
                        <span className="text-va-text-muted text-[10px] ml-auto">L{v.line}</span>
                      )}
                    </div>
                    <p className="text-va-text leading-snug">{v.message}</p>
                    {v.path && (
                      <p className="text-va-text-muted text-[10px] mt-0.5 font-mono truncate">{v.path}</p>
                    )}
                  </button>
                ))}
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}

function FixPanel({
  fixes,
  isLoading,
  onApply,
  onClose,
}: {
  fixes: FixSuggestion[]
  isLoading: boolean
  onApply: (fix: FixSuggestion) => void
  onClose: () => void
}) {
  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-3 py-2 border-b border-va-border shrink-0">
        <span className="text-xs font-semibold text-va-text-muted uppercase tracking-wider">
          Fix Suggestions ({fixes.length})
        </span>
        <button onClick={onClose} className="text-va-text-muted hover:text-va-text text-xs">
          Close
        </button>
      </div>
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="p-4 text-xs text-va-text-muted text-center">
            <div className="w-5 h-5 border-2 border-va-accent border-t-transparent rounded-full animate-spin mx-auto mb-2" />
            Generating fix suggestions...
          </div>
        ) : fixes.length === 0 ? (
          <div className="p-4 text-xs text-va-text-muted text-center">
            No fix suggestions available
          </div>
        ) : (
          fixes.map((fix, i) => (
            <div key={i} className="px-3 py-3 border-b border-va-border">
              <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] font-mono text-va-text-muted">{fix.ruleId}</span>
                <div className="flex items-center gap-2">
                  {fix.breaking && (
                    <span className="text-[10px] px-1 py-0.5 bg-red-500/20 text-red-400 rounded">
                      Breaking
                    </span>
                  )}
                  <span className="text-[10px] text-va-text-muted">
                    {Math.round(fix.confidence * 100)}%
                  </span>
                </div>
              </div>
              <p className="text-xs text-va-text-muted mb-1.5 font-mono truncate" title={fix.path}>
                {fix.path}
              </p>
              {fix.reasoning && (
                <p className="text-xs text-va-text mb-2">{fix.reasoning}</p>
              )}
              {fix.diff && (
                <pre className="text-[10px] font-mono bg-va-bg rounded p-2 mb-2 overflow-x-auto max-h-32 text-va-text">
                  {fix.diff}
                </pre>
              )}
              {fix.suggestedValue && (
                <div className="flex gap-1">
                  <button
                    onClick={() => onApply(fix)}
                    className="px-2 py-1 text-[10px] bg-va-accent text-white rounded hover:bg-va-accent/80"
                  >
                    Apply
                  </button>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

const SAMPLE_SPEC = `openapi: "3.0.3"
info:
  title: Pet Store API
  version: "1.0.0"
  description: A sample API for managing pets
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            format: int32
      responses:
        "200":
          description: A list of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Pet"
    post:
      summary: Create a pet
      operationId: createPet
      tags:
        - pets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Pet"
      responses:
        "201":
          description: Pet created
  /pets/{petId}:
    get:
      summary: Info for a specific pet
      operationId: showPetById
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
`
