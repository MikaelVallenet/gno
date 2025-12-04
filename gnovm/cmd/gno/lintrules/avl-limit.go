package lintrules

import (
	"go/ast"
	"go/token"
	"strings"
)

type AvlLimitRule struct{}

func (r AvlLimitRule) Run(ctx *RuleContext, node ast.Node) error {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return nil
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	method := sel.Sel.Name
	if method != "Iterate" && method != "ReverseIterate" {
		return nil
	}

	if len(call.Args) < 2 || !isEmptyString(call.Args[0]) || !isEmptyString(call.Args[1]) {
		return nil
	}

	if hasNoLintAbove(ctx.FileSet, ctx.Source, node) {
		return nil
	}

	return &LintError{
		Pos:     call.Pos(),
		Message: "calling Iterate/ReverseIterate without start/end bounds",
	}
}

func isEmptyString(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	if !ok {
		return false
	}
	return lit.Kind == token.STRING && lit.Value == `""`
}

func hasNoLintAbove(fset *token.FileSet, src string, n ast.Node) bool {
	pos := fset.Position(n.Pos())
	if pos.Line <= 1 {
		return false
	}

	lines := strings.Split(src, "\n")
	prev := pos.Line - 2
	if prev < 0 || prev >= len(lines) {
		return false
	}

	return strings.HasPrefix(strings.TrimSpace(lines[prev]), "//nolint")
}
