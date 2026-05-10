// Canvas page — read-only визуализатор графа цепочки. URL: /chains/:id/canvas.
//
// Поддерживает fullscreen-режим (Portal в document.body, без app-sidebar/header)
// для больших цепочек, где обычная container-ширина не вмещает граф. Esc или
// X-кнопка возвращают в обычный режим.

import { useEffect, useState } from "react"
import { createPortal } from "react-dom"
import { Link, useParams } from "react-router-dom"
import { ArrowLeft, Maximize2, Minimize2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { CanvasFlow } from "@/components/chains/canvas-flow"
import { useChain } from "@/hooks/use-chains"
import type { ChainDetail } from "@/api/types"

export default function ChainCanvasPage() {
  const { id } = useParams<{ id: string }>()
  const chainID = id ? Number(id) : 0
  const { data: chain, isLoading } = useChain(chainID)
  const [fullscreen, setFullscreen] = useState(false)

  if (isLoading || !chain) {
    return (
      <div className="container mx-auto max-w-7xl p-6">
        <Skeleton className="h-12 w-64" />
        <Skeleton className="mt-4 h-[calc(100vh-12rem)] w-full" />
      </div>
    )
  }

  return (
    <>
      <div className="container mx-auto max-w-7xl p-6">
        <div className="mb-4 flex items-center gap-3">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/chains">
              <ArrowLeft className="h-4 w-4" />
            </Link>
          </Button>
          <div className="flex-1">
            <h1 className="text-xl font-semibold">{chain.name}</h1>
            {chain.description && (
              <p className="line-clamp-1 text-xs text-muted-foreground">{chain.description}</p>
            )}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setFullscreen(true)}
            title="Развернуть граф на весь экран"
          >
            <Maximize2 className="mr-2 h-4 w-4" />
            Во весь экран
          </Button>
        </div>

        <CanvasFlow chain={chain} />

        {chain.steps.length === 0 && (
          <div className="mt-8 rounded-md border bg-muted/20 p-6 text-center text-sm text-muted-foreground">
            Цепочка пока пуста. Добавьте шаги через классический редактор:
            <Button variant="link" asChild>
              <Link to={`/chains/${chain.id}/edit`}>Открыть редактор</Link>
            </Button>
          </div>
        )}
      </div>

      {fullscreen && <FullscreenCanvas chain={chain} onClose={() => setFullscreen(false)} />}
    </>
  )
}

function FullscreenCanvas({ chain, onClose }: { chain: ChainDetail; onClose: () => void }) {
  // Esc → выход. Лочим скролл body чтобы он не мерцал под порталом.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose()
    }
    document.addEventListener("keydown", onKey)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = "hidden"
    return () => {
      document.removeEventListener("keydown", onKey)
      document.body.style.overflow = prevOverflow
    }
  }, [onClose])

  return createPortal(
    <div className="fixed inset-0 z-50 flex flex-col bg-background">
      <div className="flex items-center gap-3 border-b px-4 py-2">
        <div className="flex-1 min-w-0">
          <p className="truncate text-sm font-semibold">{chain.name}</p>
          {chain.description && (
            <p className="truncate text-xs text-muted-foreground">{chain.description}</p>
          )}
        </div>
        <p className="hidden text-xs text-muted-foreground sm:block">
          Esc — закрыть
        </p>
        <Button variant="outline" size="sm" onClick={onClose}>
          <Minimize2 className="mr-2 h-4 w-4" />
          Свернуть
        </Button>
      </div>
      <div className="flex-1 overflow-hidden">
        <CanvasFlow chain={chain} fillParent />
      </div>
    </div>,
    document.body,
  )
}
