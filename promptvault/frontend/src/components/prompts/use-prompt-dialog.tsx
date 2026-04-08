import { useMemo, useState } from "react"
import { useForm } from "react-hook-form"
import { Copy, Check } from "lucide-react"
import { toast } from "sonner"

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { extractVariables, renderTemplate } from "@/lib/template/parse"
import { useIncrementUsage } from "@/hooks/use-prompts"
import type { Prompt } from "@/api/types"

const STORAGE_KEY_PREFIX = "promptvault:prompt-vars:"

/**
 * Loads previously-saved variable values for a prompt, filtered to the current
 * variable set (so renamed/removed variables don't leak back in). Silent on
 * any storage failure — returns empty values as the fallback.
 */
function loadSavedValues(promptId: number, variables: string[]): Record<string, string> {
  const empty = Object.fromEntries(variables.map((n) => [n, ""]))
  try {
    const raw = localStorage.getItem(STORAGE_KEY_PREFIX + promptId)
    if (!raw) return empty
    const parsed = JSON.parse(raw) as unknown
    if (typeof parsed !== "object" || parsed === null) return empty
    const obj = parsed as Record<string, unknown>
    return Object.fromEntries(
      variables.map((n) => [n, typeof obj[n] === "string" ? (obj[n] as string) : ""]),
    )
  } catch {
    return empty
  }
}

interface UsePromptDialogProps {
  prompt: Prompt
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsePromptDialog({ prompt, open, onOpenChange }: UsePromptDialogProps) {
  const variables = useMemo(() => extractVariables(prompt.content), [prompt.content])

  // Pre-fill with the last values the user entered for this prompt.
  // useForm reads defaultValues once on mount — since the parent remounts the
  // dialog on each open (`{usePromptModal && <UsePromptDialog ... />}`),
  // each open gets a fresh read.
  const defaultValues = useMemo(
    () => loadSavedValues(prompt.id, variables),
    [prompt.id, variables],
  )

  const { register, watch } = useForm<Record<string, string>>({ defaultValues })

  // Watch entire form — re-renders on every change, which is what we want for live preview.
  // eslint-disable-next-line react-hooks/incompatible-library
  const values = watch()

  const preview = useMemo(
    () => renderTemplate(prompt.content, values),
    [prompt.content, values],
  )

  // Soft validation — counts non-blank (trimmed) fields. Not a blocker:
  // empty is a legitimate choice, but we show progress and a warning toast.
  const filledCount = useMemo(
    () => variables.filter((n) => (values[n] ?? "").trim() !== "").length,
    [variables, values],
  )
  const hasEmpty = variables.length > 0 && filledCount < variables.length

  const incrementUsage = useIncrementUsage()
  const [copyState, setCopyState] = useState<"idle" | "copied">("idle")
  const [clipboardBlocked, setClipboardBlocked] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(preview)
      setClipboardBlocked(false)
      setCopyState("copied")
      // Fire-and-forget: don't await, don't surface errors to user.
      incrementUsage.mutate(prompt.id)

      // Persist non-blank values for next time. Silent on quota/disabled.
      // Skip entirely when every field is blank — avoids overwriting a
      // previously-useful set of values with emptiness.
      try {
        const toSave: Record<string, string> = {}
        for (const name of variables) {
          const v = (values[name] ?? "").trim()
          if (v) toSave[name] = v
        }
        if (Object.keys(toSave).length > 0) {
          localStorage.setItem(STORAGE_KEY_PREFIX + prompt.id, JSON.stringify(toSave))
        }
      } catch {
        /* quota exceeded or storage disabled — ignore */
      }

      toast.success("Скопировано")
      if (hasEmpty) {
        toast.warning(`Не заполнено переменных: ${variables.length - filledCount}`)
      }
      // Reset label after 2 seconds.
      setTimeout(() => setCopyState("idle"), 2000)
    } catch {
      setClipboardBlocked(true)
      toast.error("Не удалось скопировать автоматически. Выделите текст и нажмите Ctrl+C")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Использовать промпт</DialogTitle>
          <DialogDescription>
            {prompt.title} — заполните переменные ниже
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 max-h-[60vh] overflow-y-auto -mx-1 px-1">
          {/* Variable inputs */}
          {variables.length > 0 && (
            <div className="space-y-3">
              {variables.map((name) => (
                <div key={name} className="space-y-1.5">
                  <Label
                    htmlFor={`var-${name}`}
                    className="flex items-center gap-1.5 text-xs font-mono text-muted-foreground"
                  >
                    {`{{${name}}}`}
                    {(values[name] ?? "").trim() === "" && (
                      <span
                        className="text-amber-500/70"
                        aria-label="Не заполнено"
                        title="Не заполнено"
                      >
                        ○
                      </span>
                    )}
                  </Label>
                  <Textarea
                    id={`var-${name}`}
                    rows={2}
                    placeholder={`Значение для ${name}`}
                    className="text-sm"
                    {...register(name)}
                  />
                </div>
              ))}
            </div>
          )}

          {/* Preview */}
          <div className="space-y-1.5">
            <Label className="text-xs font-medium text-muted-foreground">
              Предпросмотр
            </Label>
            <Textarea
              value={preview}
              readOnly
              rows={6}
              className={`text-sm leading-relaxed ${clipboardBlocked ? "ring-2 ring-amber-500/30" : ""}`}
            />
            {clipboardBlocked && (
              <p className="text-[0.7rem] text-amber-400">
                Буфер обмена недоступен. Выделите текст выше и нажмите Ctrl+C.
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Закрыть
          </Button>
          <Button onClick={handleCopy}>
            {copyState === "copied" ? (
              <>
                <Check className="h-4 w-4" />
                Скопировано
              </>
            ) : (
              <>
                <Copy className="h-4 w-4" />
                {hasEmpty
                  ? `Скопировать (${filledCount} / ${variables.length})`
                  : "Скопировать"}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
