import puppeteer from 'puppeteer';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

async function main() {
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();

  const items = [
    { name: 'screenshot-1', w: 1280, h: 800 },
    { name: 'screenshot-2', w: 1280, h: 800 },
    { name: 'screenshot-3', w: 1280, h: 800 },
    { name: 'screenshot-4', w: 1280, h: 800 },
    { name: 'screenshot-5', w: 1280, h: 800 },
    { name: 'promo-small',  w: 440,  h: 280 },
    { name: 'promo-large',  w: 1400, h: 560 },
  ];

  for (const { name, w, h } of items) {
    await page.setViewport({ width: w, height: h, deviceScaleFactor: 1 });
    const htmlPath = path.join(__dirname, `${name}.html`);
    await page.goto(`file:///${htmlPath.replace(/\\/g, '/')}`, { waitUntil: 'load' });
    await page.screenshot({
      path: path.join(__dirname, `${name}.png`),
      type: 'png',
      clip: { x: 0, y: 0, width: w, height: h },
    });
    console.log(`${name}.png (${w}x${h})`);
  }

  await browser.close();
  console.log('Done!');
}

main().catch(console.error);
