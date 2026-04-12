import { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from './ui/button';

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('[ErrorBoundary]', error, errorInfo);
  }

  render() {
    if (!this.state.error) return this.props.children;

    return (
      <div className="flex h-full flex-col items-center justify-center gap-4 p-6 text-center">
        <div className="rounded-full bg-(--color-destructive)/10 p-3">
          <AlertTriangle className="h-6 w-6 text-(--color-destructive)" />
        </div>
        <div className="space-y-1">
          <h2 className="text-sm font-semibold">Что-то пошло не так</h2>
          <p className="text-xs text-(--color-muted-foreground)">
            Расширение столкнулось с ошибкой. Попробуйте перезагрузить панель.
          </p>
        </div>
        <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 p-2 text-left font-mono text-[10px] text-(--color-muted-foreground)">
          {this.state.error.message}
        </div>
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
