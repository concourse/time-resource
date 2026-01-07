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

	previousTime := request.Version.Time
	currentTime := time.Now().UTC()

	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		currentTime = currentTime.In((*time.Location)(specifiedLocation))
	}

	tl := lord.TimeLord{
		PreviousTime: previousTime,
		Location:     specifiedLocation,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Days:         request.Source.Days,
		StartAfter:   request.Source.StartAfter,
		Cron:         request.Source.Cron,
	}

	var versions []models.Version

	if !previousTime.IsZero() {
		versions = append(versions, models.Version{Time: previousTime})
	} else if request.Source.InitialVersion {
		// For cron with initial_version, use the cron boundary time for consistency.
		// For non-cron, use currentTime (original behavior).
		versionTime := currentTime
		if request.Source.Cron != nil {
			cronTime := tl.Latest(currentTime)
			if !cronTime.IsZero() {
				versionTime = cronTime
			}
		}
		versions = append(versions, models.Version{Time: versionTime})
		return versions, nil
	}

	if tl.Check(currentTime) {
		var versionTime time.Time

		// For cron expressions, use the actual scheduled cron time
		// instead of the check time. This ensures versions are at cron boundaries.
		// Example: cron @5minutes, check at 3:07pm → version time = 3:05pm
		if request.Source.Cron != nil {
			versionTime = tl.Latest(currentTime)
		}

		// For non-cron (interval, start/stop ranges), use currentTime
		if versionTime.IsZero() {
			versionTime = currentTime
		}

		versions = append(versions, models.Version{Time: versionTime})
	}

	return versions, nil
}
