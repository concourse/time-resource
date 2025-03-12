package resource

import (
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"time"

	"github.com/concourse/time-resource/lord"
)

const BUILD_TEAM_NAME = "BUILD_TEAM_NAME"
const BUILD_PIPELINE_NAME = "BUILD_PIPELINE_NAME"
const BUILD_PIPELINE_INSTANCE_VARS = "BUILD_PIPELINE_INSTANCE_VARS"

const maxHashValue = int64(math.MaxUint32)

var msPerMinute = time.Minute.Milliseconds()

func Offset(tl lord.TimeLord, reference time.Time) time.Time {
	str := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv(BUILD_TEAM_NAME),
		os.Getenv(BUILD_PIPELINE_NAME),
		os.Getenv(BUILD_PIPELINE_INSTANCE_VARS),
	)
	hasher := fnv.New32a()
	if _, err := hasher.Write([]byte(str)); err != nil {
		fmt.Fprintln(os.Stderr, "hash error:", err.Error())
		os.Exit(1)
	}
	hash := int64(hasher.Sum32())

	start, stop := tl.LatestRangeBefore(reference)
	rangeDuration := stop.Sub(start)

	if tl.Interval != nil {
		if intervalDuration := time.Duration(*tl.Interval); intervalDuration < rangeDuration {
			rangeDuration = intervalDuration
			start = reference.Truncate(rangeDuration)
		}
	}

	if rangeDuration <= time.Minute {
		return start
	}

	rangeMs := rangeDuration.Milliseconds()
	if rangeMs <= 0 {
		return start
	}

	minutesInRange := rangeMs / msPerMinute
	if minutesInRange <= 0 {
		minutesInRange = 1
	}

	hashPerMinute := maxHashValue / minutesInRange
	if hashPerMinute <= 0 {
		hashPerMinute = 1
	}

	minutesToOffset := hash / hashPerMinute

	// Guard against overflows
	if minutesToOffset < 0 {
		minutesToOffset = 0
	}

	// Ensure the offset doesn't exceed the range duration
	if minutesToOffset > minutesInRange {
		minutesToOffset = minutesInRange
	}

	offsetDuration := time.Duration(minutesToOffset) * time.Minute
	return start.Add(offsetDuration)
}
