package lintrules

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
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

	fmt.Printf("DEBUG: checking call to %s at %v\n", method, call.Pos())

	tv, ok := ctx.Info.Types[sel.X]
	if !ok || tv.Type == nil {
		fmt.Printf("DEBUG: no type info for receiver at %v\n", sel.X.Pos())
		return nil
	}
	recvType := unwrapPtr(tv.Type)

	fmt.Printf("DEBUG: receiver type: %T %v\n", recvType, recvType)

	named, ok := recvType.(*types.Named)
	if !ok {
		fmt.Printf("DEBUG: receiver type is not named at %v\n", sel.X.Pos())
		return nil
	}
	obj := named.Obj()
	if obj == nil {
		fmt.Printf("DEBUG: named type has no object at %v\n", sel.X.Pos())
		return nil
	}

	fmt.Printf("DEBUG: receiver named type: %s.%s\n", obj.Pkg().Path(), obj.Name())
	if obj.Name() != "Tree" {
		fmt.Printf("DEBUG: receiver named type is not Tree at %v\n", sel.X.Pos())
		return nil
	}
	fmt.Printf("DEBUG: receiver named type is Tree at %v\n", sel.X.Pos())
	if obj.Pkg() == nil || obj.Pkg().Path() != "gno.land/p/nt/avl" {
		fmt.Printf("DEBUG: receiver named type is not in gno.land/p/nt/avl at %v\n", sel.X.Pos())
		return nil
	}

	fmt.Printf("DEBUG: receiver named type is in gno.land/p/nt/avl at %v\n", sel.X.Pos())
	if len(call.Args) < 2 || !isEmptyString(call.Args[0]) || !isEmptyString(call.Args[1]) {
		fmt.Printf("DEBUG: call has start/end bounds at %v\n", call.Pos())
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

func unwrapPtr(t types.Type) types.Type {
	if pt, ok := t.(*types.Pointer); ok {
		return pt.Elem()
	}
	return t
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
