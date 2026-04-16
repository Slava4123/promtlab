package mcpserver

import (
	"sync"
	"time"
)

// listCache — простой in-memory TTL cache для list_collections / list_tags (P-11).
// LLM-клиенты (Claude Desktop, Claude Code) часто зовут эти tools подряд
// одним запросом — one-shot кеш 30-60с снижает нагрузку на БД без сложной
// инвалидации.
//
// list_prompts НЕ кешируется: слишком много параметров фильтра,
// cache hit rate был бы низкий, а риск отдать stale выше (юзер только что
// создал промпт и сразу листит — должен его увидеть).
//
// При CUD операциях явную инвалидацию не делаем — TTL 30с приемлемо.
type cacheEntry struct {
	data      any
	expiresAt time.Time
}

type listCache struct {
	m   sync.Map
	ttl time.Duration
}

func newListCache(ttl time.Duration) *listCache {
	return &listCache{ttl: ttl}
}

// Get возвращает значение и true, если оно есть и не протухло.
// Протухшие записи остаются в map'е — очистим при Set (lazy GC).
func (c *listCache) Get(key string) (any, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, false
	}
	entry := v.(*cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.m.Delete(key)
		return nil, false
	}
	return entry.data, true
}

// Set сохраняет значение с TTL.
func (c *listCache) Set(key string, data any) {
	c.m.Store(key, &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	})
}

// InvalidateUser удаляет все записи конкретного юзера. Вызывается при
// CUD операциях через тот же MCP-handler (чтобы не ждать TTL после собственных изменений).
// key-префикс: "<kind>:<userID>:" — поэтому сравниваем по содержимому ключа.
func (c *listCache) InvalidateUser(userID uint, kinds ...string) {
	// kinds example: "collections", "tags". Prefix для матчинга:
	prefixes := make([]string, 0, len(kinds))
	for _, k := range kinds {
		prefixes = append(prefixes, k+":"+uintToStr(userID)+":")
	}
	c.m.Range(func(k, _ any) bool {
		key := k.(string)
		for _, p := range prefixes {
			if hasPrefix(key, p) {
				c.m.Delete(k)
				break
			}
		}
		return true
	})
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

func uintToStr(n uint) string {
	// Быстрее чем fmt.Sprint для горячего пути.
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
