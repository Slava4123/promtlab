import { useEffect, useState } from 'react';
import {
  ArrowLeft,
  LogOut,
  RefreshCw,
  CheckCircle2,
  XCircle,
  Monitor,
  Sun,
  Moon,
} from 'lucide-react';
import { Button } from './ui/button';
import { Label } from './ui/label';
import { Input } from './ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { Badge } from './ui/badge';
import { clearApiKey, setApiBase, setApiKey, setTheme, type Theme } from '../lib/storage';
import { sendBg } from '../lib/bg-client';
import { ApiError } from '../lib/types';
import { useToast } from './ui/toaster';

interface Props {
  apiKey: string;
  apiBase: string;
  theme: Theme;
  onBack: () => void;
}

export function SettingsView({ apiKey, apiBase, theme, onBack }: Props) {
  const [newKey, setNewKey] = useState('');
  const [newBase, setNewBase] = useState(apiBase);
  const [updating, setUpdating] = useState(false);
  const [health, setHealth] = useState<'checking' | 'ok' | 'fail'>('checking');
  const { toast } = useToast();

  useEffect(() => {
    let cancelled = false;
    const check = async () => {
      setHealth('checking');
      try {
        await sendBg({ type: 'api.health' });
        if (!cancelled) setHealth('ok');
      } catch {
        if (!cancelled) setHealth('fail');
      }
    };
    void check();
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleThemeChange(next: Theme) {
    await setTheme(next);
    toast({ title: 'Тема обновлена', variant: 'success', durationMs: 1500 });
  }

  async function handleBaseChange() {
    const normalized = newBase.trim().replace(/\/$/, '');
    if (!normalized) return;
    await setApiBase(normalized);
    toast({ title: 'Адрес сервера сохранён', variant: 'success' });
  }

  async function handleKeyChange() {
    if (!newKey.trim().startsWith('pvlt_')) {
      toast({ title: 'Ключ должен начинаться с pvlt_', variant: 'error' });
      return;
    }
    setUpdating(true);
    try {
      // Сохраняем и сразу проверяем через backend
      await setApiKey(newKey.trim());
      await sendBg({ type: 'api.validateKey', key: newKey.trim() });
      toast({ title: 'Ключ обновлён', variant: 'success' });
      setNewKey('');
      setHealth('ok');
    } catch (err) {
      if (err instanceof ApiError && err.code === 'unauthorized') {
        toast({ title: 'Ключ недействителен', variant: 'error' });
      } else {
        toast({ title: 'Не удалось проверить ключ', variant: 'error' });
      }
    } finally {
      setUpdating(false);
    }
  }

  async function handleLogout() {
    await clearApiKey();
    toast({ title: 'Вы вышли', variant: 'info' });
    // App автоматически покажет ApiKeySetup при следующем render через useSettings
  }

  const keyMasked = apiKey.slice(0, 9) + '…' + apiKey.slice(-4);

  return (
    <div className="flex h-full flex-col">
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
        <h2 className="flex-1 text-sm font-semibold">Настройки</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-5">
        {/* Health status */}
        <section className="space-y-2">
          <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
            Состояние
          </div>
          <div className="flex items-center gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-3">
            {health === 'checking' ? (
              <Badge variant="outline">Проверяю…</Badge>
            ) : health === 'ok' ? (
              <>
                <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                <span className="text-xs">Backend доступен</span>
              </>
            ) : (
              <>
                <XCircle className="h-4 w-4 text-(--color-destructive)" />
                <span className="text-xs">Backend недоступен</span>
              </>
            )}
            <div className="flex-1" />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={async () => {
                setHealth('checking');
                try {
                  await sendBg({ type: 'api.health' });
                  setHealth('ok');
                } catch {
                  setHealth('fail');
                }
              }}
              aria-label="Перепроверить"
            >
              <RefreshCw className="h-3.5 w-3.5" />
            </Button>
          </div>
        </section>

        {/* Theme */}
        <section className="space-y-2">
          <Label htmlFor="theme-select">Тема</Label>
          <Select value={theme} onValueChange={(v: string) => handleThemeChange(v as Theme)}>
            <SelectTrigger id="theme-select" aria-label="Тема">
              <SelectValue placeholder="Выберите тему" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="system">
                <span className="flex items-center gap-2">
                  <Monitor className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
                  Системная
                </span>
              </SelectItem>
              <SelectItem value="light">
                <span className="flex items-center gap-2">
                  <Sun className="h-3.5 w-3.5 text-amber-500" />
                  Светлая
                </span>
              </SelectItem>
              <SelectItem value="dark">
                <span className="flex items-center gap-2">
                  <Moon className="h-3.5 w-3.5 text-(--color-muted-foreground)" />
                  Тёмная
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </section>

        {/* API base */}
        <section className="space-y-2">
          <Label htmlFor="api-base-setting">Адрес сервера</Label>
          <div className="flex gap-2">
            <Input
              id="api-base-setting"
              value={newBase}
              onChange={(e) => setNewBase(e.target.value)}
              placeholder="https://promtlabs.ru"
            />
            <Button
              type="button"
              variant="outline"
              onClick={handleBaseChange}
              disabled={newBase === apiBase || !newBase.trim()}
            >
              Ок
            </Button>
          </div>
        </section>

        {/* Current API key */}
        <section className="space-y-2">
          <Label>Текущий ключ</Label>
          <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 px-3 py-2 font-mono text-xs">
            {keyMasked}
          </div>
        </section>

        {/* Replace API key */}
        <section className="space-y-2">
          <Label htmlFor="new-key">Заменить ключ</Label>
          <Input
            id="new-key"
            type="password"
            value={newKey}
            onChange={(e) => setNewKey(e.target.value)}
            placeholder="pvlt_..."
            autoComplete="off"
          />
          <Button
            type="button"
            variant="outline"
            onClick={handleKeyChange}
            disabled={updating || !newKey.trim()}
            className="w-full"
          >
            {updating ? 'Проверяю…' : 'Обновить ключ'}
          </Button>
        </section>

        {/* Logout */}
        <section className="pt-4">
          <Button
            type="button"
            variant="destructive"
            onClick={handleLogout}
            className="w-full"
          >
            <LogOut className="mr-2 h-4 w-4" />
            Выйти
          </Button>
        </section>

        {/* Version info */}
        <div className="pt-2 text-center text-[10px] text-(--color-muted-foreground)">
          ПромтЛаб Chrome Extension • v{chrome.runtime.getManifest?.().version ?? '0.1.0'}
        </div>
      </div>
    </div>
  );
}
