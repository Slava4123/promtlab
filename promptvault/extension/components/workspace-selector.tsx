import { Users, User as UserIcon } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import type { TeamDTO } from '../lib/types';

// Radix UI Select не принимает пустую строку как valid value, поэтому sentinel.
const PERSONAL_VALUE = '__personal__';

interface Props {
  workspaceId: number | null;
  teams: TeamDTO[];
  onChange: (id: number | null) => void;
}

export function WorkspaceSelector({ workspaceId, teams, onChange }: Props) {
  const value = workspaceId === null ? PERSONAL_VALUE : String(workspaceId);

  return (
    <Select
      value={value}
      onValueChange={(v: string) => {
        onChange(v === PERSONAL_VALUE ? null : Number(v));
      }}
    >
      <SelectTrigger className="h-7 w-auto min-w-[7rem] gap-1 px-2 text-xs" aria-label="Пространство">
        <SelectValue placeholder="Выберите" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={PERSONAL_VALUE}>
          <span className="flex items-center gap-2">
            <UserIcon className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
            Личное
          </span>
        </SelectItem>
        {teams.length > 0 ? (
          <>
            <SelectSeparator />
            <SelectLabel>Команды</SelectLabel>
            {teams.map((t) => (
              <SelectItem key={t.id} value={String(t.id)}>
                <span className="flex items-center gap-2">
                  <Users className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
                  {t.name}
                </span>
              </SelectItem>
            ))}
          </>
        ) : null}
      </SelectContent>
    </Select>
  );
}
