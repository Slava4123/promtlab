import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, GitBranch, Loader2, Save } from "lucide-react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { Textarea } from "../../components/ui/textarea"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { useWorkspaceStore } from "../../stores/workspace-store"

export function ChainNewPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")

  const createMut = useMutation({
    mutationFn: () =>
      sendBg({
        type: "api.createChain",
        body: { name: name.trim(), description: description.trim(), team_id: teamId },
      }),
    onSuccess: (chain) => {
      void qc.invalidateQueries({ queryKey: ["chains"] })
      toast({ title: "Цепочка создана", variant: "success" })
      navigate(`/chains/${chain.id}/edit`, { replace: true })
    },
    onError: (err: Error) => {
      toast({
        title: "Не удалось создать",
        description: err.message,
        variant: "error",
      })
    },
  })

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Новая цепочка</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        <div className="flex items-center gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-3">
          <GitBranch className="h-4 w-4 text-(--color-primary)" />
          <p className="text-[10px] text-(--color-muted-foreground)">
            Создайте многошаговый workflow, чтобы вызывать несколько промптов по очереди.
          </p>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="chain-name">Название</Label>
          <Input
            id="chain-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="PRD по идее"
            maxLength={100}
            autoFocus
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="chain-desc">Описание</Label>
          <Textarea
            id="chain-desc"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Цепочка генерирует PRD из идеи в три шага: brief → outline → draft."
            maxLength={2000}
            rows={4}
          />
        </div>

        {teamId && (
          <p className="text-[10px] text-(--color-muted-foreground)">
            Цепочка будет создана в текущей команде.
          </p>
        )}
      </div>

      <div className="flex items-center gap-2 border-t border-(--color-border) p-2">
        <Button type="button" variant="outline" size="sm" onClick={() => navigate(-1)} className="flex-1">
          Отмена
        </Button>
        <Button
          type="button"
          size="sm"
          onClick={() => createMut.mutate()}
          disabled={createMut.isPending || !name.trim()}
          className="flex-1 gap-1.5"
        >
          {createMut.isPending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Save className="h-3.5 w-3.5" />
          )}
          Создать
        </Button>
      </div>
    </div>
  )
}
