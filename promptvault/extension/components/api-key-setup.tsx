import { useState, type FormEvent } from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { sendBg } from '../lib/bg-client';
import { setApiKey, setApiBase } from '../lib/storage';
import { ApiError } from '../lib/types';

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
      setError('Адрес сервера должен начинаться с http:// или https://');
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

  return (
    <div className="flex h-full flex-col gap-4 p-5">
      <div className="space-y-2">
        <h1 className="text-lg font-semibold">ПромтЛаб</h1>
        <p className="text-sm text-(--color-muted-foreground)">
          Чтобы расширение получило доступ к вашей библиотеке промптов, создайте API-ключ
          в настройках аккаунта и вставьте его ниже.
        </p>
      </div>

      <form onSubmit={onSubmit} className="flex flex-1 flex-col gap-4">
        <div className="space-y-2">
          <Label htmlFor="api-base">Адрес сервера</Label>
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

        <div className="mt-auto">
          <Button type="submit" disabled={loading || !key} className="w-full">
            {loading ? 'Проверяю…' : 'Подключить'}
          </Button>
        </div>
      </form>
    </div>
  );
}
