package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract_NoVariables(t *testing.T) {
	assert.Nil(t, Extract(""))
	assert.Nil(t, Extract("plain text with no placeholders"))
	assert.Nil(t, Extract("{{ name }}")) // spaces → literal
	assert.Nil(t, Extract("{{}}"))       // empty → literal
	assert.Nil(t, Extract("{{1name}}"))  // starts with digit → literal
	assert.Nil(t, Extract("{{a-b}}"))    // hyphen not allowed → literal
}

func TestExtract_Single(t *testing.T) {
	assert.Equal(t, []string{"name"}, Extract("Hello {{name}}"))
}

func TestExtract_Unicode(t *testing.T) {
	assert.Equal(t, []string{"имя"}, Extract("Привет {{имя}}"))
	assert.Equal(t, []string{"名前"}, Extract("こんにちは {{名前}}"))
	assert.Equal(t, []string{"ИмяКомпании"}, Extract("{{ИмяКомпании}}"))
	assert.Equal(t, []string{"имя_клиента"}, Extract("{{имя_клиента}}"))
	assert.Equal(t, []string{"Name2"}, Extract("{{Name2}}"))
	assert.Equal(t, []string{"_private"}, Extract("{{_private}}"))
}

func TestExtract_Dedupe(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, Extract("{{a}} and {{b}} and {{a}} again"))
}

func TestExtract_Multiple(t *testing.T) {
	assert.Equal(t,
		[]string{"тип", "язык", "задача"},
		Extract("Напиши {{тип}} на {{язык}} для {{задача}}"),
	)
}

func TestHas(t *testing.T) {
	assert.False(t, Has(""))
	assert.False(t, Has("no placeholders"))
	assert.False(t, Has("{{}}"))
	assert.False(t, Has("{{ spaced }}"))
	assert.True(t, Has("hello {{name}}"))
	assert.True(t, Has("{{имя}}"))
}

func TestRender_FullValues(t *testing.T) {
	out, missing := Render("Hello {{name}}!", map[string]string{"name": "Alice"})
	assert.Equal(t, "Hello Alice!", out)
	assert.Empty(t, missing)
}

func TestRender_PartialMissing(t *testing.T) {
	out, missing := Render(
		"Напиши {{тип}} на {{язык}} для {{задача}}",
		map[string]string{"язык": "Go"},
	)
	assert.Equal(t, "Напиши  на Go для ", out)
	assert.Equal(t, []string{"тип", "задача"}, missing)
}

func TestRender_MissingDedupe(t *testing.T) {
	_, missing := Render("{{a}} {{a}} {{b}}", nil)
	assert.Equal(t, []string{"a", "b"}, missing)
}

func TestRender_NoVariables(t *testing.T) {
	out, missing := Render("plain text", nil)
	assert.Equal(t, "plain text", out)
	assert.Empty(t, missing)
}

func TestRender_SinglePass(t *testing.T) {
	// Значение переменной содержит похожий placeholder — он НЕ должен раскрыться.
	out, missing := Render("Value: {{v}}", map[string]string{"v": "{{b}}"})
	assert.Equal(t, "Value: {{b}}", out)
	assert.Empty(t, missing)
}

func TestRender_EmptyValueIsValidSubstitution(t *testing.T) {
	// Явный пустой ключ — валидная подстановка пустой строкой, не missing.
	out, missing := Render("Hello {{name}}!", map[string]string{"name": ""})
	assert.Equal(t, "Hello !", out)
	assert.Empty(t, missing)
}

func TestRender_AbsentKeyIsMissing(t *testing.T) {
	// Ключ отсутствует в map — это missing.
	out, missing := Render("Hello {{name}}!", map[string]string{})
	assert.Equal(t, "Hello !", out)
	assert.Equal(t, []string{"name"}, missing)
}

func TestRender_InvalidPlaceholdersLeftAsLiterals(t *testing.T) {
	out, missing := Render("{{ name }} and {{1x}}", map[string]string{"name": "Alice"})
	assert.Equal(t, "{{ name }} and {{1x}}", out)
	assert.Empty(t, missing)
}

func TestRender_RegexMetacharactersInValue(t *testing.T) {
	// Значение содержит $1, $&, \ — не должно ломать replace.
	out, _ := Render("Price: {{p}}", map[string]string{"p": `$1 $& \n`})
	assert.Equal(t, `Price: $1 $& \n`, out)
}
