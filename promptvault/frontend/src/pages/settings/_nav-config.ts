import {
  User,
  Lock,
  Link2,
  CreditCard,
  Gift,
  Plug,
  Palette,
  type LucideIcon,
} from "lucide-react"

export type SettingsNavItem = {
  id: string
  title: string
  icon: LucideIcon
  to: string
}

export const NAV_ITEMS: SettingsNavItem[] = [
  { id: "profile", title: "Профиль", icon: User, to: "/settings/profile" },
  { id: "security", title: "Безопасность", icon: Lock, to: "/settings/security" },
  { id: "accounts", title: "Аккаунты", icon: Link2, to: "/settings/accounts" },
  { id: "subscription", title: "Подписка", icon: CreditCard, to: "/settings/subscription" },
  { id: "referral", title: "Рефералы", icon: Gift, to: "/settings/referral" },
  { id: "integrations", title: "Интеграции", icon: Plug, to: "/settings/integrations" },
  { id: "appearance", title: "Оформление", icon: Palette, to: "/settings/appearance" },
]
