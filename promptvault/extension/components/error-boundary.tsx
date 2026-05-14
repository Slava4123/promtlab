import { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from './ui/button';

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

// SessionStorage flag — чтобы не reload'ить бесконечно, если auto-recover
// сам провалился. После успешного mount флаг чистится в main.tsx через
// removeItem на старте (см. также frontend веб-приложения — тот же паттерн).
const CHUNK_RELOAD_FLAG = 'pv.chunkErrorReloaded';

function isChunkLoadError(err: Error): boolean {
  const msg = err.message ?? '';
  return (
    msg.includes('Failed to fetch dynamically imported module') ||
    msg.includes('Importing a module script failed') ||
    msg.includes('Loading chunk') ||
    msg.includes('Loading CSS chunk')
  );
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('[ErrorBoundary]', error, errorInfo);

    // Auto-recover при chunk-load-error (после rebuild extension старый
    // side-panel держит ссылки на исчезнувшие chunk-файлы). Один раз
    // делаем reload — sessionStorage flag предотвращает infinite loop,
    // если сама перезагрузка не помогла.
    if (isChunkLoadError(error)) {
      try {
        const already = sessionStorage.getItem(CHUNK_RELOAD_FLAG);
        if (!already) {
          sessionStorage.setItem(CHUNK_RELOAD_FLAG, '1');
          location.reload();
          return;
        }
      } catch {
        // sessionStorage недоступен — fallthrough на error UI
      }
    }
  }

  render() {
    if (!this.state.error) return this.props.children;
    const chunkError = isChunkLoadError(this.state.error);

    return (
      <div className="flex h-full flex-col items-center justify-center gap-4 p-6 text-center">
        <div className="rounded-full bg-(--color-destructive)/10 p-3">
          <AlertTriangle className="h-6 w-6 text-(--color-destructive)" />
        </div>
        <div className="space-y-1">
          <h2 className="text-sm font-semibold">
            {chunkError ? 'Расширение обновилось' : 'Что-то пошло не так'}
          </h2>
          <p className="text-xs text-(--color-muted-foreground)">
            {chunkError
              ? 'Загрузка новой версии не удалась автоматически. Закройте панель (×) и откройте заново.'
              : 'Расширение столкнулось с ошибкой. Попробуйте перезагрузить панель.'}
          </p>
        </div>
        {!chunkError && (
          <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 p-2 text-left font-mono text-[10px] text-(--color-muted-foreground)">
            {this.state.error.message}
          </div>
        )}
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => {
            this.setState({ error: null });
            location.reload();
          }}
        >
          <RefreshCw className="mr-2 h-3.5 w-3.5" />
          Перезагрузить
        </Button>
      </div>
    );
  }
}
