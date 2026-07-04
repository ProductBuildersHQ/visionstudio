import { useState, useRef, useEffect } from 'react'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

interface LLMPanelProps {
  onSendMessage: (message: string) => Promise<string>
  isLoading?: boolean
}

export function LLMPanel({ onSendMessage, isLoading = false }: LLMPanelProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSend = async () => {
    if (!input.trim() || isLoading) return

    const userMessage = input.trim()
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', content: userMessage }])

    try {
      const response = await onSendMessage(userMessage)
      setMessages((prev) => [...prev, { role: 'assistant', content: response }])
    } catch (error) {
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: `Error: ${error}` },
      ])
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="px-3 py-2 border-b border-va-border">
        <h3 className="text-sm font-semibold text-va-text">LLM Assistant</h3>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {messages.length === 0 && (
          <div className="text-xs text-va-text-muted text-center py-4">
            Ask questions about your spec or request evaluations
          </div>
        )}
        {messages.map((msg, idx) => (
          <div
            key={idx}
            className={`text-sm rounded-lg px-3 py-2 ${
              msg.role === 'user'
                ? 'bg-va-accent/20 text-va-text ml-4'
                : 'bg-va-panel text-va-text mr-4'
            }`}
          >
            {msg.content}
          </div>
        ))}
        {isLoading && (
          <div className="bg-va-panel text-va-text-muted text-sm rounded-lg px-3 py-2 mr-4 animate-pulse">
            Thinking...
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-3 border-t border-va-border">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask about this spec..."
            className="flex-1 bg-va-panel border border-va-border rounded px-3 py-2 text-sm text-va-text placeholder-va-text-muted focus:outline-none focus:border-va-accent"
            disabled={isLoading}
          />
          <button
            onClick={handleSend}
            disabled={isLoading || !input.trim()}
            className="px-3 py-2 bg-va-accent text-white rounded text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-va-accent/80 transition-colors"
          >
            Send
          </button>
        </div>
      </div>
    </div>
  )
}
