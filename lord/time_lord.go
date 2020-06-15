package lord

import (
	"errors"
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

var NotYetError = errors.New("valid time has not occurred yet")

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

func (tl TimeLord) Latest(reference time.Time) (time.Time, error) {
	if tl.Start != nil && tl.Stop != nil {
		if tl.Interval != nil {
			return tl.latestIntervalInRange(reference)
		}

		return tl.latestInRange(reference)
	}

	return tl.latestInterval(reference), nil
}

func (tl TimeLord) List(reference time.Time) []time.Time {
	if tl.Interval != nil && tl.Start != nil && tl.Stop != nil {
		return tl.listIntervalInRange(reference)
	}

	if tl.Interval != nil {
		return tl.listInterval(reference)
	}

	return tl.listInRange(reference)
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

func (tl TimeLord) latestInterval(reference time.Time) time.Time {
	loc := tl.loc()
	refInLoc := reference.In(loc)

	if tl.PreviousTime.IsZero() {
		return refInLoc
	}

	start := tl.PreviousTime.In(loc)
	tlDuration := time.Duration(*tl.Interval)

	var latestValidTime time.Time
	for intervalTime := start; !intervalTime.After(refInLoc); intervalTime = intervalTime.Add(tlDuration) {
		if tl.daysMatch(intervalTime) {
			latestValidTime = intervalTime
		}
	}
	return latestValidTime
}

func (tl TimeLord) latestIntervalInRange(reference time.Time) (time.Time, error) {
	loc := tl.loc()
	refInLoc := reference.In(loc)

	start, end := tl.latestRangeBefore(refInLoc)
	var err error

	if tl.PreviousTime.IsZero() && !refInLoc.Before(end) {
		err = NotYetError
	}

	tlDuration := time.Duration(*tl.Interval)

	var latestValidTime time.Time
	for intervalTime := start; !intervalTime.After(refInLoc) && intervalTime.Before(end); intervalTime = intervalTime.Add(tlDuration) {
		if tl.daysMatch(intervalTime) {
			latestValidTime = intervalTime
		}
	}
	return latestValidTime, err
}

func (tl TimeLord) latestInRange(reference time.Time) (time.Time, error) {
	loc := tl.loc()
	refInLoc := reference.In(loc)

	start, end := tl.latestRangeBefore(refInLoc)
	var err error

	if tl.PreviousTime.IsZero() && (!refInLoc.Before(end) || !tl.daysMatch(refInLoc)) {
		err = NotYetError
	}

	if !tl.PreviousTime.Before(start) && tl.PreviousTime.Before(refInLoc) {
		return tl.PreviousTime.In(loc), err
	}

	if refInLoc.Before(end) && tl.daysMatch(refInLoc) {
		return refInLoc, err
	}

	end = end.Add(-time.Minute)
	for !tl.daysMatch(end) {
		end = end.AddDate(0, 0, -1)
	}
	return end, err
}

func (tl TimeLord) latestRangeBefore(reference time.Time) (time.Time, time.Time) {

	if tl.Start == nil || tl.Stop == nil {
		return time.Time{}, time.Time{}
	}

	refInLoc := reference.In(tl.loc())

	start := time.Date(refInLoc.Year(), refInLoc.Month(), refInLoc.Day(),
		tl.Start.Hour(), tl.Start.Minute(), 0, 0, tl.loc())

	if start.After(refInLoc) {
		start = start.AddDate(0, 0, -1)
	}

	stop := time.Date(start.Year(), start.Month(), start.Day(),
		tl.Stop.Hour(), tl.Stop.Minute(), 0, 0, tl.loc())

	if stop.Before(start) {
		stop = stop.AddDate(0, 0, 1)
	}

	return start, stop
}

func (tl TimeLord) listInterval(reference time.Time) []time.Time {
	loc := tl.loc()
	refInLoc := reference.In(loc)

	if tl.PreviousTime.IsZero() {
		return []time.Time{reference}
	}

	start := tl.PreviousTime.In(loc)
	tlDuration := time.Duration(*tl.Interval)

	versions := []time.Time{}
	for intervalTime := start; !intervalTime.After(refInLoc); intervalTime = intervalTime.Add(tlDuration) {
		if tl.daysMatch(intervalTime) {
			versions = append(versions, intervalTime)
		}
	}
	return versions
}

func (tl TimeLord) listIntervalInRange(reference time.Time) []time.Time {
	loc := tl.loc()
	refInLoc := reference.In(loc)

	start, _ := tl.latestRangeBefore(refInLoc)
	tlDuration := time.Duration(*tl.Interval)

	versions := []time.Time{}

	if !tl.PreviousTime.IsZero() {
		start, _ = tl.latestRangeBefore(tl.PreviousTime)
	}

	for dailyInterval := start; !dailyInterval.After(refInLoc); dailyInterval = dailyInterval.AddDate(0, 0, 1) {
		if tl.daysMatch(dailyInterval) {
			dailyStart, dailyEnd := tl.latestRangeBefore(dailyInterval)
			for intervalTime := dailyStart; !intervalTime.After(refInLoc) && intervalTime.Before(dailyEnd); intervalTime = intervalTime.Add(tlDuration) {
				if !tl.PreviousTime.After(intervalTime) {
					versions = append(versions, intervalTime)
				}
			}
		}
	}

	return versions
}

func (tl TimeLord) listInRange(reference time.Time) []time.Time {
	loc := tl.loc()
	refInLoc := reference.In(loc)

	start, end := tl.latestRangeBefore(refInLoc)

	versions := []time.Time{}

	if !tl.PreviousTime.IsZero() {
		previousInLoc := tl.PreviousTime.In(loc)
		previousStart, previousEnd := tl.latestRangeBefore(previousInLoc)

		if previousInLoc.Before(previousEnd) && previousInLoc.Before(refInLoc) && tl.daysMatch(previousInLoc) {
			versions = append(versions, previousInLoc)
		}

		for intervalTime := previousStart.AddDate(0, 0, 1); intervalTime.Before(start); intervalTime = intervalTime.AddDate(0, 0, 1) {
			if tl.daysMatch(intervalTime) {
				versions = append(versions, intervalTime)
			}
		}
	}

	if tl.PreviousTime.Before(start) && refInLoc.Before(end) && tl.daysMatch(refInLoc) {
		versions = append(versions, refInLoc)
	}

	return versions
}

func (tl TimeLord) loc() *time.Location {
	if tl.Location != nil {
		return (*time.Location)(tl.Location)
	}

	return time.UTC
}
