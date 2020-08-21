package resource

import (
	"time"

	"github.com/concourse/time-resource/models"
)

type OutCommand struct {
}

func (*OutCommand) Run(request models.OutRequest) (models.OutResponse, error) {
	currentTime := time.Now().UTC()
	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		currentTime = currentTime.In((*time.Location)(specifiedLocation))
	}

	outVersion := models.Version{Time: currentTime}
	response := models.OutResponse{Version: outVersion}

	return response, nil
}
