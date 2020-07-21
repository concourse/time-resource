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
}

func (tl TimeLord) Check(now time.Time) bool {

	start, stop := tl.latestRangeBefore(now)

	if !tl.daysMatch(now) {
		return false
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

func (tl TimeLord) latestRangeBefore(reference time.Time) (time.Time, time.Time) {

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
