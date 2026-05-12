import { Link } from "react-router-dom"
import { FileQuestion } from "lucide-react"

export function NotFoundPage() {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 px-6 py-8 text-center">
      <FileQuestion className="h-12 w-12 text-(--color-muted-foreground)" />
      <div>
        <h2 className="text-base font-semibold text-(--color-foreground)">Не найдено</h2>
        <p className="mt-1 text-sm text-(--color-muted-foreground)">
          Страница не существует или ещё не реализована.
        </p>
      </div>
      <Link
        to="/"
        className="rounded-md bg-(--color-primary) px-3 py-1.5 text-sm text-(--color-primary-foreground)"
      >
        На главную
      </Link>
    </div>
  )
}
