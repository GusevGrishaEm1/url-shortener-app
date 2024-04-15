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
