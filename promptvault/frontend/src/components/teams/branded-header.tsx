import type { BrandingInfo } from "@/api/branding"
import { isSafeHttpsUrl } from "@/lib/url"

interface BrandedHeaderProps {
  branding: BrandingInfo
}

// BrandedHeader — рендерится на /s/:token если PublicPromptInfo.branding != null.
// Настраивается в Team settings (Max-only).
export function BrandedHeader({ branding }: BrandedHeaderProps) {
  const primaryColor = branding.primary_color || undefined
  const headerStyle = primaryColor ? { borderBottomColor: primaryColor } : undefined

  const logo = branding.logo_url ? (
    <img
      src={branding.logo_url}
      alt="Logo"
      referrerPolicy="no-referrer"
      className="h-12 w-auto object-contain"
      loading="lazy"
    />
  ) : null

  // Defense-in-depth: backend уже валидирует схему, но если прорвётся
  // javascript:/data:/file: — не рендерим <a>, показываем просто логотип.
  const logoContainer = isSafeHttpsUrl(branding.website) ? (
    <a href={branding.website} target="_blank" rel="noopener noreferrer">
      {logo}
    </a>
  ) : logo

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
