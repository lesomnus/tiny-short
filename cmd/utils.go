package cmd

import (
	"fmt"
	"time"
)

func DurationString(d time.Duration) string {
	if d < (48 * time.Hour) {
		return d.String()
	} else {
		return fmt.Sprintf("%ddays", int(d/time.Hour/24))
	}
}
