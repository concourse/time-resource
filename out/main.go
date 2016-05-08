package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/concourse/time-resource/models"
)

func main() {
	var request models.OutRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("unable to decode payload: ", err)
	}
	if request.Source.Location == "" {
		request.Source.Location = "UTC"
	}
	loc, err := time.LoadLocation(request.Source.Location)
	if err != nil {
		fatal("unable to load timezone", err)
	}
	currentTime := time.Now().In(loc)

	outVersion := models.Version{
		Time: currentTime,
	}

	metadata := models.Metadata{
		{"time", currentTime.String()},
		{"timezone", request.Source.Location},
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version:  outVersion,
		Metadata: metadata,
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
