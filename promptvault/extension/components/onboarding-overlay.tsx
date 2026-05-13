import { useEffect, useState } from 'react';
import { Sparkles, Search, Keyboard, Star, X, GitBranch, ArrowLeft } from 'lucide-react';
import { Button } from './ui/button';
import { isOnboardingSeen, markOnboardingSeen } from '../lib/storage';

export function OnboardingOverlay() {
  const [visible, setVisible] = useState(false);
  const [step, setStep] = useState(0);

  useEffect(() => {
    let cancelled = false;
    void isOnboardingSeen().then((seen) => {
      if (!cancelled && !seen) setVisible(true);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!visible) return null;

  const steps = [
    {
      icon: <Sparkles className="h-8 w-8 text-(--color-brand)" />,
      title: 'Добро пожаловать в ПромтЛаб',
      description:
        'Расширение хранит вашу библиотеку промптов и вставляет их прямо в ChatGPT, Claude, Gemini, Perplexity, Yandex GPT, GigaChat, DeepSeek, Mistral и Qwen — одним кликом.',
    },
    {
      icon: <Search className="h-8 w-8 text-(--color-brand)" />,
      title: 'Поиск и Cmd+K',
      description:
        'Введите запрос в шапке или нажмите Cmd+K — глобальная команд-палитра ищет по промптам, коллекциям и тегам.',
    },
    {
      icon: <Keyboard className="h-8 w-8 text-(--color-brand)" />,
      title: 'Горячие клавиши',
      description:
        'Ctrl+Shift+K — открыть панель. Esc — назад. ↑↓Enter — навигация по списку. Ctrl+R — обновить.',
    },
    {
      icon: <GitBranch className="h-8 w-8 text-(--color-brand)" />,
      title: 'Цепочки промптов',
      description:
        'Объединяйте несколько промптов в последовательность. Запускайте — копируйте готовый текст, отправляйте в AI-чат, нажимайте «Далее» для следующего шага.',
    },
    {
      icon: <Star className="h-8 w-8 text-amber-500" />,
      title: 'Закреплённые и избранное',
      description:
        'Наведите на карточку и кликните ⭐ или 📌 — промпт попадёт в соответствующий таб для быстрого доступа.',
    },
  ];

  const current = steps[step];
  const isLast = step === steps.length - 1;

  async function close() {
    setVisible(false);
    await markOnboardingSeen();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 backdrop-blur-sm">
      <div className="m-3 w-full max-w-sm rounded-xl border border-(--color-border) bg-(--color-card) p-5 shadow-xl">
        <div className="mb-3 flex items-start justify-between">
          <div className="rounded-lg bg-(--color-brand-muted) p-2">{current.icon}</div>
          <button
            type="button"
            onClick={close}
            className="rounded p-1 text-(--color-muted-foreground) hover:text-(--color-foreground)"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <h2 className="mb-1 text-base font-semibold">{current.title}</h2>
        <p className="mb-4 text-xs text-(--color-muted-foreground)">{current.description}</p>
        <div className="mb-3 flex gap-1">
          {steps.map((_, i) => (
            <div
              key={i}
              className={
                'h-1 flex-1 rounded-full transition-colors ' +
                (i <= step ? 'bg-(--color-primary)' : 'bg-(--color-border)')
              }
            />
          ))}
        </div>
        <div className="flex gap-2">
          {step > 0 ? (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setStep(step - 1)}
              aria-label="Назад"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
          ) : (
            <Button type="button" variant="ghost" onClick={close} className="flex-1">
              Пропустить
            </Button>
          )}
          <Button
            type="button"
            onClick={() => {
              if (isLast) void close();
              else setStep(step + 1);
            }}
            className="flex-1"
          >
            {isLast ? 'Начать' : 'Дальше'}
          </Button>
        </div>
      </div>
    </div>
  );
}
