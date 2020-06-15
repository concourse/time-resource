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
	var request models.OutRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	currentTime := time.Now().UTC()

	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		currentTime = currentTime.In((*time.Location)(specifiedLocation))
	}

	tl := lord.TimeLord{
		Location: specifiedLocation,
		Start:    request.Source.Start,
		Stop:     request.Source.Stop,
		Interval: request.Source.Interval,
		Days:     request.Source.Days,
	}

	latestTime, _ := tl.Latest(currentTime)

	outVersion := models.Version{
		Time: latestTime,
	}

	json.NewEncoder(os.Stdout).Encode(models.OutResponse{
		Version: outVersion,
	})
}
