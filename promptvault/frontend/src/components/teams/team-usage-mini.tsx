import { FileText, FolderOpen, Link2 } from "lucide-react"

import { useTeamUsage } from "@/hooks/use-subscription"
import type { QuotaInfo } from "@/api/types"

interface TeamUsageMiniProps {
  slug: string
}

/**
 * TeamUsageMini — компактный индикатор team-pool usage для карточки команды
 * на странице /teams. Иконки совпадают с навигацией (FileText/FolderOpen/Link2),
 * чтобы юзер сразу узнавал ресурсы без расшифровки сокращений.
 */
export function TeamUsageMini({ slug }: TeamUsageMiniProps) {
  const { data, isLoading } = useTeamUsage(slug)

  if (isLoading) {
    return <div className="h-3 w-32 animate-pulse rounded-sm bg-muted/30" />
  }

  if (!data) return null

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[0.7rem] text-muted-foreground">
      <Pair icon={<FileText className="h-3 w-3" />} info={data.prompts} aria="Промпты" />
      <Pair icon={<FolderOpen className="h-3 w-3" />} info={data.collections} aria="Коллекции" />
      {data.chains.limit > 0 && (
        <Pair icon={<Link2 className="h-3 w-3" />} info={data.chains} aria="Цепочки" />
      )}
    </div>
  )
}

function Pair({
  icon,
  info,
  aria,
}: {
  icon: React.ReactNode
  info: QuotaInfo
  aria: string
}) {
  if (info.limit <= 0) return null
  const pct = info.used / info.limit
  const colorClass =
    pct >= 0.9
      ? "text-red-500"
      : pct >= 0.75
        ? "text-amber-500"
        : "text-muted-foreground"

  return (
    <span className={`inline-flex items-center gap-1 ${colorClass}`} title={aria}>
      <span aria-hidden="true">{icon}</span>
      <span className="sr-only">{aria}: </span>
      {info.used.toLocaleString("ru-RU")}/{info.limit.toLocaleString("ru-RU")}
    </span>
  )
}
