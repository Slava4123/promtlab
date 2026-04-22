import { describe, it, expect } from "vitest"
import { renderHook, act } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import type { PropsWithChildren } from "react"
import { createElement } from "react"
import { useAnalyticsFilter } from "./use-analytics-filter"

function wrapperWithUrl(initial: string) {
  return function Wrapper({ children }: PropsWithChildren) {
    return createElement(MemoryRouter, { initialEntries: [initial] }, children)
  }
}

describe("useAnalyticsFilter", () => {
  it("читает tag и collection из URL", () => {
    const { result } = renderHook(() => useAnalyticsFilter(), {
      wrapper: wrapperWithUrl("/analytics?tag=5&collection=7"),
    })
    expect(result.current.tagId).toBe(5)
    expect(result.current.collectionId).toBe(7)
  })

  it("null при отсутствии params", () => {
    const { result } = renderHook(() => useAnalyticsFilter(), {
      wrapper: wrapperWithUrl("/analytics"),
    })
    expect(result.current.tagId).toBeNull()
    expect(result.current.collectionId).toBeNull()
  })

  it("setTagId обновляет URL", () => {
    const { result } = renderHook(() => useAnalyticsFilter(), {
      wrapper: wrapperWithUrl("/analytics"),
    })
    act(() => {
      result.current.setTagId(42)
    })
    expect(result.current.tagId).toBe(42)
  })

  it("reset очищает оба фильтра", () => {
    const { result } = renderHook(() => useAnalyticsFilter(), {
      wrapper: wrapperWithUrl("/analytics?tag=1&collection=2"),
    })
    act(() => {
      result.current.reset()
    })
    expect(result.current.tagId).toBeNull()
    expect(result.current.collectionId).toBeNull()
  })
})
