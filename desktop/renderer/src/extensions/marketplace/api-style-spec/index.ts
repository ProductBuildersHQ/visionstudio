import type { ComponentType } from 'react'
import type { Extension, ExtensionManifest, ExtensionContext, ExtensionViewProps } from '../../../types/extension'
import { DashboardView } from './DashboardView'
import { FindingsExplorer } from './FindingsView'
import { LintResultsView } from './LintResultsView'
import { RulesView } from './RulesView'
import { EditorView } from './EditorView'

const manifest: ExtensionManifest = {
  id: 'plexusone.api-style-spec',
  name: 'API Style Spec',
  version: '0.5.0',
  description: 'Review and improve API style specifications with LLM-as-Judge evaluation, dashboard visualization, and inline linting',
  publisher: 'PlexusOne',
  repository: 'https://github.com/plexusone/api-style-spec',
  scope: 'project',
  activationEvents: ['onProject'],

  contributes: {
    views: [
      { id: 'dashboard', name: 'Dashboard', icon: '📊', group: 'api-style' },
      { id: 'findings', name: 'Findings', icon: '📋', group: 'api-style' },
      { id: 'lint-results', name: 'Lint Results', icon: '🔍', group: 'api-style' },
      { id: 'rules', name: 'Rules', icon: '📖', group: 'api-style' },
      { id: 'editor', name: 'Lint Editor', icon: '✏️', group: 'api-style' },
    ],
    sidebarSections: [
      {
        id: 'api-style',
        label: 'API Style',
        icon: '🔧',
        priority: 80,
        scope: 'project',
        items: [
          { viewId: 'dashboard', label: 'Dashboard', icon: '📊' },
          { viewId: 'editor', label: 'Lint Editor', icon: '✏️' },
          { viewId: 'findings', label: 'Findings', icon: '📋' },
          { viewId: 'lint-results', label: 'Lint', icon: '🔍' },
          { viewId: 'rules', label: 'Rules', icon: '📖' },
        ],
      },
    ],
    apiRoutes: [
      '/extensions/api-style-spec/lint',
      '/extensions/api-style-spec/profiles',
      '/extensions/api-style-spec/profiles/{name}',
      '/extensions/api-style-spec/suggest-fixes',
    ],
  },
}

const views: Record<string, ComponentType<ExtensionViewProps>> = {
  dashboard: DashboardView,
  findings: FindingsExplorer,
  'lint-results': LintResultsView,
  rules: RulesView,
  editor: EditorView,
}

export const apiStyleSpecExtension: Extension = {
  manifest,

  activate(_context: ExtensionContext) {},
  deactivate() {},

  getView(viewId: string) {
    return views[viewId] ?? null
  },
}
