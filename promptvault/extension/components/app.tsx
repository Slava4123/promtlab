import { useCallback, useEffect, useState } from 'react';
import {
  QueryClient,
  QueryClientProvider,
  useQueryClient,
} from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { useSettings } from '../hooks/use-settings';
import { useApplyTheme } from '../hooks/use-theme';
import { useActiveTab } from '../hooks/use-active-tab';
import { usePrompt } from '../hooks/use-prompts';
import { ApiKeySetup } from './api-key-setup';
import { Home } from './home';
import { VariableForm } from './variable-form';
import { SettingsView } from './settings-view';
import { ErrorBoundary } from './error-boundary';
import { OnboardingOverlay } from './onboarding-overlay';
import { ToasterProvider, useToast } from './ui/toaster';
import { sendBg } from '../lib/bg-client';
import { extractVariables } from '../lib/template';
import { ApiError, type Prompt } from '../lib/types';
import { hostLabel } from '../lib/messages';
import { addLocalRecent } from '../lib/storage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        // Не retry'им 4xx ошибки (401 unauthorized, 404 not_found и т.д.) —
        // 4xx это permanent ошибки, retry не поможет.
        if (
          error instanceof ApiError &&
          error.status >= 400 &&
          error.status < 500 &&
          error.status !== 0
        ) {
          return false;
        }
        return failureCount < 2;
      },
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 3000),
      refetchOnWindowFocus: true,
      staleTime: 30_000,
    },
  },
});

export function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ToasterProvider>
          <Root />
          <OnboardingOverlay />
        </ToasterProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

function Root() {
  const settings = useSettings();
  useApplyTheme(settings?.theme ?? null);

  if (!settings) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    );
  }

  if (!settings.apiKey) {
    return <ApiKeySetup initialBase={settings.apiBase} />;
  }

  return (
    <Authenticated
      apiKey={settings.apiKey}
      apiBase={settings.apiBase}
      theme={settings.theme}
    />
  );
}

type View = 'home' | 'variable-form' | 'settings';

interface AuthenticatedProps {
  apiKey: string;
  apiBase: string;
  theme: 'light' | 'dark' | 'system';
}

function Authenticated({ apiKey, apiBase, theme }: AuthenticatedProps) {
  const [view, setView] = useState<View>('home');
  const [selected, setSelected] = useState<Prompt | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [insertError, setInsertError] = useState<string | null>(null);
  const [highlightedId, setHighlightedId] = useState<number | null>(null);
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const activeTab = useActiveTab();

  const needFullPrompt = selected !== null && selected.content.length < 20;
  const fullPromptQuery = usePrompt(needFullPrompt ? selected!.id : null);
  const activePrompt = needFullPrompt ? fullPromptQuery.data ?? null : selected;

  const handleInsert = useCallback(
    async (text: string) => {
      if (!activePrompt) return;
      setSubmitting(true);
      setInsertError(null);
      try {
        await sendBg({ type: 'cmd.insertPrompt', text });
        void sendBg({ type: 'api.incrementUsage', promptId: activePrompt.id }).catch(
          () => undefined,
        );
        void queryClient.invalidateQueries({ queryKey: ['prompts', 'recent'] });

        // Сохраняем в локальной истории (backup если backend отвалится)
        void addLocalRecent({
          promptId: activePrompt.id,
          title: activePrompt.title,
          insertedAt: Date.now(),
          targetHost: activeTab.host,
        });

        const targetLabel = hostLabel(activeTab.host) ?? 'цель';

        setHighlightedId(activePrompt.id);
        setTimeout(() => setHighlightedId(null), 900);

        toast({
          title: `Вставлено в ${targetLabel}`,
          description: activePrompt.title,
          variant: 'success',
          durationMs: 5000,
          action: {
            label: 'Отменить',
            icon: 'undo',
            onClick: async () => {
              try {
                await sendBg({ type: 'cmd.undoInsert' });
                toast({ title: 'Отменено', variant: 'info', durationMs: 1500 });
              } catch {
                toast({
                  title: 'Не получилось отменить',
                  description: 'Возможно, вы уже отредактировали поле',
                  variant: 'error',
                });
              }
            },
          },
        });

        setSelected(null);
        setView('home');
      } catch (err) {
        if (err instanceof ApiError) {
          if (err.code === 'no_target') {
            setInsertError(
              'Откройте ChatGPT, Claude, Gemini или Perplexity в активной вкладке.',
            );
          } else if (err.code === 'unauthorized') {
            setInsertError('Ключ больше не действителен.');
          } else {
            setInsertError('Не удалось вставить промпт. Попробуйте ещё раз.');
          }
        } else {
          setInsertError('Не удалось вставить промпт.');
        }
      } finally {
        setSubmitting(false);
      }
    },
    [activePrompt, activeTab.host, queryClient, toast],
  );

  const handleInsertAll = useCallback(
    async (text: string) => {
      if (!activePrompt) return;
      try {
        const result = await sendBg({ type: 'cmd.insertPromptAll', text });
        void sendBg({ type: 'api.incrementUsage', promptId: activePrompt.id }).catch(
          () => undefined,
        );
        void queryClient.invalidateQueries({ queryKey: ['prompts', 'recent'] });

        if (result.successes === 0) {
          toast({
            title: 'Нет открытых вкладок',
            description: 'Откройте ChatGPT, Claude, Gemini или Perplexity',
            variant: 'error',
          });
          return;
        }

        toast({
          title: `Вставлено в ${result.successes} ${pluralTabs(result.successes)}`,
          description: activePrompt.title,
          variant: 'success',
          durationMs: 4000,
        });

        setSelected(null);
        setView('home');
      } catch {
        toast({ title: 'Не удалось вставить во все вкладки', variant: 'error' });
      }
    },
    [activePrompt, queryClient, toast],
  );

  // Auto-insert для промптов БЕЗ переменных — пропускаем VariableForm вообще
  useEffect(() => {
    if (view !== 'variable-form') return;
    if (!activePrompt) return;
    const vars = extractVariables(activePrompt.content);
    if (vars.length === 0 && !submitting) {
      void handleInsert(activePrompt.content);
    }
  }, [view, activePrompt, submitting, handleInsert]);

  function handleSelect(p: Prompt) {
    setSelected(p);
    setInsertError(null);
    setView('variable-form');
  }

  function backToHome() {
    setSelected(null);
    setInsertError(null);
    setView('home');
  }

  // Settings view
  if (view === 'settings') {
    return (
      <SettingsView
        apiKey={apiKey}
        apiBase={apiBase}
        theme={theme}
        onBack={() => setView('home')}
      />
    );
  }

  // VariableForm (либо loading full prompt)
  if (view === 'variable-form') {
    if (needFullPrompt && fullPromptQuery.isPending) {
      return (
        <div className="flex h-full items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
        </div>
      );
    }

    if (activePrompt) {
      // Если без переменных — useEffect auto-submit уже запустился; показываем loader
      const vars = extractVariables(activePrompt.content);
      if (vars.length === 0) {
        return (
          <div className="flex h-full items-center justify-center gap-2">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
            <span className="text-xs text-(--color-muted-foreground)">Вставляю…</span>
          </div>
        );
      }

      return (
        <VariableForm
          prompt={activePrompt}
          onBack={backToHome}
          onSubmit={handleInsert}
          onInsertAll={handleInsertAll}
          submitting={submitting}
          error={insertError}
          canInsert={activeTab.supported}
          canInsertReason={
            !activeTab.supported
              ? 'Откройте ChatGPT, Claude, Gemini или Perplexity'
              : undefined
          }
        />
      );
    }
  }

  // Home
  return (
    <Home
      onSelect={handleSelect}
      onOpenSettings={() => setView('settings')}
      highlightedId={highlightedId}
    />
  );
}

function pluralTabs(n: number): string {
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod10 === 1 && mod100 !== 11) return 'вкладку';
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return 'вкладки';
  return 'вкладок';
}
