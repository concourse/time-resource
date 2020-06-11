package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

func main() {
	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	err = request.Source.Validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		os.Exit(1)
	}

	previousTime := request.Version.Time
	currentTime := time.Now().UTC()

	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		currentTime = currentTime.In((*time.Location)(specifiedLocation))
	}

	tl := lord.TimeLord{
		PreviousTime: previousTime,
		Location:     specifiedLocation,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Days:         request.Source.Days,
	}

	versions := []models.Version{}

	if previousTime.IsZero() {

		if latestIntervalTime := tl.Latest(currentTime); !latestIntervalTime.IsZero() {
			versions = append(versions, models.Version{Time: latestIntervalTime})
		}

	} else {

		latestVersionTime := previousTime

		for _, elem := range tl.List(previousTime) {
			if !elem.Before(previousTime) && !elem.After(currentTime) {
				latestVersionTime = elem
				versions = append(versions, models.Version{Time: latestVersionTime})
			}
		}

		// TODO Fill in the gap

		for _, elem := range tl.List(currentTime) {
			if elem.After(latestVersionTime) && !elem.After(currentTime) {
				latestVersionTime = elem
				versions = append(versions, models.Version{Time: latestVersionTime})
			}
		}
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
