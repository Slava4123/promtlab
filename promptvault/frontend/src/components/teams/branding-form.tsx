/* eslint-disable react-hooks/set-state-in-effect */
// Prefill async-формы через setState в useEffect — оптимальный вариант
// для случая "данные приходят асинхронно через useQuery". Альтернативы
// (useMemo, key-remount, react-hook-form reset) усложнили бы код без выгоды.
import { useState, useEffect } from "react"
import { Loader2, Save } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { UpgradeGate } from "@/components/analytics/upgrade-gate"
import { useBranding, useUpdateBranding } from "@/hooks/use-branding"
import { ApiError } from "@/api/client"
import type { LogoSource } from "@/api/branding"
import { LogoUploader } from "./logo-uploader"
import { ColorPalettePicker } from "./color-palette-picker"

interface BrandingFormProps {
  slug: string
  planId: string
}

export function BrandingForm({ slug, planId }: BrandingFormProps) {
  const isMax = planId.startsWith("max")
  const { data, isLoading } = useBranding(slug, isMax)
  const updateBranding = useUpdateBranding(slug)

  const [logoUrl, setLogoUrl] = useState("")
  const [logoSource, setLogoSource] = useState<LogoSource>("url")
  const [effectiveLogoUrl, setEffectiveLogoUrl] = useState<string>("")
  const [tagline, setTagline] = useState("")
  const [website, setWebsite] = useState("")
  const [primaryColor, setPrimaryColor] = useState("")

  // Префилл формы когда данные загрузились.
  useEffect(() => {
    if (data) {
      setLogoUrl(data.logo_url ?? "")
      setLogoSource((data.logo_source as LogoSource) ?? "url")
      setEffectiveLogoUrl(data.effective_logo_url ?? "")
      setTagline(data.tagline ?? "")
      setWebsite(data.website ?? "")
      setPrimaryColor(data.primary_color ?? "")
    }
  }, [data])

  if (!isMax) {
    return (
      <UpgradeGate
        title="Брендинг публичных ссылок — фича Max"
        description="На публичных страницах /s/:token будут видны ваш логотип, подпись и фирменный цвет."
        targetPlan="Max"
      />
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    try {
      await updateBranding.mutateAsync({
        logo_url: logoUrl,
        logo_source: logoSource,
        tagline,
        website,
        primary_color: primaryColor,
      })
      toast.success("Брендинг обновлён")
    } catch (err) {
      if (err instanceof ApiError) {
        toast.error(err.message)
      } else {
        toast.error("Не удалось сохранить")
      }
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          Брендинг публичных ссылок
          <Badge variant="outline">Max</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex justify-center py-6">
            <Loader2 className="size-5 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label>Логотип</Label>
              <LogoUploader
                slug={slug}
                logoSource={logoSource}
                logoUrl={logoUrl}
                effectiveLogoUrl={effectiveLogoUrl}
                onLogoUrlChange={setLogoUrl}
                onSourceChange={(next) => setLogoSource(next)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="tagline">Подпись</Label>
              <Input
                id="tagline"
                placeholder="Например: «Библиотека промптов Acme»"
                value={tagline}
                onChange={(e) => setTagline(e.target.value)}
                maxLength={200}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="website">Сайт (URL)</Label>
              <Input
                id="website"
                type="url"
                placeholder="https://example.com"
                value={website}
                onChange={(e) => setWebsite(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">Клик по логотипу откроет этот URL.</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="primary_color">Основной цвет</Label>
              <ColorPalettePicker
                id="primary_color"
                value={primaryColor}
                onChange={setPrimaryColor}
              />
              <p className="text-xs text-muted-foreground">
                Используется для акцентов на странице. Выберите из палитры или задайте свой #RRGGBB.
              </p>
            </div>

            <Button type="submit" disabled={updateBranding.isPending}>
              {updateBranding.isPending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Save className="size-4" />
              )}
              Сохранить
            </Button>
          </form>
        )}
      </CardContent>
    </Card>
  )
}
