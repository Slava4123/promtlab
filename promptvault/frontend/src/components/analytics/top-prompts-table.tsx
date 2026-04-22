import { Link } from "react-router-dom"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import type { PromptUsageRow } from "@/api/analytics"

interface TopPromptsTableProps {
  title: string
  prompts: PromptUsageRow[]
  metricLabel?: string
}

export function TopPromptsTable({ title, prompts, metricLabel = "Использований" }: TopPromptsTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {prompts.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            Нет данных за этот период
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Промпт</TableHead>
                <TableHead className="w-32 text-right">{metricLabel}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {prompts.map((p) => (
                <TableRow key={p.prompt_id}>
                  <TableCell className="font-medium">
                    <Link to={`/prompts/${p.prompt_id}`} className="hover:underline">
                      {p.title || `Prompt #${p.prompt_id}`}
                    </Link>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">{p.uses.toLocaleString("ru")}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
