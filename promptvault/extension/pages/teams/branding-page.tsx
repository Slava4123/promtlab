import { useEffect, useRef, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  Image as ImageIcon,
  Loader2,
  Save,
  Trash2,
  Upload,
  Palette,
} from "lucide-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { uploadTeamLogoDirect } from "../../lib/api"
import { cn } from "../../lib/utils"

const BRAND_COLORS = [
  "#7c3aed", "#dc2626", "#ea580c", "#d97706",
  "#65a30d", "#059669", "#0891b2", "#2563eb",
  "#4f46e5", "#9333ea", "#c026d3", "#db2777",
] as const

const MAX_LOGO_BYTES = 512 * 1024
const ACCEPT_MIME = "image/png,image/jpeg,image/webp"

export function TeamBrandingPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [deleteLogoOpen, setDeleteLogoOpen] = useState(false)

  const brandingQuery = useQuery({
    queryKey: ["branding", slug],
    queryFn: () => sendBg({ type: "api.getTeamBranding", slug: slug ?? "" }),
    enabled: Boolean(slug),
    staleTime: 60_000,
  })

  const [tagline, setTagline] = useState("")
  const [website, setWebsite] = useState("")
  const [primaryColor, setPrimaryColor] = useState("#7c3aed")

  useEffect(() => {
    if (brandingQuery.data) {
      setTagline(brandingQuery.data.tagline ?? "")
      setWebsite(brandingQuery.data.website ?? "")
      setPrimaryColor(brandingQuery.data.primary_color ?? "#7c3aed")
    }
  }, [brandingQuery.data])

  const saveMut = useMutation({
    mutationFn: () =>
      sendBg({
        type: "api.updateTeamBranding",
        slug: slug ?? "",
        body: { tagline, website, primary_color: primaryColor },
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["branding", slug] })
      toast({ title: "Брендинг сохранён", variant: "success" })
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось сохранить", description: err.message, variant: "error" }),
  })

  const uploadMut = useMutation({
    mutationFn: (file: File) => uploadTeamLogoDirect(slug ?? "", file),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["branding", slug] })
      toast({ title: "Логотип загружен", variant: "success" })
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось загрузить", description: err.message, variant: "error" }),
  })

  const deleteLogoMut = useMutation({
    mutationFn: () => sendBg({ type: "api.deleteTeamLogo", slug: slug ?? "" }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["branding", slug] })
      toast({ title: "Логотип удалён", variant: "info" })
      setDeleteLogoOpen(false)
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось удалить", description: err.message, variant: "error" }),
  })

  async function handleFile(file: File) {
    if (file.size > MAX_LOGO_BYTES) {
      toast({
        title: "Слишком большой файл",
        description: `Максимум ${(MAX_LOGO_BYTES / 1024).toFixed(0)} КБ`,
        variant: "error",
      })
      return
    }
    if (!ACCEPT_MIME.split(",").includes(file.type)) {
      toast({ title: "Неподдерживаемый формат", description: "PNG, JPEG или WebP", variant: "error" })
      return
    }
    uploadMut.mutate(file)
  }

  if (!slug) return null

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Брендинг команды</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {brandingQuery.isPending ? (
          <div className="flex justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (
          <>
            {/* Logo */}
            <section className="space-y-2">
              <label className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                Логотип
              </label>
              <div className="flex items-center gap-3 rounded-md border border-(--color-border) bg-(--color-card) p-3">
                <div
                  className="flex h-14 w-14 shrink-0 items-center justify-center rounded-md border border-(--color-border) bg-(--color-muted)/30"
                  style={{ borderColor: primaryColor }}
                >
                  {brandingQuery.data?.effective_logo_url ? (
                    <img
                      src={brandingQuery.data.effective_logo_url}
                      alt="Logo"
                      className="h-12 w-12 rounded object-cover"
                    />
                  ) : (
                    <ImageIcon className="h-5 w-5 text-(--color-muted-foreground)" />
                  )}
                </div>
                <div className="flex-1 space-y-1">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={uploadMut.isPending}
                    className="gap-1.5 w-full"
                  >
                    {uploadMut.isPending ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <Upload className="h-3.5 w-3.5" />
                    )}
                    Загрузить
                  </Button>
                  {brandingQuery.data?.logo_source === "file" && (
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      onClick={() => setDeleteLogoOpen(true)}
                      className="gap-1.5 w-full text-(--color-destructive)"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Удалить
                    </Button>
                  )}
                </div>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept={ACCEPT_MIME}
                  className="hidden"
                  onChange={(e) => {
                    const f = e.target.files?.[0]
                    if (f) void handleFile(f)
                    e.target.value = ""
                  }}
                />
              </div>
              <p className="text-[9px] text-(--color-muted-foreground)">
                PNG / JPEG / WebP, до {(MAX_LOGO_BYTES / 1024).toFixed(0)} КБ. Max-only фича.
              </p>
            </section>

            {/* Tagline */}
            <section className="space-y-1.5">
              <Label htmlFor="tagline" className="text-xs">Слоган</Label>
              <Input
                id="tagline"
                value={tagline}
                onChange={(e) => setTagline(e.target.value)}
                placeholder="AI-команда мечты"
                maxLength={120}
              />
            </section>

            {/* Website */}
            <section className="space-y-1.5">
              <Label htmlFor="website" className="text-xs">Сайт</Label>
              <Input
                id="website"
                type="url"
                value={website}
                onChange={(e) => setWebsite(e.target.value)}
                placeholder="https://example.com"
              />
            </section>

            {/* Color picker */}
            <section className="space-y-2">
              <div className="flex items-center gap-1.5">
                <Palette className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
                <label className="text-xs font-medium">Основной цвет</label>
              </div>
              <div className="grid grid-cols-6 gap-1.5">
                {BRAND_COLORS.map((c) => (
                  <button
                    key={c}
                    type="button"
                    onClick={() => setPrimaryColor(c)}
                    className={cn(
                      "h-7 w-7 rounded-md border-2 transition-all",
                      primaryColor === c
                        ? "border-(--color-foreground) scale-105"
                        : "border-transparent hover:border-(--color-muted)",
                    )}
                    style={{ backgroundColor: c }}
                    aria-label={`Цвет ${c}`}
                  />
                ))}
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={primaryColor}
                  onChange={(e) => setPrimaryColor(e.target.value)}
                  className="h-7 w-12 cursor-pointer rounded border border-(--color-border) bg-(--color-card)"
                />
                <span className="font-mono text-[10px] text-(--color-muted-foreground)">
                  {primaryColor}
                </span>
              </div>
            </section>
          </>
        )}
      </div>

      <div className="border-t border-(--color-border) p-2">
        <Button
          type="button"
          onClick={() => saveMut.mutate()}
          disabled={saveMut.isPending || brandingQuery.isPending}
          className="w-full gap-1.5"
        >
          {saveMut.isPending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Save className="h-3.5 w-3.5" />
          )}
          Сохранить
        </Button>
      </div>

      <ConfirmDialog
        open={deleteLogoOpen}
        title="Удалить логотип?"
        description="Логотип удалится из команды."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={() => deleteLogoMut.mutate()}
        onClose={() => setDeleteLogoOpen(false)}
      />
    </div>
  )
}
