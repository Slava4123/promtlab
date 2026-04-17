import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { MemoryRouter } from "react-router-dom"

import { HelpPopover } from "../help-popover"

function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>)
}

describe("HelpPopover", () => {
  it("рендерит trigger с aria-label", () => {
    renderWithRouter(
      <HelpPopover title="Title" ariaLabel="Подсказка по полю">
        <p>Body</p>
      </HelpPopover>,
    )
    expect(screen.getByRole("button", { name: /Подсказка по полю/i })).toBeInTheDocument()
    expect(screen.queryByText("Title")).not.toBeInTheDocument()
  })

  it("клик по trigger открывает popover с заголовком и контентом", async () => {
    const user = userEvent.setup()
    renderWithRouter(
      <HelpPopover title="Создание промпта">
        <p>Промпт — это шаблон</p>
      </HelpPopover>,
    )
    await user.click(screen.getByRole("button", { name: /Подсказка/i }))
    expect(await screen.findByText("Создание промпта")).toBeInTheDocument()
    expect(screen.getByText("Промпт — это шаблон")).toBeInTheDocument()
  })

  it("Esc закрывает popover", async () => {
    const user = userEvent.setup()
    renderWithRouter(
      <HelpPopover title="X"><p>Content</p></HelpPopover>,
    )
    await user.click(screen.getByRole("button", { name: /Подсказка/i }))
    expect(await screen.findByText("Content")).toBeInTheDocument()
    await user.keyboard("{Escape}")
    // base-ui Popover убирает контент из DOM при закрытии (Portal unmount)
    expect(screen.queryByText("Content")).not.toBeInTheDocument()
  })

  it("отображает learnMore-ссылку, если задана", async () => {
    const user = userEvent.setup()
    renderWithRouter(
      <HelpPopover title="X" learnMoreHref="/help/mcp" learnMoreLabel="К FAQ">
        <p>Body</p>
      </HelpPopover>,
    )
    await user.click(screen.getByRole("button", { name: /Подсказка/i }))
    const link = await screen.findByRole("link", { name: /К FAQ/i })
    expect(link).toHaveAttribute("href", "/help/mcp")
  })
})
