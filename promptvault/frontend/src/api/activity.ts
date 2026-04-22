import { api } from "./client"

// --- Types ---

export interface ActivityItem {
  id: number
  actor_id?: number
  actor_email: string
  actor_name?: string
  event_type: string // prompt.created | prompt.updated | ... (see models/team_activity.go)
  target_type: string // prompt | collection | tag | share | member
  target_id?: number
  target_label?: string
  metadata?: Record<string, unknown>
  created_at: string
}

export interface ActivityPage {
  items: ActivityItem[]
  page: number
  page_size: number
  has_more: boolean
}

export interface ActivityFilters {
  event_type?: string
  actor_id?: number
  target_type?: string
  target_id?: number
  from?: string
  to?: string
}

export interface PromptHistory {
  versions: PromptVersionWithActor[]
  activity: ActivityItem[]
}

export interface PromptVersionWithActor {
  id: number
  version_number: number
  title: string
  content: string
  model?: string
  change_note?: string
  created_at: string
  changed_by_id?: number
  changed_by_email?: string
  changed_by_name?: string
}

// --- API functions ---

export function fetchTeamActivity(
  slug: string,
  page = 1,
  pageSize = 50,
  filters?: ActivityFilters,
): Promise<ActivityPage> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  })
  if (filters?.event_type) params.set("event_type", filters.event_type)
  if (filters?.actor_id) params.set("actor_id", String(filters.actor_id))
  if (filters?.target_type) params.set("target_type", filters.target_type)
  if (filters?.target_id) params.set("target_id", String(filters.target_id))
  if (filters?.from) params.set("from", filters.from)
  if (filters?.to) params.set("to", filters.to)

  return api<ActivityPage>(`/teams/${encodeURIComponent(slug)}/activity?${params.toString()}`)
}

export function fetchPromptHistory(promptId: number): Promise<PromptHistory> {
  return api<PromptHistory>(`/prompts/${promptId}/history`)
}
