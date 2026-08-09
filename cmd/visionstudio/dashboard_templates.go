package main

const dashboardCSS = `
  :root {
    --bg: #0f172a;
    --surface: #1e293b;
    --surface2: #334155;
    --border: #475569;
    --text: #f1f5f9;
    --text2: #94a3b8;
    --accent: #3b82f6;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    background: var(--bg);
    color: var(--text);
    line-height: 1.5;
    padding: 24px;
  }
  a { color: var(--accent); text-decoration: none; }
  a:hover { text-decoration: underline; }
  h1 {
    font-size: 1.5rem;
    font-weight: 600;
    margin-bottom: 8px;
  }
  .subtitle {
    color: var(--text2);
    font-size: 0.875rem;
    margin-bottom: 32px;
  }
  .breadcrumb {
    font-size: 0.85rem;
    color: var(--text2);
    margin-bottom: 16px;
  }
  .breadcrumb a { color: var(--accent); }
  .stats-bar {
    display: flex;
    gap: 24px;
    margin-bottom: 24px;
    flex-wrap: wrap;
  }
  .stat-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 16px 20px;
    min-width: 140px;
  }
  .stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--accent);
  }
  .stat-label {
    font-size: 0.75rem;
    color: var(--text2);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .badge {
    display: inline-block;
    padding: 2px 10px;
    border-radius: 9999px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
    color: #fff;
  }
  .status-tooltip {
    cursor: default;
    position: relative;
  }
  .status-tooltip[title]:hover::after {
    content: attr(title);
    position: absolute;
    top: calc(100% + 6px);
    left: 50%;
    transform: translateX(-50%);
    background: #1e293b;
    color: #f1f5f9;
    border: 1px solid #475569;
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 0.72rem;
    font-weight: 400;
    text-transform: none;
    letter-spacing: normal;
    white-space: pre;
    z-index: 100;
    pointer-events: none;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  }

  /* --- summary-specific --- */
  .status-chart {
    display: flex;
    gap: 4px;
    align-items: flex-end;
    height: 80px;
    margin-bottom: 4px;
  }
  .status-bar {
    flex: 1;
    min-width: 40px;
    max-width: 100px;
    border-radius: 4px 4px 0 0;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    color: #fff;
    font-size: 0.7rem;
    font-weight: 700;
    padding-bottom: 4px;
    min-height: 8px;
  }
  .status-labels {
    display: flex;
    gap: 4px;
  }
  .status-label {
    flex: 1;
    min-width: 40px;
    max-width: 100px;
    text-align: center;
    font-size: 0.6rem;
    color: var(--text2);
    text-transform: uppercase;
  }
  .chart-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 20px 24px;
    margin-bottom: 32px;
  }
  .chart-card h2 {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 1.125rem;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text);
  }
  .program-pill {
    display: inline-block;
    padding: 2px 10px;
    background: var(--surface2);
    border: 1px solid var(--accent);
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--accent);
    margin-left: 8px;
  }
  .hidden-toggle a {
    display: inline-block;
    padding: 6px 14px;
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 8px;
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--text);
    text-decoration: none;
  }
  .hidden-toggle a:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  /* initiative cards */
  .init-cards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
    gap: 16px;
    margin-bottom: 32px;
  }
  .init-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    transition: border-color 0.15s;
  }
  .init-card:hover { border-color: var(--accent); }
  .init-card-header {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .init-card-title {
    font-size: 0.95rem;
    font-weight: 600;
    flex: 1;
  }
  .init-card-title a { color: var(--text); }
  .init-card-title a:hover { color: var(--accent); text-decoration: none; }
  .init-card-id {
    font-family: 'SF Mono', 'Fira Code', monospace;
    font-size: 0.72rem;
    color: var(--text2);
  }
  .init-card-meta {
    font-size: 0.78rem;
    color: var(--text2);
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }
  .progress-bar-bg {
    width: 100%;
    height: 8px;
    background: var(--surface2);
    border-radius: 4px;
    overflow: hidden;
  }
  .progress-bar-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.3s;
  }
  .progress-text {
    font-size: 0.72rem;
    color: var(--text2);
    text-align: right;
  }
  .repo-pills {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .repo-pill {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 9999px;
    font-size: 0.68rem;
    color: var(--text);
  }
  .repo-pill-count {
    background: var(--accent);
    color: #fff;
    font-size: 0.6rem;
    font-weight: 700;
    padding: 0 5px;
    border-radius: 9999px;
    min-width: 16px;
    text-align: center;
  }
  .repo-pill-label {
    font-family: 'SF Mono', 'Fira Code', monospace;
  }

  /* --- detail-specific --- */
  .initiative {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    margin-bottom: 32px;
    overflow: hidden;
  }
  .init-header {
    padding: 20px 24px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }
  .init-header h2 {
    font-size: 1.125rem;
    font-weight: 600;
    flex: 1;
    min-width: 200px;
  }
  .init-meta {
    display: flex;
    gap: 16px;
    font-size: 0.8rem;
    color: var(--text2);
  }
  .workspace-pill {
    display: inline-block;
    padding: 2px 12px;
    background: var(--surface2);
    border: 1px solid var(--accent);
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    font-family: 'SF Mono', 'Fira Code', monospace;
    color: var(--accent);
    vertical-align: middle;
    margin-left: 8px;
  }
  .phase {
    border-bottom: 1px solid var(--border);
  }
  .phase:last-child { border-bottom: none; }
  .phase-header {
    padding: 14px 24px;
    background: var(--surface2);
    display: flex;
    align-items: center;
    gap: 12px;
    cursor: pointer;
    user-select: none;
  }
  .phase-header:hover { background: #3b4d66; }
  .phase-title {
    font-size: 0.95rem;
    font-weight: 500;
    flex: 1;
  }
  .phase-seq {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--accent);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 700;
    flex-shrink: 0;
  }
  .phase-count {
    font-size: 0.8rem;
    color: var(--text2);
  }
  .rmi-list { padding: 0; }
  .rmi-row {
    display: grid;
    grid-template-columns: 80px 180px 1fr 90px 80px 80px;
    gap: 12px;
    padding: 10px 24px 10px 64px;
    align-items: center;
    border-top: 1px solid rgba(71,85,105,0.4);
    font-size: 0.85rem;
  }
  .rmi-row:hover { background: rgba(59,130,246,0.05); }
  .rmi-id {
    font-family: 'SF Mono', 'Fira Code', monospace;
    font-size: 0.78rem;
    color: var(--accent);
    white-space: nowrap;
  }
  .rmi-title { color: var(--text); }
  .rmi-repo { font-size: 0.75rem; color: var(--text2); text-align: center; }
  .rmi-type { font-size: 0.75rem; color: var(--text2); text-align: center; }
  .rmi-required { text-align: center; font-size: 0.75rem; }
  .required-yes { color: #f59e0b; }
  .required-no { color: var(--text2); }
  .toggle-arrow {
    transition: transform 0.2s;
    font-size: 0.7rem;
    color: var(--text2);
  }
  .phase.collapsed .toggle-arrow { transform: rotate(-90deg); }
  .phase.collapsed .rmi-list { display: none; }
  .rmi-header-row {
    display: flex;
    align-items: center;
    padding: 8px 24px;
    font-size: 0.7rem;
    color: var(--text2);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid var(--border);
  }
  .rmi-header-toggle {
    width: 40px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
  }
  .rmi-header-cols {
    flex: 1;
    display: grid;
    grid-template-columns: 80px 180px 1fr 90px 80px 80px;
    gap: 12px;
  }
  .rmi-header-cols span:nth-child(4),
  .rmi-header-cols span:nth-child(5),
  .rmi-header-cols span:nth-child(6) { text-align: center; }
  .toggle-all-arrow {
    font-size: 0.7rem;
    cursor: pointer;
    transition: transform 0.2s;
    color: var(--text2);
  }
  .toggle-all-arrow:hover { color: var(--accent); }
  .toggle-all-arrow.all-collapsed { transform: rotate(-90deg); }

  .dep-section {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 20px 24px;
    margin-bottom: 32px;
  }
  .dep-section h2 {
    font-size: 1.125rem;
    font-weight: 600;
    margin-bottom: 16px;
  }
  .dep-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 8px;
  }
  .dep-edge {
    font-family: 'SF Mono', 'Fira Code', monospace;
    font-size: 0.78rem;
    padding: 6px 12px;
    background: var(--surface2);
    border-radius: 6px;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .dep-arrow { color: var(--accent); }
  .dep-rel {
    font-size: 0.65rem;
    color: var(--text2);
    text-transform: uppercase;
  }

  /* token/cost display */
  .token-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    background: var(--surface2);
    border: 1px solid #10b981;
    border-radius: 9999px;
    font-size: 0.7rem;
    color: #10b981;
    font-family: 'SF Mono', 'Fira Code', monospace;
  }
  .cost-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    background: var(--surface2);
    border: 1px solid #f59e0b;
    border-radius: 9999px;
    font-size: 0.7rem;
    color: #f59e0b;
    font-family: 'SF Mono', 'Fira Code', monospace;
  }
  .token-section {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }
  .token-note {
    font-size: 0.72rem;
    color: var(--text2);
    font-style: italic;
  }
  .description {
    font-size: 0.85rem;
    color: var(--text2);
    line-height: 1.5;
    margin-bottom: 16px;
  }
  .init-description {
    font-size: 0.8rem;
    color: var(--text2);
    line-height: 1.4;
    margin-top: 4px;
  }
`

const dashboardJS = `
var STORAGE_KEY = 'prism-dashboard-collapsed';

function getCollapsedState() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY)) || {};
  } catch(e) { return {}; }
}

function saveCollapsedState() {
  var state = {};
  document.querySelectorAll('.phase').forEach(function(p) {
    var id = p.getAttribute('data-phase-id');
    if (id) state[id] = p.classList.contains('collapsed');
  });
  try { localStorage.setItem(STORAGE_KEY, JSON.stringify(state)); } catch(e) {}
}

function togglePhase(el) {
  el.classList.toggle('collapsed');
  saveCollapsedState();
}

function toggleAllPhases(arrow) {
  var init = arrow.closest('.initiative');
  var phases = init.querySelectorAll('.phase');
  var allCollapsed = arrow.classList.contains('all-collapsed');
  phases.forEach(function(p) {
    if (allCollapsed) {
      p.classList.remove('collapsed');
    } else {
      p.classList.add('collapsed');
    }
  });
  arrow.classList.toggle('all-collapsed');
  saveCollapsedState();
}

document.addEventListener('DOMContentLoaded', function() {
  var saved = getCollapsedState();
  var hasSaved = Object.keys(saved).length > 0;

  document.querySelectorAll('.phase').forEach(function(p) {
    var id = p.getAttribute('data-phase-id');
    if (hasSaved && id && saved[id] !== undefined) {
      if (saved[id]) p.classList.add('collapsed');
      else p.classList.remove('collapsed');
    }
  });

  if (!hasSaved) {
    var inits = document.querySelectorAll('.initiative');
    if (inits.length > 1) {
      inits.forEach(function(init, idx) {
        if (idx > 0) {
          init.querySelectorAll('.phase').forEach(function(p) {
            p.classList.add('collapsed');
          });
        }
      });
    }
    saveCollapsedState();
  }
});
`

// summaryHTML is the landing page with initiative cards grouped by program.
const summaryHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="5">
<title>PRISM Dashboard</title>
<style>` + dashboardCSS + `</style>
</head>
<body>

<h1>PRISM Dashboard</h1>
<p class="subtitle">Product Delivery Control Plane — initiative summary</p>

{{- $totalRMIs := 0 }}
{{- $totalPhases := 0 }}
{{- $totalCompleted := 0 }}
{{- range .Initiatives }}
  {{- $totalRMIs = (add $totalRMIs .TotalRMIs) }}
  {{- $totalCompleted = (add $totalCompleted .CompletedRMIs) }}
  {{- range .Phases }}
    {{- $totalPhases = (add $totalPhases 1) }}
  {{- end }}
{{- end }}

<div class="stats-bar">
  <div class="stat-card">
    <div class="stat-value">{{ len .Initiatives }}</div>
    <div class="stat-label">Initiatives</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ len .VisiblePrograms }}</div>
    <div class="stat-label">Programs</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ $totalPhases }}</div>
    <div class="stat-label">Phases</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ $totalRMIs }}</div>
    <div class="stat-label">RMIs</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ $totalCompleted }}/{{ $totalRMIs }}</div>
    <div class="stat-label">Completed</div>
  </div>
  {{- if .TotalTokens }}
  <div class="stat-card">
    <div class="stat-value" style="color:#10b981">{{ formatTokens .TotalTokens.TotalTokens }}</div>
    <div class="stat-label">Total Tokens</div>
  </div>
  <div class="stat-card">
    <div class="stat-value" style="color:#f59e0b">{{ formatCost .TotalTokens.CostUSD }}</div>
    <div class="stat-label">Total Cost</div>
  </div>
  {{- end }}
</div>

{{- if .TokenDataNote }}
<p class="token-note" style="margin-bottom:16px">{{ .TokenDataNote }}</p>
{{- end }}

{{- if .StatusDist }}
<div class="chart-card">
  <h2>RMI Status Distribution</h2>
  {{- $maxCount := 0 }}
  {{- range .StatusDist }}
    {{- if gt .Count $maxCount }}{{ $maxCount = .Count }}{{ end }}
  {{- end }}
  <div class="status-chart">
    {{- range .StatusDist }}
    <div class="status-bar" style="background:{{ statusColor .Status }};height:{{ if gt $maxCount 0 }}{{ pct .Count $maxCount }}%{{ else }}0%{{ end }}">{{ .Count }}</div>
    {{- end }}
  </div>
  <div class="status-labels">
    {{- range .StatusDist }}
    <div class="status-label">{{ displayStatus .Status }}</div>
    {{- end }}
  </div>
</div>
{{- end }}

{{- $hiddenCount := .HiddenProgramCount }}
{{- if or $hiddenCount .ShowHidden }}
<div class="hidden-toggle" style="margin-bottom:20px">
  {{- if .ShowHidden }}
  <a href="/">&#128065; Hiding {{ $hiddenCount }} hidden program{{ if ne $hiddenCount 1 }}s{{ end }} — click to hide</a>
  {{- else }}
  <a href="/?show_hidden=1">&#128065; Show {{ $hiddenCount }} hidden program{{ if ne $hiddenCount 1 }}s{{ end }}</a>
  {{- end }}
</div>
{{- end }}

{{- range .VisiblePrograms }}
<div style="margin-bottom:32px">
  <h2 class="section-title">
    <a href="/program/{{ .ID }}">{{ .Name }}</a>
    <span class="program-pill">{{ len .Initiatives }} initiatives</span>
    <span class="init-card-id" style="margin-left:8px;vertical-align:middle">{{ .ID }}</span>
    {{- if .Hidden }}
    <span class="program-pill" style="background:#ef4444;color:#fff;margin-left:8px">hidden</span>
    {{- end }}
    {{- if .Tokens }}
    <span class="token-badge" style="margin-left:8px">{{ formatTokens .Tokens.TotalTokens }}</span>
    <span class="cost-badge">{{ formatCost .Tokens.CostUSD }}</span>
    {{- end }}
  </h2>
  <div class="init-cards">
    {{- range .Initiatives }}
    <div class="init-card">
      <div class="init-card-header">
        <div class="init-card-title"><a href="/initiative/{{ .Initiative.ID }}">{{ .Initiative.Title }}</a></div>
        <span class="badge" style="background:{{ statusColor .Initiative.Status }}">{{ displayStatus .Initiative.Status }}</span>
      </div>
      <div class="init-card-id">{{ .Initiative.ID }}</div>
      {{- if .Initiative.Description }}
      <div class="init-description">{{ .Initiative.Description }}</div>
      {{- end }}
      <div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill" style="width:{{ pct .CompletedRMIs .TotalRMIs }}%;background:{{ statusColor "completed" }}"></div>
        </div>
        <div class="progress-text">{{ .CompletedRMIs }}/{{ .TotalRMIs }} RMIs completed ({{ pct .CompletedRMIs .TotalRMIs }}%)</div>
      </div>
      <div class="init-card-meta">
        {{- if .Initiative.Priority }}<span>Priority: {{ .Initiative.Priority }}</span>{{ end }}
        {{- if .Initiative.HomeRepo }}<span>Home: {{ shortRepo .Initiative.HomeRepo }}</span>{{ end }}
        {{- if .Initiative.Workspace }}<span class="workspace-pill">{{ .Initiative.Workspace }}</span>{{ end }}
      </div>
      {{- if .Tokens }}
      <div class="token-section">
        <span class="token-badge">{{ formatTokens .Tokens.TotalTokens }} tokens</span>
        <span class="cost-badge">{{ formatCost .Tokens.CostUSD }}</span>
      </div>
      {{- end }}
      {{- if .Repos }}
      <div class="repo-pills">
        {{- range .Repos }}
        <span class="repo-pill">
          <span class="repo-pill-label">{{ .Name }}</span>
          <span class="repo-pill-count">{{ .Count }}</span>
        </span>
        {{- end }}
      </div>
      {{- end }}
    </div>
    {{- end }}
  </div>
</div>
{{- end }}

{{- if .Standalone }}
<div style="margin-bottom:32px">
  <h2 class="section-title">Standalone Initiatives</h2>
  <div class="init-cards">
    {{- range .Standalone }}
    <div class="init-card">
      <div class="init-card-header">
        <div class="init-card-title"><a href="/initiative/{{ .Initiative.ID }}">{{ .Initiative.Title }}</a></div>
        <span class="badge" style="background:{{ statusColor .Initiative.Status }}">{{ displayStatus .Initiative.Status }}</span>
      </div>
      <div class="init-card-id">{{ .Initiative.ID }}</div>
      {{- if .Initiative.Description }}
      <div class="init-description">{{ .Initiative.Description }}</div>
      {{- end }}
      <div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill" style="width:{{ pct .CompletedRMIs .TotalRMIs }}%;background:{{ statusColor "completed" }}"></div>
        </div>
        <div class="progress-text">{{ .CompletedRMIs }}/{{ .TotalRMIs }} RMIs completed ({{ pct .CompletedRMIs .TotalRMIs }}%)</div>
      </div>
      <div class="init-card-meta">
        {{- if .Initiative.Priority }}<span>Priority: {{ .Initiative.Priority }}</span>{{ end }}
        {{- if .Initiative.HomeRepo }}<span>Home: {{ shortRepo .Initiative.HomeRepo }}</span>{{ end }}
        {{- if .Initiative.Workspace }}<span class="workspace-pill">{{ .Initiative.Workspace }}</span>{{ end }}
      </div>
      {{- if .Tokens }}
      <div class="token-section">
        <span class="token-badge">{{ formatTokens .Tokens.TotalTokens }} tokens</span>
        <span class="cost-badge">{{ formatCost .Tokens.CostUSD }}</span>
      </div>
      {{- end }}
      {{- if .Repos }}
      <div class="repo-pills">
        {{- range .Repos }}
        <span class="repo-pill">
          <span class="repo-pill-label">{{ .Name }}</span>
          <span class="repo-pill-count">{{ .Count }}</span>
        </span>
        {{- end }}
      </div>
      {{- end }}
    </div>
    {{- end }}
  </div>
</div>
{{- end }}

{{- if .InitDeps }}
<div class="dep-section">
  <h2>Initiative Dependencies ({{ len .InitDeps }} edges)</h2>
  <div class="dep-grid">
    {{- range .InitDeps }}
    <div class="dep-edge">
      <span><a href="/initiative/{{ .SourceInitiativeID }}">{{ .SourceInitiativeID }}</a></span>
      <span class="dep-arrow">&rarr;</span>
      <span><a href="/initiative/{{ .TargetInitiativeID }}">{{ .TargetInitiativeID }}</a></span>
      <span class="dep-rel">{{ .Relationship }}</span>
    </div>
    {{- end }}
  </div>
</div>
{{- end }}

</body>
</html>
`

// detailHTML is the initiative drill-down view with full phase/RMI detail.
const detailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="5">
<title>{{ .Init.Initiative.Title }} — PRISM</title>
<style>` + dashboardCSS + `</style>
</head>
<body>

<div class="breadcrumb"><a href="/">Dashboard</a>
{{- if .Init.Initiative.ProgramID }} / <a href="/program/{{ .Init.Initiative.ProgramID }}">{{ .Init.ProgramName }}</a>{{ end }}
 / {{ .Init.Initiative.ID }}</div>

<div class="initiative">
  <div class="init-header">
    <h2>{{ .Init.Initiative.Title }}</h2>
    <span class="badge" style="background:{{ statusColor .Init.Initiative.Status }}">{{ displayStatus .Init.Initiative.Status }}</span>
    <div class="init-meta">
      <span>{{ .Init.Initiative.ID }}</span>
      {{- if .Init.Initiative.Priority }}
      <span>Priority: {{ .Init.Initiative.Priority }}</span>
      {{- end }}
      {{- if .Init.Initiative.HomeRepo }}
      <span>Home: {{ shortRepo .Init.Initiative.HomeRepo }}</span>
      {{- end }}
      {{- if .Init.Initiative.ProgramID }}
      <span class="program-pill"><a href="/program/{{ .Init.Initiative.ProgramID }}" style="color:inherit">{{ .Init.ProgramName }}</a></span>
      <span class="init-card-id">{{ .Init.Initiative.ProgramID }}</span>
      {{- end }}
      {{- if .Init.Initiative.Workspace }}
      <span class="workspace-pill">{{ .Init.Initiative.Workspace }}</span>
      {{- end }}
    </div>
  </div>

  {{- if .Init.Initiative.Description }}
  <div class="description" style="padding:0 24px 12px">{{ .Init.Initiative.Description }}</div>
  {{- end }}

  {{- if .Init.Repos }}
  <div class="repo-pills" style="padding:16px 24px">
    {{- range .Init.Repos }}
    <span class="repo-pill">
      <span class="repo-pill-label">{{ .Name }}</span>
      <span class="repo-pill-count">{{ .Count }}</span>
    </span>
    {{- end }}
  </div>
  {{- end }}

  <div style="padding:0 24px 16px">
    <div class="progress-bar-bg">
      <div class="progress-bar-fill" style="width:{{ pct .Init.CompletedRMIs .Init.TotalRMIs }}%;background:{{ statusColor "completed" }}"></div>
    </div>
    <div class="progress-text">{{ .Init.CompletedRMIs }}/{{ .Init.TotalRMIs }} RMIs completed ({{ pct .Init.CompletedRMIs .Init.TotalRMIs }}%)</div>
    {{- if .Init.Tokens }}
    <div class="token-section" style="margin-top:8px">
      <span class="token-badge">{{ formatTokens .Init.Tokens.TotalTokens }} tokens</span>
      <span class="cost-badge">{{ formatCost .Init.Tokens.CostUSD }}</span>
    </div>
    {{- end }}
  </div>

  {{- if .Init.Phases }}
  <div class="rmi-header-row">
    <div class="rmi-header-toggle">
      <span class="toggle-all-arrow" onclick="toggleAllPhases(this)">&#9660;</span>
    </div>
    <div class="rmi-header-cols">
      <span>Status</span>
      <span>ID</span>
      <span>Title</span>
      <span>Repo</span>
      <span>Type</span>
      <span>Tokens</span>
    </div>
  </div>
  {{- end }}

  {{- range .Init.Phases }}
  <div class="phase" data-phase-id="{{ .Phase.ID }}" onclick="togglePhase(this)">
    <div class="phase-header">
      <span class="toggle-arrow">&#9660;</span>
      <span class="phase-seq">{{ .Phase.SequenceNumber }}</span>
      <span class="phase-title">{{ .Phase.Title }}</span>
      {{- $phaseTooltip := phaseTooltip .RMIs }}
      {{- range phaseStatusCounts .RMIs }}
      <span class="badge status-tooltip" style="background:{{ statusColor .Status }}" title="{{ $phaseTooltip }}">{{ displayStatus .Status }} {{ .Count }}</span>
      {{- end }}
      <span class="phase-count">{{ len .RMIs }} RMIs</span>
      {{- if .Tokens }}
      <span class="token-badge" style="margin-left:auto">{{ formatTokens .Tokens.TotalTokens }}</span>
      <span class="cost-badge">{{ formatCost .Tokens.CostUSD }}</span>
      {{- end }}
    </div>
    <div class="rmi-list">
      {{- range .RMIs }}
      <div class="rmi-row">
        <span><span class="badge status-tooltip" style="background:{{ statusColor .RMI.Status }};font-size:0.6rem;padding:1px 6px" title="{{ .Tooltip }}">{{ displayStatus .RMI.Status }}</span></span>
        <span class="rmi-id">{{ .RMI.ID }}</span>
        <span class="rmi-title">{{ typeIcon .RMI.ItemType | safeHTML }} {{ .RMI.Title }}</span>
        <span class="rmi-repo">{{ shortRepo .RMI.RepositoryID }}</span>
        <span class="rmi-type">{{ .RMI.ItemType }}</span>
        <span class="rmi-type">{{ if .Tokens }}{{ formatTokens .Tokens.TotalTokens }}{{ else }}-{{ end }}</span>
      </div>
      {{- end }}
    </div>
  </div>
  {{- end }}
</div>

{{- if .InitDeps }}
<div class="dep-section">
  <h2>Initiative Dependencies</h2>
  <div class="dep-grid">
    {{- range .InitDeps }}
    <div class="dep-edge">
      <span><a href="/initiative/{{ .SourceInitiativeID }}">{{ .SourceInitiativeID }}</a></span>
      <span class="dep-arrow">&rarr;</span>
      <span><a href="/initiative/{{ .TargetInitiativeID }}">{{ .TargetInitiativeID }}</a></span>
      <span class="dep-rel">{{ .Relationship }}</span>
    </div>
    {{- end }}
  </div>
</div>
{{- end }}

{{- if .RMIDeps }}
<div class="dep-section">
  <h2>RMI Dependencies ({{ len .RMIDeps }} edges)</h2>
  <div class="dep-grid">
    {{- range .RMIDeps }}
    <div class="dep-edge">
      <span>{{ .SourceRMIID }}</span>
      <span class="dep-arrow">&rarr;</span>
      <span>{{ .TargetRMIID }}</span>
      <span class="dep-rel">{{ .Relationship }}</span>
    </div>
    {{- end }}
  </div>
</div>
{{- end }}

<script>` + dashboardJS + `</script>
</body>
</html>
`

// programHTML is the program drill-down showing all initiatives in a program.
const programHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="5">
<title>{{ .Program.Name }} — PRISM</title>
<style>` + dashboardCSS + `</style>
</head>
<body>

<div class="breadcrumb"><a href="/">Dashboard</a> / {{ .Program.Name }}</div>

<h1>{{ .Program.Name }}</h1>
<div class="init-card-id" style="margin-bottom:4px">{{ .Program.ID }}</div>
{{- if .Program.Description }}
<p class="description">{{ .Program.Description }}</p>
{{- end }}
<p class="subtitle">{{ len .Program.Initiatives }} initiatives in this program</p>

{{- $totalRMIs := 0 }}
{{- $totalCompleted := 0 }}
{{- $totalPhases := 0 }}
{{- range .Program.Initiatives }}
  {{- $totalRMIs = (add $totalRMIs .TotalRMIs) }}
  {{- $totalCompleted = (add $totalCompleted .CompletedRMIs) }}
  {{- range .Phases }}
    {{- $totalPhases = (add $totalPhases 1) }}
  {{- end }}
{{- end }}

<div class="stats-bar">
  <div class="stat-card">
    <div class="stat-value">{{ len .Program.Initiatives }}</div>
    <div class="stat-label">Initiatives</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ $totalPhases }}</div>
    <div class="stat-label">Phases</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ $totalRMIs }}</div>
    <div class="stat-label">RMIs</div>
  </div>
  <div class="stat-card">
    <div class="stat-value">{{ $totalCompleted }}/{{ $totalRMIs }}</div>
    <div class="stat-label">Completed</div>
  </div>
  {{- if .Program.Tokens }}
  <div class="stat-card">
    <div class="stat-value" style="color:#10b981">{{ formatTokens .Program.Tokens.TotalTokens }}</div>
    <div class="stat-label">Total Tokens</div>
  </div>
  <div class="stat-card">
    <div class="stat-value" style="color:#f59e0b">{{ formatCost .Program.Tokens.CostUSD }}</div>
    <div class="stat-label">Total Cost</div>
  </div>
  {{- end }}
</div>

{{- if .InitDeps }}
<div class="dep-section">
  <h2>Initiative Dependencies</h2>
  <div class="dep-grid">
    {{- range .InitDeps }}
    <div class="dep-edge">
      <span><a href="/initiative/{{ .SourceInitiativeID }}">{{ .SourceInitiativeID }}</a></span>
      <span class="dep-arrow">&rarr;</span>
      <span><a href="/initiative/{{ .TargetInitiativeID }}">{{ .TargetInitiativeID }}</a></span>
      <span class="dep-rel">{{ .Relationship }}</span>
    </div>
    {{- end }}
  </div>
</div>
{{- end }}

{{- range .Program.Initiatives }}
<div class="initiative">
  <div class="init-header">
    <h2><a href="/initiative/{{ .Initiative.ID }}" style="color:var(--text)">{{ .Initiative.Title }}</a></h2>
    <span class="badge" style="background:{{ statusColor .Initiative.Status }}">{{ displayStatus .Initiative.Status }}</span>
    <div class="init-meta">
      <span>{{ .Initiative.ID }}</span>
      {{- if .Initiative.Priority }}
      <span>Priority: {{ .Initiative.Priority }}</span>
      {{- end }}
      {{- if .Initiative.HomeRepo }}
      <span>Home: {{ shortRepo .Initiative.HomeRepo }}</span>
      {{- end }}
      {{- if .Initiative.Workspace }}
      <span class="workspace-pill">{{ .Initiative.Workspace }}</span>
      {{- end }}
    </div>
  </div>

  {{- if .Initiative.Description }}
  <div class="description" style="padding:0 24px 12px">{{ .Initiative.Description }}</div>
  {{- end }}

  {{- if .Repos }}
  <div class="repo-pills" style="padding:16px 24px">
    {{- range .Repos }}
    <span class="repo-pill">
      <span class="repo-pill-label">{{ .Name }}</span>
      <span class="repo-pill-count">{{ .Count }}</span>
    </span>
    {{- end }}
  </div>
  {{- end }}

  <div style="padding:0 24px 16px">
    <div class="progress-bar-bg">
      <div class="progress-bar-fill" style="width:{{ pct .CompletedRMIs .TotalRMIs }}%;background:{{ statusColor "completed" }}"></div>
    </div>
    <div class="progress-text">{{ .CompletedRMIs }}/{{ .TotalRMIs }} RMIs completed ({{ pct .CompletedRMIs .TotalRMIs }}%)</div>
  </div>

  {{- if .Phases }}
  <div class="rmi-header-row">
    <div class="rmi-header-toggle">
      <span class="toggle-all-arrow" onclick="toggleAllPhases(this)">&#9660;</span>
    </div>
    <div class="rmi-header-cols">
      <span>Status</span>
      <span>ID</span>
      <span>Title</span>
      <span>Repo</span>
      <span>Type</span>
      <span>Req</span>
    </div>
  </div>
  {{- end }}

  {{- range .Phases }}
  <div class="phase" data-phase-id="{{ .Phase.ID }}" onclick="togglePhase(this)">
    <div class="phase-header">
      <span class="toggle-arrow">&#9660;</span>
      <span class="phase-seq">{{ .Phase.SequenceNumber }}</span>
      <span class="phase-title">{{ .Phase.Title }}</span>
      {{- $phaseTooltip := phaseTooltip .RMIs }}
      {{- range phaseStatusCounts .RMIs }}
      <span class="badge status-tooltip" style="background:{{ statusColor .Status }}" title="{{ $phaseTooltip }}">{{ displayStatus .Status }} {{ .Count }}</span>
      {{- end }}
      <span class="phase-count">{{ len .RMIs }} RMIs</span>
    </div>
    <div class="rmi-list">
      {{- range .RMIs }}
      <div class="rmi-row">
        <span><span class="badge status-tooltip" style="background:{{ statusColor .RMI.Status }};font-size:0.6rem;padding:1px 6px" title="{{ .Tooltip }}">{{ displayStatus .RMI.Status }}</span></span>
        <span class="rmi-id">{{ .RMI.ID }}</span>
        <span class="rmi-title">{{ typeIcon .RMI.ItemType | safeHTML }} {{ .RMI.Title }}</span>
        <span class="rmi-repo">{{ shortRepo .RMI.RepositoryID }}</span>
        <span class="rmi-type">{{ .RMI.ItemType }}</span>
        <span class="rmi-type">{{ if .Tokens }}{{ formatTokens .Tokens.TotalTokens }}{{ else }}-{{ end }}</span>
      </div>
      {{- end }}
    </div>
  </div>
  {{- end }}
</div>
{{- end }}

<script>` + dashboardJS + `</script>
</body>
</html>
`
