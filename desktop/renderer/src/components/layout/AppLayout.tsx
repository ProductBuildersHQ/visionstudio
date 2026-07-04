import { ReactNode } from 'react'

interface AppLayoutProps {
  sidebar: ReactNode
  main: ReactNode
  terminal?: ReactNode
}

export function AppLayout({ sidebar, main, terminal }: AppLayoutProps) {
  return (
    <div className="flex h-screen w-screen overflow-hidden">
      {/* Left column: Sidebar */}
      <div className="flex flex-col w-72 bg-va-sidebar border-r border-va-border">
        <div className="flex-1 overflow-y-auto">
          {sidebar}
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
