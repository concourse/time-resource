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

	cron string

	prev    string
	prevDay time.Weekday

	now       string
	extraTime time.Duration
	nowDay    time.Weekday

	result bool
	latest expectedTime
	list   []expectedTime
}

const exampleFormatWithTZ = "3:04 PM -0700 2006"
const exampleFormatWithoutTZ = "3:04 PM 2006"

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

	if tc.cron != "" {
		cronExpr := models.Cron{Expression: tc.cron}
		err := cronExpr.Validate()
		Expect(err).NotTo(HaveOccurred())

		tl.Cron = &cronExpr
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

	// Add any extra time if specified
	now = now.Add(tc.extraTime)

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
		Expect(actual.Second()).To(Equal(0))
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

var _ = DescribeTable("A cron expression without a previous time", (testCase).Run,
	Entry("exactly at the cron time (minute)", testCase{
		cron:   "30 * * * *", // Every hour at 30 minutes
		now:    "1:30 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 13, minute: 30, weekday: time.Monday},
	}),
	Entry("not at the cron time", testCase{
		cron:   "30 * * * *", // Every hour at 30 minutes
		now:    "1:29 PM +0000",
		nowDay: time.Monday,
		result: false,
		latest: expectedTime{isZero: true},
	}),
	Entry("at a specific hour and minute", testCase{
		cron:   "15 9 * * *", // Every day at 9:15
		now:    "9:15 AM +0000",
		nowDay: time.Tuesday,
		result: true,
		latest: expectedTime{hour: 9, minute: 15, weekday: time.Tuesday},
	}),
	Entry("at a specific hour and minute on specific days", testCase{
		cron:   "15 9 * * 1-5", // Every weekday at 9:15
		now:    "9:15 AM +0000",
		nowDay: time.Wednesday,
		result: true,
		latest: expectedTime{hour: 9, minute: 15, weekday: time.Wednesday},
	}),
	Entry("at a specific hour and minute but not on specified days", testCase{
		cron:   "15 9 * * 1-5", // Every weekday at 9:15
		now:    "9:15 AM +0000",
		nowDay: time.Sunday,
		result: false,
		latest: expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("A cron expression with a timezone", (testCase).Run,
	Entry("at the cron time with fixed offset", testCase{
		cron:     "0 9 * * *", // Every day at 9am
		location: "Etc/GMT+5",
		now:      "2:00 PM +0000", // 9am UTC-5
		nowDay:   time.Thursday,
		result:   true,
		latest:   expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
)

var _ = DescribeTable("A cron expression with complex patterns", (testCase).Run,
	Entry("with range of hours", testCase{
		cron:   "0 9-17 * * *", // Every hour from 9am to 5pm
		now:    "2:00 PM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
	Entry("with step values", testCase{
		cron:   "0 */2 * * *", // Every even hour
		now:    "2:00 PM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
	Entry("with step values not matching", testCase{
		cron:   "0 */2 * * *", // Every even hour
		now:    "3:00 PM +0000",
		nowDay: time.Thursday,
		result: false,
		latest: expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("A cron expression with edge cases", (testCase).Run,
	Entry("at midnight", testCase{
		cron:   "0 0 * * *", // Every day at midnight
		now:    "12:00 AM +0000",
		nowDay: time.Friday,
		result: true,
		latest: expectedTime{hour: 0, minute: 0, weekday: time.Friday},
	}),
	Entry("with last day of week", testCase{
		cron:   "0 12 * * 6", // Noon every Saturday
		now:    "12:00 PM +0000",
		nowDay: time.Saturday,
		result: true,
		latest: expectedTime{hour: 12, minute: 0, weekday: time.Saturday},
	}),
	Entry("with last minute of hour", testCase{
		cron:   "59 * * * *", // Every hour at 59 minutes
		now:    "1:59 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 13, minute: 59, weekday: time.Monday},
	}),
	Entry("with name of weekday", testCase{
		cron:   "0 12 * * MON", // Noon every Monday
		now:    "12:00 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 12, minute: 0, weekday: time.Monday},
	}),
)
