import { createElement, lazy, Suspense } from 'react'
import type { ComponentType } from 'react'
import type { Extension, ExtensionManifest, ExtensionContext, ExtensionViewProps } from '../../types/extension'

const manifest: ExtensionManifest = {
  id: 'visionstudio.aidlc',
  name: 'AI Development Lifecycle',
  version: '1.0.0',
  description: 'AWS AI Development Lifecycle workflow with three-phase document management',
  publisher: 'ProductBuildersHQ',
  scope: 'project',
  activationEvents: ['onProject:implementationMethodology=aidlc'],

  contributes: {
    views: [
      { id: 'aidlc-workflow', name: 'AIDLC Workflow', group: 'aidlc' },
      { id: 'aidlc-sync', name: 'AIDLC Sync', group: 'aidlc' },
    ],
    sidebarSections: [
      {
        id: 'aidlc',
        label: 'AI DLC',
        priority: 90,
        scope: 'project',
        items: [
          { viewId: 'aidlc-workflow', label: 'Workflow' },
          { viewId: 'aidlc-sync', label: 'Sync' },
        ],
      },
    ],
    apiRoutes: [
      '/projects/{project}/aidlc/state',
      '/projects/{project}/aidlc/workflow',
      '/projects/{project}/aidlc/documents',
      '/projects/{project}/aidlc/documents/{docId}',
      '/projects/{project}/aidlc/documents/create',
      '/projects/{project}/aidlc/sync/diff',
      '/projects/{project}/aidlc/sync',
      '/projects/{project}/aidlc/phase/requirements',
      '/projects/{project}/aidlc/phase/transition',
      '/projects/{project}/aidlc/templates',
      '/projects/{project}/aidlc/templates/{docType}',
    ],
  },
}

const LazyAIDLCWorkflowView = lazy(() =>
  import('../../components/aidlc/AIDLCWorkflowView').then(m => ({ default: m.AIDLCWorkflowView }))
)
const LazyAIDLCSyncPanel = lazy(() =>
  import('../../components/aidlc/AIDLCSyncPanel').then(m => ({ default: m.AIDLCSyncPanel }))
)

function WorkflowView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyAIDLCWorkflowView))
}

function SyncView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyAIDLCSyncPanel))
}

const views: Record<string, ComponentType<ExtensionViewProps>> = {
  'aidlc-workflow': WorkflowView,
  'aidlc-sync': SyncView,
}

export const aidlcExtension: Extension = {
  manifest,

  activate(_context: ExtensionContext) {},
  deactivate() {},

  getView(viewId: string) {
    return views[viewId] ?? null
  },
}
