interface MarkdownSourceProps {
  content: string
  className?: string
}

export function MarkdownSource({ content, className = '' }: MarkdownSourceProps) {
  return (
    <div className={`mde-source ${className}`}>
      <pre className="mde-source-pre">
        <code>{content}</code>
      </pre>
    </div>
  )
}
