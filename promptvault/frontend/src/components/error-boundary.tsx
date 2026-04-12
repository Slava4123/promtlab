import { Component, type ErrorInfo, type ReactNode } from "react"
import { captureException } from "@/lib/sentry"

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  // Отправляет необработанную React ошибку в Sentry/GlitchTip с componentStack.
  // Noop если Sentry не инициализирован (SDK проверяет клиента внутри).
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    captureException(error, {
      contexts: {
        react: {
          componentStack: errorInfo.componentStack,
        },
      },
    })
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-background p-6">
          <div className="max-w-md text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-red-500/[0.08] ring-1 ring-red-500/10">
              <span className="text-2xl">!</span>
            </div>
            <h1 className="text-lg font-semibold text-foreground">Что-то пошло не так</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Произошла непредвиденная ошибка. Попробуйте перезагрузить страницу.
            </p>
            {this.state.error && (
              <pre className="mt-4 max-h-32 overflow-auto rounded-lg bg-foreground/[0.04] p-3 text-left text-[0.7rem] text-red-400">
                {this.state.error.message}
              </pre>
            )}
            <button
              onClick={() => window.location.reload()}
              className="mt-5 rounded-lg bg-violet-600 px-5 py-2 text-sm font-medium text-white transition-colors hover:bg-violet-500"
            >
              Перезагрузить
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
