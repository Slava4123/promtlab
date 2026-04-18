// Package template парсит и рендерит шаблоны промптов вида `{{name}}`.
//
// Grammar (BNF, Unicode-aware), идентично frontend/src/lib/template/parse.ts:
//
//	template   ::= (literal | variable)*
//	variable   ::= "{{" identifier "}}"
//	identifier ::= (letter | "_") (letter | digit | "_")*
//	letter     ::= any Unicode letter (\p{L}) — latin, cyrillic, CJK, etc.
//	digit      ::= any Unicode decimal digit (\p{N})
//
// Всё, что не попадает под variable, считается literal — включая формы
// `{{ name }}` (с пробелами), `{{1name}}` (первый символ — цифра), `{{}}`
// (пусто). Это даёт неявный escape: оберни в пробел, чтобы отобразить буквально.
//
// Инварианты:
//   - Extract де-дуплицирует имена, сохраняя порядок первого вхождения.
//   - Render делает single-pass: подстановленные значения НЕ ре-сканируются.
//   - Missing keys возвращаются отдельным срезом `missing`; caller решает,
//     считать ли это ошибкой (mcpserver — считает).
package template

import "regexp"

// variableRegex — тот же синтаксис, что и фронтовый VARIABLE_REGEX.
// Go regexp по умолчанию поддерживает \p{L} и \p{N}.
var variableRegex = regexp.MustCompile(`\{\{([\p{L}_][\p{L}\p{N}_]*)\}\}`)

// Extract возвращает уникальные имена переменных в порядке первого вхождения.
// Возвращает пустой срез, если переменных нет.
func Extract(content string) []string {
	matches := variableRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}

// Has возвращает true, если content содержит хотя бы одну валидную переменную.
// Быстрее Extract, когда важно только наличие.
func Has(content string) bool {
	return variableRegex.MatchString(content)
}

// Render подставляет values[name] вместо каждого `{{name}}`.
// Возвращает rendered и список имён, для которых ключа нет в values.
// Явный пустой ключ (values["name"] = "") считается валидной подстановкой
// пустой строкой и НЕ попадает в missing.
// Повторы в missing исключены (де-дуп в порядке первого вхождения).
func Render(content string, values map[string]string) (rendered string, missing []string) {
	if !Has(content) {
		return content, nil
	}
	seenMissing := make(map[string]struct{})
	rendered = variableRegex.ReplaceAllStringFunc(content, func(match string) string {
		sub := variableRegex.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		name := sub[1]
		v, ok := values[name]
		if !ok {
			if _, dup := seenMissing[name]; !dup {
				seenMissing[name] = struct{}{}
				missing = append(missing, name)
			}
			return ""
		}
		return v
	})
	return rendered, missing
}
