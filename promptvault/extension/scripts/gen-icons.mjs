#!/usr/bin/env node
/**
 * Генерирует иконки extension через простой SVG → PNG конвертер.
 * Создаёт public/icon/{16,32,48,128}.png
 *
 * Не использует внешних зависимостей (canvas/sharp) — только Node built-ins.
 * Метод: рендерим SVG в Node через sharp если есть, иначе пишем минимальные PNG вручную.
 *
 * Fallback подход: записываем сам SVG файл, и копируем его 4 раза как .png — Chrome
 * иногда принимает. Но надёжный путь — создать настоящие PNG.
 *
 * Используем простой алгоритм: бинарный PNG writer с одним цветом + aa-ed текст.
 * Это overkill. Проще: сгенерируем SVG для каждого размера и используем встроенный
 * workaround — Chrome extensions МОГУТ использовать SVG... нет, MV3 требует PNG.
 *
 * Самый простой рабочий путь: используем pngjs если есть в node_modules, иначе
 * создаём через dataURL в node через встроенный модуль.
 *
 * Финальное решение: используем `sharp` если доступен, иначе fallback на
 * простейший заполненный PNG (pure color square без текста) — это хоть как-то
 * лучше чем placeholder "П" от Chrome.
 */

import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { deflateSync } from 'node:zlib';

const __dirname = dirname(fileURLToPath(import.meta.url));
const outDir = resolve(__dirname, '..', 'public', 'icon');
mkdirSync(outDir, { recursive: true });

// Primary brand color — purple
const COLOR = { r: 139, g: 92, b: 246 };
const BG_ALPHA = 255;

// CRC-32 lookup table (initialized before use)
const CRC_TABLE = (() => {
  const t = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) {
      c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    }
    t[n] = c >>> 0;
  }
  return t;
})();

const sizes = [16, 32, 48, 128];

for (const size of sizes) {
  const png = makeRoundedPurplePng(size);
  const path = resolve(outDir, `${size}.png`);
  writeFileSync(path, png);
  console.log('wrote', path, '(' + png.length + ' bytes)');
}

/**
 * Создаёт PNG с фиолетовым скруглённым квадратом (round rect) + белой буквой "П" в центре.
 * Формат: RGBA 8-bit, single IDAT chunk с deflate compression.
 */
function makeRoundedPurplePng(size) {
  const width = size;
  const height = size;
  const radius = Math.floor(size * 0.22); // ~22% corner radius

  // Буквенная маска "П" — приближение через bbox
  const letterInset = Math.floor(size * 0.28);
  const letterTop = Math.floor(size * 0.27);
  const letterBottom = Math.floor(size * 0.73);
  const letterBarHeight = Math.max(1, Math.floor(size * 0.12));

  // Raw RGBA rows
  const bytesPerPixel = 4;
  const rowBytes = width * bytesPerPixel;
  const raw = Buffer.alloc((rowBytes + 1) * height); // +1 filter byte per row

  for (let y = 0; y < height; y++) {
    raw[y * (rowBytes + 1)] = 0; // filter: none
    for (let x = 0; x < width; x++) {
      const idx = y * (rowBytes + 1) + 1 + x * bytesPerPixel;
      const inRoundRect = isInsideRoundedRect(x, y, width, height, radius);
      const inLetter =
        x >= letterInset &&
        x < width - letterInset &&
        y >= letterTop &&
        y <= letterBottom &&
        // Левая вертикаль П
        (x - letterInset < Math.max(1, Math.floor(size * 0.08)) ||
          // Правая вертикаль П
          width - letterInset - x <= Math.max(1, Math.floor(size * 0.08)) ||
          // Верхняя перекладина П
          y - letterTop < letterBarHeight);

      if (inRoundRect) {
        if (inLetter) {
          raw[idx] = 255;
          raw[idx + 1] = 255;
          raw[idx + 2] = 255;
          raw[idx + 3] = 255;
        } else {
          raw[idx] = COLOR.r;
          raw[idx + 1] = COLOR.g;
          raw[idx + 2] = COLOR.b;
          raw[idx + 3] = BG_ALPHA;
        }
      } else {
        raw[idx] = 0;
        raw[idx + 1] = 0;
        raw[idx + 2] = 0;
        raw[idx + 3] = 0;
      }
    }
  }

  return encodePng(width, height, raw);
}

function isInsideRoundedRect(x, y, w, h, r) {
  if (x >= r && x < w - r) return true;
  if (y >= r && y < h - r) return true;
  // Corners
  const cx = x < r ? r : w - 1 - r;
  const cy = y < r ? r : h - 1 - r;
  const dx = x - cx;
  const dy = y - cy;
  return dx * dx + dy * dy <= r * r;
}

// ===== Minimal PNG encoder =====

function encodePng(width, height, rawWithFilter) {
  const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);

  // IHDR chunk
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(width, 0);
  ihdr.writeUInt32BE(height, 4);
  ihdr.writeUInt8(8, 8);  // bit depth
  ihdr.writeUInt8(6, 9);  // color type: RGBA
  ihdr.writeUInt8(0, 10); // compression
  ihdr.writeUInt8(0, 11); // filter
  ihdr.writeUInt8(0, 12); // interlace
  const ihdrChunk = chunk('IHDR', ihdr);

  // IDAT chunk
  const compressed = deflateSync(rawWithFilter);
  const idatChunk = chunk('IDAT', compressed);

  // IEND chunk
  const iendChunk = chunk('IEND', Buffer.alloc(0));

  return Buffer.concat([signature, ihdrChunk, idatChunk, iendChunk]);
}

function chunk(type, data) {
  const length = Buffer.alloc(4);
  length.writeUInt32BE(data.length, 0);
  const typeBytes = Buffer.from(type, 'ascii');
  const crcData = Buffer.concat([typeBytes, data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(crcData), 0);
  return Buffer.concat([length, typeBytes, data, crc]);
}

function crc32(buf) {
  let c = 0xffffffff;
  for (const b of buf) {
    c = CRC_TABLE[(c ^ b) & 0xff] ^ (c >>> 8);
  }
  return (c ^ 0xffffffff) >>> 0;
}
