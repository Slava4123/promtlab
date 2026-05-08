import { useRef, useState } from "react"
import { Loader2, Trash2, Upload } from "lucide-react"
import { toast } from "sonner"

import { ApiError } from "@/api/client"
import { useDeleteLogo, useUploadLogo } from "@/hooks/use-branding"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"

// Лимит и whitelist соответствуют backend (usecases/team/logo.go).
const MAX_LOGO_BYTES = 1 << 20
const ALLOWED_MIME = ["image/png", "image/jpeg", "image/webp"] as const
type AllowedMime = (typeof ALLOWED_MIME)[number]

function isAllowedMime(t: string): t is AllowedMime {
  return (ALLOWED_MIME as readonly string[]).includes(t)
}

interface LogoUploaderProps {
  slug: string
  // logoSource приходит из BrandingInfo: 'url' | 'file' | 'none' | undefined.
  logoSource?: string
  // Текущий внешний URL — для отображения preview в URL-режиме и
  // отправки в saveBranding при переключении обратно.
  logoUrl: string
  // EffectiveLogoURL — готовый src для <img> с сервера (либо URL, либо
  // /api/teams/.../branding/logo). Префилл preview в file-режиме без
  // дополнительной логики на стороне фронта.
  effectiveLogoUrl?: string
  onLogoUrlChange: (next: string) => void
  // onSourceChange — фронт меняет source при upload (→ 'file') и delete (→ 'none').
  // Bаренда может вызвать save родительской формы, чтобы поля branding
  // согласовались с серверным состоянием.
  onSourceChange?: (next: "url" | "file" | "none") => void
}

type Mode = "url" | "file"

// LogoUploader — двухрежимный контрол.
//
// Режим «URL»: внешний CDN-link в Input. onLogoUrlChange ведёт значение к
// родителю — финальный save вместе со всеми остальными полями делает
// существующая branding-форма.
//
// Режим «Файл»: drag-and-drop / click-to-pick. На select валидируем MIME и
// размер на клиенте (быстрая обратная связь без сетевого round-trip), затем
// мутацией useUploadLogo шлём multipart на бэк. На успех — preview через
// effective_logo_url из ответа. На ошибку — toast (сообщение от бэка
// или локальный fallback).
export function LogoUploader({
  slug,
  logoSource,
  logoUrl,
  effectiveLogoUrl,
  onLogoUrlChange,
  onSourceChange,
}: LogoUploaderProps) {
  const initialMode: Mode = logoSource === "file" ? "file" : "url"
  const [mode, setMode] = useState<Mode>(initialMode)
  const [prevLogoSource, setPrevLogoSource] = useState(logoSource)
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const upload = useUploadLogo(slug)
  const remove = useDeleteLogo(slug)

  // MJ-1: «storing prop in state» pattern с setState during render вместо
  // useEffect — рекомендуемый React idiom для синхронизации с пропом
  // (https://react.dev/learn/you-might-not-need-an-effect#adjusting-some-state-when-a-prop-changes).
  // Если родитель перезагружает branding (после save), mode синхронизируется
  // с серверным logo_source — иначе юзер мог застрять в URL-режиме после
  // успешной загрузки файла.
  if (prevLogoSource !== logoSource) {
    setPrevLogoSource(logoSource)
    if (logoSource === "file") setMode("file")
    else if (logoSource === "url") setMode("url")
  }

  const handleFile = (file: File) => {
    if (!isAllowedMime(file.type)) {
      toast.error("Поддерживаются только PNG, JPEG и WebP")
      return
    }
    if (file.size > MAX_LOGO_BYTES) {
      toast.error("Файл больше 1 МБ")
      return
    }
    upload.mutate(file, {
      onSuccess: (data) => {
        toast.success("Логотип загружен")
        onSourceChange?.(data.logo_source)
      },
      onError: (err) => {
        // 402 уже показан quota-store; toast не дублируем.
        if (err instanceof ApiError && err.status === 402) return
        toast.error(err instanceof Error ? err.message : "Не удалось загрузить")
      },
    })
  }

  const handleDelete = () => {
    remove.mutate(undefined, {
      onSuccess: (data) => {
        toast.success("Логотип удалён")
        onSourceChange?.(data.logo_source)
      },
      onError: (err) => {
        toast.error(err instanceof Error ? err.message : "Не удалось удалить")
      },
    })
  }

  const onDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    if (!dragOver) setDragOver(true)
  }
  const onDragLeave = () => setDragOver(false)
  const onDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files?.[0]
    if (file) handleFile(file)
  }

  const isUploading = upload.isPending
  const isDeleting = remove.isPending
  const previewUrl = mode === "file" ? effectiveLogoUrl : logoUrl

  return (
    <div className="space-y-3">
      {/* Mode tabs — radiogroup не используем (нативные inputs создают шум
          screen-reader'ам), плоские button'ы с aria-pressed читаются чисто. */}
      <div className="inline-flex rounded-md border border-input p-0.5">
        <Button
          type="button"
          variant={mode === "url" ? "secondary" : "ghost"}
          size="sm"
          aria-pressed={mode === "url"}
          onClick={() => setMode("url")}
          className="h-7 px-3 text-[0.78rem]"
        >
          URL
        </Button>
        <Button
          type="button"
          variant={mode === "file" ? "secondary" : "ghost"}
          size="sm"
          aria-pressed={mode === "file"}
          onClick={() => setMode("file")}
          className="h-7 px-3 text-[0.78rem]"
        >
          Загрузить файл
        </Button>
      </div>

      {mode === "url" ? (
        <div className="space-y-1.5">
          <Label htmlFor="logo_url" className="sr-only">
            Логотип (URL)
          </Label>
          <Input
            id="logo_url"
            type="url"
            placeholder="https://cdn.example.com/logo.png"
            value={logoUrl}
            onChange={(e) => onLogoUrlChange(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            HTTPS-URL изображения. Рекомендуемый размер 200×60px.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          <div
            onDragOver={onDragOver}
            onDragLeave={onDragLeave}
            onDrop={onDrop}
            onClick={() => fileInputRef.current?.click()}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault()
                fileInputRef.current?.click()
              }
            }}
            className={cn(
              "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-6 text-center transition-colors",
              dragOver ? "border-ring bg-muted/40" : "border-input hover:bg-muted/20",
              isUploading && "pointer-events-none opacity-60",
            )}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept={ALLOWED_MIME.join(",")}
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) handleFile(file)
                // Сбрасываем value, чтобы повторный выбор того же файла
                // (после удаления) тоже триггернул onChange.
                e.target.value = ""
              }}
            />
            {isUploading ? (
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            ) : (
              <Upload className="size-6 text-muted-foreground" />
            )}
            <p className="text-sm font-medium text-foreground">
              {isUploading ? "Загружаем…" : "Перетащите файл или нажмите для выбора"}
            </p>
            <p className="text-xs text-muted-foreground">
              PNG, JPEG или WebP, до 1 МБ, не больше 1024×1024 px
            </p>
          </div>

          {previewUrl && !isUploading && (
            <div className="flex items-center gap-3 rounded-md border border-input bg-muted/20 p-3">
              <img
                src={previewUrl}
                alt="Превью логотипа"
                className="h-10 max-w-[120px] rounded-sm object-contain"
              />
              <span className="flex-1 truncate text-xs text-muted-foreground">
                {logoSource === "file" ? "Файл загружен на сервер" : "Внешний URL"}
              </span>
              {logoSource === "file" && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="h-7 text-xs text-muted-foreground hover:text-destructive"
                >
                  {isDeleting ? (
                    <Loader2 className="size-3.5 animate-spin" />
                  ) : (
                    <Trash2 className="size-3.5" />
                  )}
                  Удалить
                </Button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
