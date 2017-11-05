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

	tl := lord.TimeLord{
		PreviousTime: previousTime,
		Location:     request.Source.Location,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Skew:         request.Source.Skew,
		Days:         request.Source.Days,
	}

	versions := []models.Version{}

	if !previousTime.IsZero() {
		versions = append(versions, models.Version{Time: previousTime})
	}

	if tl.Check(currentTime) {
		versions = append(versions, models.Version{Time: currentTime})
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
