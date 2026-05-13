import { useState, type FormEvent } from 'react';
import { ExternalLink, Sparkles, KeyRound } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { sendBg } from '../lib/bg-client';
import { setApiKey, setApiBase } from '../lib/storage';
import { ApiError } from '../lib/types';
import { openWebPage, deriveFrontendUrl } from '../lib/utils';

interface Props {
  initialBase: string;
}

export function ApiKeySetup({ initialBase }: Props) {
  const [key, setKey] = useState('');
  const [base, setBase] = useState(initialBase);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmedKey = key.trim();
    const trimmedBase = base.trim().replace(/\/$/, '');

    if (!trimmedKey.startsWith('pvlt_')) {
      setError('Ключ должен начинаться с pvlt_');
      return;
    }
    if (!/^https?:\/\//.test(trimmedBase)) {
      setError('Адрес API должен начинаться с http:// или https://');
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await setApiBase(trimmedBase);
      await setApiKey(trimmedKey);
      await sendBg({ type: 'api.validateKey', key: trimmedKey });
    } catch (err) {
      if (err instanceof ApiError && err.code === 'unauthorized') {
        setError('Ключ недействителен. Проверьте что скопировали его целиком, без лишних пробелов.');
      } else if (err instanceof ApiError && err.code === 'network') {
        setError(
          `Нет соединения с ${trimmedBase}. Проверьте адрес и что сервер запущен. ` +
            'Для локальной разработки используйте http://localhost:8080.',
        );
      } else {
        setError('Не удалось проверить ключ. Попробуйте ещё раз или откройте DevTools (F12) → Console.');
      }
      setLoading(false);
    }
  }

  // Web app живёт на frontendUrl (в prod: тот же host что apiBase;
  // в dev :8080 backend → :5173 frontend).
  const frontendUrl = deriveFrontendUrl(base);
  const baseHost = frontendUrl.replace(/^https?:\/\//, '').replace(/\/$/, '') || 'promtlabs.ru';

  function openExternal(path: string) {
    openWebPage(base, path);
  }

  return (
    <div className="flex h-full flex-col gap-4 p-5 overflow-y-auto">
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-(--color-brand)" />
          <h1 className="text-lg font-semibold">ПромтЛаб</h1>
        </div>
        <p className="text-sm text-(--color-muted-foreground)">
          Подключите расширение к вашему аккаунту через API-ключ.
        </p>
      </div>

      {/* Quick links для нового юзера */}
      <div className="grid grid-cols-2 gap-2">
        <button
          type="button"
          onClick={() => openExternal('/sign-up?from=extension')}
          className="flex flex-col items-start gap-1 rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-left hover:bg-(--color-muted)/40"
        >
          <ExternalLink className="h-3.5 w-3.5 text-(--color-brand)" />
          <div className="text-xs font-medium">Создать аккаунт</div>
          <div className="text-[10px] text-(--color-muted-foreground)">На {baseHost}</div>
        </button>
        <button
          type="button"
          onClick={() => openExternal('/settings/integrations?from=extension')}
          className="flex flex-col items-start gap-1 rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-left hover:bg-(--color-muted)/40"
        >
          <KeyRound className="h-3.5 w-3.5 text-(--color-brand)" />
          <div className="text-xs font-medium">Получить ключ</div>
          <div className="text-[10px] text-(--color-muted-foreground)">Настройки → API-ключи</div>
        </button>
      </div>

      <form onSubmit={onSubmit} className="flex flex-1 flex-col gap-4">
        <div className="space-y-2">
          <Label htmlFor="api-base">Адрес API</Label>
          <Input
            id="api-base"
            type="url"
            value={base}
            onChange={(e) => setBase(e.target.value)}
            placeholder="https://promtlabs.ru"
            disabled={loading}
          />
          <p className="text-xs text-(--color-muted-foreground)">
            По умолчанию промышленный сервер. Для локальной разработки — http://localhost:8080.
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="api-key">API-ключ</Label>
          <Input
            id="api-key"
            type="password"
            autoComplete="off"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="pvlt_..."
            disabled={loading}
          />
          <p className="text-xs text-(--color-muted-foreground)">
            Создать ключ: Настройки → API-ключи → Создать.
          </p>
        </div>

        {error ? (
          <div className="rounded-md border border-(--color-destructive) bg-(--color-destructive)/10 p-3 text-sm text-(--color-destructive)">
            {error}
          </div>
        ) : null}

        <div className="mt-auto space-y-2">
          <Button type="submit" variant="brand" disabled={loading || !key} className="w-full">
            {loading ? 'Проверяю…' : 'Подключить'}
          </Button>
          <button
            type="button"
            onClick={() => openExternal('/forgot-password')}
            className="block w-full text-center text-[10px] text-(--color-muted-foreground) hover:underline"
          >
            Забыли пароль?
          </button>
        </div>
      </form>
    </div>
  );
}
