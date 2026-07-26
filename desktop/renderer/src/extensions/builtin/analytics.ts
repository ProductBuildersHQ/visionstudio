import { createElement, lazy, Suspense } from 'react'
import type { ComponentType } from 'react'
import type { Extension, ExtensionManifest, ExtensionContext, ExtensionViewProps } from '../../types/extension'

const manifest: ExtensionManifest = {
  id: 'visionstudio.analytics',
  name: 'Analytics & Insights',
  version: '1.0.0',
  description: 'Maturity model, capability stack, roadmap, and DevX dashboard',
  publisher: 'ProductBuildersHQ',
  scope: 'both',
  activationEvents: ['onProject'],

  contributes: {
    views: [
      { id: 'maturity-model', name: 'Maturity Model', icon: '📈', group: 'analytics' },
      { id: 'capabilities', name: 'Capabilities', icon: '🧱', group: 'analytics' },
      { id: 'roadmap', name: 'Roadmap', icon: '🗺️', group: 'analytics' },
      { id: 'devx-dashboard', name: 'DevX Dashboard', icon: '📊', group: 'analytics' },
    ],
    sidebarSections: [
      {
        id: 'devx',
        label: 'DevX',
        icon: '📊',
        priority: 190,
        scope: 'global',
        items: [
          { viewId: 'devx-dashboard', label: 'Usage Dashboard', icon: '📊' },
        ],
      },
      {
        id: 'analytics',
        label: 'Analytics',
        priority: 60,
        scope: 'project',
        items: [
          { viewId: 'maturity-model', label: 'Maturity Model', icon: '📈' },
          { viewId: 'capabilities', label: 'Capabilities', icon: '🧱' },
          { viewId: 'roadmap', label: 'Roadmap', icon: '🗺️' },
        ],
      },
    ],
    apiRoutes: [
      '/projects/{project}/maturity/models',
      '/projects/{project}/maturity/models/{modelId}',
      '/projects/{project}/maturity/dashboard',
      '/projects/{project}/capabilities',
      '/projects/{project}/capabilities/{capabilityId}',
      '/projects/{project}/roadmap',
      '/devx/dashboard',
    ],
  },
}

const LazyMaturityModelView = lazy(() =>
  import('../../components/maturity-model/MaturityModelView').then(m => ({ default: m.MaturityModelView }))
)
const LazyCapabilityStackView = lazy(() =>
  import('../../components/capability-stack/CapabilityStackView').then(m => ({ default: m.CapabilityStackView }))
)
const LazyRoadmapView = lazy(() =>
  import('../../components/roadmap/RoadmapView').then(m => ({ default: m.RoadmapView }))
)
const LazyDevXDashboardView = lazy(() =>
  import('../../components/devx/DevXDashboardView').then(m => ({ default: m.DevXDashboardView }))
)

function MaturityModelView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyMaturityModelView))
}

function CapabilitiesView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyCapabilityStackView))
}

function RoadmapView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyRoadmapView))
}

function DevXView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyDevXDashboardView))
}

const views: Record<string, ComponentType<ExtensionViewProps>> = {
  'maturity-model': MaturityModelView,
  capabilities: CapabilitiesView,
  roadmap: RoadmapView,
  'devx-dashboard': DevXView,
}

export const analyticsExtension: Extension = {
  manifest,

  activate(_context: ExtensionContext) {},
  deactivate() {},

  getView(viewId: string) {
    return views[viewId] ?? null
  },
}
