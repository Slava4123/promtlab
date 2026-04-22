// jschardet-обёртка для детекции кодировки .txt/.md файлов. Вызывается только
// когда UTF-8 decode дал "кракозябры" (≥10% replacement chars). Если jschardet
// уверенно определил другой encoding — пере-декодируем через TextDecoder.
//
// Поддерживаемые encodings из jschardet нас интересуют: UTF-8, windows-1251
// (основная cp для RU), KOI8-R, IBM866 (cp866), ISO-8859-5.

import jschardet from "jschardet"

const RU_ENCODINGS = ["UTF-8", "windows-1251", "KOI8-R", "IBM866", "ISO-8859-5"]
const CONFIDENCE_THRESHOLD = 0.5

export interface EncodingDetectionResult {
  encoding: string
  confidence: number
  recovered: boolean  // true если мы смогли пере-декодировать
  content: string     // итоговый текст (либо пере-декодированный, либо оригинал)
}

// Пытается пере-декодировать `bytes` используя детекцию. Если detect не уверен
// или encoding не поддерживается TextDecoder'ом — возвращает original (с UTF-8).
export function detectAndDecode(
  bytes: Uint8Array,
  utf8Fallback: string,
): EncodingDetectionResult {
  // jschardet работает с бинарной строкой — сконвертируем Uint8Array
  // (используем только первые 64 KB для скорости).
  const sample = bytes.slice(0, 64 * 1024)
  let binaryString = ""
  for (let i = 0; i < sample.length; i++) {
    binaryString += String.fromCharCode(sample[i])
  }

  let detection: { encoding: string; confidence: number } | null
  try {
    detection = jschardet.detect(binaryString, {
      minimumThreshold: CONFIDENCE_THRESHOLD,
      detectEncodings: RU_ENCODINGS,
    })
  } catch {
    detection = null
  }

  if (
    !detection ||
    !detection.encoding ||
    detection.confidence < CONFIDENCE_THRESHOLD
  ) {
    return {
      encoding: "utf-8",
      confidence: 0,
      recovered: false,
      content: utf8Fallback,
    }
  }

  const normalized = detection.encoding.toLowerCase()
  // Если jschardet сказал UTF-8 — наш оригинал уже валидный, ничего не делаем.
  if (normalized === "utf-8" || normalized === "ascii") {
    return {
      encoding: "utf-8",
      confidence: detection.confidence,
      recovered: false,
      content: utf8Fallback,
    }
  }

  // Пере-декодируем в определённой кодировке. TextDecoder понимает
  // windows-1251/koi8-r/ibm866/iso-8859-5 нативно (Encoding Standard).
  try {
    const decoder = new TextDecoder(normalized, { fatal: false })
    const content = decoder.decode(bytes)
    return {
      encoding: normalized,
      confidence: detection.confidence,
      recovered: true,
      content,
    }
  } catch {
    // Неподдерживаемый браузером encoding — возвращаем UTF-8 fallback.
    return {
      encoding: "utf-8",
      confidence: detection.confidence,
      recovered: false,
      content: utf8Fallback,
    }
  }
}
