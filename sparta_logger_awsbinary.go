//go:build lambdabinary
// +build lambdabinary

package sparta

import "fmt"

var (
	durationUnitLabel = "ms"
)

func updateLoggerGlobals() {
	// NOP
}

func colorize(s interface{}, c int, disabled bool) string {
	return fmt.Sprintf("%s", s)
}
