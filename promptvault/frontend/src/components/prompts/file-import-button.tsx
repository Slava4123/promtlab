import { useRef } from "react"
import { Upload, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ACCEPTED_FILE_EXTENSIONS } from "@/lib/file-import/constants"

interface FileImportButtonProps {
  onFileSelect: (file: File) => void
  isImporting: boolean
  disabled?: boolean
}

/**
 * Кнопка загрузки файла для редактора промпта. Скрытый `<input type="file">`
 * триггерится программным кликом. После выбора — вызывает `onFileSelect(file)`.
 * Во время парсинга (`isImporting=true`) — spinner и disabled.
 *
 * Accessibility: явный label, `aria-busy`, keyboard-friendly (Space/Enter).
 */
export function FileImportButton({
  onFileSelect,
  isImporting,
  disabled = false,
}: FileImportButtonProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  const handleClick = () => {
    inputRef.current?.click()
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    // Reset значения input, чтобы выбор того же файла повторно триггерил onChange.
    e.target.value = ""
    if (file) {
      onFileSelect(file)
    }
  }

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={handleClick}
        disabled={disabled || isImporting}
        aria-busy={isImporting}
        aria-label="Загрузить файл с содержимым промпта"
        className="h-7 gap-1.5 text-[0.75rem]"
      >
        {isImporting ? (
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
        ) : (
          <Upload className="h-3.5 w-3.5" />
        )}
        {isImporting ? "Парсинг…" : "Загрузить файл"}
      </Button>
      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_FILE_EXTENSIONS}
        onChange={handleChange}
        className="hidden"
        aria-hidden="true"
        tabIndex={-1}
      />
    </>
  )
}
