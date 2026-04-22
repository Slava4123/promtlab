import { useEffect, useState, useCallback } from "react"
import { FileDown } from "lucide-react"

interface FileImportDropZoneProps {
  onFileDrop: (file: File) => void
  disabled?: boolean
}

/**
 * Полноэкранный оверлей, показывается когда пользователь перетаскивает файл на
 * страницу. При drop — вызывает onFileDrop(file). Слушает document-level события,
 * так что drop работает куда угодно на странице (не только над редактором).
 *
 * Mobile: не активируется (touch-drag не триггерит dragover в браузерах).
 */
export function FileImportDropZone({
  onFileDrop,
  disabled = false,
}: FileImportDropZoneProps) {
  const [visible, setVisible] = useState(false)
  // Счётчик вложенных dragenter/dragleave событий — без него overlay мигает
  // когда курсор проходит над дочерними элементами.
  const [, setDragDepth] = useState(0)

  const handleDragEnter = useCallback(
    (e: DragEvent) => {
      if (disabled) return
      if (!e.dataTransfer?.types?.includes("Files")) return
      e.preventDefault()
      setDragDepth((d) => d + 1)
      setVisible(true)
    },
    [disabled],
  )

  const handleDragOver = useCallback(
    (e: DragEvent) => {
      if (disabled) return
      if (!e.dataTransfer?.types?.includes("Files")) return
      // preventDefault обязателен чтобы drop сработал.
      e.preventDefault()
      if (e.dataTransfer) {
        e.dataTransfer.dropEffect = "copy"
      }
    },
    [disabled],
  )

  const handleDragLeave = useCallback(
    (e: DragEvent) => {
      if (disabled) return
      if (!e.dataTransfer?.types?.includes("Files")) return
      e.preventDefault()
      setDragDepth((d) => {
        const next = d - 1
        if (next <= 0) setVisible(false)
        return Math.max(0, next)
      })
    },
    [disabled],
  )

  const handleDrop = useCallback(
    (e: DragEvent) => {
      if (disabled) return
      e.preventDefault()
      setVisible(false)
      setDragDepth(0)
      const file = e.dataTransfer?.files?.[0]
      if (file) onFileDrop(file)
    },
    [disabled, onFileDrop],
  )

  useEffect(() => {
    document.addEventListener("dragenter", handleDragEnter)
    document.addEventListener("dragover", handleDragOver)
    document.addEventListener("dragleave", handleDragLeave)
    document.addEventListener("drop", handleDrop)
    return () => {
      document.removeEventListener("dragenter", handleDragEnter)
      document.removeEventListener("dragover", handleDragOver)
      document.removeEventListener("dragleave", handleDragLeave)
      document.removeEventListener("drop", handleDrop)
    }
  }, [handleDragEnter, handleDragOver, handleDragLeave, handleDrop])

  if (!visible) return null

  return (
    <div
      className="pointer-events-none fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm"
      role="region"
      aria-label="Зона загрузки файла"
    >
      <div className="flex flex-col items-center gap-3 rounded-2xl border-2 border-dashed border-violet-500/60 bg-card/90 px-10 py-8 shadow-2xl">
        <FileDown className="h-12 w-12 text-violet-400" />
        <p className="text-lg font-semibold text-foreground">Отпустите для загрузки</p>
        <p className="text-sm text-muted-foreground">
          Поддерживаются .txt, .md, .json, .pdf, .docx и другие
        </p>
      </div>
    </div>
  )
}
