import { api } from "./client"

// --- Types (mirror backend usecases/prompt_insights/types.go) ---

// PromptInsightRow — общий формат строки во всех листингах insight'ов
// (unused / trending / declining / most-edited). Один shape — один UI-компонент.
export interface PromptInsightRow {
  prompt_id: number
  title: string
  uses: number
  updated_at?: string
}

// DuplicatePair — пара похожих промптов с similarity score [0..1] из pg_trgm.
export interface DuplicatePair {
  prompt_a: PromptInsightRow
  prompt_b: PromptInsightRow
  similarity: number
}

interface ItemsEnvelope<T> {
  items?: T[]
}

// getItems — общий helper: бэк возвращает {items: [...]} envelope для всех
// listing endpoints, защищаемся от отсутствия поля (server может вернуть {} на
// пустой результат — реальный кейс для unused при чистом DB).
async function getItems<T>(path: string): Promise<T[]> {
  const env = await api<ItemsEnvelope<T>>(path)
  return env.items ?? []
}

// --- Listing endpoints ---

export const fetchUnused = () =>
  getItems<PromptInsightRow>("/prompts/insights/unused")

export const fetchDuplicates = () =>
  getItems<DuplicatePair>("/prompts/insights/duplicates")

export const fetchTrending = () =>
  getItems<PromptInsightRow>("/prompts/insights/trending")

export const fetchDeclining = () =>
  getItems<PromptInsightRow>("/prompts/insights/declining")

export const fetchMostEdited = () =>
  getItems<PromptInsightRow>("/prompts/insights/most-edited")

// --- Mutation ---

export interface MergeResult {
  kept_id: number
  merged_id: number
}

// mergePrompts — keepID остаётся, mergeID мягко удаляется (soft-delete в trash);
// usage stats, теги, коллекции, share-ссылки переезжают на keepID. Бэк отдаёт
// {kept_id, merged_id} для подтверждения операции.
export function mergePrompts(keepID: number, mergeID: number): Promise<MergeResult> {
  return api<MergeResult>(`/prompts/${keepID}/merge-with/${mergeID}`, {
    method: "POST",
  })
}
