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
	timeFormats = append(timeFormats, "3:04 PM MST")
	timeFormats = append(timeFormats, "3PM MST")
	timeFormats = append(timeFormats, "3 PM MST")
	timeFormats = append(timeFormats, "15:04 MST")
	timeFormats = append(timeFormats, "1504 MST")
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

func ParseTime(timeString string) (time.Time, error) {
	for _, format := range timeFormats {
		parsedTime, err := time.Parse(format, timeString)
		if err != nil {
			continue
		}

		return parsedTime.UTC(), nil
	}

	return time.Time{}, errors.New("could not parse time")
}

func main() {
	currentTime := time.Now().UTC()
	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error decoding payload: "+err.Error())
		os.Exit(1)
	}

	versions := []models.Version{}
	start := request.Source.Between.Start
	stop := request.Source.Between.Stop
	interval := request.Source.Interval

	lastCheckedAt := request.Version.Time.UTC()

	validateConfig(start, stop, interval)

	if start != "" && stop != "" {
		startTime, err := ParseTime(start)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid start time: "+start+"; "+err.Error())
			os.Exit(1)
		}

		stopTime, err := ParseTime(stop)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid stop time: "+stop+"; "+err.Error())
			os.Exit(1)
		}

		if between.Between(startTime, stopTime, currentTime) {
			if lastCheckedAt.IsZero() {
				versions = append(versions, models.Version{Time: currentTime})
			} else {
				maxInterval := stopTime.Sub(startTime)

				if currentTime.Sub(request.Version.Time) > maxInterval {
					versions = append(versions, models.Version{Time: currentTime})
				}
			}
		}

	} else if interval != "" {
		parsedInterval, err := time.ParseDuration(interval)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid interval: "+interval+"; "+err.Error())
			os.Exit(1)
		}

		if currentTime.Sub(request.Version.Time) > parsedInterval {
			versions = append(versions, models.Version{Time: currentTime})
		}
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
