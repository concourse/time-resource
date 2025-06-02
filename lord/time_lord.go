package lord

import (
	"time"

	"github.com/concourse/time-resource/models"
)

var DEFAULT_TIME_OF_DAY = models.TimeOfDay(time.Duration(0))

type TimeLord struct {
	PreviousTime time.Time
	Location     *models.Location
	Start        *models.TimeOfDay
	Stop         *models.TimeOfDay
	Interval     *models.Interval
	Days         []models.Weekday
	StartAfter   *models.StartAfter
}

func (tl TimeLord) Check(now time.Time) bool {

	start, stop := tl.LatestRangeBefore(now)

	if !tl.daysMatch(now) {
		return false
	}

	if tl.StartAfter != nil {
		startAfter := time.Time(*tl.StartAfter)
		startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
			startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())

		if !startInLoc.Before(now) {
			return false
		}
	}

	if !start.IsZero() && (now.Before(start) || !now.Before(stop)) {
		return false
	}

	if tl.PreviousTime.IsZero() {
		return true
	}

	if tl.Interval != nil {
		if now.Sub(tl.PreviousTime) >= time.Duration(*tl.Interval) {
			return true
		}
	} else if !start.IsZero() {
		return tl.PreviousTime.Before(start)
	}

	return false
}

func (tl TimeLord) Latest(reference time.Time) time.Time {
	if tl.PreviousTime.After(reference) {
		return time.Time{}
	}

	refInLoc := reference.In(tl.loc())
	for !tl.daysMatch(refInLoc) {
		refInLoc = refInLoc.AddDate(0, 0, -1)
	}

	start, stop := tl.LatestRangeBefore(refInLoc)

	if tl.PreviousTime.IsZero() && !reference.Before(stop) {
		return time.Time{}
	}

	if tl.Interval == nil {
		if tl.PreviousTime.After(start) {
			return time.Time{}
		}
		return start
	}

	tlDuration := time.Duration(*tl.Interval)

	var latestValidTime time.Time
	for intervalTime := start; !intervalTime.After(reference) && intervalTime.Before(stop); intervalTime = intervalTime.Add(tlDuration) {
		latestValidTime = intervalTime
	}
	return latestValidTime
}

func (tl TimeLord) List(reference time.Time) []time.Time {
	start := tl.PreviousTime

	var addForRange func(time.Time, time.Time)
	versions := []time.Time{}

	if tl.Interval == nil {

		if start.IsZero() {
			refRangeStart, refRangeEnd := tl.LatestRangeBefore(reference)
			if !reference.Before(refRangeEnd) {
				return versions
			}
			start = refRangeStart
		}

		addForRange = func(dailyStart, _ time.Time) {
			if !dailyStart.Before(start) && !dailyStart.After(reference) {
				versions = append(versions, dailyStart)
			}
		}

	} else {
		tlDuration := time.Duration(*tl.Interval)

		if start.IsZero() {
			start = reference
		}

		addForRange = func(dailyStart, dailyEnd time.Time) {
			intervalTime := dailyStart.Truncate(tlDuration)

			for intervalTime.Before(dailyStart) || intervalTime.Before(start) {
				intervalTime = intervalTime.Add(tlDuration)
			}

			for !intervalTime.After(reference) && intervalTime.Before(dailyEnd) {
				versions = append(versions, intervalTime)
				intervalTime = intervalTime.Add(tlDuration)
			}
		}
	}

	var dailyStart, dailyEnd time.Time
	for dailyInterval := start; !dailyStart.After(reference); dailyInterval = dailyInterval.AddDate(0, 0, 1) {
		if tl.daysMatch(dailyInterval) {
			dailyStart, dailyEnd = tl.LatestRangeBefore(dailyInterval)
			if dailyStart.After(reference) {
				break
			}
			addForRange(dailyStart, dailyEnd)
		}
	}
	return versions
}

func (tl TimeLord) daysMatch(now time.Time) bool {
	if len(tl.Days) == 0 {
		return true
	}

	todayInLoc := models.Weekday(now.In(tl.loc()).Weekday())

	for _, day := range tl.Days {
		if day == todayInLoc {
			return true
		}
	}

	return false
}

func (tl TimeLord) LatestRangeBefore(reference time.Time) (time.Time, time.Time) {

	tlStart := DEFAULT_TIME_OF_DAY
	if tl.Start != nil {
		tlStart = *tl.Start
	}
	tlStop := DEFAULT_TIME_OF_DAY
	if tl.Stop != nil {
		tlStop = *tl.Stop
	}

	refInLoc := reference.In(tl.loc())

	start := time.Date(refInLoc.Year(), refInLoc.Month(), refInLoc.Day(),
		tlStart.Hour(), tlStart.Minute(), 0, 0, tl.loc())

	if start.After(refInLoc) {
		start = start.AddDate(0, 0, -1)
	}

	stop := time.Date(start.Year(), start.Month(), start.Day(),
		tlStop.Hour(), tlStop.Minute(), 0, 0, tl.loc())

	if !stop.After(start) {
		stop = stop.AddDate(0, 0, 1)
	}

	return start, stop
}

func (tl TimeLord) loc() *time.Location {
	if tl.Location != nil {
		return (*time.Location)(tl.Location)
	}

	return time.UTC
}
