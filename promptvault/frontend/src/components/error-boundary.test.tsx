import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { ErrorBoundary } from "./error-boundary"

function ThrowingChild(): React.ReactNode {
  throw new Error("Test explosion")
}

describe("ErrorBoundary", () => {
  it("renders children when no error", () => {
    render(
      <ErrorBoundary>
        <p>Hello</p>
      </ErrorBoundary>
    )
    expect(screen.getByText("Hello")).toBeDefined()
  })

  it("renders error UI when child throws", () => {
    // Suppress console.error from React during expected error
    const spy = vi.spyOn(console, "error").mockImplementation(() => {})

    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>
    )

    expect(screen.getByText("Что-то пошло не так")).toBeDefined()
    expect(screen.getByText("Test explosion")).toBeDefined()
    expect(screen.getByText("Перезагрузить")).toBeDefined()

    spy.mockRestore()
  })
})
