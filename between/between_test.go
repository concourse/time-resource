package between_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/time-resource/between"
)

type testCase struct {
	description string

	start         time.Time
	stop          time.Time
	timeToCompare time.Time

	result bool
}

var _ = Describe("Between", func() {
	pst, _ := time.LoadLocation("America/Los_Angeles")
	mst, _ := time.LoadLocation("America/Denver")

	var cases = []testCase{
		{
			description:   "between the start and stop time",
			start:         time.Date(2010, 1, 5, 2, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 4, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 3, 0, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "between the start and stop time down to the minute",
			start:         time.Date(2010, 1, 5, 2, 1, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 2, 3, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 2, 2, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "not between the start and stop time",
			start:         time.Date(2010, 1, 5, 2, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 4, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 5, 0, 0, 0, time.UTC),
			result:        false,
		},
		{
			description:   "not between the start and stop time down to the minute",
			start:         time.Date(2010, 1, 5, 2, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 4, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 4, 10, 0, 0, time.UTC),
			result:        false,
		},

		{
			description:   "not between the start and stop time down to the minute",
			start:         time.Date(2010, 1, 5, 11, 7, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 11, 10, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 11, 5, 0, 0, time.UTC),
			result:        false,
		},
		{
			description:   "not between the start and stop time down to the minute",
			start:         time.Date(2010, 1, 5, 10, 7, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 11, 10, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 11, 5, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "not between the start and stop time down to the minute",
			start:         time.Date(2010, 1, 5, 2, 20, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 4, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 5, 3, 10, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "one nanosecond before the start time",
			start:         time.Date(2010, 1, 2, 3, 4, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 2, 3, 7, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 2, 3, 3, 59, 999999999, time.UTC),
			result:        false,
		},
		{
			description:   "equal to the start time",
			start:         time.Date(2010, 1, 2, 3, 4, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 2, 3, 7, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 2, 3, 4, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "one nanosecond before the stop time",
			start:         time.Date(2010, 1, 2, 3, 4, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 2, 3, 7, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 2, 3, 6, 59, 999999999, time.UTC),
			result:        true,
		},
		{
			description:   "equal to the stop time",
			start:         time.Date(2010, 1, 2, 3, 4, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 2, 3, 7, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 1, 2, 3, 7, 0, 0, time.UTC),
			result:        false,
		},
		{
			description:   "between the start and stop time but on a different day",
			start:         time.Date(2010, 1, 5, 2, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 4, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 2, 5, 3, 0, 0, 0, time.UTC),
			result:        true,
		},

		// Our date parsing library always returns the date as 1/1 since we only
		// give it a time. If the stop time is before the start time then assume
		// that the stop is in the next day.
		{
			description:   "between the start and stop time but the stop time is before the start time",
			start:         time.Date(2010, 1, 5, 5, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 1, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 2, 5, 6, 0, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "between the start and stop time but the stop time is before the start time (ignoring the date)",
			start:         time.Date(2010, 1, 5, 5, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 2, 1, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 2, 5, 6, 0, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "between the start and stop time but the stop time is before the start time (when the time to compare is in the early hours)",
			start:         time.Date(2010, 1, 5, 20, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 8, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 2, 5, 1, 0, 0, 0, time.UTC),
			result:        true,
		},
		{
			description:   "between the start and stop time but the stop time is before the start time",
			start:         time.Date(2010, 1, 5, 5, 0, 0, 0, time.UTC),
			stop:          time.Date(2010, 1, 5, 1, 0, 0, 0, time.UTC),
			timeToCompare: time.Date(2010, 2, 5, 4, 0, 0, 0, time.UTC),
			result:        false,
		},

		{
			description:   "between the start and stop time but the compare time is in a different timezone",
			start:         time.Date(2010, 1, 5, 2, 0, 0, 0, mst),
			stop:          time.Date(2010, 1, 5, 6, 0, 0, 0, mst),
			timeToCompare: time.Date(2010, 2, 5, 1, 1, 0, 0, pst),
			result:        true,
		},
	}

	for _, testCase := range cases {
		capturedTestCase := testCase // closures (╯°□°）╯︵ ┻━┻
		description := fmt.Sprintf("returns %t if the time to compare is %s", capturedTestCase.result, capturedTestCase.description)

		It(description, func() {
			result := between.Between(capturedTestCase.start, capturedTestCase.stop, capturedTestCase.timeToCompare)
			Expect(result).To(Equal(capturedTestCase.result))
		})
	}
})
