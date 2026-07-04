import { ReactNode } from 'react'

interface AppLayoutProps {
  sidebar: ReactNode
  llmPanel: ReactNode
  main: ReactNode
  terminal?: ReactNode
}

export function AppLayout({ sidebar, llmPanel, main, terminal }: AppLayoutProps) {
  return (
    <div className="flex h-screen w-screen overflow-hidden">
      {/* Left column: Sidebar + LLM Panel */}
      <div className="flex flex-col w-72 bg-va-sidebar border-r border-va-border">
        {/* Sidebar: Projects, Profile, Workflow, Specs */}
        <div className="flex-1 overflow-y-auto">
          {sidebar}
        </div>

        {/* LLM Panel */}
        <div className="h-80 border-t border-va-border">
          {llmPanel}
        </div>
      </div>

      {/* Main content area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Main panel */}
        <div className="flex-1 overflow-hidden">
          {main}
        </div>

        {/* Terminal panel */}
        {terminal}
      </div>
    </div>
  )
}
