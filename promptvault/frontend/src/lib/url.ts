// isSafeHttpsUrl проверяет, что URL имеет схему https: и непустой host.
// Используется для defense-in-depth перед рендером user-submitted ссылок в
// <a href>. Backend валидирует при записи (usecases/team/branding), но если
// прорвётся javascript:/data:/file:/https:backslash-path — frontend не
// должен его открывать.
//
// Проверка host отсекает вырожденные случаи типа "https:\\\\evil.com"
// (URL-парсер трактует \\ как path, host=пустой).
export function isSafeHttpsUrl(raw: string | null | undefined): boolean {
  if (!raw) return false
  try {
    const u = new URL(raw)
    return u.protocol === "https:" && u.host.length > 0
  } catch {
    return false
  }
}
