package lintrules

import (
	"go/ast"
	"go/token"
	"go/types"
)

type LintError struct {
	Pos     token.Pos
	Message string
}

func (e *LintError) Error() string {
	return e.Message
}

type RuleContext struct {
	FileSet *token.FileSet
	Source  string
	Info    *types.Info
}

type LintRule interface {
	Run(ctx *RuleContext, node ast.Node) error
}
