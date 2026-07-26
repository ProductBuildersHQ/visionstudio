import type { ComponentType } from 'react'
import type { Rubric } from '@plexusone/structured-evaluation'

// ---------------------------------------------------------------------------
// Extension Manifest
// ---------------------------------------------------------------------------

export interface ExtensionManifest {
  id: string
  name: string
  version: string
  description: string
  publisher: string
  icon?: string
  repository?: string
  license?: string

  scope: ExtensionScope
  activationEvents?: string[]

  contributes: ExtensionContributions
}

export type ExtensionScope = 'project' | 'global' | 'both'

export interface ExtensionContributions {
  views?: ViewContribution[]
  sidebarSections?: SidebarSectionContribution[]
  commands?: CommandContribution[]
  settings?: SettingContribution[]
  apiRoutes?: string[]
}

// ---------------------------------------------------------------------------
// View Contributions
// ---------------------------------------------------------------------------

export interface ViewContribution {
  id: string
  name: string
  icon?: string
  when?: string
  group?: string
}

export interface SidebarSectionContribution {
  id: string
  label: string
  icon?: string
  priority?: number
  scope?: 'global' | 'project'
  items: SidebarItemContribution[]
}

export interface SidebarItemContribution {
  viewId: string
  label: string
  icon?: string
  badge?: string
  when?: string
}

// ---------------------------------------------------------------------------
// Commands & Settings
// ---------------------------------------------------------------------------

export interface CommandContribution {
  id: string
  title: string
  icon?: string
  keybinding?: string
}

export interface SettingContribution {
  id: string
  title: string
  description?: string
  type: 'string' | 'boolean' | 'number' | 'select'
  default?: unknown
  options?: { label: string; value: string }[]
}

// ---------------------------------------------------------------------------
// Extension Runtime API
// ---------------------------------------------------------------------------

export interface ExtensionContext {
  extensionId: string
  projectName?: string
  projectPath?: string

  api: ExtensionAPI
  ui: ExtensionUI
}

export interface ExtensionAPI {
  fetch<T>(path: string, options?: RequestInit): Promise<T>
  fetchText(path: string, options?: RequestInit): Promise<string>

  getProjectData<T>(key: string): Promise<T | null>
  setProjectData(key: string, value: unknown): Promise<void>

  evaluate(specType: string, content: string): Promise<Rubric>

  onFileChanged(callback: (event: FileChangeEvent) => void): () => void
  onEvalComplete(callback: (event: EvalCompleteEvent) => void): () => void
}

export interface FileChangeEvent {
  type: 'created' | 'modified' | 'deleted' | 'renamed'
  path: string
  specType?: string
}

export interface EvalCompleteEvent {
  specType: string
  result: Rubric
}

export interface ExtensionUI {
  showNotification(message: string, type?: 'info' | 'success' | 'warning' | 'error'): void
  showProgress(title: string): ProgressHandle
  openTerminal(command: string): void
}

export interface ProgressHandle {
  update(message: string, percent?: number): void
  done(): void
}

// ---------------------------------------------------------------------------
// Extension Registration (what the extension module exports)
// ---------------------------------------------------------------------------

export interface Extension {
  manifest: ExtensionManifest

  activate(context: ExtensionContext): void | Promise<void>
  deactivate?(): void | Promise<void>

  getView(viewId: string): ComponentType<ExtensionViewProps> | null
}

export interface ExtensionViewProps {
  context: ExtensionContext
}

// ---------------------------------------------------------------------------
// Marketplace
// ---------------------------------------------------------------------------

export interface MarketplaceManifest {
  version: string
  lastUpdated: string
  extensions: MarketplaceEntry[]
}

export interface MarketplaceEntry {
  id: string
  name: string
  description: string
  publisher: string
  version: string
  repository: string
  icon?: string
  downloads?: number
  rating?: number
  tags?: string[]
  builtIn?: boolean
}

// ---------------------------------------------------------------------------
// Extension Registry (internal to VisionStudio)
// ---------------------------------------------------------------------------

export interface RegisteredExtension {
  manifest: ExtensionManifest
  extension: Extension
  isActive: boolean
  isInstalled: boolean
  isBuiltIn: boolean
  source: 'builtin' | 'marketplace' | 'local'
}

export type ExtensionRegistry = Map<string, RegisteredExtension>
