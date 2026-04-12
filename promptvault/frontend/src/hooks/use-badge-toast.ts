import { toast } from "sonner"
import { useQueryClient } from "@tanstack/react-query"
import type { BadgeSummary } from "@/api/types"

/**
 * useBadgeUnlocks — helper-хук для обработки newly_unlocked_badges из ответа
 * mutating API (create prompt/collection, increment usage).
 *
 * Возвращает одну функцию handleBadgeUnlocks, которую нужно вызвать в
 * onSuccess mutation с полем response.newly_unlocked_badges.
 *
 * UX:
 * - 0 badges → no-op
 * - 1 badge → toast.success со специфичным title и description
 * - >1 badges → один toast со списком через description
 *
 * Помимо toast — invalidate ["badges"] чтобы страница /badges обновилась.
 */
export function useBadgeUnlocks() {
  const qc = useQueryClient()

  return (badges: BadgeSummary[] | undefined) => {
    if (!badges || badges.length === 0) return

    qc.invalidateQueries({ queryKey: ["badges"] })

    if (badges.length === 1) {
      const b = badges[0]
      toast.success(`${b.icon} ${b.title}`, {
        description: "Новое достижение разблокировано!",
        duration: 6000,
      })
      return
    }

    const list = badges.map((b) => `${b.icon} ${b.title}`).join(", ")
    toast.success(`Разблокировано ${badges.length} достижения`, {
      description: list,
      duration: 8000,
    })
  }
}
