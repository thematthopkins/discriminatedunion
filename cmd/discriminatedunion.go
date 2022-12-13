// Exhaustiveness checks for discriminated unions.
//
// # Usage
//
// The command line usage is:
//
//	discriminatedunion [flags] [packages]
//
// Checks to see that switch statements address all the possible
// types of an interface.
package main

import (
	"github.com/thematthopkins/discriminatedunion"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(discriminatedunion.Analyzer) }
