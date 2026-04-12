import type { ReactNode } from 'react';
import { Button } from './ui/button';
import { getSettings } from '../lib/storage';

interface Props {
  title: string;
  description?: string;
  action?: ReactNode;
  /** Показать кнопку "Открыть ПромтЛаб" ведущую на apiBase */
  showOpenWebLink?: boolean;
}

export function EmptyState({ title, description, action, showOpenWebLink }: Props) {
  const openWebApp = async () => {
    const { apiBase } = await getSettings();
    void chrome.tabs.create({ url: apiBase });
  };

  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
      <div className="text-sm font-medium text-(--color-foreground)">{title}</div>
      {description ? (
        <div className="text-xs text-(--color-muted-foreground)">{description}</div>
      ) : null}
      {action}
      {showOpenWebLink ? (
        <Button type="button" variant="outline" size="sm" onClick={openWebApp}>
          Открыть ПромтЛаб
        </Button>
      ) : null}
    </div>
  );
}
