import { useState, useEffect } from 'react'
import { ExecutionPanel } from './panels/ExecutionPanel'
import { SpendPanel } from './panels/SpendPanel'
import { MaturityPanel } from './panels/MaturityPanel'
import { SpecsPanel } from './panels/SpecsPanel'

type Tab = 'execution' | 'spend' | 'maturity' | 'specs'

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>('execution')
  const [apiStatus, setApiStatus] = useState<'loading' | 'connected' | 'error'>('loading')

  useEffect(() => {
    fetch('/api/execution')
      .then((res) => {
        if (res.ok) {
          setApiStatus('connected')
        } else {
          setApiStatus('error')
        }
      })
      .catch(() => setApiStatus('error'))
  }, [])

  const tabs: { id: Tab; label: string }[] = [
    { id: 'execution', label: 'Execution' },
    { id: 'spend', label: 'Spend' },
    { id: 'maturity', label: 'Maturity' },
    { id: 'specs', label: 'Specs' },
  ]

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      <header className="border-b border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold">VisionStudio</h1>
          <div className="flex items-center gap-2">
            <span
              className={`h-2 w-2 rounded-full ${
                apiStatus === 'connected'
                  ? 'bg-green-500'
                  : apiStatus === 'error'
                  ? 'bg-red-500'
                  : 'bg-yellow-500'
              }`}
            />
            <span className="text-sm text-gray-400">
              {apiStatus === 'connected'
                ? 'API Connected'
                : apiStatus === 'error'
                ? 'API Error'
                : 'Connecting...'}
            </span>
          </div>
        </div>
        <nav className="mt-4 flex gap-1">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 rounded-t-lg text-sm font-medium transition-colors ${
                activeTab === tab.id
                  ? 'bg-gray-800 text-white'
                  : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/50'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </header>
      <main className="p-6">
        {activeTab === 'execution' && <ExecutionPanel />}
        {activeTab === 'spend' && <SpendPanel />}
        {activeTab === 'maturity' && <MaturityPanel />}
        {activeTab === 'specs' && <SpecsPanel />}
      </main>
    </div>
  )
}
