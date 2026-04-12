// Обёртка над chrome.storage.local — единственное место для пользовательских данных.
// Хранит: API-ключ, base URL, theme, сохранённые значения переменных, недавнее использование.

const KEY_API_KEY = 'pv.apiKey';
const KEY_API_BASE = 'pv.apiBase';
const KEY_THEME = 'pv.theme';
const KEY_SAVED_VARS = 'pv.savedVars';      // { [promptId]: { [varName]: value } }
const KEY_LAST_INSERT = 'pv.lastInsert';    // { promptId, text, ts } — для undo
const KEY_WORKSPACE = 'pv.workspaceId';     // null = Личное, number = team id
const KEY_COLLECTION = 'pv.collectionId';   // null = все, number = collection id
const KEY_LOCAL_RECENT = 'pv.localRecent';  // LocalRecentEntry[]
const KEY_ONBOARDING = 'pv.onboardingSeen'; // boolean
const KEY_PROMPT_CACHE = 'pv.promptCache';  // { items: Prompt[], ts: number }

const DEFAULT_API_BASE = import.meta.env?.WXT_API_BASE ?? 'https://promtlabs.ru';

export type Theme = 'light' | 'dark' | 'system';

export interface StoredSettings {
  apiKey: string | null;
  apiBase: string;
  theme: Theme;
}

export interface LastInsert {
  promptId: number;
  text: string;
  ts: number;
}

export async function getSettings(): Promise<StoredSettings> {
  const data = await chrome.storage.local.get([KEY_API_KEY, KEY_API_BASE, KEY_THEME]);
  return {
    apiKey: data[KEY_API_KEY] ?? null,
    apiBase: data[KEY_API_BASE] ?? DEFAULT_API_BASE,
    theme: (data[KEY_THEME] as Theme) ?? 'system',
  };
}

export async function setApiKey(key: string): Promise<void> {
  await chrome.storage.local.set({ [KEY_API_KEY]: key });
}

export async function clearApiKey(): Promise<void> {
  await chrome.storage.local.remove(KEY_API_KEY);
}

export async function setApiBase(base: string): Promise<void> {
  await chrome.storage.local.set({ [KEY_API_BASE]: base });
}

export async function setTheme(theme: Theme): Promise<void> {
  await chrome.storage.local.set({ [KEY_THEME]: theme });
}

export function onSettingsChanged(cb: (settings: StoredSettings) => void): () => void {
  const listener = (
    changes: { [key: string]: chrome.storage.StorageChange },
    area: chrome.storage.AreaName,
  ) => {
    if (area !== 'local') return;
    if (KEY_API_KEY in changes || KEY_API_BASE in changes || KEY_THEME in changes) {
      void getSettings().then(cb);
    }
  };
  chrome.storage.onChanged.addListener(listener);
  return () => chrome.storage.onChanged.removeListener(listener);
}

// ===== Saved variable values (per-prompt history) =====

export async function getSavedVars(promptId: number): Promise<Record<string, string>> {
  const data = await chrome.storage.local.get(KEY_SAVED_VARS);
  const all = (data[KEY_SAVED_VARS] ?? {}) as Record<string, Record<string, string>>;
  return all[String(promptId)] ?? {};
}

export async function setSavedVars(
  promptId: number,
  values: Record<string, string>,
): Promise<void> {
  const data = await chrome.storage.local.get(KEY_SAVED_VARS);
  const all = (data[KEY_SAVED_VARS] ?? {}) as Record<string, Record<string, string>>;
  all[String(promptId)] = values;
  await chrome.storage.local.set({ [KEY_SAVED_VARS]: all });
}

// ===== Last insertion (for undo) =====

export async function getLastInsert(): Promise<LastInsert | null> {
  const data = await chrome.storage.local.get(KEY_LAST_INSERT);
  return (data[KEY_LAST_INSERT] as LastInsert | undefined) ?? null;
}

export async function setLastInsert(entry: LastInsert): Promise<void> {
  await chrome.storage.local.set({ [KEY_LAST_INSERT]: entry });
}

export async function clearLastInsert(): Promise<void> {
  await chrome.storage.local.remove(KEY_LAST_INSERT);
}

// ===== Workspace + Collection =====

export interface WorkspaceSelection {
  workspaceId: number | null;  // null = personal, number = team id
  collectionId: number | null; // null = all, number = specific collection
}

export async function getWorkspace(): Promise<WorkspaceSelection> {
  const data = await chrome.storage.local.get([KEY_WORKSPACE, KEY_COLLECTION]);
  return {
    workspaceId: data[KEY_WORKSPACE] ?? null,
    collectionId: data[KEY_COLLECTION] ?? null,
  };
}

export async function setWorkspace(workspaceId: number | null): Promise<void> {
  await chrome.storage.local.set({ [KEY_WORKSPACE]: workspaceId });
  // При смене workspace — сброс выбранной коллекции
  await chrome.storage.local.remove(KEY_COLLECTION);
}

export async function setCollection(collectionId: number | null): Promise<void> {
  await chrome.storage.local.set({ [KEY_COLLECTION]: collectionId });
}

export function onWorkspaceChanged(cb: (sel: WorkspaceSelection) => void): () => void {
  const listener = (
    changes: { [key: string]: chrome.storage.StorageChange },
    area: chrome.storage.AreaName,
  ) => {
    if (area !== 'local') return;
    if (KEY_WORKSPACE in changes || KEY_COLLECTION in changes) {
      void getWorkspace().then(cb);
    }
  };
  chrome.storage.onChanged.addListener(listener);
  return () => chrome.storage.onChanged.removeListener(listener);
}

// ===== Local recent (backup history) =====

export interface LocalRecentEntry {
  promptId: number;
  title: string;
  insertedAt: number;
  targetHost: string | null;
}

const MAX_LOCAL_RECENT = 20;

export async function getLocalRecent(): Promise<LocalRecentEntry[]> {
  const data = await chrome.storage.local.get(KEY_LOCAL_RECENT);
  return (data[KEY_LOCAL_RECENT] as LocalRecentEntry[] | undefined) ?? [];
}

export async function addLocalRecent(entry: LocalRecentEntry): Promise<void> {
  const existing = await getLocalRecent();
  const filtered = existing.filter((e) => e.promptId !== entry.promptId);
  const updated = [entry, ...filtered].slice(0, MAX_LOCAL_RECENT);
  await chrome.storage.local.set({ [KEY_LOCAL_RECENT]: updated });
}

// ===== Onboarding =====

export async function isOnboardingSeen(): Promise<boolean> {
  const data = await chrome.storage.local.get(KEY_ONBOARDING);
  return Boolean(data[KEY_ONBOARDING]);
}

export async function markOnboardingSeen(): Promise<void> {
  await chrome.storage.local.set({ [KEY_ONBOARDING]: true });
}

// ===== Prompts offline cache =====

export interface PromptCacheEntry<T = unknown> {
  items: T;
  ts: number;
}

const PROMPT_CACHE_TTL_MS = 5 * 60 * 1000; // 5 минут fresh, потом stale но ещё используется как fallback

export async function getCachedPrompts<T>(): Promise<PromptCacheEntry<T> | null> {
  const data = await chrome.storage.local.get(KEY_PROMPT_CACHE);
  return (data[KEY_PROMPT_CACHE] as PromptCacheEntry<T> | undefined) ?? null;
}

export async function setCachedPrompts<T>(items: T): Promise<void> {
  await chrome.storage.local.set({
    [KEY_PROMPT_CACHE]: { items, ts: Date.now() } as PromptCacheEntry<T>,
  });
}

export function isCacheFresh(entry: PromptCacheEntry<unknown>): boolean {
  return Date.now() - entry.ts < PROMPT_CACHE_TTL_MS;
}
