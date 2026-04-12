import type { Prompt } from '../lib/types';
import { PromptCard } from './prompt-card';

interface Props {
  prompts: Prompt[];
  onSelect: (p: Prompt) => void;
  highlightedId?: number | null;
  focusedId?: number | null;
}

export function PromptList({ prompts, onSelect, highlightedId, focusedId }: Props) {
  return (
    <div className="flex flex-col gap-2">
      {prompts.map((p) => (
        <PromptCard
          key={p.id}
          prompt={p}
          onClick={() => onSelect(p)}
          highlighted={highlightedId === p.id}
          focused={focusedId === p.id}
        />
      ))}
    </div>
  );
}
