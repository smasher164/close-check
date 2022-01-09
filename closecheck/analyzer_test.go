package closecheck_test

import (
	"testing"

	"github.com/smasher164/close-check/closecheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) { analysistest.Run(t, analysistest.TestData(), closecheck.Analyzer, "a") }
