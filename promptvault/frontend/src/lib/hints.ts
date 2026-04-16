// M-13: In-app education. Dismissible hints. Состояние хранится в localStorage,
// потому что сервер об этом знать не должен — UX-onboarding независимый от
// устройств (на новом девайсе юзер может захотеть освежить подсказки).
import { useCallback, useSyncExternalStore } from "react"

const STORAGE_KEY = "pv_hints_dismissed"

export type HintId =
  | "ai_button"
  | "dashboard_extension"
  | "dashboard_mcp"
  | "settings_referral"
  | "team_empty"

// Наивный pub/sub для переподписки всех hints после dismiss/reset.
const listeners = new Set<() => void>()
const notify = () => listeners.forEach((l) => l())

function readDismissed(): Set<HintId> {
  if (typeof localStorage === "undefined") return new Set()
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return new Set()
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? new Set(arr as HintId[]) : new Set()
  } catch {
    return new Set()
  }
}

function writeDismissed(set: Set<HintId>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(Array.from(set)))
  } catch {
    /* quota exceeded / storage disabled */
  }
}

// Кешируем snapshot — useSyncExternalStore требует референсную стабильность между
// вызовами, когда значение не менялось. Пересоздаём только на notify().
let snapshotCache: ReadonlySet<HintId> = readDismissed()

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

function getSnapshot(): ReadonlySet<HintId> {
  return snapshotCache
}

function getServerSnapshot(): ReadonlySet<HintId> {
  return new Set<HintId>() // SSR-safe default
}

export function useHintDismissed(id: HintId): [boolean, () => void] {
  const dismissedSet = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)
  const dismiss = useCallback(() => {
    const next = new Set(dismissedSet)
    next.add(id)
    snapshotCache = next
    writeDismissed(next)
    notify()
  }, [id, dismissedSet])
  return [dismissedSet.has(id), dismiss]
}

export function resetHints() {
  snapshotCache = new Set()
  writeDismissed(snapshotCache)
  notify()
}
