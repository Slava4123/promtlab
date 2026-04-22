import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import type { ContributorRow } from "@/api/analytics"

function initials(name: string, email: string): string {
  const src = name || email || "?"
  const parts = src.split(/\s+/).filter(Boolean)
  if (parts.length === 0) return "?"
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase()
  return (parts[0]![0]! + parts[1]![0]!).toUpperCase()
}

interface ContributorsLeaderboardProps {
  contributors: ContributorRow[]
}

export function ContributorsLeaderboard({ contributors }: ContributorsLeaderboardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Топ контрибьюторов</CardTitle>
      </CardHeader>
      <CardContent>
        {contributors.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            Нет активности за этот период
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Участник</TableHead>
                <TableHead className="w-24 text-right">Создал</TableHead>
                <TableHead className="w-24 text-right">Правил</TableHead>
                <TableHead className="w-24 text-right">Использ.</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {contributors.map((c) => (
                <TableRow key={c.user_id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Avatar className="size-7">
                        <AvatarFallback className="text-xs">{initials(c.name ?? "", c.email)}</AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col">
                        <span className="text-sm font-medium">{c.name || c.email}</span>
                        {c.name && <span className="text-xs text-muted-foreground">{c.email}</span>}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">{c.prompts_created}</TableCell>
                  <TableCell className="text-right tabular-nums">{c.prompts_edited}</TableCell>
                  <TableCell className="text-right tabular-nums">{c.uses}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
