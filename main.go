package main

import (
	"github.com/smasher164/close-check/closecheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(closecheck.Analyzer)
}
