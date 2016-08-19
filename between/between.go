package between

import "time"

const day = 24 * time.Hour

func Between(startBase time.Time, stopBase time.Time, now time.Time) bool {
	startDuration := dayOffset(startBase.UTC())
	stopDuration := dayOffset(stopBase.UTC())

	if stopBase.Before(startBase) {
		if startDuration > (day / 2) {
			startDuration -= day
		} else {
			stopDuration += day
		}
	}

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

func dayOffset(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
}
