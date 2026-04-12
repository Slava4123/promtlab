// Минимальный toaster без внешних зависимостей.
// Использование: useToast().toast({ title, description, variant, action })

import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { CheckCircle2, XCircle, Info, Undo2 } from 'lucide-react';
import { cn } from '../../lib/utils';

export type ToastVariant = 'success' | 'error' | 'info';

export interface ToastPayload {
  id?: number;
  title: string;
  description?: string;
  variant?: ToastVariant;
  action?: { label: string; onClick: () => void; icon?: 'undo' };
  durationMs?: number;
}

interface ToastItem extends Required<Omit<ToastPayload, 'action' | 'description'>> {
  description?: string;
  action?: ToastPayload['action'];
}

interface ToastContextValue {
  toast: (payload: ToastPayload) => void;
  dismiss: (id: number) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

let counter = 0;

export function ToasterProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);

  const dismiss = useCallback((id: number) => {
    setItems((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = useCallback(
    (payload: ToastPayload) => {
      const id = ++counter;
      const item: ToastItem = {
        id,
        title: payload.title,
        description: payload.description,
        variant: payload.variant ?? 'info',
        action: payload.action,
        durationMs: payload.durationMs ?? 3000,
      };
      setItems((prev) => [...prev, item]);
      if (item.durationMs > 0) {
        setTimeout(() => dismiss(id), item.durationMs);
      }
    },
    [dismiss],
  );

  const ctx = useMemo(() => ({ toast, dismiss }), [toast, dismiss]);

  return (
    <ToastContext.Provider value={ctx}>
      {children}
      <ToastViewport items={items} onDismiss={dismiss} />
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used inside ToasterProvider');
  return ctx;
}

function ToastViewport({ items, onDismiss }: { items: ToastItem[]; onDismiss: (id: number) => void }) {
  return (
    <div className="pointer-events-none fixed bottom-3 left-3 right-3 z-50 flex flex-col gap-2">
      {items.map((t) => (
        <ToastCard key={t.id} item={t} onDismiss={() => onDismiss(t.id)} />
      ))}
    </div>
  );
}

function ToastCard({ item, onDismiss }: { item: ToastItem; onDismiss: () => void }) {
  const [visible, setVisible] = useState(false);
  useEffect(() => {
    const t = requestAnimationFrame(() => setVisible(true));
    return () => cancelAnimationFrame(t);
  }, []);

  const Icon = item.variant === 'success' ? CheckCircle2 : item.variant === 'error' ? XCircle : Info;
  const iconClass =
    item.variant === 'success'
      ? 'text-emerald-400'
      : item.variant === 'error'
      ? 'text-(--color-destructive)'
      : 'text-(--color-muted-foreground)';

  return (
    <div
      className={cn(
        'pointer-events-auto flex items-start gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-3 shadow-lg transition-all',
        visible ? 'translate-y-0 opacity-100' : 'translate-y-2 opacity-0',
      )}
    >
      <Icon className={cn('mt-0.5 h-4 w-4 shrink-0', iconClass)} />
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium text-(--color-card-foreground)">{item.title}</div>
        {item.description ? (
          <div className="mt-0.5 text-xs text-(--color-muted-foreground)">{item.description}</div>
        ) : null}
      </div>
      {item.action ? (
        <button
          type="button"
          onClick={() => {
            item.action!.onClick();
            onDismiss();
          }}
          className="shrink-0 flex items-center gap-1 rounded px-2 py-1 text-xs font-medium text-(--color-primary) hover:bg-(--color-accent)"
        >
          {item.action.icon === 'undo' ? <Undo2 className="h-3 w-3" /> : null}
          {item.action.label}
        </button>
      ) : null}
    </div>
  );
}
