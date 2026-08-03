import { useMemo } from 'react'
import MarkdownIt from 'markdown-it'

interface MarkdownPreviewProps {
  content: string
  className?: string
}

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true,
})

export function MarkdownPreview({ content, className = '' }: MarkdownPreviewProps) {
  const renderedHTML = useMemo(() => {
    if (!content?.trim()) return ''
    return md.render(content)
  }, [content])

  if (!renderedHTML) {
    return (
      <div className={`mde-preview-empty ${className}`}>
        <span className="text-gray-500 italic">No content to preview</span>
      </div>
    )
  }

  return (
    <div
      className={`mde-prose ${className}`}
      dangerouslySetInnerHTML={{ __html: renderedHTML }}
    />
  )
}
