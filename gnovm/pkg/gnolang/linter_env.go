package gnolang

type LintEnv struct {
	Store    Store
	Last     BlockNode
	File     *FileSet
	FileBody string // File source string for no-lint
}

func (e *LintEnv) EvalStaticTypeOf(expr Expr) Type {
	return evalStaticTypeOf(e.Store, e.Last, expr)
}

func (e *LintEnv) UnwrapPointerType(t Type) Type {
	return unwrapPointerType(t)
}

func (e *LintEnv) FileSource() string {
	return e.FileBody
}
