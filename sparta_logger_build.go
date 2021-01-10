// +build !lambdabinary

package sparta

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

var (
	durationUnitLabel = "ms"
)

func updateLoggerGlobals() {
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
}

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
