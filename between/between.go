package between

import "time"

var formatBase, _ = time.Parse("3:04 PM -0700", "12:00 AM +0000")

const day = 24 * time.Hour

func Between(startBase time.Time, stopBase time.Time, now time.Time) bool {
	if stopBase.Before(startBase) {
		return Between(startBase, stopBase.AddDate(0, 0, 1), now) || Between(startBase.AddDate(0, 0, -1), stopBase, now)
	}

	startDuration := startBase.Sub(formatBase)
	stopDuration := stopBase.Sub(formatBase)

	start := now.Truncate(day).Add(startDuration)
	stop := now.Truncate(day).Add(stopDuration)

	if now.Equal(start) {
		return true
	}

	if now.After(start) && now.Before(stop) {
		return true
	}

	return false
}
