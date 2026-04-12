import { type ReactNode } from "react"

interface HighlightMatchProps {
  text: string
  query: string
  className?: string
}

function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
}

export function HighlightMatch({ text, query, className }: HighlightMatchProps): ReactNode {
  if (!query || !text) return text ?? null

  const escaped = escapeRegExp(query.trim())
  if (!escaped) return text

  const parts = text.split(new RegExp(`(${escaped})`, "gi"))

  if (parts.length === 1) return text

  return (
    <span className={className}>
      {parts.map((part, i) =>
        part.toLowerCase() === query.trim().toLowerCase() ? (
          <mark
            key={i}
            className="bg-brand-muted text-brand-muted-foreground rounded-sm px-0.5"
          >
            {part}
          </mark>
        ) : (
          part
        ),
      )}
    </span>
  )
}
