package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/nilfunc"

	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Стандартные статические анализаторы из golang.org/x/tools/go/analysis/passes
	// analyzers := []*analysis.Analyzer{
	// 	atomic.Analyzer,
	// 	copylock.Analyzer,
	// 	nilfunc.Analyzer,
	// }

	// // Добавление анализаторов из staticcheck.io
	// for _, a := range staticcheck.Analyzers {
	// 	if a.Analyzer.Name == "SA" {
	// 		analyzers = append(analyzers, a.Analyzer)
	// 	}
	// 	if a.Analyzer.Name == "ST1000" {
	// 		analyzers = append(analyzers, a.Analyzer)
	// 	}
	// 	if a.Analyzer.Name == "S1004" {
	// 		analyzers = append(analyzers, a.Analyzer)
	// 	}
	// 	if a.Analyzer.Name == "QF1004" {
	// 		analyzers = append(analyzers, a.Analyzer)
	// 	}
	// }

	// // Мой кастомный анализатор
	// //analyzers = append(analyzers, OSExitAnalyzer)

	// // Создание нового анализатора, который включает в себя все выбранные анализаторы
	// myAnalyzer := &analysis.Analyzer{
	// 	Name: "myAnalyzer",
	// 	Doc:  "Custom multianalyzer with various analyzers",
	// 	Run: func(pass *analysis.Pass) (interface{}, error) {
	// 		for _, a := range analyzers {
	// 			a.Run(pass)
	// 		}
	// 		return nil, nil
	// 	},
	// }

	// multichecker.Main(myAnalyzer)
	// определяем map подключаемых правил
	mychecks := []*analysis.Analyzer{
		atomic.Analyzer,
		copylock.Analyzer,
		nilfunc.Analyzer,
	}
	checks := map[string]bool{
		"ST1000": true,
		"S1004":  true,
		"QF1004": true,
	}
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") || checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	mychecks = append(mychecks, OSExitAnalyzer)

	multichecker.Main(
		mychecks...,
	)
}
