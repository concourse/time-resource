package lord_test

import (
	"time"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type testCase struct {
	interval string

	location string

	start string
	stop  string

	days []time.Weekday

	prev    string
	prevDay time.Weekday

	now       string
	extraTime time.Duration
	nowDay    time.Weekday

	result bool
}

const exampleFormatWithTZ = "3:04 PM -0700"
const exampleFormatWithoutTZ = "3:04 PM"

func (tc testCase) Run() {
	var tl lord.TimeLord

	if tc.location != "" {
		loc, err := time.LoadLocation(tc.location)
		Expect(err).NotTo(HaveOccurred())

		tl.Location = (*models.Location)(loc)
	}

	var format string
	if tl.Location != nil {
		format = exampleFormatWithoutTZ
	} else {
		format = exampleFormatWithTZ
	}

	if tc.start != "" {
		startTime, err := time.Parse(format, tc.start)
		Expect(err).NotTo(HaveOccurred())

		start := models.NewTimeOfDay(startTime.UTC())
		tl.Start = &start
	}

	if tc.stop != "" {
		stopTime, err := time.Parse(format, tc.stop)
		Expect(err).NotTo(HaveOccurred())

		stop := models.NewTimeOfDay(stopTime.UTC())
		tl.Stop = &stop
	}

	if tc.interval != "" {
		interval, err := time.ParseDuration(tc.interval)
		Expect(err).NotTo(HaveOccurred())

		tl.Interval = (*models.Interval)(&interval)
	}

	tl.Days = make([]models.Weekday, len(tc.days))
	for i, d := range tc.days {
		tl.Days[i] = models.Weekday(d)
	}

	now, err := time.Parse(exampleFormatWithTZ, tc.now)
	Expect(err).NotTo(HaveOccurred())

	for now.Weekday() != tc.nowDay {
		now = now.AddDate(0, 0, 1)
	}

	if tc.prev != "" {
		prev, err := time.Parse(exampleFormatWithTZ, tc.prev)
		Expect(err).NotTo(HaveOccurred())

		for prev.Weekday() != tc.prevDay {
			prev = prev.AddDate(0, 0, 1)
		}

		tl.PreviousTime = prev
	}

	result := tl.Check(now.UTC())
	Expect(result).To(Equal(tc.result))
}

var _ = DescribeTable("A range without a previous time", (testCase).Run,
	Entry("between the start and stop time", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "3:00 AM +0000",
		result: true,
	}),
	Entry("between the start and stop time down to the minute", testCase{
		start:  "2:01 AM +0000",
		stop:   "2:03 AM +0000",
		now:    "2:02 AM +0000",
		result: true,
	}),
	Entry("not between the start and stop time", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "5:00 AM +0000",
		result: false,
	}),
	Entry("after the stop time, down to the minute", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "4:10 AM +0000",
		result: false,
	}),
	Entry("before the start time, down to the minute", testCase{
		start:  "11:07 AM +0000",
		stop:   "11:10 AM +0000",
		now:    "11:05 AM +0000",
		result: false,
	}),
	Entry("one nanosecond before the start time", testCase{
		start:     "3:04 AM +0000",
		stop:      "3:07 AM +0000",
		now:       "3:03 AM +0000",
		extraTime: time.Minute - time.Nanosecond,
		result:    false,
	}),
	Entry("equal to the start time", testCase{
		start:  "3:04 AM +0000",
		stop:   "3:07 AM +0000",
		now:    "3:04 AM +0000",
		result: true,
	}),
	Entry("one nanosecond before the stop time", testCase{
		start:     "3:04 AM +0000",
		stop:      "3:07 AM +0000",
		now:       "3:06 AM +0000",
		extraTime: time.Minute - time.Nanosecond,
		result:    true,
	}),
	Entry("equal to the stop time", testCase{
		start:  "3:04 AM +0000",
		stop:   "3:07 AM +0000",
		now:    "3:07 AM +0000",
		result: false,
	}),

	Entry("between the start and stop time but the stop time is before the start time, spanning more than a day", testCase{
		start:  "5:00 AM +0000",
		stop:   "1:00 AM +0000",
		now:    "6:00 AM +0000",
		result: true,
	}),
	Entry("between the start and stop time but the stop time is before the start time, spanning half a day", testCase{
		start:  "8:00 PM +0000",
		stop:   "8:00 AM +0000",
		now:    "1:00 AM +0000",
		result: true,
	}),

	Entry("between the start and stop time but the compare time is in a different timezone", testCase{
		start:  "2:00 AM -0600",
		stop:   "6:00 AM -0600",
		now:    "1:00 AM -0700",
		result: true,
	}),
)

var _ = DescribeTable("A range with a location and no previous time", (testCase).Run,
	Entry("between the start and stop time in a given location", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		now:      "6:00 PM +0000",
		result:   true,
	}),
	Entry("between the start and stop time in a given location on a matching day", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		days:     []time.Weekday{time.Wednesday},
		now:      "6:00 PM +0000",
		nowDay:   time.Wednesday,
		result:   true,
	}),
	Entry("not between the start and stop time in a given location", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		now:      "8:00 PM +0000",
		result:   false,
	}),
	Entry("between the start and stop time in a given location but not on a matching day", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		days:     []time.Weekday{time.Wednesday},
		now:      "6:00 PM +0000",
		nowDay:   time.Thursday,
		result:   false,
	}),
	Entry("between the start and stop time in a given location and on a matching day compared to UTC", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "10:00 PM",
		stop:     "11:00 PM",
		days:     []time.Weekday{time.Wednesday},
		now:      "2:00 AM +0000",
		nowDay:   time.Thursday,
		result:   true,
	}),
)

var _ = DescribeTable("A range with a location and a previous time", (testCase).Run,
	Entry("between the start and stop time in a given location, on a new day", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",

		prev:    "6:00 PM +0000",
		prevDay: time.Wednesday,
		now:     "6:00 PM +0000",
		nowDay:  time.Thursday,

		result: true,
	}),
	Entry("not between the start and stop time in a given location, on the same day", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",

		prev:    "6:00 PM +0000",
		prevDay: time.Wednesday,
		now:     "6:01 PM +0000",
		nowDay:  time.Wednesday,

		result: false,
	}),
)

var _ = DescribeTable("An interval", (testCase).Run,
	Entry("without a previous time", testCase{
		interval: "2m",
		now:      "12:00 PM +0000",
		result:   true,
	}),
	Entry("with a previous time that has not elapsed", testCase{
		interval: "2m",
		prev:     "12:00 PM +0000",
		now:      "12:01 PM +0000",
		result:   false,
	}),
	Entry("with a previous time that has elapsed", testCase{
		interval: "2m",
		prev:     "12:00 PM +0000",
		now:      "12:02 PM +0000",
		result:   true,
	}),
)

var _ = DescribeTable("A range with an interval and a previous time", (testCase).Run,
	Entry("between the start and stop time, on a new day", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev:    "2:58 PM +0000",
		prevDay: time.Wednesday,
		now:     "1:00 PM +0000",
		nowDay:  time.Thursday,

		result: true,
	}),
	Entry("between the start and stop time, elapsed", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev: "1:02 PM +0000",
		now:  "1:04 PM +0000",

		result: true,
	}),
	Entry("between the start and stop time, not elapsed", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev: "1:02 PM +0000",
		now:  "1:03 PM +0000",

		result: false,
	}),
	Entry("not between the start and stop time, elapsed", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev: "2:58 PM +0000",
		now:  "3:02 PM +0000",

		result: false,
	}),
)
