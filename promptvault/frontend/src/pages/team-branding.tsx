import { Link, useParams, useNavigate } from "react-router-dom"
import { ArrowLeft } from "lucide-react"
import { buttonVariants } from "@/components/ui/button"
import { toast } from "sonner"
import { useAuthStore } from "@/stores/auth-store"
import { useTeam } from "@/hooks/use-teams"
import { BrandingForm } from "@/components/teams/branding-form"
import { ApiError } from "@/api/client"

// Phase 14 D: /teams/:slug/branding — страница настройки брендинга для owner команды.
export default function TeamBrandingPage() {
  const { slug = "" } = useParams()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const planId = user?.plan_id ?? "free"

  const { data: team, isLoading, error } = useTeam(slug)

  if (error instanceof ApiError && error.status === 403) {
    toast.error("Нет доступа к команде")
    navigate("/teams")
    return null
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <span className="text-sm text-muted-foreground">Загрузка...</span>
      </div>
    )
  }

  if (!team) return null

  if (team.role !== "owner") {
    toast.error("Только владелец команды может настраивать брендинг")
    navigate(`/teams/${slug}`)
    return null
  }

  return (
    <div className="container mx-auto max-w-2xl space-y-6 px-4 py-8">
      <div className="flex items-center gap-3">
        <Link
          to={`/teams/${slug}`}
          className={buttonVariants({ variant: "ghost", size: "sm" })}
        >
          <ArrowLeft className="size-4" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold">Брендинг команды</h1>
          <p className="text-sm text-muted-foreground">
            Настройте внешний вид публичных share-ссылок
          </p>
        </div>
      </div>

      <BrandingForm slug={slug} planId={planId} />
    </div>
  )
}
