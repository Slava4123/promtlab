import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { ArrowLeft, Copy, Send, Share2, Zap } from 'lucide-react';
import { Button } from './ui/button';
import { Label } from './ui/label';
import { Textarea } from './ui/textarea';
import { useToast } from './ui/toaster';
import type { Prompt } from '../lib/types';
import { extractVariables, renderTemplate } from '../lib/template';
import { getSavedVars, setSavedVars } from '../lib/storage';
import { useKeyboardShortcuts } from '../hooks/use-keyboard-shortcuts';
import { sendBg } from '../lib/bg-client';
import { ApiError } from '../lib/types';

interface Props {
  prompt: Prompt;
  onBack: () => void;
  onSubmit: (finalText: string) => void;
  onInsertAll?: (finalText: string) => void;
  submitting: boolean;
  error: string | null;
  canInsert: boolean;
  canInsertReason?: string;
}

export function VariableForm({
  prompt,
  onBack,
  onSubmit,
  onInsertAll,
  submitting,
  error,
  canInsert,
  canInsertReason,
}: Props) {
  const variables = useMemo(() => extractVariables(prompt.content), [prompt.content]);
  const [values, setValues] = useState<Record<string, string>>(() =>
    Object.fromEntries(variables.map((v) => [v, ''])),
  );
  const [loadedSaved, setLoadedSaved] = useState(false);
  const { toast } = useToast();

  // Загружаем сохранённые значения переменных из storage
  useEffect(() => {
    let cancelled = false;
    void getSavedVars(prompt.id).then((saved) => {
      if (cancelled) return;
      setValues((prev) => {
        const next = { ...prev };
        for (const v of variables) {
          if (saved[v] && !next[v]) next[v] = saved[v];
        }
        return next;
      });
      setLoadedSaved(true);
    });
    return () => {
      cancelled = true;
    };
  }, [prompt.id, variables]);

  const preview = useMemo(
    () => renderTemplate(prompt.content, values),
    [prompt.content, values],
  );

  const charCount = preview.length;
  const tokenEstimate = Math.round(charCount / 4); // rough 1 token ≈ 4 chars

  function update(name: string, value: string) {
    setValues((prev) => ({ ...prev, [name]: value }));
  }

  async function submit(e?: FormEvent) {
    e?.preventDefault();
    if (loadedSaved) {
      await setSavedVars(prompt.id, values);
    }
    onSubmit(preview);
  }

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(preview);
      toast({
        title: 'Скопировано',
        description: `${preview.length.toLocaleString('ru-RU')} симв. в буфере`,
        variant: 'success',
        durationMs: 2000,
      });
      if (loadedSaved) {
        await setSavedVars(prompt.id, values);
      }
    } catch {
      toast({ title: 'Не удалось скопировать', variant: 'error' });
    }
  }

  async function share() {
    try {
      const result = await sendBg({ type: 'api.createShareLink', promptId: prompt.id });
      await navigator.clipboard.writeText(result.url);
      toast({
        title: 'Ссылка скопирована',
        description: result.url,
        variant: 'success',
        durationMs: 3000,
      });
    } catch (err) {
      const msg =
        err instanceof ApiError && err.code === 'unauthorized'
          ? 'Нет прав на публичную ссылку'
          : 'Не удалось создать ссылку';
      toast({ title: msg, variant: 'error' });
    }
  }

  // Keyboard shortcuts
  useKeyboardShortcuts([
    {
      key: 'Escape',
      handler: () => {
        onBack();
      },
    },
    {
      key: 'Enter',
      ctrlOrCmd: true,
      handler: () => {
        if (canInsert && !submitting) void submit();
      },
    },
    {
      key: 'c',
      ctrlOrCmd: true,
      shift: true,
      handler: () => {
        void copyToClipboard();
      },
    },
  ]);

  return (
    <form onSubmit={submit} className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-3">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={onBack}
          aria-label="Назад"
          title="Назад (Esc)"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">{prompt.title}</h2>
      </div>

      <div className="flex-1 overflow-y-auto">
        {variables.length > 0 ? (
          <div className="space-y-3 border-b border-(--color-border) p-3">
            <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
              Переменные
            </div>
            {variables.map((v) => (
              <div key={v} className="space-y-1.5">
                <Label htmlFor={`var-${v}`} className="font-mono text-xs text-(--color-primary)">
                  {'{{'}
                  {v}
                  {'}}'}
                </Label>
                <Textarea
                  id={`var-${v}`}
                  value={values[v] ?? ''}
                  onChange={(e) => update(v, e.target.value)}
                  rows={2}
                  placeholder={`Значение для ${v}`}
                />
              </div>
            ))}
          </div>
        ) : null}

        <div className="space-y-2 p-3">
          <div className="flex items-center justify-between">
            <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
              Предпросмотр
            </div>
            <div className="text-[10px] text-(--color-muted-foreground)">
              {charCount.toLocaleString('ru-RU')} симв • ~{tokenEstimate.toLocaleString('ru-RU')} ток
            </div>
          </div>
          <div className="whitespace-pre-wrap rounded-md border border-(--color-border) bg-(--color-muted)/40 p-3 text-xs">
            <HighlightedPreview content={prompt.content} values={values} />
          </div>
        </div>

        {error ? (
          <div className="mx-3 mb-3 rounded-md border border-(--color-destructive) bg-(--color-destructive)/10 p-3 text-xs text-(--color-destructive)">
            {error}
          </div>
        ) : null}
      </div>

      <div className="border-t border-(--color-border) p-3 space-y-2">
        {!canInsert && canInsertReason ? (
          <div className="text-[10px] text-amber-500">{canInsertReason}</div>
        ) : null}
        <div className="flex gap-2">
          <Button
            type="submit"
            disabled={submitting || !canInsert}
            className="flex-1"
            title="Вставить в активную вкладку (⌘↵)"
          >
            <Send className="mr-1.5 h-4 w-4" />
            {submitting ? 'Вставляю…' : 'Вставить'}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={copyToClipboard}
            title="Скопировать в буфер (⌘⇧C)"
            aria-label="Скопировать"
          >
            <Copy className="h-4 w-4" />
          </Button>
          {onInsertAll ? (
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={async () => {
                if (loadedSaved) {
                  await setSavedVars(prompt.id, values);
                }
                onInsertAll(preview);
              }}
              title="Вставить во все открытые поддерживаемые вкладки"
              aria-label="Вставить во все"
            >
              <Zap className="h-4 w-4" />
            </Button>
          ) : null}
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={share}
            title="Создать публичную ссылку"
            aria-label="Поделиться"
          >
            <Share2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </form>
  );
}

/**
 * Рендерит preview с подсветкой `{{переменных}}`:
 *   - незаполненные (values[name] === '') → purple mono placeholder
 *   - заполненные → подсвеченный фон со значением
 */
function HighlightedPreview({
  content,
  values,
}: {
  content: string;
  values: Record<string, string>;
}) {
  const VAR_RE = /\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}/gu;
  const nodes: React.ReactNode[] = [];
  let lastIdx = 0;
  let i = 0;
  for (const match of content.matchAll(VAR_RE)) {
    const start = match.index ?? 0;
    if (start > lastIdx) {
      nodes.push(<span key={`t${i}`}>{content.slice(lastIdx, start)}</span>);
    }
    const name = match[1];
    const val = values[name];
    if (val && val.length > 0) {
      nodes.push(
        <span
          key={`v${i}`}
          className="rounded bg-(--color-primary)/15 px-0.5 text-(--color-foreground)"
          title={`{{${name}}}`}
        >
          {val}
        </span>,
      );
    } else {
      nodes.push(
        <span
          key={`p${i}`}
          className="font-mono text-(--color-primary)/70"
        >
          {'{{'}
          {name}
          {'}}'}
        </span>,
      );
    }
    lastIdx = start + match[0].length;
    i++;
  }
  if (lastIdx < content.length) {
    nodes.push(<span key="tail">{content.slice(lastIdx)}</span>);
  }
  return <>{nodes}</>;
}
