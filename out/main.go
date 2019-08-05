package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/concourse/time-resource/models"
)

func main() {
	var request models.OutRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	versionTime := time.Now().UTC()

	specifiedDelay := request.Params.After
	if specifiedDelay != nil {
		versionTime = versionTime.Add((time.Duration)(*specifiedDelay))
	}

	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		versionTime = versionTime.In((*time.Location)(specifiedLocation))
	}

	outVersion := models.Version{
		Time: versionTime,
	}

	json.NewEncoder(os.Stdout).Encode(models.OutResponse{
		Version: outVersion,
	})
}
