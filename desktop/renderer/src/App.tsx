import { useState, useEffect, useCallback } from 'react'
import { AppLayout, Sidebar, WorkflowDiagram, SpecEditor, LLMPanel, TerminalPanel, DEFAULT_TERMINAL_HEIGHT } from './components'
import { api } from './services/api'
import type { Project, Spec } from './types'

type ActiveView = 'workflow' | 'spec'

function App() {
  const [projects, setProjects] = useState<Project[]>([])
  const [activeProject, setActiveProject] = useState<Project | null>(null)
  const [activeView, setActiveView] = useState<ActiveView>('workflow')
  const [activeSpec, setActiveSpec] = useState<Spec | null>(null)
  const [specContent, setSpecContent] = useState<string>('')
  const [isDirty, setIsDirty] = useState(false)
  const [isLLMLoading, setIsLLMLoading] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [terminalHeight, setTerminalHeight] = useState(DEFAULT_TERMINAL_HEIGHT)

  // Load projects on mount
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

  const handleProjectSelect = (project: Project) => {
    setActiveProject(project)
    setActiveView('workflow')
    setActiveSpec(null)
  }

  const handleSpecSelect = async (spec: Spec) => {
    if (!activeProject) return

    try {
      const fullSpec = await api.getSpec(activeProject.name, spec.type)
      setActiveSpec(fullSpec)
      setSpecContent(fullSpec.content || '')
      setIsDirty(false)
      setActiveView('spec')
    } catch (err) {
      console.error('Failed to load spec:', err)
      // Fall back to the spec we have
      setActiveSpec(spec)
      setSpecContent(spec.content || '')
      setIsDirty(false)
      setActiveView('spec')
    }
  }

  const handleWorkflowClick = () => {
    setActiveView('workflow')
    setActiveSpec(null)
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

  const handleLLMMessage = async (message: string): Promise<string> => {
    setIsLLMLoading(true)
    try {
      const response = await api.chat(message, specContent)
      return response
    } catch (err) {
      return `Error: ${err}`
    } finally {
      setIsLLMLoading(false)
    }
  }

  const handleTerminalHeightChange = useCallback((height: number) => {
    setTerminalHeight(height)
  }, [])

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
    <AppLayout
      sidebar={
        <Sidebar
          projects={projects}
          activeProject={activeProject}
          onProjectSelect={handleProjectSelect}
          onSpecSelect={handleSpecSelect}
          onWorkflowClick={handleWorkflowClick}
          activeSpec={activeSpec}
        />
      }
      llmPanel={
        <LLMPanel onSendMessage={handleLLMMessage} isLoading={isLLMLoading} />
      }
      main={
        activeProject ? (
          activeView === 'workflow' ? (
            <WorkflowDiagram
              project={activeProject}
              onSpecClick={handleSpecSelect}
            />
          ) : activeSpec ? (
            <SpecEditor
              spec={{ ...activeSpec, content: specContent }}
              onContentChange={handleContentChange}
              onSave={handleSave}
              isDirty={isDirty}
            />
          ) : (
            <EmptyState message="Select a spec to edit" />
          )
        ) : (
          <EmptyState message="Select a project to get started" />
        )
      }
      terminal={
        <TerminalPanel
          height={terminalHeight}
          onHeightChange={handleTerminalHeightChange}
          projectPath={activeProject?.path}
          projectName={activeProject?.name}
        />
      }
    />
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
