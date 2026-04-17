package seo

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image/color"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	ogWidth  = 1200
	ogHeight = 630
	// Палитра ПромтЛаб: фиолетовый бренд + светлые тона.
	ogTitleSize  = 64
	ogFooterSize = 28
	ogPaddingX   = 80
	ogPaddingY   = 100
	ogMaxLines   = 4
)

var (
	bgTopColor    = color.RGBA{R: 0x1a, G: 0x10, B: 0x32, A: 0xff} // #1a1032 deep purple
	bgBottomColor = color.RGBA{R: 0x4c, G: 0x1d, B: 0x95, A: 0xff} // #4c1d95 violet-900
	titleColor    = color.RGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff} // off-white
	footerColor   = color.RGBA{R: 0xa7, G: 0x8b, B: 0xfa, A: 0xff} // violet-300

	titleFace  font.Face // lazy-init на первом render
	footerFace font.Face
)

func init() {
	titleFace = mustFace(gobold.TTF, ogTitleSize)
	footerFace = mustFace(goregular.TTF, ogFooterSize)
}

func mustFace(ttf []byte, size float64) font.Face {
	parsed, err := truetype.Parse(ttf)
	if err != nil {
		panic("seo: cannot parse embedded font: " + err.Error())
	}
	return truetype.NewFace(parsed, &truetype.Options{Size: size})
}

// renderOGImage отдаёт PNG 1200×630 с заголовком промпта на фирменном фоне.
// Без external assets (фон через градиент в коде, шрифт из gofont).
func renderOGImage(title string) ([]byte, error) {
	dc := gg.NewContext(ogWidth, ogHeight)

	// Вертикальный градиент фон.
	grad := gg.NewLinearGradient(0, 0, 0, ogHeight)
	grad.AddColorStop(0, bgTopColor)
	grad.AddColorStop(1, bgBottomColor)
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, ogWidth, ogHeight)
	dc.Fill()

	// Заголовок: word-wrap по ширине, центр по вертикали.
	dc.SetColor(titleColor)
	dc.SetFontFace(titleFace)
	maxWidth := float64(ogWidth - 2*ogPaddingX)
	lines := dc.WordWrap(title, maxWidth)
	if len(lines) > ogMaxLines {
		// Truncate с ellipsis на последней разрешённой строке.
		lines = lines[:ogMaxLines]
		lines[ogMaxLines-1] = trimToWidth(dc, lines[ogMaxLines-1]+"…", maxWidth)
	}
	lineHeight := ogTitleSize * 1.2
	totalHeight := float64(len(lines)) * lineHeight
	startY := (float64(ogHeight)-totalHeight)/2 - lineHeight/2

	for i, line := range lines {
		y := startY + float64(i+1)*lineHeight
		dc.DrawStringAnchored(line, ogPaddingX, y, 0, 0)
	}

	// Footer: «ПромтЛаб» внизу по центру.
	dc.SetColor(footerColor)
	dc.SetFontFace(footerFace)
	dc.DrawStringAnchored("ПромтЛаб · promtlabs.ru", ogWidth/2, ogHeight-ogPaddingY/2, 0.5, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// trimToWidth урезает строку справа пока ширина не поместится. Используется
// только для последней строки multi-line wrap (не критично к производительности).
func trimToWidth(dc *gg.Context, s string, maxWidth float64) string {
	for len(s) > 1 {
		w, _ := dc.MeasureString(s)
		if w <= maxWidth {
			return s
		}
		// убираем последний символ, добавляем ellipsis заново
		runes := []rune(strings.TrimSuffix(s, "…"))
		if len(runes) == 0 {
			return "…"
		}
		s = string(runes[:len(runes)-1]) + "…"
	}
	return s
}

// ogETag — детерминированный ETag по slug+updated_at. Используется для
// 304 Not Modified — клиент кеширует PNG, мы не рендерим повторно.
func ogETag(slug string, updatedAt time.Time) string {
	h := sha256.Sum256([]byte(slug + "|" + updatedAt.UTC().Format(time.RFC3339Nano)))
	return `"` + hex.EncodeToString(h[:8]) + `"`
}
