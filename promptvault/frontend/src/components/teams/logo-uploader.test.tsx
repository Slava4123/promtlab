import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

import { LogoUploader } from "./logo-uploader"

vi.mock("@/api/branding", () => ({
  uploadLogo: vi.fn(async () => ({
    logo_source: "file",
    effective_logo_url: "/api/teams/x/branding/logo",
    size_bytes: 100,
    content_type: "image/png",
  })),
  deleteLogo: vi.fn(async () => ({ logo_source: "none" })),
}))
vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

import { uploadLogo } from "@/api/branding"
import { toast } from "sonner"

function wrap(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  return <QueryClientProvider client={qc}>{ui}</QueryClientProvider>
}

describe("LogoUploader", () => {
  it("показывает URL-инпут в режиме url", () => {
    render(
      wrap(
        <LogoUploader
          slug="x"
          logoSource="url"
          logoUrl="https://cdn.example/logo.png"
          onLogoUrlChange={() => {}}
        />,
      ),
    )
    expect(screen.getByPlaceholderText(/cdn\.example/i)).toBeDefined()
  })

  it("переключение на «Загрузить файл» показывает drop-зону", () => {
    render(
      wrap(
        <LogoUploader
          slug="x"
          logoSource="url"
          logoUrl=""
          onLogoUrlChange={() => {}}
        />,
      ),
    )
    fireEvent.click(screen.getByRole("button", { name: /Загрузить файл/i }))
    expect(screen.getByText(/Перетащите файл/i)).toBeDefined()
  })

  it("файл >1MB → toast.error без вызова upload", async () => {
    render(
      wrap(
        <LogoUploader
          slug="x"
          logoSource="file"
          logoUrl=""
          onLogoUrlChange={() => {}}
        />,
      ),
    )
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    const tooBig = new File([new Uint8Array(1024 * 1024 + 1)], "big.png", { type: "image/png" })
    fireEvent.change(input, { target: { files: [tooBig] } })
    expect(toast.error).toHaveBeenCalledWith("Файл больше 1 МБ")
    expect(uploadLogo).not.toHaveBeenCalled()
  })

  it("неверный тип (text/plain) → toast.error без вызова upload", async () => {
    render(
      wrap(
        <LogoUploader
          slug="x"
          logoSource="file"
          logoUrl=""
          onLogoUrlChange={() => {}}
        />,
      ),
    )
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    const txt = new File(["hello"], "a.txt", { type: "text/plain" })
    fireEvent.change(input, { target: { files: [txt] } })
    expect(toast.error).toHaveBeenCalledWith("Поддерживаются только PNG, JPEG и WebP")
    expect(uploadLogo).not.toHaveBeenCalled()
  })

  it("валидный PNG → uploadLogo вызван с File", async () => {
    render(
      wrap(
        <LogoUploader
          slug="acme"
          logoSource="file"
          logoUrl=""
          onLogoUrlChange={() => {}}
        />,
      ),
    )
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    const png = new File([new Uint8Array([0x89, 0x50, 0x4e, 0x47])], "logo.png", { type: "image/png" })
    fireEvent.change(input, { target: { files: [png] } })
    await waitFor(() => expect(uploadLogo).toHaveBeenCalledWith("acme", png))
  })
})
