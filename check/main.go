package main

import (
	"encoding/json"
	"errors"
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
	timeFormats = append(timeFormats, "3:04 PM")
	timeFormats = append(timeFormats, "3PM")
	timeFormats = append(timeFormats, "3 PM")
	timeFormats = append(timeFormats, "15:04")
	timeFormats = append(timeFormats, "1504")
}

func validateConfig(start string, stop string, interval string) {
	if start == "" && stop == "" && interval == "" {
		fmt.Fprintln(os.Stderr, "one of 'interval' or 'between' must be specified")
		os.Exit(1)
	}

	if start == "" && stop != "" {
		fmt.Fprintln(os.Stderr, "empty start time!")
		os.Exit(1)
	}

	if stop == "" && start != "" {
		fmt.Fprintln(os.Stderr, "empty stop time!")
		os.Exit(1)
	}
}

func ParseTime(timeString string, location *time.Location) (time.Time, error) {
	for _, format := range timeFormats {
		parsedTime, err := time.Parse(format, timeString)
		if err != nil {
			continue
		}

		return parsedTime.In(location), nil
	}

	return time.Time{}, errors.New("could not parse time")
}

func ParseWeekdays(daysList []string) ([]time.Weekday, error) {
	days := []time.Weekday{}

	for _, d := range daysList {
		switch d {
		case "Sunday":
			days = append(days, time.Sunday)
		case "Monday":
			days = append(days, time.Monday)
		case "Tuesday":
			days = append(days, time.Tuesday)
		case "Wednesday":
			days = append(days, time.Wednesday)
		case "Thursday":
			days = append(days, time.Thursday)
		case "Friday":
			days = append(days, time.Friday)
		case "Saturday":
			days = append(days, time.Saturday)
		default:
			return []time.Weekday{}, errors.New(fmt.Sprintf("invalid day '%s'", d))
		}
	}
	return days, nil
}

func IsInDays(currentTime time.Time, daysList []time.Weekday) bool {
	if len(daysList) == 0 {
		return true
	}

	currentDay := currentTime.Weekday()
	for _, d := range daysList {
		if d == currentDay {
			return true
		}
	}

	return false
}

func IntervalHasPassed(interval string, versionTime time.Time, currentTime time.Time) (bool, error) {
	parsedInterval, err := time.ParseDuration(interval)

	if err != nil {
		return false, err
	}

	if currentTime.Sub(versionTime) > parsedInterval {
		return true, nil
	}

	return false, nil
}

func main() {
	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error decoding payload: "+err.Error())
		os.Exit(1)
	}
	if request.Source.Location == "" {
		request.Source.Location = "UTC"
	}
	loc, err := time.LoadLocation(request.Source.Location)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	currentTime := time.Now().In(loc)

	start := request.Source.Start
	stop := request.Source.Stop
	interval := request.Source.Interval
	incrementVersion := false

	lastCheckedAt := request.Version.Time.In(loc)

	validateConfig(start, stop, interval)

	days, err := ParseWeekdays(request.Source.Days)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if start != "" && stop != "" {
		startTime, err := ParseTime(start, loc)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid start time: "+start+"; "+err.Error())
			os.Exit(1)
		}

		stopTime, err := ParseTime(stop, loc)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid stop time: "+stop+"; "+err.Error())
			os.Exit(1)
		}

		if between.Between(startTime, stopTime, currentTime) {
			if lastCheckedAt.IsZero() {
				incrementVersion = true
			} else {
				// This means we have a config that runs once within a given time range.
				// In that case, we set our interval to be the max time from that range
				// so it only runs once.
				if interval == "" {
					if startTime.After(stopTime) {
						stopTime = stopTime.Add(24 * time.Hour)
					}

					interval = stopTime.Sub(startTime).String()
				}

				intervalHasPassed, err := IntervalHasPassed(interval, request.Version.Time, currentTime)

				if err != nil {
					fmt.Fprintln(os.Stderr, "invalid interval: "+interval+"; "+err.Error())
					os.Exit(1)
				}

				if intervalHasPassed {
					incrementVersion = true
				}
			}
		}
	} else if interval != "" {
		intervalHasPassed, err := IntervalHasPassed(interval, request.Version.Time, currentTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid interval: "+interval+"; "+err.Error())
			os.Exit(1)
		}

		if intervalHasPassed {
			incrementVersion = true
		}
	}

	versions := []models.Version{}

	if !lastCheckedAt.IsZero() {
		versions = append(versions, models.Version{Time: request.Version.Time})
	}

	if incrementVersion && IsInDays(currentTime, days) {
		versions = append(versions, models.Version{Time: currentTime})
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
