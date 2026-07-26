import { createElement, lazy, Suspense } from 'react'
import type { ComponentType } from 'react'
import type { Extension, ExtensionManifest, ExtensionContext, ExtensionViewProps } from '../../types/extension'

const manifest: ExtensionManifest = {
  id: 'visionstudio.visionspec',
  name: 'VisionSpec Workflow',
  version: '1.0.0',
  description: 'Product and feature specification workflow with LLM-as-Judge evaluation',
  publisher: 'ProductBuildersHQ',
  scope: 'project',
  activationEvents: ['onProject'],

  contributes: {
    views: [
      { id: 'workflow', name: 'Workflow', icon: '📊', group: 'visionspec' },
      { id: 'findings', name: 'All Findings', icon: '📋', group: 'visionspec' },
    ],
    sidebarSections: [
      {
        id: 'visionspec',
        label: 'VisionSpec',
        priority: 100,
        scope: 'project',
        items: [
          { viewId: 'workflow', label: 'Workflow', icon: '📊' },
          { viewId: 'findings', label: 'All Findings', icon: '📋' },
        ],
      },
    ],
    apiRoutes: [
      '/projects/{project}/workflow',
      '/projects/{project}/workflow/status',
      '/projects/{project}/specs/{specType}',
      '/projects/{project}/specs/{specType}/evaluate',
      '/projects/{project}/lint',
    ],
  },
}

const LazyWorkflowDiagram = lazy(() =>
  import('../../components/project/WorkflowDiagram').then(m => ({ default: m.WorkflowDiagram }))
)
const LazyFindingsView = lazy(() =>
  import('../../components/project/FindingsView').then(m => ({ default: m.FindingsView }))
)

function WorkflowView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyWorkflowDiagram))
}

function FindingsView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyFindingsView))
}

const views: Record<string, ComponentType<ExtensionViewProps>> = {
  workflow: WorkflowView,
  findings: FindingsView,
}

export const visionspecExtension: Extension = {
  manifest,

  activate(_context: ExtensionContext) {},
  deactivate() {},

  getView(viewId: string) {
    return views[viewId] ?? null
  },
}
