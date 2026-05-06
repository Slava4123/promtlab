import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"

import { ColorPalettePicker } from "./color-palette-picker"
import { BRAND_COLORS } from "@/lib/branding/colors"

describe("ColorPalettePicker", () => {
  it("рендерит все 12 preset-цветов", () => {
    render(<ColorPalettePicker value="#0066CC" onChange={() => {}} />)
    expect(screen.getAllByRole("radio")).toHaveLength(BRAND_COLORS.length)
  })

  it("отмечает текущий выбранный цвет aria-checked", () => {
    render(<ColorPalettePicker value="#6366F1" onChange={() => {}} />)
    const active = screen.getByRole("radio", { checked: true })
    expect(active).toHaveAttribute("aria-label", expect.stringContaining("#6366F1"))
  })

  it("вызывает onChange при клике на preset", () => {
    const onChange = vi.fn()
    render(<ColorPalettePicker value="" onChange={onChange} />)
    fireEvent.click(screen.getByRole("radio", { name: /Корпоративный синий/i }))
    expect(onChange).toHaveBeenCalledWith("#0066CC")
  })

  it("открывает custom-секцию для не-preset value", () => {
    render(<ColorPalettePicker value="#abcdef" onChange={() => {}} />)
    // когда значение не из палитры — секция должна быть открыта при mount,
    // и в hex-инпуте видна сохранённая величина
    const hexInput = screen.getByPlaceholderText("#0066CC") as HTMLInputElement
    expect(hexInput.value).toBe("#abcdef")
  })

  it("custom hex: невалидный ввод не вызывает onChange", () => {
    const onChange = vi.fn()
    render(<ColorPalettePicker value="" onChange={onChange} />)
    fireEvent.click(screen.getByRole("button", { name: /Свой цвет/i }))
    const hexInput = screen.getByPlaceholderText("#0066CC")
    fireEvent.change(hexInput, { target: { value: "#XYZ" } })
    fireEvent.blur(hexInput)
    expect(onChange).not.toHaveBeenCalled()
  })

  it("custom hex: валидный ввод вызывает onChange", () => {
    const onChange = vi.fn()
    render(<ColorPalettePicker value="" onChange={onChange} />)
    fireEvent.click(screen.getByRole("button", { name: /Свой цвет/i }))
    const hexInput = screen.getByPlaceholderText("#0066CC")
    fireEvent.change(hexInput, { target: { value: "#a1b2c3" } })
    fireEvent.blur(hexInput)
    expect(onChange).toHaveBeenCalledWith("#a1b2c3")
  })

  it("disabled блокирует preset-клики", () => {
    const onChange = vi.fn()
    render(<ColorPalettePicker value="" onChange={onChange} disabled />)
    fireEvent.click(screen.getByRole("radio", { name: /Корпоративный синий/i }))
    expect(onChange).not.toHaveBeenCalled()
  })
})
