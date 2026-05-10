import { useState } from "react"
import { Link } from "react-router-dom"
import { Plus, Link2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { useChains, useDeleteChain } from "@/hooks/use-chains"
import { useCurrentTeamRole } from "@/hooks/use-team-role"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { ChainCard } from "@/components/chains/chain-card"

export default function ChainsPage() {
  const team = useWorkspaceStore((s) => s.team)
  const teamId = team?.teamId ?? null
  const teamName = team?.teamName ?? null
  const { canWrite, isViewer } = useCurrentTeamRole()

  const { data, isLoading } = useChains({ teamId })
  const deleteChain = useDeleteChain()
  // Локальный state для красивого ConfirmDialog (вместо native confirm()).
  const [pendingDelete, setPendingDelete] = useState<{ id: number; name: string } | null>(null)

  const isEmpty = !isLoading && data && data.items.length === 0

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6 flex items-center justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold">
            {teamName ? `Цепочки — ${teamName}` : "Цепочки промптов"}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Связывайте промпты в последовательности — ответ одного шага становится переменной для следующего.
          </p>
        </div>
        {canWrite && (
          <Button variant="brand" asChild>
            <Link to="/chains/new">
              <Plus className="mr-2 h-4 w-4" />
              Создать цепочку
            </Link>
          </Button>
        )}
      </div>

      {isLoading && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-44" />
          ))}
        </div>
      )}

      {isEmpty && canWrite && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <Link2 className="mb-4 h-12 w-12 text-muted-foreground" aria-hidden="true" />
            <p className="mb-2 text-base font-medium text-foreground">
              {teamName ? `В команде «${teamName}» пока нет цепочек` : "Пока нет цепочек"}
            </p>
            <p className="mb-6 max-w-md text-sm text-muted-foreground">
              Связывайте промпты в последовательности — ответ одного шага становится переменной для следующего.
            </p>
            <Button variant="brand" asChild>
              <Link to="/chains/new">
                <Plus className="mr-2 h-4 w-4" />
                Создать первую цепочку
              </Link>
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Viewer в команде: создавать не может — показываем плейсхолдер
          с подсказкой обратиться к owner/editor'у. */}
      {isEmpty && !canWrite && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Link2 className="mb-4 h-12 w-12 text-muted-foreground" aria-hidden="true" />
            <p className="mb-2 text-base font-medium text-foreground">
              {teamName ? `В команде «${teamName}» пока нет цепочек` : "Пока нет цепочек"}
            </p>
            <p className="text-sm text-muted-foreground">
              Попросите владельца или редактора команды создать первую цепочку.
            </p>
          </CardContent>
        </Card>
      )}

      {!isLoading && data && data.items.length > 0 && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {data.items.map((chain) => (
            <ChainCard
              key={chain.id}
              chain={chain}
              canWrite={canWrite}
              isViewer={isViewer}
              onDelete={(id, name) => setPendingDelete({ id, name })}
            />
          ))}
        </div>
      )}

      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(v) => !v && setPendingDelete(null)}
        title="Удалить цепочку?"
        description={
          pendingDelete
            ? `Цепочка «${pendingDelete.name}» удалится вместе со всеми шагами и историей запусков. Это действие нельзя отменить.`
            : ""
        }
        confirmLabel="Удалить"
        isPending={deleteChain.isPending}
        onConfirm={() => {
          if (!pendingDelete) return
          deleteChain.mutate(pendingDelete.id, { onSettled: () => setPendingDelete(null) })
        }}
      />
    </div>
  )
}
