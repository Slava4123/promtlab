import { useNavigate } from "react-router-dom"
import { Home } from "../components/home"
import { Plus } from "lucide-react"
import { Button } from "../components/ui/button"

// Главная страница — список промптов с pinned/recent секциями.
// Реализация в components/home.tsx; здесь router-wrapper.
// Клик на карточку ведёт на detail (а не сразу на use), чтобы юзер видел весь
// промпт, мог Edit/Delete/Share. "Использовать" — кнопка в detail.
export function DashboardPage() {
  const navigate = useNavigate()
  return (
    <div className="relative h-full">
      <Home
        onSelect={(p) => navigate(`/prompts/${p.id}`)}
        onOpenSettings={() => navigate("/settings")}
        highlightedId={null}
      />
      {/* Floating "+" button — создать новый промпт. Brand-CTA: фиолетовый
          из identity, тень с brand-shadow. 44×44 (WCAG touch-target). */}
      <Button
        type="button"
        variant="brand"
        onClick={() => navigate("/prompts/new")}
        className="absolute bottom-4 right-4 h-11 w-11 rounded-full p-0"
        aria-label="Создать промпт"
        title="Создать промпт"
      >
        <Plus className="h-5 w-5" />
      </Button>
    </div>
  )
}
