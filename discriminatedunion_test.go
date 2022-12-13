package discriminatedunion_test

import (
	"github.com/thematthopkins/discriminatedunion"
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestExhaustive(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), discriminatedunion.Analyzer, "./...")
}
