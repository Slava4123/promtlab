// Re-export шаблонного парсера из @pv/shared. Грамматика идентична backend
// (см. promptvault/backend/internal/template/template.go).

export {
  extractVariables,
  hasVariables,
  renderTemplate,
} from "@pv/shared/template"
