import { useState, useEffect, useCallback, createElement } from 'react'
import { AppLayout, Sidebar, SpecEditor, TerminalPanel, DEFAULT_TERMINAL_HEIGHT, AddProjectModal } from './components'
import { MethodologySelector } from './components/layout/MethodologySelector'
import { ExtensionManagerView } from './components/extensions/ExtensionManagerView'
import { api } from './services/api'
import { useProjectEvents, FileEvent } from './hooks/useProjectEvents'
import { extensionRegistry, registerBuiltinExtensions, registerMarketplaceExtensions } from './extensions'
import { AppProvider } from './contexts/AppContext'
import type { Project, Spec, ProjectMethodologyConfig } from './types'
import type { ExtensionContext } from './types/extension'

interface ActiveView {
  extensionId: string
  viewId: string
}

function buildExtensionContext(extensionId: string, project?: Project | null): ExtensionContext {
  return {
    extensionId,
    projectName: project?.name,
    projectPath: project?.path,
    api: {
      async fetch<T>(path: string, options?: RequestInit): Promise<T> {
        const url = `http://127.0.0.1:8765/api/extensions/${extensionId}${path}`
        const response = await fetch(url, {
          ...options,
          headers: { 'Content-Type': 'application/json', ...options?.headers },
        })
        if (!response.ok) throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        return response.json()
      },
      async fetchText(path: string, options?: RequestInit): Promise<string> {
        const url = `http://127.0.0.1:8765/api/extensions/${extensionId}${path}`
        const response = await fetch(url, options)
        if (!response.ok) throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        return response.text()
      },
      async getProjectData<T>(key: string): Promise<T | null> {
        if (!project?.name) return null
        try {
          const data = await api.getSpec(project.name, `${extensionId}:${key}`)
          return JSON.parse(data.content ?? 'null')
        } catch {
          return null
        }
      },
      async setProjectData(key: string, value: unknown): Promise<void> {
        if (!project?.name) return
        await api.saveSpec(project.name, `${extensionId}:${key}`, JSON.stringify(value))
      },
      async evaluate(specType: string, _content: string) {
        if (!project?.name) throw new Error('No active project')
        return api.evaluateSpec(project.name, specType)
      },
      onFileChanged(_callback) { return () => {} },
      onEvalComplete(_callback) { return () => {} },
    },
    ui: {
      showNotification(message: string, type = 'info') {
        console.log(`[${extensionId}] ${type}: ${message}`)
      },
      showProgress(title: string) {
        console.log(`[${extensionId}] progress: ${title}`)
        return {
          update(message: string, _percent?: number) { console.log(`[${extensionId}] progress: ${message}`) },
          done() { console.log(`[${extensionId}] progress done`) },
        }
      },
      openTerminal(command: string) {
        console.log(`[${extensionId}] terminal: ${command}`)
      },
    },
  }
}

const DEFAULT_VIEW: ActiveView = { extensionId: 'visionstudio.unified-dashboard', viewId: 'initiative' }

function App() {
  const [projects, setProjects] = useState<Project[]>([])
  const [activeProject, setActiveProject] = useState<Project | null>(null)
  const [activeView, setActiveView] = useState<ActiveView>(DEFAULT_VIEW)
  const [activeSpec, setActiveSpec] = useState<Spec | null>(null)
  const [specContent, setSpecContent] = useState<string>('')
  const [isDirty, setIsDirty] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [terminalHeight, setTerminalHeight] = useState(DEFAULT_TERMINAL_HEIGHT)
  const [showAddProjectModal, setShowAddProjectModal] = useState(false)
  const [showMethodologyModal, setShowMethodologyModal] = useState(false)
  const [isConnected, setIsConnected] = useState(false)
  const [extensionsReady, setExtensionsReady] = useState(false)

  // Register extensions once, then activate based on project
  useEffect(() => {
    registerBuiltinExtensions()
    registerMarketplaceExtensions()
    setExtensionsReady(true)
  }, [])

  // Activate/deactivate extensions when the active project changes
  useEffect(() => {
    if (!extensionsReady) return
    extensionRegistry.activateForProject(activeProject)
  }, [activeProject, extensionsReady])

  const handleFileChanged = useCallback((event: FileEvent) => {
    console.log('File changed:', event)
    if (activeSpec && event.specType === activeSpec.type && event.type === 'file_modified') {
      if (activeProject && !isDirty) {
        api.getSpec(activeProject.name, activeSpec.type).then(fullSpec => {
          setActiveSpec(fullSpec)
          setSpecContent(fullSpec.content || '')
        }).catch(console.error)
      }
    }
    if (activeProject && event.project === activeProject.name) {
      loadProjects()
    }
  }, [activeSpec, activeProject, isDirty])

  const handleEvalComplete = useCallback((event: FileEvent) => {
    console.log('Eval complete:', event)
    if (activeProject && event.project === activeProject.name) {
      loadProjects()
    }
    if (event.data) {
      const score = event.data.score as number
      const decision = event.data.decision as string
      console.log(`Evaluation complete for ${event.specType}: score=${score}, decision=${decision}`)
    }
  }, [activeProject])

  const handleWorkflowChanged = useCallback((_event: FileEvent) => {
    if (activeProject) {
      loadProjects()
    }
  }, [activeProject])

  useProjectEvents(activeProject?.name, {
    onFileChanged: handleFileChanged,
    onEvalComplete: handleEvalComplete,
    onWorkflowChanged: handleWorkflowChanged,
    onConnected: () => setIsConnected(true),
    onDisconnected: () => setIsConnected(false),
    enabled: !!activeProject,
  })

  useEffect(() => {
    loadProjects()
  }, [])

  const loadProjects = async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await api.listProjects()
      setProjects(data)
      if (data.length > 0) {
        setActiveProject(data[0])
      }
    } catch (err) {
      setError(`Failed to connect to daemon: ${err}`)
    } finally {
      setIsLoading(false)
    }
  }

  const navigateToView = useCallback((extensionId: string, viewId: string) => {
    setActiveView({ extensionId, viewId })
    setActiveSpec(null)
  }, [])

  const navigateToSpec = useCallback(async (spec: Spec) => {
    if (!activeProject) return
    try {
      const fullSpec = await api.getSpec(activeProject.name, spec.type)
      setActiveSpec(fullSpec)
      setSpecContent(fullSpec.content || '')
      setIsDirty(false)
    } catch (err) {
      console.error('Failed to load spec:', err)
      setActiveSpec(spec)
      setSpecContent(spec.content || '')
      setIsDirty(false)
    }
  }, [activeProject])

  const handleProjectSelect = (project: Project) => {
    setActiveProject(project)
    setActiveView(DEFAULT_VIEW)
    setActiveSpec(null)
  }

  const handleMethodologySave = (config: ProjectMethodologyConfig) => {
    if (activeProject) {
      const updatedProject = {
        ...activeProject,
        requirementsMethodology: config.requirementsMethodology,
        implementationMethodology: config.implementationMethodology,
      }
      setActiveProject(updatedProject)
      setProjects(prevProjects =>
        prevProjects.map(p =>
          p.name === activeProject.name ? updatedProject : p
        )
      )
    }
    setShowMethodologyModal(false)
  }

  const handleContentChange = (content: string) => {
    setSpecContent(content)
    setIsDirty(content !== (activeSpec?.content || ''))
  }

  const handleSave = async () => {
    if (!activeProject || !activeSpec) return
    try {
      await api.saveSpec(activeProject.name, activeSpec.type, specContent)
      setIsDirty(false)
    } catch (err) {
      console.error('Failed to save spec:', err)
    }
  }

  const handleTerminalHeightChange = useCallback((height: number) => {
    setTerminalHeight(height)
  }, [])

  const handleAddProject = async (name: string, path: string, profile: string, initialize: boolean) => {
    const newProject = await api.addProject(name, path, profile, initialize)
    setProjects([...projects, newProject])
    setActiveProject(newProject)
    setActiveView(DEFAULT_VIEW)
  }

  const handleRemoveProject = async (projectName: string) => {
    try {
      await api.removeProject(projectName)
      const updatedProjects = projects.filter(p => p.name !== projectName)
      setProjects(updatedProjects)
      if (activeProject?.name === projectName) {
        setActiveProject(updatedProjects.length > 0 ? updatedProjects[0] : null)
        setActiveView(DEFAULT_VIEW)
        setActiveSpec(null)
      }
    } catch (err) {
      console.error('Failed to remove project:', err)
    }
  }

  const renderMain = () => {
    if (activeSpec) {
      return (
        <SpecEditor
          spec={{ ...activeSpec, content: specContent }}
          onContentChange={handleContentChange}
          onSave={handleSave}
          isDirty={isDirty}
        />
      )
    }

    if (activeView.extensionId === '_system' && activeView.viewId === 'extensions') {
      return <ExtensionManagerView onViewSelect={navigateToView} />
    }

    if (extensionsReady) {
      const ViewComponent = extensionRegistry.getView(activeView.extensionId, activeView.viewId)
      if (ViewComponent) {
        const ctx = buildExtensionContext(activeView.extensionId, activeProject)
        return createElement(ViewComponent, { context: ctx })
      }
    }

    if (!activeProject) {
      return <EmptyState message="Select a project to get started" />
    }

    return <EmptyState message="View not found" />
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-va-bg text-va-text">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full mx-auto mb-4" />
          <p>Connecting to daemon...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen bg-va-bg text-va-text">
        <div className="text-center max-w-md">
          <p className="text-va-error mb-4">{error}</p>
          <button
            onClick={loadProjects}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <AppProvider value={{ activeProject, navigateToSpec, navigateToView }}>
      <AppLayout
        sidebar={
          <>
            <Sidebar
              projects={projects}
              activeProject={activeProject}
              onProjectSelect={handleProjectSelect}
              onSpecSelect={navigateToSpec}
              onViewSelect={navigateToView}
              onMethodologyClick={() => setShowMethodologyModal(true)}
              onExtensionsClick={() => navigateToView('_system', 'extensions')}
              activeView={activeView}
              activeSpec={activeSpec}
              onAddProjectClick={() => setShowAddProjectModal(true)}
              onRemoveProject={handleRemoveProject}
              isConnected={isConnected}
            />
            {showAddProjectModal && (
              <AddProjectModal
                onClose={() => setShowAddProjectModal(false)}
                onAdd={handleAddProject}
              />
            )}
            {showMethodologyModal && activeProject && (
              <MethodologySelector
                projectName={activeProject.name}
                currentConfig={{
                  requirementsMethodology: activeProject.requirementsMethodology || activeProject.profile.name,
                  implementationMethodology: activeProject.implementationMethodology || 'none',
                }}
                onClose={() => setShowMethodologyModal(false)}
                onSave={handleMethodologySave}
              />
            )}
          </>
        }
        main={renderMain()}
        terminal={
          <TerminalPanel
            height={terminalHeight}
            onHeightChange={handleTerminalHeightChange}
            projectPath={activeProject?.path}
            projectName={activeProject?.name}
          />
        }
      />
    </AppProvider>
  )
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="flex items-center justify-center h-full text-va-text-muted">
      {message}
    </div>
  )
}

export default App
