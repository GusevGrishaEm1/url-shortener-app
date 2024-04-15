package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OSExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "check for direct os.Exit in main function",
	Run:      run,
	Requires: []*analysis.Analyzer{},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if fn.Name.Name == "main" {
				ast.Inspect(fn, func(n ast.Node) bool {
					callExpr, ok := n.(*ast.CallExpr)
					if ok {
						if isOSExitCall(callExpr) {
							pass.Report(analysis.Diagnostic{
								Pos:     callExpr.Pos(),
								Message: fmt.Sprintf("direct os.Exit in main function"),
							})
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}

func isOSExitCall(expr *ast.CallExpr) bool {
	if ident, ok := expr.Fun.(*ast.SelectorExpr); ok {
		if ident.Sel.Name == "Exit" {
			if ident.X.(*ast.Ident).Name == "os" {
				return true
			}
		}
	}
	return false
}
