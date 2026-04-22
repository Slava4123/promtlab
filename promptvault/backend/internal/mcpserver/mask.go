package mcpserver

import "strings"

// maskEmail — локальная копия utils.MaskEmail для mcpserver (избегаем
// обратного import'а из delivery-слоя в mcpserver-слой). При изменении
// логики обновлять обе копии или вынести в общий pkg/.
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return ""
	}
	return email[:1] + "***" + email[at:]
}
