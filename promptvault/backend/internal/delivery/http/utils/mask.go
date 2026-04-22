package utils

import "strings"

// MaskEmail превращает "alice@example.com" в "a***@example.com". Используется
// для GDPR-ограничений: viewer команды видит маску, owner/editor — полный email.
//
// Правила:
//   - пустая строка → пустая строка (NULL-семантика: нет email = не маскируем);
//   - без "@" → пустая строка (нечего маскировать, не раскрываем raw-значение);
//   - local-part из одного символа → этот символ + "***" + "@domain" (все равно
//     скрываем количество символов).
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return ""
	}
	return email[:1] + "***" + email[at:]
}
