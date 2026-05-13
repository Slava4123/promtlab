import type { ReactNode } from 'react';
import { Button } from './ui/button';
import { getSettings } from '../lib/storage';
import { openWebPage } from '../lib/utils';

interface Props {
  title: string;
  description?: string;
  action?: ReactNode;
  /** Показать кнопку "Открыть ПромтЛаб" ведущую на frontend (не API). */
  showOpenWebLink?: boolean;
}

export function EmptyState({ title, description, action, showOpenWebLink }: Props) {
  // openWebPage делает derive backend→frontend (localhost:8080 → :5173 в dev,
  // promtlabs.ru остаётся promtlabs.ru в prod). Раньше дергали chrome.tabs.create
  // прямо с apiBase, что в dev открывало backend :8080 (404 page not found).
  const openWebApp = async () => {
    const { apiBase } = await getSettings();
    openWebPage(apiBase, '/');
  };

  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
      <div className="text-sm font-medium text-(--color-foreground)">{title}</div>
      {description ? (
        <div className="text-xs text-(--color-muted-foreground)">{description}</div>
      ) : null}
      {action}
      {showOpenWebLink ? (
        <Button type="button" variant="brand" size="sm" onClick={openWebApp}>
          Открыть ПромтЛаб
        </Button>
      ) : null}
    </div>
  );
}
