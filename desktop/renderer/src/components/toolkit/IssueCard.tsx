import { SeverityBadge, getSeverityStyle } from './SeverityBadge'

interface Issue {
  severity?: string
  category?: string
  code?: string
  message?: string
  location?: string
  suggestion?: string
  title?: string
  description?: string
  recommendation?: string
}

interface IssueCardProps {
  issue: Issue
}

export function IssueCard({ issue }: IssueCardProps) {
  const style = getSeverityStyle(issue.severity)

  return (
    <div className={`px-4 py-3 border-l-4 ${style.border} bg-va-panel/50`}>
      <div className="flex items-start gap-3">
        {issue.severity && <SeverityBadge severity={issue.severity} />}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            {issue.category && (
              <span className="text-xs text-va-text-muted capitalize">{issue.category}</span>
            )}
            {issue.code && (
              <span className="text-[10px] px-1.5 py-0.5 bg-va-panel border border-va-border rounded font-mono">
                {issue.code}
              </span>
            )}
            {issue.location && (
              <span className="text-[10px] text-va-text-muted">@ {issue.location}</span>
            )}
          </div>
          <p className="text-sm text-va-text mt-0.5 font-medium">
            {issue.title || issue.message}
          </p>
          {issue.description && (
            <p className="text-sm text-va-text-muted mt-0.5">{issue.description}</p>
          )}
          {(issue.recommendation || issue.suggestion) && (
            <p className="text-xs text-va-accent mt-1">
              {issue.recommendation || issue.suggestion}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
