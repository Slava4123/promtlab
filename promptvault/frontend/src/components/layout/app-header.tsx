import { SidebarTrigger } from "@/components/ui/sidebar"

export function AppHeader() {
  return (
    <header className="flex h-11 items-center border-b border-white/[0.04] px-4">
      <SidebarTrigger />
    </header>
  )
}
