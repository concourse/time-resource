package resource

import (
	"time"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type OutCommand struct {
}

func (*OutCommand) Run(request models.OutRequest) (models.OutResponse, error) {
	currentTime := GetCurrentTime()

	tl := lord.TimeLord{
		PreviousTime: time.Time{},
		Location:     request.Source.Location,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Days:         request.Source.Days,
	}
	latestTime := tl.Latest(currentTime)
	offsetTime := Offset(tl, latestTime)

	outVersion := models.Version{Time: offsetTime}
	response := models.OutResponse{Version: outVersion}

	return response, nil
}
