package resource

import (
	"time"
)

var GetCurrentTime = func() time.Time {
	return time.Now().UTC()
}
