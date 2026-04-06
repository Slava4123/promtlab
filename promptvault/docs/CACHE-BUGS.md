# Баги кэширования и инвалидации TanStack Query

## P0 — Критические

### 1. Workspace store — нет очистки кэша при переключении команды
**Файл:** `frontend/src/components/layout/app-sidebar.tsx`
При переключении команды/личного пространства — все team-scoped запросы остаются в кэше. Пользователь видит промпты/коллекции/теги старой команды.
**Фикс:** Импортировать `useQueryClient` и инвалидировать при переключении.

### 2. useAcceptInvitation — неполная инвалидация
**Файл:** `frontend/src/hooks/use-teams.ts`
Не инвалидируются: `["team-invitations"]`, `["team"]`. Владелец не видит что приглашение принято.

### 3. useDeclineInvitation — неполная инвалидация
**Файл:** `frontend/src/hooks/use-teams.ts`
Не инвалидируется: `["team-invitations"]`. Владелец не видит отклонение.

### 4. 403 редирект — team-view.tsx
**Файл:** `frontend/src/pages/team-view.tsx`
Удалённый пользователь видит страницу команды. Нужен редирект на `/teams` при ошибке.

### 5. 403 distinction в api/client.ts
**Файл:** `frontend/src/api/client.ts`
Нет различения 403 от других ошибок. Нужен специфический error message.

## P1 — Важные

### 6. useToggleFavorite — не инвалидирует prompt detail
**Файл:** `frontend/src/hooks/use-prompts.ts`
Добавить: `["prompt", id]` в инвалидацию.

### 7. useUpdateCollection — не инвалидирует конкретную коллекцию
**Файл:** `frontend/src/hooks/use-collections.ts`
Добавить: `["collection", id]` в инвалидацию.

### 8. useDeleteCollection — не обновляет промпты
**Файл:** `frontend/src/hooks/use-collections.ts`
Промпты показывают удалённую коллекцию. Добавить: `["prompts"]`.

### 9. useCreatePrompt — не обновляет теги
**Файл:** `frontend/src/hooks/use-prompts.ts`
Добавить: `["tags"]` в инвалидацию.

### 10. Query key inconsistency — collections и tags
`useCollections` ключ: `["collections", { teamId }]`, но инвалидация: `["collections"]`.
`useTags` ключ: `["tags", { teamId }]`, но инвалидация: `["tags"]`.
Инвалидация по partial key работает (prefix match), но лучше быть explicit.

### 11. collection-view.tsx — добавление промптов в коллекцию
**Файл:** `frontend/src/pages/collection-view.tsx`
После добавления промптов через диалог — список промптов коллекции не обновляется.

## Файлы для изменения

1. `frontend/src/hooks/use-teams.ts` — инвалидации accept/decline
2. `frontend/src/hooks/use-prompts.ts` — toggle favorite + create
3. `frontend/src/hooks/use-collections.ts` — update/delete инвалидации
4. `frontend/src/pages/team-view.tsx` — 403 редирект
5. `frontend/src/api/client.ts` — 403 error
6. `frontend/src/components/layout/app-sidebar.tsx` — очистка кэша при переключении
