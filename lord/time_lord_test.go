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
		startAfter, err := time.Parse(iso8601Format, tc.start_after)
		Expect(err).NotTo(HaveOccurred())
		startAfterModel := models.StartAfter(startAfter.UTC())
		tl.StartAfter = &startAfterModel
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
		prev:    "7:00 AM +0000",
		prevDay: time.Monday,
		result:  true,
		latest:  expectedTime{hour: 10, weekday: time.Monday},
	}),
	Entry("with stop before start and prev in the start day and now after the stop time", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "6:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "11:00 AM +0000",
		prevDay: time.Monday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
	Entry("with stop before start and prev before the range and now after the start time", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "11:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "8:00 AM +0000",
		prevDay: time.Tuesday,
		result:  true,
		latest:  expectedTime{hour: 10, weekday: time.Tuesday},
	}),
	Entry("with different days where now is correct but prev is incorrect day", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "11:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "11:00 AM +0000",
		prevDay: time.Monday,
		result:  true,
		latest:  expectedTime{hour: 10, weekday: time.Tuesday},
	}),
	Entry("with stop before start and prev in the stop day", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "6:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "4:00 AM +0000",
		prevDay: time.Tuesday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
	Entry("with stop before start and prev in the stop day and now in the start day", testCase{
		start:   "10:00 AM +0000",
		stop:    "5:00 AM +0000",
		now:     "11:00 AM +0000",
		nowDay:  time.Tuesday,
		prev:    "4:00 AM +0000",
		prevDay: time.Tuesday,
		result:  true,
		latest:  expectedTime{hour: 10, weekday: time.Tuesday},
	}),
	Entry("in the range", testCase{
		start:   "2:00 AM +0000",
		stop:    "4:00 AM +0000",
		now:     "3:00 AM +0000",
		nowDay:  time.Monday,
		prev:    "2:30 AM +0000",
		prevDay: time.Monday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
	Entry("behind the range", testCase{
		start:   "2:00 AM +0000",
		stop:    "4:00 AM +0000",
		now:     "4:30 AM +0000",
		nowDay:  time.Monday,
		prev:    "2:00 AM +0000",
		prevDay: time.Monday,
		result:  false,
		latest:  expectedTime{isZero: true},
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
		start:    "1:00 PM +0000",
		stop:     "3:00 PM +0000",
		prev:     "2:58 PM +0000",
		prevDay:  time.Wednesday,
		now:      "1:00 PM +0000",
		nowDay:   time.Thursday,
		result:   true,
		latest:   expectedTime{hour: 13, weekday: time.Thursday},
		list: []expectedTime{
			{hour: 14, minute: 58, weekday: time.Wednesday},
			{hour: 13, weekday: time.Thursday},
		},
	}),
	Entry("between the start and stop time, elapsed", testCase{
		interval: "2m",
		start:    "1:00 PM +0000",
		stop:     "3:00 PM +0000",
		prev:     "1:02 PM +0000",
		now:      "1:04 PM +0000",
		result:   true,
		latest:   expectedTime{hour: 13, minute: 4},
		list: []expectedTime{
			{hour: 13, minute: 2},
			{hour: 13, minute: 4},
		},
	}),
	Entry("between the start and stop time, not elapsed", testCase{
		interval: "2m",
		start:    "1:00 PM +0000",
		stop:     "3:00 PM +0000",
		prev:     "1:02 PM +0000",
		now:      "1:03 PM +0000",
		result:   false,
		latest:   expectedTime{hour: 13, minute: 2},
	}),
	Entry("not between the start and stop time, elapsed", testCase{
		interval: "2m",
		start:    "1:00 PM +0000",
		stop:     "3:00 PM +0000",
		prev:     "2:58 PM +0000",
		now:      "3:02 PM +0000",
		result:   false,
		latest:   expectedTime{hour: 14, minute: 58},
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

var _ = DescribeTable("A cron expression without a previous time", (testCase).Run,
	Entry("exactly at the cron time (minute)", testCase{
		cron:   "30 * * * *",
		now:    "1:30 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 13, minute: 30, weekday: time.Monday},
	}),
	Entry("one minute before the cron time", testCase{
		cron:   "30 * * * *",
		now:    "1:29 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 12, minute: 30, weekday: time.Monday},
	}),
	Entry("at a specific hour and minute", testCase{
		cron:   "15 9 * * *",
		now:    "9:15 AM +0000",
		nowDay: time.Tuesday,
		result: true,
		latest: expectedTime{hour: 9, minute: 15, weekday: time.Tuesday},
	}),
	Entry("at a specific hour and minute on specific days", testCase{
		cron:   "15 9 * * 1-5",
		now:    "9:15 AM +0000",
		nowDay: time.Wednesday,
		result: true,
		latest: expectedTime{hour: 9, minute: 15, weekday: time.Wednesday},
	}),
	Entry("at a specific hour and minute but not on specified days", testCase{
		cron:   "15 9 * * 1-5",
		now:    "9:15 AM +0000",
		nowDay: time.Sunday,
		result: true,
		latest: expectedTime{hour: 9, minute: 15, weekday: time.Friday},
	}),
)

var _ = DescribeTable("A cron expression with a timezone", (testCase).Run,
	Entry("at the cron time with fixed offset", testCase{
		cron:     "0 9 * * *",
		location: "Etc/GMT+5",
		now:      "2:00 PM +0000",
		nowDay:   time.Thursday,
		result:   true,
		latest:   expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
)

var _ = DescribeTable("A cron expression with complex patterns", (testCase).Run,
	Entry("with range of hours", testCase{
		cron:   "0 9-17 * * *",
		now:    "2:00 PM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
	Entry("with step values", testCase{
		cron:   "0 */2 * * *",
		now:    "2:00 PM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
	Entry("with step values at odd hour", testCase{
		cron:   "0 */2 * * *",
		now:    "3:00 PM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 14, minute: 0, weekday: time.Thursday},
	}),
)

var _ = DescribeTable("A cron expression with edge cases", (testCase).Run,
	Entry("at midnight", testCase{
		cron:   "0 0 * * *",
		now:    "12:00 AM +0000",
		nowDay: time.Friday,
		result: true,
		latest: expectedTime{hour: 0, minute: 0, weekday: time.Friday},
	}),
	Entry("with last day of week", testCase{
		cron:   "0 12 * * 6",
		now:    "12:00 PM +0000",
		nowDay: time.Saturday,
		result: true,
		latest: expectedTime{hour: 12, minute: 0, weekday: time.Saturday},
	}),
	Entry("with last minute of hour", testCase{
		cron:   "59 * * * *",
		now:    "1:59 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 13, minute: 59, weekday: time.Monday},
	}),
	Entry("with name of weekday", testCase{
		cron:   "0 12 * * MON",
		now:    "12:00 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 12, minute: 0, weekday: time.Monday},
	}),
)

var _ = DescribeTable("A cron expression with a previous time", (testCase).Run,
	Entry("next cron time has passed", testCase{
		cron:    "30 * * * *",
		prev:    "12:30 PM +0000",
		prevDay: time.Monday,
		now:     "1:31 PM +0000",
		nowDay:  time.Monday,
		result:  true,
		latest:  expectedTime{hour: 13, minute: 30, weekday: time.Monday},
	}),
	Entry("next cron time has not passed", testCase{
		cron:    "30 * * * *",
		prev:    "12:30 PM +0000",
		prevDay: time.Monday,
		now:     "1:29 PM +0000",
		nowDay:  time.Monday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
	Entry("exactly at next cron time", testCase{
		cron:    "30 * * * *",
		prev:    "12:30 PM +0000",
		prevDay: time.Monday,
		now:     "1:30 PM +0000",
		nowDay:  time.Monday,
		result:  true,
		latest:  expectedTime{hour: 13, minute: 30, weekday: time.Monday},
	}),
	Entry("multiple cron times have passed", testCase{
		cron:    "30 * * * *",
		prev:    "12:30 PM +0000",
		prevDay: time.Monday,
		now:     "3:45 PM +0000",
		nowDay:  time.Monday,
		result:  true,
		latest:  expectedTime{hour: 15, minute: 30, weekday: time.Monday},
		list: []expectedTime{
			{hour: 13, minute: 30, weekday: time.Monday},
			{hour: 14, minute: 30, weekday: time.Monday},
			{hour: 15, minute: 30, weekday: time.Monday},
		},
	}),
)

var _ = DescribeTable("A cron expression with start_after", (testCase).Run,
	Entry("start_after in future, no previous time", testCase{
		cron:        "0 9 * * *",
		start_after: "2025-01-01T00:00:00",
		now:         "9:00 AM +0000",
		nowDay:      time.Monday,
		result:      false,
		latest:      expectedTime{isZero: true},
		list:        []expectedTime{},
	}),
	Entry("start_after in past, no previous time, at cron time", testCase{
		cron:        "0 9 * * *",
		start_after: "2017-12-31T00:00:00",
		now:         "9:00 AM +0000",
		nowDay:      time.Monday,
		result:      true,
		latest:      expectedTime{hour: 9, minute: 0, weekday: time.Monday},
	}),
	Entry("start_after in past, with previous time, next cron passed", testCase{
		cron:        "0 * * * *",
		start_after: "2017-12-31T00:00:00",
		prev:        "9:00 AM +0000",
		prevDay:     time.Monday,
		now:         "10:05 AM +0000",
		nowDay:      time.Monday,
		result:      true,
		latest:      expectedTime{hour: 10, minute: 0, weekday: time.Monday},
	}),
	Entry("start_after one second before now", testCase{
		cron:        "0 9 * * *",
		start_after: "2018-01-01T08:59:59",
		now:         "9:00 AM +0000",
		nowDay:      time.Monday,
		result:      true,
		latest:      expectedTime{hour: 9, minute: 0, weekday: time.Monday},
	}),
)

var _ = DescribeTable("A cron expression with timezone and previous time", (testCase).Run,
	Entry("timezone affects next calculation", testCase{
		cron:     "0 9 * * *",
		location: "America/New_York",
		prev:     "2:00 PM +0000",
		prevDay:  time.Monday,
		now:      "2:05 PM +0000",
		nowDay:   time.Tuesday,
		result:   true,
		latest:   expectedTime{hour: 14, minute: 0, weekday: time.Tuesday},
	}),
	Entry("timezone with prev, next not passed", testCase{
		cron:     "0 9 * * *",
		location: "America/New_York",
		prev:     "2:00 PM +0000",
		prevDay:  time.Monday,
		now:      "1:00 PM +0000",
		nowDay:   time.Tuesday,
		result:   false,
		latest:   expectedTime{isZero: true},
	}),
	Entry("timezone crossing midnight", testCase{
		cron:     "0 23 * * *",
		location: "America/Los_Angeles",
		now:      "7:00 AM +0000",
		nowDay:   time.Wednesday,
		result:   true,
		latest:   expectedTime{hour: 7, minute: 0, weekday: time.Wednesday},
	}),
)

var _ = DescribeTable("A range with days filter", (testCase).Run,
	Entry("on allowed day", testCase{
		start:  "9:00 AM +0000",
		stop:   "5:00 PM +0000",
		days:   []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		now:    "10:00 AM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 9, minute: 0, weekday: time.Monday},
	}),
	Entry("not on allowed day", testCase{
		start:  "9:00 AM +0000",
		stop:   "5:00 PM +0000",
		days:   []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		now:    "10:00 AM +0000",
		nowDay: time.Tuesday,
		result: false,
		latest: expectedTime{isZero: true},
	}),
	Entry("empty days means all days allowed", testCase{
		start:  "9:00 AM +0000",
		stop:   "5:00 PM +0000",
		days:   []time.Weekday{},
		now:    "10:00 AM +0000",
		nowDay: time.Saturday,
		result: true,
		latest: expectedTime{hour: 9, minute: 0, weekday: time.Saturday},
	}),
	Entry("days filter with previous time", testCase{
		start:   "9:00 AM +0000",
		stop:    "5:00 PM +0000",
		days:    []time.Weekday{time.Monday, time.Wednesday},
		prev:    "10:00 AM +0000",
		prevDay: time.Monday,
		now:     "10:00 AM +0000",
		nowDay:  time.Wednesday,
		result:  true,
		latest:  expectedTime{hour: 9, minute: 0, weekday: time.Wednesday},
	}),
)

var _ = DescribeTable("Days filter with location", (testCase).Run,
	Entry("timezone makes it different day", testCase{
		location: "Pacific/Auckland",
		start:    "9:00 AM",
		stop:     "5:00 PM",
		days:     []time.Weekday{time.Tuesday},
		now:      "8:00 PM +0000",
		nowDay:   time.Monday,
		result:   true,
		latest:   expectedTime{hour: 9, minute: 0, weekday: time.Tuesday},
	}),
	Entry("timezone makes it wrong day", testCase{
		location: "Pacific/Auckland",
		start:    "9:00 AM",
		stop:     "5:00 PM",
		days:     []time.Weekday{time.Monday},
		now:      "8:00 PM +0000",
		nowDay:   time.Monday,
		result:   false,
		latest:   expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("Interval with days filter", (testCase).Run,
	Entry("interval on allowed day", testCase{
		interval: "1h",
		start:    "9:00 AM +0000",
		stop:     "5:00 PM +0000",
		days:     []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		now:      "10:00 AM +0000",
		nowDay:   time.Monday,
		result:   true,
		latest:   expectedTime{hour: 10, minute: 0, weekday: time.Monday},
	}),
	Entry("interval on non-allowed day", testCase{
		interval: "1h",
		start:    "9:00 AM +0000",
		stop:     "5:00 PM +0000",
		days:     []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		now:      "10:00 AM +0000",
		nowDay:   time.Tuesday,
		result:   false,
		latest:   expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("Every minute cron expression", (testCase).Run,
	Entry("no previous, triggers immediately", testCase{
		cron:   "* * * * *",
		now:    "3:47 PM +0000",
		nowDay: time.Tuesday,
		result: true,
		latest: expectedTime{hour: 15, minute: 47, weekday: time.Tuesday},
	}),
	Entry("with previous 1 minute ago", testCase{
		cron:    "* * * * *",
		prev:    "3:46 PM +0000",
		prevDay: time.Tuesday,
		now:     "3:47 PM +0000",
		nowDay:  time.Tuesday,
		result:  true,
		latest:  expectedTime{hour: 15, minute: 47, weekday: time.Tuesday},
	}),
	Entry("with previous same minute - no trigger", testCase{
		cron:      "* * * * *",
		prev:      "3:47 PM +0000",
		prevDay:   time.Tuesday,
		now:       "3:47 PM +0000",
		extraTime: 30 * time.Second,
		nowDay:    time.Tuesday,
		result:    false,
		latest:    expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("Special cron expressions", (testCase).Run,
	Entry("@hourly at exact hour", testCase{
		cron:   "@hourly",
		now:    "3:00 PM +0000",
		nowDay: time.Wednesday,
		result: true,
		latest: expectedTime{hour: 15, minute: 0, weekday: time.Wednesday},
	}),
	Entry("@hourly between hours", testCase{
		cron:   "@hourly",
		now:    "3:30 PM +0000",
		nowDay: time.Wednesday,
		result: true,
		latest: expectedTime{hour: 15, minute: 0, weekday: time.Wednesday},
	}),
	Entry("@daily at midnight", testCase{
		cron:   "@daily",
		now:    "12:00 AM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 0, minute: 0, weekday: time.Thursday},
	}),
	Entry("@daily in afternoon finds today's midnight", testCase{
		cron:   "@daily",
		now:    "3:00 PM +0000",
		nowDay: time.Thursday,
		result: true,
		latest: expectedTime{hour: 0, minute: 0, weekday: time.Thursday},
	}),
	Entry("@weekly on Sunday midnight", testCase{
		cron:   "@weekly",
		now:    "12:00 AM +0000",
		nowDay: time.Sunday,
		result: true,
		latest: expectedTime{hour: 0, minute: 0, weekday: time.Sunday},
	}),
)

var _ = DescribeTable("List edge cases", (testCase).Run,
	Entry("list with no matches returns empty", testCase{
		start:  "9:00 AM +0000",
		stop:   "5:00 PM +0000",
		now:    "6:00 PM +0000",
		nowDay: time.Monday,
		result: false,
		latest: expectedTime{isZero: true},
		list:   []expectedTime{},
	}),
)

var _ = DescribeTable("Previous time edge cases", (testCase).Run,
	Entry("previous time after now returns zero latest", testCase{
		start:   "9:00 AM +0000",
		stop:    "5:00 PM +0000",
		prev:    "10:00 AM +0000",
		prevDay: time.Tuesday,
		now:     "9:30 AM +0000",
		nowDay:  time.Monday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("Multi-day scenarios", (testCase).Run,
	Entry("overnight range with prev in evening, now in morning", testCase{
		start:   "10:00 PM +0000",
		stop:    "6:00 AM +0000",
		prev:    "11:00 PM +0000",
		prevDay: time.Monday,
		now:     "3:00 AM +0000",
		nowDay:  time.Tuesday,
		result:  false,
		latest:  expectedTime{isZero: true},
	}),
	Entry("overnight range with prev before range, now in morning", testCase{
		start:   "10:00 PM +0000",
		stop:    "6:00 AM +0000",
		prev:    "9:00 PM +0000",
		prevDay: time.Monday,
		now:     "3:00 AM +0000",
		nowDay:  time.Tuesday,
		result:  true,
		latest:  expectedTime{hour: 22, minute: 0, weekday: time.Monday},
	}),
)

var _ = DescribeTable("Cron with comma-separated values", (testCase).Run,
	Entry("multiple minutes", testCase{
		cron:   "0,15,30,45 * * * *",
		now:    "3:17 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 15, minute: 15, weekday: time.Monday},
	}),
	Entry("multiple hours", testCase{
		cron:   "0 9,12,15,18 * * *",
		now:    "1:00 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 12, minute: 0, weekday: time.Monday},
	}),
	Entry("multiple days of week", testCase{
		cron:   "0 12 * * 1,3,5",
		now:    "12:00 PM +0000",
		nowDay: time.Wednesday,
		result: true,
		latest: expectedTime{hour: 12, minute: 0, weekday: time.Wednesday},
	}),
)

var _ = DescribeTable("Only start time specified", (testCase).Run,
	Entry("after start time, no prev", testCase{
		start:  "9:00 AM +0000",
		now:    "10:00 AM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 9, minute: 0, weekday: time.Monday},
	}),
	Entry("before start time, no prev", testCase{
		start:  "9:00 AM +0000",
		now:    "8:00 AM +0000",
		nowDay: time.Monday,
		result: false,
		latest: expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("Only stop time specified", (testCase).Run,
	Entry("before stop time, no prev", testCase{
		stop:   "5:00 PM +0000",
		now:    "3:00 PM +0000",
		nowDay: time.Monday,
		result: true,
		latest: expectedTime{hour: 0, minute: 0, weekday: time.Monday},
	}),
	Entry("after stop time, no prev", testCase{
		stop:   "5:00 PM +0000",
		now:    "6:00 PM +0000",
		nowDay: time.Monday,
		result: false,
		latest: expectedTime{isZero: true},
	}),
)

var _ = DescribeTable("StartAfter with non-cron configurations", (testCase).Run,
	// Bug #2 fix: Check() non-cron path with StartAfter
	Entry("start_after with range, now before start_after", testCase{
		start:       "9:00 AM +0000",
		stop:        "5:00 PM +0000",
		start_after: "2025-01-15T12:00:00",
		now:         "10:00 AM +0000",
		nowDay:      time.Wednesday,
		// now (2018) is before start_after (2025), should not trigger
		result: false,
		latest: expectedTime{isZero: true},
		list:   []expectedTime{},
	}),
	Entry("start_after with range, now after start_after", testCase{
		start:       "9:00 AM +0000",
		stop:        "5:00 PM +0000",
		start_after: "2017-01-01T00:00:00",
		now:         "10:00 AM +0000",
		nowDay:      time.Wednesday,
		// now (2018) is after start_after (2017), should trigger
		result: true,
		latest: expectedTime{hour: 9, minute: 0, weekday: time.Wednesday},
	}),
	Entry("start_after with interval, now before start_after", testCase{
		interval:    "1h",
		start:       "9:00 AM +0000",
		stop:        "5:00 PM +0000",
		start_after: "2025-01-15T12:00:00",
		now:         "11:00 AM +0000",
		nowDay:      time.Wednesday,
		result:      false,
		latest:      expectedTime{isZero: true},
		list:        []expectedTime{},
	}),
	Entry("start_after with interval, now after start_after", testCase{
		interval:    "1h",
		start:       "9:00 AM +0000",
		stop:        "5:00 PM +0000",
		start_after: "2017-01-01T00:00:00",
		now:         "11:00 AM +0000",
		nowDay:      time.Wednesday,
		result:      true,
		latest:      expectedTime{hour: 11, minute: 0, weekday: time.Wednesday},
	}),

	// Bug #2 fix: timezone handling in non-cron StartAfter check
	Entry("start_after with location, boundary test", testCase{
		location:    "America/New_York",
		start:       "9:00 AM",
		stop:        "5:00 PM",
		start_after: "2018-01-03T10:00:00", // interpreted as 10:00 AM New York
		now:         "3:00 PM +0000",       // 10:00 AM New York
		nowDay:      time.Wednesday,
		// now equals start_after, should not trigger (needs to be strictly after)
		result: false,
		latest: expectedTime{isZero: true},
		list:   []expectedTime{},
	}),
	Entry("start_after with location, one minute after", testCase{
		location:    "America/New_York",
		start:       "9:00 AM",
		stop:        "5:00 PM",
		start_after: "2018-01-03T10:00:00",
		now:         "3:01 PM +0000",
		nowDay:      time.Wednesday,
		result:      true,
		latest:      expectedTime{hour: 9, minute: 0, weekday: time.Wednesday}, // local hour, not UTC
	}),

	// With previous time
	Entry("start_after with range and prev, now after start_after", testCase{
		start:       "9:00 AM +0000",
		stop:        "5:00 PM +0000",
		start_after: "2017-01-01T00:00:00",
		prev:        "8:00 AM +0000",
		prevDay:     time.Tuesday,
		now:         "10:00 AM +0000",
		nowDay:      time.Wednesday,
		result:      true,
		latest:      expectedTime{hour: 9, minute: 0, weekday: time.Wednesday},
	}),
	Entry("start_after with range and prev, now before start_after", testCase{
		start:       "9:00 AM +0000",
		stop:        "5:00 PM +0000",
		start_after: "2025-01-15T12:00:00",
		prev:        "8:00 AM +0000",
		prevDay:     time.Tuesday,
		now:         "10:00 AM +0000",
		nowDay:      time.Wednesday,
		result:      false,
		latest:      expectedTime{isZero: true},
		list:        []expectedTime{},
	}),
)

var _ = DescribeTable("StartAfter edge cases", (testCase).Run,
	// start_after alone (no range, no interval, no cron)
	Entry("start_after only, now before", testCase{
		start_after: "2025-01-15T00:00:00",
		now:         "10:00 AM +0000",
		nowDay:      time.Wednesday,
		result:      false,
		latest:      expectedTime{isZero: true},
		list:        []expectedTime{},
	}),
	Entry("start_after only, now after", testCase{
		start_after: "2017-01-01T00:00:00",
		now:         "10:00 AM +0000",
		nowDay:      time.Wednesday,
		result:      true,
		latest:      expectedTime{hour: 0, minute: 0, weekday: time.Wednesday},
	}),
)
