import { createElement, lazy, Suspense } from 'react'
import type { ComponentType } from 'react'
import type { Extension, ExtensionManifest, ExtensionContext, ExtensionViewProps } from '../../types/extension'

const manifest: ExtensionManifest = {
  id: 'visionstudio.v2mom',
  name: 'V2MOM Strategic Alignment',
  version: '1.0.0',
  description: 'V2MOM cascade visualization and alignment tracking',
  publisher: 'ProductBuildersHQ',
  scope: 'both',
  activationEvents: ['onProject'],

  contributes: {
    views: [
      { id: 'v2mom', name: 'V2MOM', icon: '🎯', group: 'strategy' },
      { id: 'organization', name: 'Organization', icon: '🏢', group: 'strategy' },
    ],
    sidebarSections: [
      {
        id: 'organization',
        label: 'Organization',
        icon: '🏢',
        priority: 200,
        scope: 'global',
        items: [
          { viewId: 'organization', label: 'Strategy & V2MOMs', icon: '🏢' },
        ],
      },
      {
        id: 'strategy',
        label: 'Strategy',
        priority: 70,
        scope: 'project',
        items: [
          { viewId: 'v2mom', label: 'V2MOM Cascade', icon: '🎯' },
        ],
      },
    ],
    apiRoutes: [
      '/projects/{project}/v2moms',
      '/projects/{project}/v2moms/cascade',
      '/projects/{project}/v2moms/{v2momId}',
      '/organization',
      '/organization/v2moms',
      '/organization/v2moms/{v2momId}',
      '/organization/cascade',
    ],
  },
}

const LazyV2MOMView = lazy(() =>
  import('../../components/v2mom/V2MOMView').then(m => ({ default: m.V2MOMView }))
)
const LazyOrganizationView = lazy(() =>
  import('../../components/organization/OrganizationView').then(m => ({ default: m.OrganizationView }))
)

function V2MOMView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyV2MOMView))
}

function OrganizationView(_props: ExtensionViewProps) {
  return createElement(Suspense, { fallback: null }, createElement(LazyOrganizationView))
}

const views: Record<string, ComponentType<ExtensionViewProps>> = {
  v2mom: V2MOMView,
  organization: OrganizationView,
}

export const v2momExtension: Extension = {
  manifest,

  activate(_context: ExtensionContext) {},
  deactivate() {},

  getView(viewId: string) {
    return views[viewId] ?? null
  },
}
