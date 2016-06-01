package between

import "time"

func Between(start time.Time, stop time.Time, timeToCompare time.Time) bool {
	utcStart := start.UTC()
	utcStop := stop.UTC()
	utcCompare := timeToCompare.UTC()

	var stopHour int
	var timeToCompareHour int
	startHour := 0

	if utcStart.Hour() > utcStop.Hour() {
		stopHour = utcStop.Hour() + 24 - utcStart.Hour()
	} else {
		stopHour = utcStop.Hour() - utcStart.Hour()
	}

	if utcStart.Hour() > utcCompare.Hour() {
		timeToCompareHour = utcCompare.Hour() + 24 - utcStart.Hour()
	} else {
		timeToCompareHour = utcCompare.Hour() - utcStart.Hour()
	}

	hoursInRange := (timeToCompareHour >= startHour) &&
		(timeToCompareHour <= stopHour)

	if !hoursInRange {
		return false
	}

	if timeToCompareHour == stopHour && utcCompare.Minute() > utcStop.Minute() {
		return false
	}
	if timeToCompareHour == startHour && utcCompare.Minute() < utcStart.Minute() {
		return false
	}
	return true
}
