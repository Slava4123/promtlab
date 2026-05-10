// MN-61 — exhaustive switch helper для discriminated unions.
//
// Use:
//
//   switch (result.kind) {
//     case "ok": ...
//     case "totp_required": ...
//     default: assertNever(result, "LoginResult")
//   }
//
// Если позже в LoginResult добавится `{ kind: "rate_limited" }` и забыть
// добавить case в switch, TypeScript отдаст compile-error в `assertNever(result)`
// — невозможно cast'нуть `{ kind: "rate_limited" }` в `never`.
//
// Без assertNever default-ветка молча проглатывала бы новый вариант — баг
// проявлялся бы только в runtime ("ничего не происходит при rate-limited login").

export function assertNever(value: never, context = "discriminated union"): never {
  throw new Error(
    `assertNever(${context}): unhandled case ${JSON.stringify(value)}`,
  )
}
