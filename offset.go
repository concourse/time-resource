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
const offsetHashFormat = "%s/%s"

const maxHashValue = int64(math.MaxUint32)

var msPerMinute = time.Minute.Milliseconds()

func Offset(tl lord.TimeLord, reference time.Time) time.Time {
	str := fmt.Sprintf(offsetHashFormat, os.Getenv(BUILD_TEAM_NAME), os.Getenv(BUILD_PIPELINE_NAME))
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

	hashPerMinute := maxHashValue / (rangeDuration.Milliseconds() / msPerMinute)
	offsetDuration := time.Duration(hash/hashPerMinute) * time.Minute

	return start.Add(offsetDuration)
}
