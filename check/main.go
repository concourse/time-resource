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

	var versions []models.Version

	if previousTime.IsZero() {

		if latestIntervalTime := tl.Latest(currentTime); !latestIntervalTime.IsZero() {
			versions = []models.Version{{Time: latestIntervalTime}}
		}

	} else {
		timeList := tl.List(currentTime)
		versions = make([]models.Version, len(timeList))

		for idx, elem := range timeList {
			versions[idx] = models.Version{Time: elem}
		}
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
