import type { ComponentType } from 'react'
import type {
  Extension,
  ExtensionContext,
  ExtensionViewProps,
  RegisteredExtension,
  SidebarSectionContribution,
  ViewContribution,
} from '../types/extension'
import { api } from '../services/api'

const INSTALLED_EXTENSIONS_KEY = 'visionstudio:installed-extensions'

interface ProjectInfo {
  name: string
  path?: string
  implementationMethodology?: string
  requirementsMethodology?: string
}

function matchesActivationEvent(event: string, project?: ProjectInfo | null): boolean {
  if (event === 'onProject') {
    return !!project
  }
  if (event.startsWith('onProject:')) {
    if (!project) return false
    const condition = event.slice('onProject:'.length)
    const eqIdx = condition.indexOf('=')
    if (eqIdx === -1) return false
    const key = condition.slice(0, eqIdx)
    const value = condition.slice(eqIdx + 1)
    const projectRecord = project as unknown as Record<string, unknown>
    return projectRecord[key] === value
  }
  return true
}

function shouldActivate(events: string[] | undefined, project?: ProjectInfo | null): boolean {
  if (!events || events.length === 0) return true
  return events.some(e => matchesActivationEvent(e, project))
}

function loadInstalledIds(): Set<string> {
  try {
    const raw = localStorage.getItem(INSTALLED_EXTENSIONS_KEY)
    if (raw) return new Set(JSON.parse(raw))
  } catch { /* ignore corrupt data */ }
  return new Set()
}

function saveInstalledIds(ids: Set<string>): void {
  localStorage.setItem(INSTALLED_EXTENSIONS_KEY, JSON.stringify([...ids]))
}

class ExtensionRegistryImpl {
  private extensions = new Map<string, RegisteredExtension>()
  private installedIds = loadInstalledIds()

  register(extension: Extension, source: 'builtin' | 'marketplace' | 'local'): void {
    const id = extension.manifest.id
    if (this.extensions.has(id)) {
      console.warn(`Extension ${id} already registered, skipping`)
      return
    }

    const isBuiltIn = source === 'builtin'

    this.extensions.set(id, {
      manifest: extension.manifest,
      extension,
      isActive: false,
      isInstalled: isBuiltIn || this.installedIds.has(id),
      isBuiltIn,
      source,
    })
  }

  install(extensionId: string): void {
    const reg = this.extensions.get(extensionId)
    if (!reg) throw new Error(`Extension ${extensionId} not found`)
    if (reg.isInstalled) return

    reg.isInstalled = true
    this.installedIds.add(extensionId)
    saveInstalledIds(this.installedIds)
  }

  async uninstall(extensionId: string): Promise<void> {
    const reg = this.extensions.get(extensionId)
    if (!reg) throw new Error(`Extension ${extensionId} not found`)
    if (reg.isBuiltIn) throw new Error('Cannot uninstall built-in extensions')

    if (reg.isActive) {
      await this.deactivate(extensionId)
    }
    reg.isInstalled = false
    this.installedIds.delete(extensionId)
    saveInstalledIds(this.installedIds)
  }

  async activate(extensionId: string, projectName?: string, projectPath?: string): Promise<void> {
    const reg = this.extensions.get(extensionId)
    if (!reg) throw new Error(`Extension ${extensionId} not found`)
    if (!reg.isInstalled) return
    if (reg.isActive) return

    const context = this.createContext(extensionId, projectName, projectPath)
    await reg.extension.activate(context)
    reg.isActive = true
  }

  async deactivate(extensionId: string): Promise<void> {
    const reg = this.extensions.get(extensionId)
    if (!reg || !reg.isActive) return

    await reg.extension.deactivate?.()
    reg.isActive = false
  }

  async activateForProject(project: ProjectInfo | null): Promise<void> {
    for (const [id, reg] of this.extensions) {
      if (!reg.isInstalled) continue
      const shouldBeActive = shouldActivate(reg.manifest.activationEvents, project)
      if (shouldBeActive && !reg.isActive) {
        await this.activate(id, project?.name, project?.path)
      } else if (!shouldBeActive && reg.isActive) {
        await this.deactivate(id)
      }
    }
  }

  getView(extensionId: string, viewId: string): ComponentType<ExtensionViewProps> | null {
    const reg = this.extensions.get(extensionId)
    if (!reg) return null
    return reg.extension.getView(viewId)
  }

  getAllViews(): Array<{ extensionId: string; view: ViewContribution }> {
    const result: Array<{ extensionId: string; view: ViewContribution }> = []
    for (const [id, reg] of this.extensions) {
      for (const view of reg.manifest.contributes.views ?? []) {
        result.push({ extensionId: id, view })
      }
    }
    return result
  }

  getAllSidebarSections(): Array<{ extensionId: string; section: SidebarSectionContribution }> {
    const result: Array<{ extensionId: string; section: SidebarSectionContribution }> = []
    for (const [id, reg] of this.extensions) {
      if (!reg.isActive) continue
      for (const section of reg.manifest.contributes.sidebarSections ?? []) {
        result.push({ extensionId: id, section })
      }
    }
    result.sort((a, b) => (b.section.priority ?? 0) - (a.section.priority ?? 0))
    return result
  }

  getGlobalSidebarSections(): Array<{ extensionId: string; section: SidebarSectionContribution }> {
    return this.getAllSidebarSections().filter(s => s.section.scope === 'global')
  }

  getProjectSidebarSections(): Array<{ extensionId: string; section: SidebarSectionContribution }> {
    return this.getAllSidebarSections().filter(s => s.section.scope !== 'global')
  }

  getRegistered(extensionId: string): RegisteredExtension | undefined {
    return this.extensions.get(extensionId)
  }

  listAll(): RegisteredExtension[] {
    return Array.from(this.extensions.values())
  }

  listInstalled(): RegisteredExtension[] {
    return this.listAll().filter(r => r.isInstalled)
  }

  listActive(): RegisteredExtension[] {
    return this.listAll().filter(r => r.isActive)
  }

  private createContext(
    extensionId: string,
    projectName?: string,
    projectPath?: string,
  ): ExtensionContext {
    return {
      extensionId,
      projectName,
      projectPath,
      api: {
        async fetch<T>(path: string, options?: RequestInit): Promise<T> {
          const url = `http://127.0.0.1:8765/api/extensions/${extensionId}${path}`
          const response = await fetch(url, {
            ...options,
            headers: { 'Content-Type': 'application/json', ...options?.headers },
          })
          if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`)
          }
          return response.json()
        },
        async fetchText(path: string, options?: RequestInit): Promise<string> {
          const url = `http://127.0.0.1:8765/api/extensions/${extensionId}${path}`
          const response = await fetch(url, options)
          if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`)
          }
          return response.text()
        },
        async getProjectData<T>(key: string): Promise<T | null> {
          if (!projectName) return null
          try {
            const data = await api.getSpec(projectName, `${extensionId}:${key}`)
            return JSON.parse(data.content ?? 'null')
          } catch {
            return null
          }
        },
        async setProjectData(key: string, value: unknown): Promise<void> {
          if (!projectName) return
          await api.saveSpec(projectName, `${extensionId}:${key}`, JSON.stringify(value))
        },
        async evaluate(specType: string, _content: string) {
          if (!projectName) throw new Error('No active project')
          return api.evaluateSpec(projectName, specType)
        },
        onFileChanged(_callback) {
          return () => {}
        },
        onEvalComplete(_callback) {
          return () => {}
        },
      },
      ui: {
        showNotification(message: string, type = 'info') {
          console.log(`[${extensionId}] ${type}: ${message}`)
        },
        showProgress(title: string) {
          console.log(`[${extensionId}] progress: ${title}`)
          return {
            update(message: string, _percent?: number) {
              console.log(`[${extensionId}] progress: ${message}`)
            },
            done() {
              console.log(`[${extensionId}] progress done`)
            },
          }
        },
        openTerminal(command: string) {
          console.log(`[${extensionId}] terminal: ${command}`)
        },
      },
    }
  }
}

export const extensionRegistry = new ExtensionRegistryImpl()
