package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/concourse/time-resource/models"
	"fmt"
)

func main() {
	var request models.OutRequest

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

	outVersion := models.Version{
		Time: currentTime,
	}

	metadata := models.Metadata{
		{"time", currentTime.String()},
		{"location", request.Source.Location},
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version:  outVersion,
		Metadata: metadata,
	})
}
