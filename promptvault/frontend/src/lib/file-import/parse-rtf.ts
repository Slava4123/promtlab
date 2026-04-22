import { MAX_PROMPT_CONTENT_LENGTH } from "@/lib/constants"
import { FileImportError, type ParseResult } from "./types"
import { normalizeText } from "./parse-text"

// Простой regex-based RTF → plain text stripper. Обрабатывает:
//  - группы `{...}` — удаляются целиком если это header/font/color table
//  - control words `\something` — вырезаются, некоторые семантические (\par, \line, \tab) → символы
//  - hex-escapes `\'XX` — декодируются как Windows-1251 (стандарт для RU RTF)
//  - unicode escapes `\uN` — декодируются
//  - \* и \\ \{ \} — литералы
//
// Это НЕ полноценный RTF-парсер. Покрывает 90% RU-кейсов (.rtf из Word).
// Для сложных с изображениями/таблицами нужен отдельный parser — откладываем.

// Mapping Windows-1251 → UTF-16 для hex-escape декодирования. Диапазон 0x80-0xFF.
// 0x00-0x7F идентичны ASCII.
// Источник: https://en.wikipedia.org/wiki/Windows-1251
const CP1251_TO_UNICODE: Record<number, number> = {
  0x80: 0x0402, 0x81: 0x0403, 0x82: 0x201a, 0x83: 0x0453, 0x84: 0x201e,
  0x85: 0x2026, 0x86: 0x2020, 0x87: 0x2021, 0x88: 0x20ac, 0x89: 0x2030,
  0x8a: 0x0409, 0x8b: 0x2039, 0x8c: 0x040a, 0x8d: 0x040c, 0x8e: 0x040b,
  0x8f: 0x040f, 0x90: 0x0452, 0x91: 0x2018, 0x92: 0x2019, 0x93: 0x201c,
  0x94: 0x201d, 0x95: 0x2022, 0x96: 0x2013, 0x97: 0x2014, 0x99: 0x2122,
  0x9a: 0x0459, 0x9b: 0x203a, 0x9c: 0x045a, 0x9d: 0x045c, 0x9e: 0x045b,
  0x9f: 0x045f, 0xa0: 0x00a0, 0xa1: 0x040e, 0xa2: 0x045e, 0xa3: 0x0408,
  0xa4: 0x00a4, 0xa5: 0x0490, 0xa6: 0x00a6, 0xa7: 0x00a7, 0xa8: 0x0401,
  0xa9: 0x00a9, 0xaa: 0x0404, 0xab: 0x00ab, 0xac: 0x00ac, 0xad: 0x00ad,
  0xae: 0x00ae, 0xaf: 0x0407, 0xb0: 0x00b0, 0xb1: 0x00b1, 0xb2: 0x0406,
  0xb3: 0x0456, 0xb4: 0x0491, 0xb5: 0x00b5, 0xb6: 0x00b6, 0xb7: 0x00b7,
  0xb8: 0x0451, 0xb9: 0x2116, 0xba: 0x0454, 0xbb: 0x00bb, 0xbc: 0x0458,
  0xbd: 0x0405, 0xbe: 0x0455, 0xbf: 0x0457,
}

// Cyrillic 0xC0-0xFF → U+0410-U+044F (A-Я, а-я).
function cp1251ToChar(byte: number): string {
  if (byte < 0x80) return String.fromCharCode(byte)
  if (byte >= 0xc0 && byte <= 0xff) return String.fromCharCode(0x0410 + (byte - 0xc0))
  const mapped = CP1251_TO_UNICODE[byte]
  if (mapped !== undefined) return String.fromCharCode(mapped)
  return "" // неизвестный байт — пропускаем
}

// Группы которые вырезаем целиком (служебные метаданные Word).
const DESTINATION_GROUPS = [
  "fonttbl",
  "colortbl",
  "stylesheet",
  "info",
  "pict", // картинки — не извлекаем
  "bin",
  "object",
  "generator",
  "header",
  "footer",
  "revtbl",
  "listtable",
  "listoverridetable",
  "rsidtbl",
  "mmathPr",
]

export async function parseRtfFile(file: File): Promise<ParseResult> {
  const text = await file.text()

  if (!text.trimStart().startsWith("{\\rtf")) {
    throw new FileImportError(
      "PARSE_FAILED",
      "Файл не похож на RTF (должен начинаться с {\\rtf)",
      "rtf",
    )
  }

  const stripped = stripRtf(text)
  const normalized = normalizeText(stripped).trim()

  if (normalized.length === 0) {
    throw new FileImportError("EMPTY_RESULT", "RTF не содержит текста", "rtf")
  }

  const { content, truncated } = truncate(normalized)

  return {
    content,
    kind: "rtf",
    filename: file.name,
    originalBytes: file.size,
    truncated,
    warnings: [],
  }
}

// Главный стриппер. State-machine c depth-counter для групп.
function stripRtf(input: string): string {
  let output = ""
  let i = 0
  const len = input.length

  // Пропускаем первую `{` и initial `\rtf1...` header — до первого space/CRLF.
  while (i < len) {
    const ch = input[i]

    // Группа '{' — проверим, не destination-группа ли (напр. {\fonttbl …}).
    if (ch === "{") {
      const groupEnd = findMatchingBrace(input, i)
      if (groupEnd === -1) {
        i++
        continue
      }
      const groupContent = input.slice(i + 1, groupEnd)
      const firstControl = groupContent.match(/^\s*\\\*?\s*\\?([a-zA-Z]+)/)
      if (firstControl && DESTINATION_GROUPS.includes(firstControl[1])) {
        // Пропускаем группу целиком.
        i = groupEnd + 1
        continue
      }
      // Иначе — рекурсия на содержимое.
      output += stripRtf(groupContent)
      i = groupEnd + 1
      continue
    }

    if (ch === "}") {
      i++
      continue
    }

    // Control word: \something, может быть с цифровым параметром.
    if (ch === "\\") {
      const next = input[i + 1]

      // Литералы.
      if (next === "\\" || next === "{" || next === "}") {
        output += next
        i += 2
        continue
      }

      // hex-escape \'XX — 2 hex-цифры.
      if (next === "'") {
        const hex = input.slice(i + 2, i + 4)
        if (/^[0-9a-fA-F]{2}$/.test(hex)) {
          const byte = parseInt(hex, 16)
          output += cp1251ToChar(byte)
          i += 4
          continue
        }
        i += 2
        continue
      }

      // Unicode \uN или \uN?  (N — signed 16-bit; может быть отрицательным).
      if (next === "u") {
        const match = input.slice(i + 2).match(/^(-?\d+)/)
        if (match) {
          let code = parseInt(match[1], 10)
          if (code < 0) code = 65536 + code
          output += String.fromCharCode(code)
          i += 2 + match[1].length
          // Пропустить следующий символ (fallback для не-Unicode readers).
          if (input[i] === "?") i++
          else if (input[i] === " ") i++
          else if (input[i] === "\\") {
            // другой control
          } else if (/[a-zA-Z0-9]/.test(input[i] ?? "")) i++
          continue
        }
      }

      // Обычный control word: \word или \wordN, заканчивается non-alphanumeric.
      const wordMatch = input.slice(i + 1).match(/^([a-zA-Z]+)(-?\d+)?/)
      if (wordMatch) {
        const word = wordMatch[1]
        // Семантические control words → символы.
        if (word === "par" || word === "line" || word === "sect") output += "\n"
        else if (word === "tab") output += "\t"
        else if (word === "emdash") output += "—"
        else if (word === "endash") output += "–"
        else if (word === "lquote") output += "\u2018"
        else if (word === "rquote") output += "\u2019"
        else if (word === "ldblquote") output += "\u201c"
        else if (word === "rdblquote") output += "\u201d"
        else if (word === "bullet") output += "•"
        // Остальные — вырезаем (форматирование: \b \i \fs24 \cf1 и т.д.)
        i += 1 + word.length + (wordMatch[2]?.length ?? 0)
        // Space/newline после control word — delimiter, не литерал.
        if (input[i] === " ") i++
        continue
      }

      // Unknown escape — пропускаем \.
      i++
      continue
    }

    // Newlines в RTF — обычно служебные, не данные. Игнорируем.
    if (ch === "\r" || ch === "\n") {
      i++
      continue
    }

    output += ch
    i++
  }

  return output
}

function findMatchingBrace(input: string, start: number): number {
  if (input[start] !== "{") return -1
  let depth = 1
  let i = start + 1
  while (i < input.length) {
    const ch = input[i]
    if (ch === "\\" && (input[i + 1] === "{" || input[i + 1] === "}" || input[i + 1] === "\\")) {
      i += 2
      continue
    }
    if (ch === "{") depth++
    else if (ch === "}") {
      depth--
      if (depth === 0) return i
    }
    i++
  }
  return -1
}

function truncate(input: string): { content: string; truncated: boolean } {
  if (input.length <= MAX_PROMPT_CONTENT_LENGTH) {
    return { content: input, truncated: false }
  }
  return { content: input.slice(0, MAX_PROMPT_CONTENT_LENGTH), truncated: true }
}
