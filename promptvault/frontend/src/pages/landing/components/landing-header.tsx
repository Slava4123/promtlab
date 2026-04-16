import { useState, useEffect, useCallback } from "react"
import { Link } from "react-router-dom"
import { Lock, Menu } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Sheet, SheetTrigger, SheetContent } from "@/components/ui/sheet"
import { cn } from "@/lib/utils"
import { navLinks } from "../data/landing-content"

export function LandingHeader() {
  const [scrolled, setScrolled] = useState(false)
  const [sheetOpen, setSheetOpen] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    window.addEventListener("scroll", onScroll, { passive: true })
    return () => window.removeEventListener("scroll", onScroll)
  }, [])

  const handleNavClick = useCallback(() => {
    setSheetOpen(false)
  }, [])

  return (
    <header
      className={cn(
        "fixed inset-x-0 top-0 z-50 transition-all duration-300",
        scrolled
          ? "border-b border-border/50 bg-background/80 backdrop-blur-xl"
          : "bg-transparent",
      )}
    >
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        {/* Лого */}
        <Link to="/" className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500/25 to-violet-600/5 ring-1 ring-violet-500/15">
            <Lock className="h-4 w-4 text-violet-400" />
          </div>
          <span className="text-[0.95rem] font-semibold tracking-tight">ПромтЛаб</span>
        </Link>

        {/* Десктоп навигация */}
        <nav className="hidden items-center gap-6 text-sm text-muted-foreground sm:flex">
          {navLinks.map(link => (
            <a key={link.href} href={link.href} className="transition-colors hover:text-foreground">
              {link.label}
            </a>
          ))}
          <Link to="/help" className="transition-colors hover:text-foreground">
            Помощь
          </Link>
        </nav>

        {/* Десктоп кнопки */}
        <div className="hidden items-center gap-2 sm:flex">
          <Button variant="ghost" size="sm" nativeButton={false} render={<Link to="/sign-in" />}>
            Войти
          </Button>
          <Button variant="brand" size="sm" nativeButton={false} render={<Link to="/sign-up" />}>
            Начать бесплатно
          </Button>
        </div>

        {/* Мобильное меню */}
        <div className="flex items-center gap-2 sm:hidden">
          <Button variant="brand" size="sm" nativeButton={false} render={<Link to="/sign-up" />}>
            Начать
          </Button>
          <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
            <SheetTrigger render={<Button variant="ghost" size="icon-sm" />}>
              <Menu className="h-4 w-4" />
            </SheetTrigger>
            <SheetContent side="right" className="dark w-64 pt-12">
              <nav className="flex flex-col gap-4 px-2">
                {navLinks.map(link => (
                  <a
                    key={link.href}
                    href={link.href}
                    onClick={handleNavClick}
                    className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                  >
                    {link.label}
                  </a>
                ))}
                <Link
                  to="/help"
                  onClick={handleNavClick}
                  className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                >
                  Помощь
                </Link>
                <hr className="border-border/30" />
                <Link
                  to="/sign-in"
                  onClick={handleNavClick}
                  className="text-sm text-muted-foreground hover:text-foreground"
                >
                  Войти
                </Link>
              </nav>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  )
}
