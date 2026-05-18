import { Link } from "react-router-dom"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import type { PromptUsageRow } from "@/api/analytics"

interface TopPromptsTableProps {
  title: string
  prompts: PromptUsageRow[]
  metricLabel?: string
}

// Top-10 с ранжированием (# колонка слева), striped rows (alternating) и
// hover-подсветкой — стандартный паттерн analytics-таблиц (Tremor, Vercel
// Analytics, Linear). Без полос таблица выглядит «плоской», глаз скользит
// по строкам и теряет связь между промптом и метрикой. Чередующийся фон
// делит данные визуально на пары и заземляет числа правой колонки.
export function TopPromptsTable({ title, prompts, metricLabel = "Использований" }: TopPromptsTableProps) {
  return (
    <Card className="min-w-0">
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="px-0 sm:px-6">
        {prompts.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            Нет данных за этот период
          </div>
        ) : (
          <div className="overflow-x-auto">
            <Table className="w-full table-fixed">
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead className="w-12 text-center text-muted-foreground">#</TableHead>
                  <TableHead>Промпт</TableHead>
                  <TableHead className="w-32 text-right">{metricLabel}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {prompts.map((p, idx) => {
                  const display = p.title || `Prompt #${p.prompt_id}`
                  const rank = idx + 1
                  return (
                    <TableRow
                      key={p.prompt_id}
                      className="border-b-0 transition-colors odd:bg-foreground/[0.025] hover:bg-violet-500/5"
                    >
                      <TableCell className="w-12 py-2.5 text-center text-xs font-medium tabular-nums text-muted-foreground">
                        {rank}
                      </TableCell>
                      <TableCell className="max-w-0 truncate py-2.5 font-medium">
                        <Link
                          to={`/prompts/${p.prompt_id}`}
                          className="block truncate hover:underline"
                          title={display}
                        >
                          {display}
                        </Link>
                      </TableCell>
                      <TableCell className="w-32 py-2.5 text-right font-mono text-sm tabular-nums">
                        {p.uses.toLocaleString("ru")}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
