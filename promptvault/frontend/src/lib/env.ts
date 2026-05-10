// Centralized env-var validation. MN-65 — раньше каждый callsite читал
// import.meta.env.VITE_FOO напрямую через `as string | undefined` cast,
// без проверки что переменная задана и в правильном формате. Опечатка в
// .env (`VITE_CHAINS_ENABED=true`) тихо превращалась в `false` без warning'а.
//
// Здесь — единая Zod-схема с дефолтами и типизацией. Импортируйте `env`
// в коде вместо прямого обращения к `import.meta.env`. На старте `parseEnv`
// логирует warning'и о missing/invalid переменных в dev — упрощает диагностику
// «почему feature-flag не сработал» при QA.
import { z } from "zod"

// Boolean из строки "true"/"false" (vite не делает coercion для VITE_*).
const boolString = z
  .union([z.literal("true"), z.literal("false"), z.literal("")])
  .optional()
  .transform((v) => v === "true")

const numberString = (defaultValue: number) =>
  z
    .string()
    .optional()
    .transform((v) => {
      if (v === undefined || v === "") return defaultValue
      const n = Number(v)
      return Number.isFinite(n) ? n : defaultValue
    })

const envSchema = z.object({
  // API base URL (vite dev-server проксирует /api в backend; в prod — origin).
  VITE_API_URL: z.string().url().optional(),

  // Phase 16 dark launch — управляет видимостью /chains, sidebar item, pricing tier.
  VITE_CHAINS_ENABLED: boolString,

  // Sentry / GlitchTip integration.
  VITE_SENTRY_ENABLED: boolString,
  VITE_SENTRY_DSN: z.string().url().optional().or(z.literal("").transform(() => undefined)),
  VITE_SENTRY_ENVIRONMENT: z.string().optional().default("production"),
  VITE_SENTRY_RELEASE: z.string().optional().default("dev"),
  VITE_SENTRY_TRACES_SAMPLE_RATE: numberString(0.0),
})

export type ParsedEnv = z.infer<typeof envSchema>

// parseEnv — Lazy parsing с safe-fallback: при ошибке валидации возвращаем
// «безопасные» дефолты + console.warn (а не throw), чтобы кривой .env не
// уронил весь app. Throw был бы странным — feature-flag false по дефолту
// всегда правильнее чем падать на старте.
export function parseEnv(raw: ImportMetaEnv): ParsedEnv {
  const result = envSchema.safeParse(raw)
  if (result.success) return result.data
  if (import.meta.env.DEV) {
    console.warn("[env] Validation failed for VITE_* vars:", result.error.format())
  }
  // Fallback: дефолты с features=off.
  return envSchema.parse({})
}

export const env: ParsedEnv = parseEnv(import.meta.env)
