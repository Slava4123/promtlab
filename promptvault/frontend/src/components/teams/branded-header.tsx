import type { BrandingInfo } from "@/api/branding"
import { isSafeHttpsUrl } from "@/lib/url"

interface BrandedHeaderProps {
  branding: BrandingInfo
}

// BrandedHeader — рендерится на /s/:token если PublicPromptInfo.branding != null.
// Настраивается в Team settings (Max-only).
//
// Phase 16-X: используем branding.effective_logo_url (резолвинг между
// внешним URL и uploaded-file делает backend в buildBrandingInfo).
// Fallback на legacy logo_url для backward compat: старый клиент ещё в
// деплое или legacy ссылки без logo_source.
export function BrandedHeader({ branding }: BrandedHeaderProps) {
  const primaryColor = branding.primary_color || undefined
  const headerStyle = primaryColor ? { borderBottomColor: primaryColor } : undefined

  const logoSrc = branding.effective_logo_url || branding.logo_url || ""
  const logo = logoSrc ? (
    <img
      src={logoSrc}
      alt="Logo"
      referrerPolicy="no-referrer"
      className="h-12 w-auto object-contain"
      loading="lazy"
    />
  ) : null

  // Defense-in-depth: backend уже валидирует схему, но если прорвётся
  // javascript:/data:/file: — не рендерим <a>, показываем просто логотип.
  // Если логотипа нет, но website задан — показываем кликабельный текст «Сайт»,
  // чтобы юзер не терял ссылку при file-upload без preview-ошибок.
  const websiteSafe = isSafeHttpsUrl(branding.website)
  let logoContainer: React.ReactNode = logo
  if (logo && websiteSafe) {
    logoContainer = (
      <a href={branding.website} target="_blank" rel="noopener noreferrer">
        {logo}
      </a>
    )
  } else if (!logo && websiteSafe) {
    logoContainer = (
      <a
        href={branding.website}
        target="_blank"
        rel="noopener noreferrer"
        className="text-sm font-medium text-foreground underline-offset-2 hover:underline"
      >
        Перейти на сайт
      </a>
    )
  }

  return (
    <div
      className="mb-6 flex items-center gap-4 border-b-2 pb-4"
      style={headerStyle}
    >
      {logoContainer}
      {branding.tagline && (
        <p className="text-sm text-muted-foreground">{branding.tagline}</p>
      )}
    </div>
  )
}
