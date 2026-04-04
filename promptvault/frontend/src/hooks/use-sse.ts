import { useState, useRef, useCallback, useEffect } from "react"
import { getAccessToken, ensureFreshToken } from "@/api/client"

interface UseSSEReturn {
  data: string
  isStreaming: boolean
  error: string | null
  start: (path: string, body: object) => Promise<void>
  abort: () => void
}

export function useSSE(): UseSSEReturn {
  const [data, setData] = useState("")
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const controllerRef = useRef<AbortController | null>(null)

  const abort = useCallback(() => {
    controllerRef.current?.abort()
    controllerRef.current = null
    setIsStreaming(false)
  }, [])

  const start = useCallback(async (path: string, body: object) => {
    // Abort any existing stream
    controllerRef.current?.abort()

    const controller = new AbortController()
    controllerRef.current = controller

    setData("")
    setError(null)
    setIsStreaming(true)

    try {
      const makeRequest = () => {
        const token = getAccessToken()
        return fetch(`/api${path}`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify(body),
          signal: controller.signal,
        })
      }

      let res = await makeRequest()

      // Auto-refresh on 401 and retry
      if (res.status === 401) {
        try {
          await ensureFreshToken()
          res = await makeRequest()
        } catch {
          throw new Error("Сессия истекла. Войдите заново")
        }
      }

      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: "Ошибка запроса" }))
        throw new Error(body.error || `HTTP ${res.status}`)
      }

      const reader = res.body?.getReader()
      if (!reader) {
        throw new Error("Streaming не поддерживается")
      }

      const decoder = new TextDecoder()
      let buffer = ""

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        const lines = buffer.split("\n")
        buffer = lines.pop() ?? ""

        let currentEvent = ""
        let dataLines: string[] = []

        for (const line of lines) {
          if (line.startsWith("event: ")) {
            currentEvent = line.slice(7).trim()
            continue
          }

          if (line.startsWith("data: ")) {
            const payload = line.slice(6)

            if (payload === "[DONE]") {
              // Flush any remaining data
              if (dataLines.length > 0) {
                const text = dataLines.join("\n")
                setData((prev) => prev + text)
                dataLines = []
              }
              setIsStreaming(false)
              return
            }

            if (currentEvent === "error") {
              currentEvent = ""
              throw new Error(payload)
            }

            dataLines.push(payload)
            continue
          }

          // Empty line = SSE event separator → flush accumulated data lines
          if (line === "" && dataLines.length > 0) {
            const text = dataLines.join("\n")
            setData((prev) => prev + text)
            dataLines = []
          }
          currentEvent = ""
        }
      }

      setIsStreaming(false)
    } catch (err: unknown) {
      if (err instanceof Error && err.name === "AbortError") {
        // User-initiated cancellation — not an error
        return
      }
      setError(err instanceof Error ? err.message : "Неизвестная ошибка")
      setIsStreaming(false)
    }
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      controllerRef.current?.abort()
    }
  }, [])

  return { data, isStreaming, error, start, abort }
}
