import "@testing-library/jest-dom/vitest"
import { vi } from "vitest"

// jsdom не имеет matchMedia (используется в theme-store для prefers-color-scheme).
// Стандартный мок из MDN/Vitest docs.
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})
