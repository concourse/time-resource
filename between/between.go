package between

import "time"

const day = 24 * time.Hour

func Between(startOffset time.Duration, stopOffset time.Duration, now time.Time) bool {
	if stopOffset < startOffset {
		if startOffset > (day / 2) {
			startOffset -= day
		} else {
			stopOffset += day
		}
	}

	start := now.Truncate(day).Add(startOffset)
	stop := now.Truncate(day).Add(stopOffset)

	if now.Equal(start) {
		return true
	}

	if now.After(start) && now.Before(stop) {
		return true
	}

	return false
}
