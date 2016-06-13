package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/concourse/time-resource/models"
)

func main() {
	currentTime := time.Now().UTC()

	outVersion := models.Version{
		Time: currentTime,
	}

	metadata := models.Metadata{
		{"time", currentTime.String()},
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version:  outVersion,
		Metadata: metadata,
	})
}
