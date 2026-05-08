import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import rehypeSanitize, { defaultSchema } from "rehype-sanitize"
import rehypeHighlight from "rehype-highlight"
import javascript from "highlight.js/lib/languages/javascript"
import typescript from "highlight.js/lib/languages/typescript"
import python from "highlight.js/lib/languages/python"
import go from "highlight.js/lib/languages/go"
import bash from "highlight.js/lib/languages/bash"
import json from "highlight.js/lib/languages/json"
import sql from "highlight.js/lib/languages/sql"
import markdown from "highlight.js/lib/languages/markdown"
import css from "highlight.js/lib/languages/css"
import xml from "highlight.js/lib/languages/xml"
import type { ComponentProps } from "react"
import "./prompt-content.css"
import { cn } from "@/lib/utils"

// MJ-17: до фикса rehype-highlight без options автоопределял язык по
// 192 встроенным грамматикам — bundle-чанк vendor-markdown ~325 KB.
// Передавая explicit languages map (10 языков покрывают 95% реальных
// промптов), tree-shake режет неиспользуемые грамматики до ~20-30 KB.
const highlightLanguages = {
  javascript,
  typescript,
  python,
  go,
  bash,
  json,
  sql,
  markdown,
  css,
  xml, // включает HTML, SVG, plist (Highlight.js под общей грамматикой)
}

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
        rehypePlugins={[
          [rehypeSanitize, sanitizeSchema],
          [rehypeHighlight, { languages: highlightLanguages, detect: false }],
        ]}
        components={{ a: AnchorLink }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
