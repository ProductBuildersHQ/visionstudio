import { createElement, lazy, Suspense } from 'react'
import type { ComponentType } from 'react'
import type { Extension, ExtensionManifest, ExtensionContext, ExtensionViewProps } from '../../types/extension'

const manifest: ExtensionManifest = {
  id: 'visionstudio.unified-dashboard',
  name: 'Unified Dashboard',
  version: '1.0.0',
  description: 'Multi-domain dashboard: execution, spend, maturity, and specs',
  publisher: 'ProductBuildersHQ',
  scope: 'global',
  activationEvents: [],

  contributes: {
    views: [
      { id: 'initiative', name: 'Initiative', icon: '🎯', group: 'dashboard' },
      { id: 'execution', name: 'Execution', icon: '📊', group: 'dashboard' },
      { id: 'spend', name: 'Spend', icon: '💰', group: 'dashboard' },
      { id: 'maturity', name: 'Maturity', icon: '📈', group: 'dashboard' },
      { id: 'specs', name: 'Specs', icon: '📋', group: 'dashboard' },
    ],
    sidebarSections: [
      {
        id: 'dashboard',
        label: 'Dashboard',
        priority: 200,
        scope: 'global',
        items: [
          { viewId: 'initiative', label: 'Initiative', icon: '🎯' },
          { viewId: 'execution', label: 'Execution', icon: '📊' },
          { viewId: 'spend', label: 'Spend', icon: '💰' },
          { viewId: 'maturity', label: 'Maturity', icon: '📈' },
          { viewId: 'specs', label: 'Specs', icon: '📋' },
        ],
      },
    ],
    apiRoutes: [
      '/execution',
      '/spend',
      '/maturity',
      '/specs',
    ],
  },
}

const LazyInitiativeView = lazy(() =>
  import('../../unified/panels/InitiativeView').then(m => ({ default: m.InitiativeView }))
)
const LazyExecutionPanel = lazy(() =>
  import('../../unified/panels/ExecutionPanel').then(m => ({ default: m.ExecutionPanel }))
)
const LazySpendPanel = lazy(() =>
  import('../../unified/panels/SpendPanel').then(m => ({ default: m.SpendPanel }))
)
const LazyMaturityPanel = lazy(() =>
  import('../../unified/panels/MaturityPanel').then(m => ({ default: m.MaturityPanel }))
)
const LazySpecsPanel = lazy(() =>
  import('../../unified/panels/SpecsPanel').then(m => ({ default: m.SpecsPanel }))
)

function InitiativeViewWrapper(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyInitiativeView))
}

function ExecutionPanelWrapper(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyExecutionPanel))
}

function SpendPanelWrapper(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazySpendPanel))
}

function MaturityPanelWrapper(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyMaturityPanel))
}

function SpecsPanelWrapper(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazySpecsPanel))
}

const views: Record<string, ComponentType<ExtensionViewProps>> = {
  initiative: InitiativeViewWrapper,
  execution: ExecutionPanelWrapper,
  spend: SpendPanelWrapper,
  maturity: MaturityPanelWrapper,
  specs: SpecsPanelWrapper,
}

export const unifiedDashboardExtension: Extension = {
  manifest,

  activate(_context: ExtensionContext) {},
  deactivate() {},

  getView(viewId: string) {
    return views[viewId] ?? null
  },
}
