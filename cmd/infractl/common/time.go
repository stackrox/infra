package common

import (
	"fmt"
	"time"
)

// FormatExpiration formats the given duration in a human-friendly way,
// indicating how long in the future the event will happen.
func FormatExpiration(delta time.Duration) string {
	var (
		days    = int(delta.Hours() / 24)
		hours   = int(delta.Hours()) % 24
		minutes = int(delta.Minutes()) % 60
		seconds = int(delta.Seconds()) % 60
	)

	switch {
	case delta >= 24*time.Hour:
		return fmt.Sprintf("%dd remaining", days)

	case delta > 8*time.Hour:
		return fmt.Sprintf("%dh remaining", hours)

	case delta > time.Hour:
		return fmt.Sprintf("%dh%dm remaining", hours, minutes)

	case delta > 5*time.Minute:
		return fmt.Sprintf("%dm remaining", minutes)

	case delta > time.Minute:
		return fmt.Sprintf("%dm%ds remaining", minutes, seconds)

	case delta > 15*time.Second:
		return fmt.Sprintf("%ds remaining", seconds)

	case delta > 0:
		return "expiring now"

	default:
		return "expired"
	}
}

// FormatTime formats the given time in a human-friendly way, indicating how
// long in the past the event happened.
func FormatTime(moment time.Time) string {
	var (
		delta   = time.Now().Local().Sub(moment.Local())
		hours   = int(delta.Hours())
		minutes = int(delta.Minutes()) % 60
		seconds = int(delta.Seconds()) % 60
	)

	switch {
	case delta >= 24*time.Hour:
		return moment.Local().Format("Mon Jan 2 15:04:05 -0700 MST 2006")

	case delta > 8*time.Hour:
		return fmt.Sprintf("%dh ago", hours)

	case delta > time.Hour:
		return fmt.Sprintf("%dh%dm ago", hours, minutes)

	case delta > 5*time.Minute:
		return fmt.Sprintf("%dm ago", minutes)

	case delta > time.Minute:
		return fmt.Sprintf("%dm%ds ago", minutes, seconds)

	case delta > 15*time.Second:
		return fmt.Sprintf("%ds ago", seconds)

	default:
		return "just now"
	}
}
