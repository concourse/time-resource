package lord_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type expectedTime struct {
	isZero  bool
	hour    int
	minute  int
	weekday time.Weekday
}

type testCase struct {
	interval string

	location string

	start string
	stop  string

	days []time.Weekday

	start_after string
	prev        string
	prevDay     time.Weekday

	now       string
	extraTime time.Duration
	nowDay    time.Weekday

	result bool
	latest expectedTime
	list   []expectedTime
}

const exampleFormatWithTZ = "3:04 PM -0700 2006"
const exampleFormatWithoutTZ = "3:04 PM 2006"
const iso8601Format = "2006-01-02T15:04:05"

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

	if tc.start_after != "" {
		startTime, err := time.Parse(iso8601Format, tc.start_after)
		Expect(err).NotTo(HaveOccurred())
		startTimeModel := models.StartAfter(startTime.UTC())
		tl.StartAfter = &startTimeModel
	}

	if tc.start != "" {
		tc.start += " 2018"
		startTime, err := time.Parse(format, tc.start)
		Expect(err).NotTo(HaveOccurred())

		start := models.NewTimeOfDay(startTime.UTC())
		tl.Start = &start
	}

	if tc.stop != "" {
		tc.stop += " 2018"
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

	now, err := time.Parse(exampleFormatWithTZ, tc.now+" 2018")
	Expect(err).NotTo(HaveOccurred())

	for now.Weekday() != tc.nowDay {
		now = now.AddDate(0, 0, 1)
	}

	if tc.prev != "" {
		tc.prev += " 2018"
		prev, err := time.Parse(exampleFormatWithTZ, tc.prev)
		Expect(err).NotTo(HaveOccurred())

		for prev.Weekday() != tc.prevDay {
			prev = prev.AddDate(0, 0, 1)
		}

		tl.PreviousTime = prev
	}

	result := tl.Check(now.UTC())
	Expect(result).To(Equal(tc.result))

	latest := tl.Latest(now.UTC())
	Expect(latest.IsZero()).To(Equal(tc.latest.isZero))
	if !tc.latest.isZero {
		Expect(latest.Hour()).To(Equal(tc.latest.hour))
		Expect(latest.Minute()).To(Equal(tc.latest.minute))
		Expect(latest.Second()).To(Equal(0))
		Expect(latest.Weekday()).To(Equal(tc.latest.weekday))
		if tc.list == nil {
			tc.list = []expectedTime{tc.latest}
		}
	}

	list := tl.List(now.UTC())
	Expect(len(list)).To(Equal(len(tc.list)))
	for idx, actual := range list {
		expected := tc.list[idx]
		Expect(actual.Hour()).To(Equal(expected.hour))
		Expect(actual.Minute()).To(Equal(expected.minute))
		Expect(latest.Second()).To(Equal(0))
		Expect(actual.Weekday()).To(Equal(expected.weekday))
	}
}

var _ = DescribeTable("A range without a previous time", (testCase).Run,
	Entry("between the start and stop time", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "3:00 AM +0000",
		result: true,
		latest: expectedTime{hour: 2},
	}),
	Entry("between the start and stop time down to the minute", testCase{
		start:  "2:01 AM +0000",
		stop:   "2:03 AM +0000",
		now:    "2:02 AM +0000",
		result: true,
		latest: expectedTime{hour: 2, minute: 1},
	}),
	Entry("not between the start and stop time", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "5:00 AM +0000",
		result: false,
		latest: expectedTime{isZero: true},
	}),
	Entry("after the stop time, down to the minute", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "4:10 AM +0000",
		result: false,
		latest: expectedTime{isZero: true},
	}),
	Entry("before the start time, down to the minute", testCase{
		start:  "11:07 AM +0000",
		stop:   "11:10 AM +0000",
		now:    "11:05 AM +0000",
		result: false,
		latest: expectedTime{isZero: true},
	}),
	Entry("one nanosecond before the start time", testCase{
		start:     "3:04 AM +0000",
		stop:      "3:07 AM +0000",
		now:       "3:03 AM +0000",
		extraTime: time.Minute - time.Nanosecond,
		result:    false,
		latest:    expectedTime{isZero: true},
	}),
	Entry("equal to the start time", testCase{
		start:  "3:04 AM +0000",
		stop:   "3:07 AM +0000",
		now:    "3:04 AM +0000",
		result: true,
		latest: expectedTime{hour: 3, minute: 4},
	}),
	Entry("one nanosecond before the stop time", testCase{
		start:     "3:04 AM +0000",
		stop:      "3:07 AM +0000",
		now:       "3:06 AM +0000",
		extraTime: time.Minute - time.Nanosecond,
		result:    true,
		latest:    expectedTime{hour: 3, minute: 4},
	}),
	Entry("equal to the stop time", testCase{
		start:  "3:04 AM +0000",
		stop:   "3:07 AM +0000",
		now:    "3:07 AM +0000",
		result: false,
		latest: expectedTime{isZero: true},
	}),

	Entry("between the start and stop time but the stop time is before the start time, spanning more than a day", testCase{
		start:  "5:00 AM +0000",
		stop:   "1:00 AM +0000",
		now:    "6:00 AM +0000",
		result: true,
		latest: expectedTime{hour: 5},
	}),
	Entry("between the start and stop time but the stop time is before the start time, spanning half a day", testCase{
		start:  "8:00 PM +0000",
		stop:   "8:00 AM +0000",
		now:    "1:00 AM +0000",
		result: true,
		latest: expectedTime{hour: 20, weekday: time.Saturday},
	}),
	Entry("between the start and stop time but the stop time is before the start time and now is in the stop day", testCase{
		start:  "8:00 PM +0000",
		stop:   "8:00 AM +0000",
		now:    "7:00 AM +0000",
		result: true,
		latest: expectedTime{hour: 20, weekday: time.Saturday},
	}),

	Entry("between the start and stop time but the compare time is in a different timezone", testCase{
		start:  "2:00 AM -0600",
		stop:   "6:00 AM -0600",
		now:    "1:00 AM -0700",
		result: true,
		latest: expectedTime{hour: 8},
	}),

	Entry("covering almost a full day", testCase{
		start:  "12:01 AM -0700",
		stop:   "11:59 PM -0700",
		now:    "1:10 AM +0000",
		result: true,
		latest: expectedTime{hour: 7, minute: 1, weekday: time.Saturday},
	}),
)

var _ = DescribeTable("A range with a previous time", (testCase).Run,
	Entry("an hour before start", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "3:00 AM +0000",
		prev:   "1:00 AM +0000",
		result: true,
		latest: expectedTime{hour: 2},
	}),
	Entry("with stop before start and prev in the start day and now in the stop day", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "4:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "11:00 AM +0000",
		prevDay: time.Monday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
	Entry("with stop before start and prev outside the range and now in the stop day", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "4:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "9:00 AM +0000",
		prevDay: time.Monday,
		result:  true,
		latest:  expectedTime{hour: 10, weekday: time.Monday},
	}),
	Entry("after now and in range on same day as now", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "3:00 AM +0000",
		prev:   "3:30 AM +0000",
		result: false,
		latest: expectedTime{isZero: true},
	}),
	Entry("after now and out of range on same day as now", testCase{
		start:  "2:00 AM +0000",
		stop:   "4:00 AM +0000",
		now:    "3:00 AM +0000",
		prev:   "5:00 AM +0000",
		result: false,
		latest: expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("A range with a location and no previous time", (testCase).Run,
	Entry("between the start and stop time in a given location", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		now:      "6:00 PM +0000",
		result:   true,
		latest:   expectedTime{hour: 13},
	}),
	Entry("between the start and stop time in a given location on a matching day", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		days:     []time.Weekday{time.Wednesday},
		now:      "6:00 PM +0000",
		nowDay:   time.Wednesday,
		result:   true,
		latest:   expectedTime{hour: 13, weekday: time.Wednesday},
	}),
	Entry("not between the start and stop time in a given location", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		now:      "8:00 PM +0000",
		result:   false,
		latest:   expectedTime{isZero: true},
	}),
	Entry("between the start and stop time in a given location but not on a matching day", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "1:00 PM",
		stop:     "3:00 PM",
		days:     []time.Weekday{time.Wednesday},
		now:      "6:00 PM +0000",
		nowDay:   time.Thursday,
		result:   false,
		latest:   expectedTime{isZero: true},
	}),
	Entry("between the start and stop time in a given location and on a matching day compared to UTC", testCase{
		location: "America/Indiana/Indianapolis",
		start:    "9:00 PM",
		stop:     "11:00 PM",
		days:     []time.Weekday{time.Wednesday},
		now:      "2:00 AM +0000",
		nowDay:   time.Thursday,
		result:   true,
		latest:   expectedTime{hour: 21, weekday: time.Wednesday},
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
		latest: expectedTime{hour: 13, weekday: time.Thursday},
		list: []expectedTime{
			{hour: 13, weekday: time.Wednesday},
			{hour: 13, weekday: time.Thursday},
		},
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
		latest: expectedTime{hour: 13, weekday: time.Wednesday},
	}),
)

var _ = DescribeTable("An interval", (testCase).Run,
	Entry("without a previous time", testCase{
		interval: "2m",
		now:      "12:00 PM +0000",
		result:   true,
		latest:   expectedTime{hour: 12},
	}),
	Entry("with a previous time that has not elapsed", testCase{
		interval: "2m",
		prev:     "12:00 PM +0000",
		now:      "12:01 PM +0000",
		result:   false,
		latest:   expectedTime{hour: 12},
	}),
	Entry("with a previous time that has elapsed", testCase{
		interval: "2m",
		prev:     "12:00 PM +0000",
		now:      "12:02 PM +0000",
		result:   true,
		latest:   expectedTime{hour: 12, minute: 2},
		list: []expectedTime{
			{hour: 12},
			{hour: 12, minute: 2},
		},
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
		latest: expectedTime{hour: 13, weekday: time.Thursday},
		list: []expectedTime{
			{hour: 14, minute: 58, weekday: time.Wednesday},
			{hour: 13, weekday: time.Thursday},
		},
	}),
	Entry("between the start and stop time, elapsed", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev: "1:02 PM +0000",
		now:  "1:04 PM +0000",

		result: true,
		latest: expectedTime{hour: 13, minute: 4},
		list: []expectedTime{
			{hour: 13, minute: 2},
			{hour: 13, minute: 4},
		},
	}),
	Entry("between the start and stop time, not elapsed", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev: "1:02 PM +0000",
		now:  "1:03 PM +0000",

		result: false,
		latest: expectedTime{hour: 13, minute: 2},
	}),
	Entry("not between the start and stop time, elapsed", testCase{
		interval: "2m",

		start: "1:00 PM +0000",
		stop:  "3:00 PM +0000",

		prev: "2:58 PM +0000",
		now:  "3:02 PM +0000",

		result: false,
		latest: expectedTime{hour: 14, minute: 58},
	}),
)

var _ = DescribeTable("Start time with a range and interval", (testCase).Run,
	Entry("start_after is in the future, now is before start_after", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2025-01-01T00:00:00",
		now:         "3:06 AM +0000",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
	Entry("start_after is in the future, now is after start_after but before range", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2025-01-01T00:00:00",
		now:         "12:00 PM +0000",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
	Entry("start_after is in the past, now is within the range", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2017-12-31T00:00:00",
		now:         "1:30 PM +0000",
		result:      true,
		latest:      expectedTime{hour: 13, minute: 30},
	}),
	Entry("start_after is in the past, now is outside the range", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2017-12-31T00:00:00",
		now:         "4:00 PM +0000",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
	Entry("start_after is in the past, now is before the range", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2017-12-31T00:00:00",
		now:         "12:00 PM +0000",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
	Entry("start_after is in the past, now is exactly at the start of the range", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2017-12-31T00:00:00",
		now:         "1:00 PM +0000",
		result:      true,
		latest:      expectedTime{hour: 13, minute: 0},
	}),
	Entry("start_after is in the past, now is exactly at the stop of the range", testCase{
		interval:    "2m",
		start:       "1:00 PM +0000",
		stop:        "3:00 PM +0000",
		start_after: "2017-12-31T00:00:00",
		now:         "3:00 PM +0000",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
	Entry("start_after is in the past, now is in the range with location", testCase{
		location:    "America/Indiana/Indianapolis",
		start:       "1:00 PM",
		stop:        "3:00 PM",
		now:         "6:05 PM +0000",
		start_after: "2017-12-31T00:00:00",
		result:      true,
		latest:      expectedTime{hour: 13, minute: 0},
	}),
	Entry("start_after is in the past, now is before the range with location", testCase{
		location:    "America/Indiana/Indianapolis",
		start:       "1:00 PM",
		stop:        "3:00 PM",
		now:         "4:05 PM +0000",
		start_after: "2017-12-31T00:00:00",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
	Entry("start_after is in the past, now is after the range with location", testCase{
		location:    "America/Indiana/Indianapolis",
		start:       "1:00 PM",
		stop:        "3:00 PM",
		now:         "10:05 PM +0000",
		start_after: "2017-12-31T00:00:00",
		result:      false,
		latest:      expectedTime{isZero: true},
	}),
)
