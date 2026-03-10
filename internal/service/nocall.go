package service

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const Doc = "nocall find callings specified by nocall.functions flag"

var flagFunctions string

var Analyzer = &analysis.Analyzer{
	Name: "nocall",
	Doc:  Doc,
	Run:  run,
}

func SetupAnalyzer() {
	Analyzer.Flags.StringVar(&flagFunctions, "functions", "", "comma separated functions which are restricted")
}

func run(pass *analysis.Pass) (any, error) {
	if flagFunctions == "" {
		return nil, nil
	}
	functionNames := strings.Split(flagFunctions, ",")

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				var name string

				switch fun := call.Fun.(type) {
				case *ast.Ident:
					name = fun.Name
				case *ast.SelectorExpr:
					pkg, ok := fun.X.(*ast.Ident)
					if !ok {
						return true
					}

					name = fmt.Sprintf("%s.%s", pkg.Name, fun.Sel.Name)
				}

				if slices.Contains(functionNames, name) {
					pass.Reportf(call.Pos(), "%s function is restricted", name)
				}
			}

			return true
		})
	}

	return nil, nil
}
