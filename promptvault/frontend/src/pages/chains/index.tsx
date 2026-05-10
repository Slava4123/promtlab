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
import { ChainTemplateCard } from "@/components/chains/chain-template-card"
import { ChainCard } from "@/components/chains/chain-card"
import { CHAIN_TEMPLATES } from "@/lib/chain-templates"

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
            Связывайте промпты в последовательности — output одного шага становится переменной для следующего.
          </p>
        </div>
        {/* В empty-state верхнюю CTA скрываем — основное действие есть в галерее
            шаблонов и «Создать с нуля». В filled-state CTA остаётся как primary. */}
        {canWrite && !isEmpty && (
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
        <div className="space-y-6">
          <div>
            <h2 className="mb-3 text-sm font-medium text-foreground">Начните с шаблона</h2>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {CHAIN_TEMPLATES.map((tpl) => (
                <ChainTemplateCard key={tpl.id} template={tpl} teamId={teamId} />
              ))}
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="h-px flex-1 bg-border/50" />
            <span className="text-[0.72rem] uppercase tracking-wider text-muted-foreground">или</span>
            <div className="h-px flex-1 bg-border/50" />
          </div>

          <div className="flex justify-center">
            <Button variant="outline" asChild>
              <Link to="/chains/new">
                <Plus className="mr-2 h-4 w-4" />
                Создать с нуля
              </Link>
            </Button>
          </div>
        </div>
      )}

      {/* Viewer в команде: галерею шаблонов скрываем — у viewer'а нет canWrite.
          Показываем «команда без цепочек» с подсказкой к owner/editor'у. */}
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
