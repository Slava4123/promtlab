import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import rehypeSanitize, { defaultSchema } from "rehype-sanitize"
import rehypeHighlight from "rehype-highlight"
import type { ComponentProps } from "react"
import "./prompt-content.css"
import { cn } from "@/lib/utils"

// sanitizeSchema расширяет defaultSchema чтобы:
//   - пропустить classNames `language-*`, `hljs`, `hljs-*` на <code> (иначе rehypeHighlight классы удаляются)
//   - разрешить <input type="checkbox" disabled checked?> для GFM task-lists
// Протоколы href оставляем дефолтные (http/https/mailto/...) — javascript: блокируется.
const sanitizeSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    code: [
      ...(defaultSchema.attributes?.code ?? []),
      ["className", /^language-/, "hljs", /^hljs-/],
    ],
    pre: [
      ...(defaultSchema.attributes?.pre ?? []),
      ["className", "hljs"],
    ],
    span: [
      ...(defaultSchema.attributes?.span ?? []),
      ["className", /^hljs-/],
    ],
    input: [
      ...(defaultSchema.attributes?.input ?? []),
      ["type"],
      ["checked"],
      ["disabled"],
    ],
  },
  tagNames: [...(defaultSchema.tagNames ?? []), "input"],
}

interface PromptContentProps {
  content: string
  className?: string
}

// AnchorLink — принудительно target="_blank" + rel="noopener noreferrer"
// для всех ссылок внутри промпта, защита от tabnabbing.
function AnchorLink(props: ComponentProps<"a">) {
  return <a {...props} target="_blank" rel="noopener noreferrer" />
}

/**
 * PromptContent — production-grade Markdown рендер содержимого промпта.
 * Поддержка: GFM (таблицы, task-lists, strikethrough), syntax highlight через highlight.js,
 * XSS-санитайзинг через rehype-sanitize, типографика через @tailwindcss/typography.
 *
 * Используется в shared-prompt, public-prompt, preview-диалогах.
 */
export function PromptContent({ content, className }: PromptContentProps) {
  return (
    <div className={cn("prose prose-neutral dark:prose-invert max-w-none break-words", className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[[rehypeSanitize, sanitizeSchema], rehypeHighlight]}
        components={{ a: AnchorLink }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
