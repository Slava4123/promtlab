import { Folder } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import type { CollectionDTO } from '../lib/types';

const ALL_VALUE = '__all__';

interface Props {
  collectionId: number | null;
  collections: CollectionDTO[];
  onChange: (id: number | null) => void;
}

export function CollectionSelector({ collectionId, collections, onChange }: Props) {
  if (collections.length === 0) return null;

  const value = collectionId === null ? ALL_VALUE : String(collectionId);

  return (
    <Select
      value={value}
      onValueChange={(v: string) => onChange(v === ALL_VALUE ? null : Number(v))}
    >
      <SelectTrigger className="h-8 text-xs" aria-label="Коллекция">
        <SelectValue placeholder="Все коллекции" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={ALL_VALUE}>
          <span className="flex items-center gap-2">
            <Folder className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
            Все коллекции
          </span>
        </SelectItem>
        {collections.length > 0 ? <SelectSeparator /> : null}
        {collections.map((c) => (
          <SelectItem key={c.id} value={String(c.id)}>
            <span className="flex items-center gap-2">
              {c.icon ? (
                <span className="text-xs">{c.icon}</span>
              ) : (
                <Folder
                  className="h-3.5 w-3.5"
                  style={{ color: c.color ?? 'var(--color-muted-foreground)' }}
                />
              )}
              <span className="flex-1">{c.name}</span>
              {c.prompts_count !== undefined ? (
                <span className="text-[10px] text-(--color-muted-foreground)">
                  {c.prompts_count}
                </span>
              ) : null}
            </span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
