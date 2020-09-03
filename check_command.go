package resource

import (
	"time"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type CheckCommand struct {
}

func (*CheckCommand) Run(request models.CheckRequest) ([]models.Version, error) {
	err := request.Source.Validate()
	if err != nil {
		return nil, err
	}

	currentTime := GetCurrentTime()

	tl := lord.TimeLord{
		Location: request.Source.Location,
		Start:    request.Source.Start,
		Stop:     request.Source.Stop,
		Interval: request.Source.Interval,
		Days:     request.Source.Days,
	}

	tl.PreviousTime = tl.Latest(request.Version.Time)

	var rawTimeList []time.Time
	if tl.PreviousTime.IsZero() {
		tl.PreviousTime = currentTime.AddDate(0, 0, -7)
		rawTimeList = []time.Time{tl.Latest(currentTime)}
	} else {
		rawTimeList = tl.List(currentTime)
	}

	var versions []models.Version
	for _, elem := range rawTimeList {
		offsetTime := Offset(tl, elem)
		if offsetTime.Before(request.Version.Time) {
			continue
		}
		if offsetTime.After(currentTime) {
			break
		}
		versions = append(versions, models.Version{Time: offsetTime})
	}

	return versions, nil
}
