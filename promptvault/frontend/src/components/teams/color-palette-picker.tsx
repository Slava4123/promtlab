import { useState } from "react"
import { Check, Pipette } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { BRAND_COLORS, isValidHex } from "@/lib/branding/colors"

interface ColorPalettePickerProps {
  value: string
  onChange: (next: string) => void
  // id — для связки <label htmlFor>; кейс «обернёт всю секцию в <Label>»
  // тоже работает, но лучше явный id для accessibility scanner'ов.
  id?: string
  // disabled — чтобы синхронно с form-state блокировать interactions.
  disabled?: boolean
}

// ColorPalettePicker — визуальный выбор #RRGGBB:
//   1. 12 preset-чипов из BRAND_COLORS, кликабельные. Активный — с галочкой.
//   2. Свёрнутая секция «Свой цвет» с native <input type="color"> + текстовое
//      поле hex. Пользователь вводит либо палитрой, либо native picker'ом —
//      обе ветки пишут одно значение через onChange.
//
// onChange вызывается ТОЛЬКО на валидный hex. Невалидный текст в кастом-поле
// не уходит наверх (юзер видит aria-invalid стиль).
export function ColorPalettePicker({ value, onChange, id, disabled }: ColorPalettePickerProps) {
  const [customOpen, setCustomOpen] = useState(() => {
    // Открыть кастом-секцию, если сохранённое значение не из палитры —
    // иначе юзер не поймёт, откуда взялся незнакомый цвет.
    if (!value) return false
    return !BRAND_COLORS.some((c) => c.value.toLowerCase() === value.toLowerCase())
  })
  const [customDraft, setCustomDraft] = useState(value)

  const handlePresetClick = (next: string) => {
    if (disabled) return
    onChange(next)
    setCustomDraft(next)
  }

  const handleCustomBlur = () => {
    const v = customDraft.trim()
    if (v === value) return
    if (isValidHex(v)) onChange(v)
  }

  const handleNativePicker = (e: React.ChangeEvent<HTMLInputElement>) => {
    // <input type="color"> всегда отдаёт валидный #rrggbb (lowercase).
    const next = e.target.value.toUpperCase()
    setCustomDraft(next)
    onChange(next)
  }

  return (
    <div className="space-y-3">
      {/* Preset palette */}
      <div
        role="radiogroup"
        aria-label="Выбор brand-цвета из палитры"
        id={id}
        className="flex flex-wrap gap-2"
      >
        {BRAND_COLORS.map((c) => {
          const active = c.value.toLowerCase() === value.toLowerCase()
          return (
            <button
              key={c.value}
              type="button"
              role="radio"
              aria-checked={active}
              aria-label={`${c.label} (${c.value})`}
              title={`${c.label} (${c.value})`}
              disabled={disabled}
              onClick={() => handlePresetClick(c.value)}
              style={{ backgroundColor: c.value }}
              className={cn(
                "size-8 rounded-full ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
                active ? "ring-2 ring-ring ring-offset-2" : "hover:scale-110",
              )}
            >
              {active && <Check className="mx-auto size-4 text-white drop-shadow" aria-hidden="true" />}
            </button>
          )
        })}
      </div>

      {/* Toggle для custom */}
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={() => setCustomOpen((o) => !o)}
        disabled={disabled}
        className="h-7 px-2 text-[0.78rem] text-muted-foreground hover:text-foreground"
      >
        <Pipette className="mr-1.5 size-3.5" aria-hidden="true" />
        {customOpen ? "Скрыть свой цвет" : "Свой цвет"}
      </Button>

      {/* Custom row: native color picker + текстовый hex */}
      {customOpen && (
        <div className="flex items-center gap-2">
          <Label className="sr-only" htmlFor="brand-color-native">
            Picker цвета
          </Label>
          <input
            id="brand-color-native"
            type="color"
            value={isValidHex(customDraft) ? customDraft : "#000000"}
            onChange={handleNativePicker}
            disabled={disabled}
            className="size-10 cursor-pointer rounded-md border border-input bg-transparent disabled:cursor-not-allowed disabled:opacity-50"
            aria-label="Выбор цвета через системный picker"
          />
          <Input
            type="text"
            value={customDraft}
            onChange={(e) => setCustomDraft(e.target.value)}
            onBlur={handleCustomBlur}
            disabled={disabled}
            placeholder="#0066CC"
            aria-invalid={customDraft !== "" && !isValidHex(customDraft)}
            className="w-32 font-mono text-sm uppercase"
            maxLength={7}
          />
        </div>
      )}
    </div>
  )
}
