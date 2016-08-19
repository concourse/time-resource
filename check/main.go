package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/concourse/time-resource/between"
	"github.com/concourse/time-resource/models"
)

var timeFormats []string

func init() {
	timeFormats = append(timeFormats, "3:04 PM -0700")
	timeFormats = append(timeFormats, "3PM -0700")
	timeFormats = append(timeFormats, "3 PM -0700")
	timeFormats = append(timeFormats, "15:04 -0700")
	timeFormats = append(timeFormats, "1504 -0700")
}

func IsInDays(currentTime time.Time, daysList []models.Weekday) bool {
	if len(daysList) == 0 {
		return true
	}

	currentDay := currentTime.Weekday()
	for _, d := range daysList {
		if time.Weekday(d) == currentDay {
			return true
		}
	}

	return false
}

func IntervalHasPassed(interval models.Interval, versionTime time.Time, currentTime time.Time) bool {
	return currentTime.Sub(versionTime) > time.Duration(interval)
}

func main() {
	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	start := request.Source.Start
	stop := request.Source.Stop
	interval := request.Source.Interval
	incrementVersion := false

	lastCheckedAt := request.Version.Time.UTC()

	err = request.Source.Validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		os.Exit(1)
	}

	currentTime := time.Now().UTC()

	if start != nil && stop != nil {
		if between.Between(time.Duration(*start), time.Duration(*stop), currentTime) {
			if lastCheckedAt.IsZero() {
				incrementVersion = true
			} else {
				if interval == nil {
					delta := models.Interval(*stop - *start)
					interval = &delta
				}

				if IntervalHasPassed(*interval, lastCheckedAt, currentTime) {
					incrementVersion = true
				}
			}
		}
	} else if interval != nil {
		if IntervalHasPassed(*interval, lastCheckedAt, currentTime) {
			incrementVersion = true
		}
	}

	versions := []models.Version{}

	if !lastCheckedAt.IsZero() {
		versions = append(versions, models.Version{Time: request.Version.Time})
	}

	if incrementVersion && IsInDays(currentTime, request.Source.Days) {
		versions = append(versions, models.Version{Time: currentTime})
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
