package mcpserver

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestAllQuotaSensitiveHandlersCallCheckMCPQuota — AST-lint, защищающий от
// забываемости. SDK v1.5 не поддерживает глобальное middleware для tools,
// а декоратор добавляет хрупкую зависимость от типа handler'а. Лучший
// компромисс — статический анализ tools.go: каждый метод toolHandlers
// с именем, соответствующим write/destructive tool, должен содержать
// вызов t.checkMCPQuota.
//
// Если добавлен новый write-tool, но забыт quota check — тест упадёт
// с явным сообщением, названием метода и предложением что добавить.
//
// Исключения (idempotent UX-toggles, не едят квоту по BACKLOG):
//   - promptFavorite, promptPin, promptIncrementUsage
//
// Также пропускаются read-only (они не должны вызывать quota check).
func TestAllQuotaSensitiveHandlersCallCheckMCPQuota(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "tools.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse tools.go: %v", err)
	}

	// Quota-sensitive tool-methods (имя на toolHandlers).
	// Список синхронизирован с apikey/constants.go write/destructive + BACKLOG
	// (исключая idempotent favorite/pin/incrementUsage).
	quotaSensitive := map[string]bool{
		"createPrompt":     true,
		"updatePrompt":     true,
		"deletePrompt":     true,
		"createTag":        true,
		"createCollection": true,
		"deleteCollection": true,
		"tagDelete":        true,
		"collectionUpdate": true,
		"promptRevert":     true,
		"shareCreate":      true,
		"shareDeactivate":  true,
		"restorePrompt":    true,
		"purgePrompt":      true,
	}

	seen := make(map[string]bool)
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			return true
		}
		if !methodReceiverIs(fn, "toolHandlers") {
			return true
		}
		name := fn.Name.Name
		if !quotaSensitive[name] {
			return true
		}
		seen[name] = containsCall(fn, "checkMCPQuota")
		return false
	})

	for name := range quotaSensitive {
		hasCheck, presented := seen[name], seen[name]
		if !presented {
			t.Errorf("lint: ожидался handler (*toolHandlers).%s, но не найден в tools.go", name)
			continue
		}
		if !hasCheck {
			t.Errorf("lint: handler (*toolHandlers).%s должен вызвать t.checkMCPQuota(ctx) — это write/destructive tool, квотируется по BACKLOG", name)
		}
	}
}

func methodReceiverIs(fn *ast.FuncDecl, typeName string) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return false
	}
	switch r := fn.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return r.Name == typeName
	case *ast.StarExpr:
		if id, ok := r.X.(*ast.Ident); ok {
			return id.Name == typeName
		}
	}
	return false
}

func containsCall(fn *ast.FuncDecl, methodName string) bool {
	found := false
	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == methodName {
			found = true
			return false
		}
		return true
	})
	_ = strings.TrimSpace("")
	return found
}
