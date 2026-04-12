import { Globe } from 'lucide-react';
import { Badge } from './ui/badge';
import { hostLabel } from '../lib/messages';
import type { ActiveTabState } from '../hooks/use-active-tab';

export function ActiveTabBadge({ state }: { state: ActiveTabState }) {
  if (!state.host) {
    return (
      <Badge variant="outline" title="Нет активной вкладки">
        <Globe className="h-2.5 w-2.5" />
        Нет вкладки
      </Badge>
    );
  }

  if (state.supported) {
    return (
      <Badge variant="success" title={`Готов вставить в ${state.host}`}>
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
        {hostLabel(state.host)}
      </Badge>
    );
  }

  return (
    <Badge variant="warning" title={`${state.host} не поддерживается`}>
      <Globe className="h-2.5 w-2.5" />
      Не поддерживается
    </Badge>
  );
}
