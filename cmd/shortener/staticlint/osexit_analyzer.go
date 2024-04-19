// Пакет main предоставляет инструмент для проверки прямых вызовов os.Exit в функции main.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Анализатор OSExitAnalyzer для проверки прямых вызовов os.Exit в функции main.
var OSExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "check for direct os.Exit in main function",
	Run:      run,
	Requires: []*analysis.Analyzer{},
}

// run проверяет наличие прямого вызова os.Exit в функции main.
//
// Функция принимает анализ Pass в качестве параметра и возвращает interface{} и ошибку.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if fn.Name.Name != "main" {
				return true
			}
			ast.Inspect(fn, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if isOSExitCall(callExpr) {
					pass.Report(analysis.Diagnostic{
						Pos:     callExpr.Pos(),
						Message: "direct os.Exit in main function",
					})
					return false
				}
				return true
			})
			return true
		})
	}
	return nil, nil
}

func isOSExitCall(expr *ast.CallExpr) bool {
	if ident, ok := expr.Fun.(*ast.SelectorExpr); ok {
		if ident.Sel.Name != "Exit" {
			return false
		}
		if ident.X.(*ast.Ident).Name == "os" {
			return true
		}
		return false
	}
	return false
}
