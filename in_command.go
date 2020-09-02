package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type InCommand struct {
}

func (*InCommand) Run(destination string, request models.InRequest) (models.InResponse, error) {
	err := os.MkdirAll(destination, 0755)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("creating destination: %w", err)
	}

	file, err := os.Create(filepath.Join(destination, "input"))
	if err != nil {
		return models.InResponse{}, fmt.Errorf("creating input file: %w", err)
	}

	defer file.Close()

	err = json.NewEncoder(file).Encode(request)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("writing input file: %w", err)
	}

	requestedVersionTime := request.Version.Time
	if requestedVersionTime.IsZero() {
		requestedVersionTime = GetCurrentTime()
	}

	tl := lord.TimeLord{
		PreviousTime: time.Time{},
		Location:     request.Source.Location,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Days:         request.Source.Days,
	}
	latestTime := tl.Latest(requestedVersionTime)
	offsetTime := Offset(tl, latestTime)

	inVersion := models.Version{Time: offsetTime}
	response := models.InResponse{Version: inVersion}

	return response, nil
}
