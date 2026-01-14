package lord

import (
	"time"

	"github.com/adhocore/gronx"
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
	Cron         *models.Cron
}

func (tl TimeLord) Check(now time.Time) bool {
	if tl.Cron != nil {
		nowInLoc := now.In(tl.loc())

		// Check start_after constraint first
		if tl.StartAfter != nil {
			startAfter := time.Time(*tl.StartAfter)
			startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
				startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())

			if !startInLoc.Before(nowInLoc) {
				return false
			}
		}

		if tl.PreviousTime.IsZero() {
			// No previous version exists.
			// Find the most recent scheduled cron time <= now.
			// This ensures the resource emits an initial version even if the check
			// doesn't run at the exact cron minute.
			//
			// Example: @daily, check at 3pm → prevTick = today 00:00 → trigger
			// Example: @5minutes, check at 3:07pm → prevTick = 3:05pm → trigger
			prevTick, err := gronx.PrevTickBefore(tl.Cron.Expression, nowInLoc, true)
			if err != nil {
				return false
			}
			return !prevTick.IsZero()
		}

		// Previous version exists.
		// Check if the next scheduled cron time after prevTime has passed.
		// This handles late checks - if cron is :30 and check runs at :31,
		// we still trigger because :30 has passed.
		prevInLoc := tl.PreviousTime.In(tl.loc())
		nextTime, err := tl.Cron.Next(prevInLoc)
		if err != nil {
			return false
		}

		return !nextTime.IsZero() && !nextTime.After(nowInLoc)
	}

	start, stop := tl.LatestRangeBefore(now)

	if !tl.daysMatch(now) {
		return false
	}

	if tl.StartAfter != nil {
		startAfter := time.Time(*tl.StartAfter)
		startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
			startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())

		if !startInLoc.Before(now.In(tl.loc())) {
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
		return now.Sub(tl.PreviousTime) >= time.Duration(*tl.Interval)
	}

	return !start.IsZero() && tl.PreviousTime.Before(start)
}

func (tl TimeLord) Latest(reference time.Time) time.Time {
	if tl.Cron != nil {
		refInLoc := reference.In(tl.loc())

		if tl.StartAfter != nil {
			startAfter := time.Time(*tl.StartAfter)
			startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
				startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())

			if !startInLoc.Before(refInLoc) {
				return time.Time{}
			}
		}

		if tl.PreviousTime.IsZero() {
			// Return the most recent scheduled cron time <= reference
			prevTick, err := gronx.PrevTickBefore(tl.Cron.Expression, refInLoc, true)
			if err != nil || prevTick.IsZero() {
				return time.Time{}
			}
			return prevTick.UTC()
		}

		// Find the most recent cron time <= reference
		// This handles cases where multiple cron times have passed since prev
		latest, err := gronx.PrevTickBefore(tl.Cron.Expression, refInLoc, true)
		if err != nil || latest.IsZero() {
			return time.Time{}
		}

		// Only return if this is after the previous time (a new trigger)
		prevInLoc := tl.PreviousTime.In(tl.loc())
		if latest.After(prevInLoc) {
			return latest.UTC()
		}

		return time.Time{}
	}

	if tl.PreviousTime.After(reference) {
		return time.Time{}
	}

	if tl.StartAfter != nil {
		startAfter := time.Time(*tl.StartAfter)
		startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
			startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())
		if !startInLoc.Before(reference.In(tl.loc())) {
			return time.Time{}
		}
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
		// If we're past the stop time, return zero (no new version to emit)
		if !reference.Before(stop) {
			return time.Time{}
		}
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
	if tl.Cron != nil {
		refInLoc := reference.In(tl.loc())

		if tl.StartAfter != nil {
			startAfter := time.Time(*tl.StartAfter)
			startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
				startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())

			if !startInLoc.Before(reference.In(tl.loc())) {
				return []time.Time{}
			}
		}

		if tl.PreviousTime.IsZero() {
			// Return the most recent scheduled cron time <= reference
			prevTick, err := gronx.PrevTickBefore(tl.Cron.Expression, refInLoc, true)
			if err != nil || prevTick.IsZero() {
				return []time.Time{}
			}
			return []time.Time{prevTick.UTC()}
		}

		// Get previous time in local timezone
		start := tl.PreviousTime.In(tl.loc())

		// Get all occurrences between previous and reference
		maxOccurrences := 5
		times := tl.Cron.NextN(start, refInLoc, maxOccurrences)

		// Convert all times back to UTC
		result := make([]time.Time, len(times))
		for i, t := range times {
			result[i] = t.UTC()
		}

		return result
	}

	start := tl.PreviousTime
	versions := []time.Time{}

	if tl.StartAfter != nil {
		startAfter := time.Time(*tl.StartAfter)
		startInLoc := time.Date(startAfter.Year(), startAfter.Month(), startAfter.Day(),
			startAfter.Hour(), startAfter.Minute(), startAfter.Second(), 0, tl.loc())
		if !startInLoc.Before(reference.In(tl.loc())) {
			return versions
		}
	}

	var addForRange func(time.Time, time.Time)

	if tl.Interval == nil {

		if start.IsZero() {
			refRangeStart, refRangeEnd := tl.LatestRangeBefore(reference)
			if !reference.Before(refRangeEnd) {
				return versions
			}
			start = refRangeStart
		}

		addForRange = func(dailyStart, dailyEnd time.Time) {
			// Don't add if reference is past the stop time
			if !reference.Before(dailyEnd) {
				return
			}
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
		tlStart.Hour(), tlStart.Minute(), tlStart.Second(), 0, tl.loc())

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
