import { useState, useMemo } from "react"
import { Link, useParams, useNavigate } from "react-router-dom"
import { ArrowLeft } from "lucide-react"
import { buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ApiError } from "@/api/client"
import { useTeam } from "@/hooks/use-teams"
import { useTeamActivity } from "@/hooks/use-team-activity"
import { ActivityTimeline } from "@/components/activity/activity-timeline"
import { ActivityFilters } from "@/components/activity/activity-filters"
import { toast } from "sonner"

// Phase 14 C.3: /teams/:slug/activity — timeline всех событий в команде.
// Доступ: любой член команды (viewer+). Backend проверит через GetBySlug.
export default function TeamActivityPage() {
  const { slug = "" } = useParams()
  const navigate = useNavigate()
  const { data: team, isLoading: teamLoading, error: teamError } = useTeam(slug)

  const [eventType, setEventType] = useState("")
  const filters = useMemo(() => ({ event_type: eventType || undefined }), [eventType])

  const { data, isLoading, isFetching, fetchNextPage, hasNextPage } = useTeamActivity(slug, filters)

  if (teamError instanceof ApiError && teamError.status === 403) {
    toast.error("Нет доступа к команде")
    navigate("/teams")
    return null
  }

  const items = data?.pages.flatMap((p) => p.items) ?? []

  return (
    <div className="container mx-auto max-w-4xl space-y-6 px-4 py-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Link
          to={`/teams/${slug}`}
          className={buttonVariants({ variant: "ghost", size: "sm" })}
        >
          <ArrowLeft className="size-4" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold">Активность: {team?.name ?? slug}</h1>
          <p className="text-sm text-muted-foreground">Кто что менял в промптах, коллекциях и составе команды</p>
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">События</CardTitle>
          <ActivityFilters eventType={eventType} onEventTypeChange={setEventType} />
        </CardHeader>
        <CardContent>
          {teamLoading || isLoading ? (
            <div className="flex flex-col gap-3 py-2">
              {[0, 1, 2, 3].map((i) => (
                <div key={i} className="flex gap-3">
                  <Skeleton className="size-8 rounded-full" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-2/3" />
                    <Skeleton className="h-3 w-24" />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <ActivityTimeline
              items={items}
              hasMore={!!hasNextPage}
              isFetching={isFetching}
              onLoadMore={() => fetchNextPage()}
              hasFilter={!!eventType}
              onClearFilter={() => setEventType("")}
            />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
