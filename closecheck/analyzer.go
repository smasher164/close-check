package closecheck

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "closecheck",
	Doc:      "closecheck checks that connections are closed properly",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var closerType = types.NewInterface([]*types.Func{
	// method named Close that returns an error
	types.NewFunc(token.NoPos, nil, "Close", types.NewSignature(
		/* receiver type: none */ nil,
		/* parameter types: none */ nil,
		/* return types: error */ types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Universe.Lookup("error").Type())),
		/* variadic? */ false,
	)),
}, nil).Complete()

var errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

func isClose(stmt ast.Stmt, id *ast.Ident) bool {
	eStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := eStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.X.(*ast.Ident).Obj == id.Obj && sel.Sel.Name == "Close"
}

func containsIdent(x ast.Node, id *ast.Ident) (found bool) {
	ast.Inspect(x, func(n ast.Node) bool {
		if got, _ := n.(*ast.Ident); got != nil {
			if got.Obj == id.Obj {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func run(p *analysis.Pass) (interface{}, error) {
	// 1. Check that the LVALUE implements io.Closer.
	//    - get typed ast
	//    - look for assignment statements
	//    - look at left hand side of assignments
	//    - determine if at least one term implements io.Closer
	// 2. Handle the error when obtaining the Closer.
	// 3. Close() the connection afterwards.
	// 4. Further iteration:
	//    - Don't double-close
	//    - Close after use
	inspect := p.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// inspect := p.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
	}
	// astutil
	inspect.WithStack(
		nodeFilter,
		func(n ast.Node, push bool, stack []ast.Node) bool {
			if push {
				return true
			}
			stmt := n.(*ast.AssignStmt)
			if stmt.Tok != token.DEFINE {
				return false
			}
			var closerVar *ast.Ident
			var errVar *ast.Ident
			for _, expr := range stmt.Lhs {
				id, _ := expr.(*ast.Ident)
				if id == nil {
					continue
				}
				typ := p.TypesInfo.Defs[id].Type()
				if types.Implements(typ, closerType) {
					closerVar = id
				}
				if types.Implements(typ, errorType) {
					errVar = id
				}
			}
			scope := stack[len(stack)-2] // [..., scope, assignstmt]
			assignIdx := 0
			list := scope.(*ast.BlockStmt).List
			for i, s := range list {
				if stmt == s {
					assignIdx = i
					break
				}
			}
			if closerVar != nil {
				if errVar != nil {
					// check that error is handled in list[assignIdx+1]
					x, ok := list[assignIdx+1].(*ast.IfStmt)
					if !ok {
						p.Reportf(errVar.NamePos, "error %v not checked immediately", errVar.Name)
						return true
					}
					if !containsIdent(x.Cond, errVar) {
						p.Reportf(errVar.NamePos, "error %v not checked immediately", errVar.Name)
						return true
					}
				}
				numCloses := 0
				for _, s := range list[assignIdx+1:] {
					if isClose(s, closerVar) {
						numCloses++
					}
				}
				if numCloses == 0 {
					p.Reportf(stmt.TokPos, "%v.Close() not called", closerVar.Name)
				} else if numCloses > 1 {
					p.Reportf(stmt.TokPos, "%v.Close() called multiple times", closerVar.Name)
				}
			}
			return true
		})
	return nil, nil
}
