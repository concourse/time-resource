package lord

import (
	"time"

	"github.com/concourse/time-resource/models"
)

type TimeLord struct {
	PreviousTime time.Time
	Location     *models.Location
	Start        *models.TimeOfDay
	Stop         *models.TimeOfDay
	Interval     *models.Interval
	Days         []models.Weekday
}

func (tl TimeLord) Check(now time.Time) bool {
	if !tl.daysMatch(now) {
		return false
	}

	if !tl.isInRange(now) {
		return false
	}

	if tl.PreviousTime.IsZero() {
		return true
	}

	if tl.Interval != nil {
		if now.Sub(tl.PreviousTime) >= time.Duration(*tl.Interval) {
			return true
		}
	} else {
		if now.UTC().YearDay() > tl.PreviousTime.UTC().YearDay() {
			return true
		}
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

func (tl TimeLord) loc() *time.Location {
	if tl.Location != nil {
		return (*time.Location)(tl.Location)
	}

	return time.UTC
}

const day = 24 * time.Hour

func (tl TimeLord) isInRange(now time.Time) bool {
	if tl.Start == nil || tl.Stop == nil {
		return true
	}

	startOffset := time.Duration(tl.Start.In(tl.Location))
	stopOffset := time.Duration(tl.Stop.In(tl.Location))

	start := now.Truncate(day).Add(startOffset)
	stop := now.Truncate(day).Add(stopOffset)

	return tl.isBetween(now, start, stop)
}

func (tl TimeLord) isBetween(now time.Time, start time.Time, stop time.Time) bool {
	if stop.Before(start) {
		return tl.isBetween(now, start.AddDate(0, 0, -1), stop) || tl.isBetween(now, start, stop.AddDate(0, 0, 1))
	}

	if now.Equal(start) {
		return true
	}

	if now.After(start) && now.Before(stop) {
		return true
	}

	return false
}

func (tl TimeLord) yesterday() TimeLord {
	yesterday := tl
	startOffset := time.Duration(*tl.Start) - day
	yesterday.Start = (*models.TimeOfDay)(&startOffset)
	return yesterday
}

func (tl TimeLord) tomorrow() TimeLord {
	tomorrow := tl
	stopOffset := time.Duration(*tl.Stop) + day
	tomorrow.Stop = (*models.TimeOfDay)(&stopOffset)
	return tomorrow
}
