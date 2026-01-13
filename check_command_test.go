package resource_test

import (
	"fmt"
	"time"

	resource "github.com/concourse/time-resource"
	"github.com/concourse/time-resource/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {
	var now time.Time

	BeforeEach(func() {
		now = time.Now().UTC()
	})

	Context("when executed", func() {
		var source models.Source
		var version models.Version
		var response models.CheckResponse

		BeforeEach(func() {
			source = models.Source{}
			version = models.Version{}
			response = models.CheckResponse{}
		})

		JustBeforeEach(func() {
			command := resource.CheckCommand{}
			var err error
			response, err = command.Run(models.CheckRequest{
				Source:  source,
				Version: version,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when nothing is specified", func() {
			Context("when no version is given", func() {
				It("outputs a version containing the current time", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})

			Context("when a version is given", func() {
				Context("when the resource has already triggered today", func() {
					BeforeEach(func() {
						version.Time = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, now.Second(), now.Nanosecond(), now.Location())
					})

					It("outputs only the supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", version.Time.Unix(), 1))
					})
				})

				Context("when the resource was triggered yesterday", func() {
					BeforeEach(func() {
						version.Time = now.Add(-24 * time.Hour)
					})

					It("outputs both previous and current versions", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", version.Time.Unix(), 1))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})
		})

		Context("when a time range is specified", func() {
			Context("when we are in the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(-1 * time.Hour)
					stop := now.Add(1 * time.Hour)
					source.Start = tod(start.Hour(), start.Minute(), 0)
					source.Stop = tod(stop.Hour(), stop.Minute(), 0)
				})

				Context("when no version is given", func() {
					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when a version is given within the current time range", func() {
					BeforeEach(func() {
						version.Time = now.Add(-30 * time.Minute)
					})

					It("outputs only the supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", version.Time.Unix(), 1))
					})
				})

				Context("when a version is given from yesterday", func() {
					BeforeEach(func() {
						version.Time = now.Add(-24 * time.Hour)
					})

					It("outputs both previous and current versions", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", version.Time.Unix(), 1))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when an interval is specified", func() {
					BeforeEach(func() {
						interval := models.Interval(time.Minute)
						source.Interval = &interval
					})

					Context("when no version is given", func() {
						It("outputs a version containing the current time", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the interval has not elapsed", func() {
						BeforeEach(func() {
							version.Time = now
						})

						It("outputs only the supplied version", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
						})
					})

					Context("when the interval has elapsed", func() {
						BeforeEach(func() {
							version.Time = now.Add(-2 * time.Minute)
						})

						It("outputs both previous and current versions", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})
				})

				Context("when the current day is specified", func() {
					BeforeEach(func() {
						source.Days = []models.Weekday{models.Weekday(now.Weekday())}
					})

					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when we are out of the specified day", func() {
					BeforeEach(func() {
						source.Days = []models.Weekday{models.Weekday(now.AddDate(0, 0, 1).Weekday())}
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})

			Context("when we are not within the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(6 * time.Hour)
					stop := now.Add(7 * time.Hour)
					source.Start = tod(start.Hour(), start.Minute(), 0)
					source.Stop = tod(stop.Hour(), stop.Minute(), 0)
				})

				It("does not output any versions", func() {
					Expect(response).To(BeEmpty())
				})

				Context("when an interval is given", func() {
					BeforeEach(func() {
						interval := models.Interval(time.Minute)
						source.Interval = &interval
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})

			Context("with a location configured", func() {
				BeforeEach(func() {
					loc, err := time.LoadLocation("America/Indiana/Indianapolis")
					Expect(err).ToNot(HaveOccurred())
					srcLoc := models.Location(*loc)
					source.Location = &srcLoc
					now = now.In(loc)

					start := now.Add(-1 * time.Hour)
					stop := now.Add(1 * time.Hour)
					source.Start = tod(start.Hour(), start.Minute(), 0)
					source.Stop = tod(stop.Hour(), stop.Minute(), 0)
				})

				It("outputs a version when in range", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})
		})

		Context("when an interval is specified", func() {
			BeforeEach(func() {
				interval := models.Interval(time.Minute)
				source.Interval = &interval
			})

			Context("when no version is given", func() {
				It("outputs a version containing the current time", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})

			Context("when the interval has not elapsed", func() {
				BeforeEach(func() {
					version.Time = now
				})

				It("outputs only the supplied version", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
				})
			})

			Context("when the interval has elapsed", func() {
				BeforeEach(func() {
					version.Time = now.Add(-2 * time.Minute)
				})

				It("outputs both previous and current versions", func() {
					Expect(response).To(HaveLen(2))
					Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
					Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})

			Context("with a longer interval", func() {
				BeforeEach(func() {
					interval := models.Interval(time.Hour)
					source.Interval = &interval
				})

				Context("when within the interval", func() {
					BeforeEach(func() {
						version.Time = now.Add(-30 * time.Minute)
					})

					It("outputs only the supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
					})
				})

				Context("when beyond the interval", func() {
					BeforeEach(func() {
						version.Time = now.Add(-61 * time.Minute)
					})

					It("outputs both versions", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})
		})

		Context("when a cron expression is specified", func() {
			Context("when no version is given", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "*/5 * * * *"}
					source.Cron = &cronExpr
				})

				It("does not output any versions", func() {
					Expect(response).To(BeEmpty())
				})
			})

			Context("when no version is given and initial_version is true", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "*/5 * * * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at the most recent cron boundary with zero seconds and nanoseconds", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Minute() % 5).To(Equal(0))
					Expect(response[0].Time.Second()).To(Equal(0))
					Expect(response[0].Time.Nanosecond()).To(Equal(0))
					Expect(response[0].Time.Unix()).To(BeNumerically("<=", time.Now().Unix()))
				})
			})

			Context("when a version is given", func() {
				Context("and next cron time has passed", func() {
					BeforeEach(func() {
						cronExpr := models.Cron{Expression: "* * * * *"}
						source.Cron = &cronExpr
						version.Time = now.Add(-2 * time.Minute)
					})

					It("outputs both previous version and new version at cron boundary", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
						Expect(response[1].Time.Second()).To(Equal(0))
						Expect(response[1].Time.Nanosecond()).To(Equal(0))
					})
				})

				Context("and next cron time has not passed", func() {
					BeforeEach(func() {
						futureMinute := (now.Minute() + 30) % 60
						cronExpr := models.Cron{Expression: fmt.Sprintf("%d * * * *", futureMinute)}
						source.Cron = &cronExpr
						version.Time = now
					})

					It("outputs only the previous version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
					})
				})
			})

			Context("with L modifier (last day of month)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 L * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the last day of a month", func() {
					Expect(response).To(HaveLen(1))
					nextDay := response[0].Time.AddDate(0, 0, 1)
					Expect(nextDay.Day()).To(Equal(1))
				})
			})

			Context("with # modifier (second Monday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 7 * * 1#2"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the second Monday (day 8-14)", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(Equal(time.Monday))
					Expect(response[0].Time.Day()).To(BeNumerically(">=", 8))
					Expect(response[0].Time.Day()).To(BeNumerically("<=", 14))
				})
			})

			Context("with # modifier (fourth Thursday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 * * 4#4"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the fourth Thursday (day 22-28)", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(Equal(time.Thursday))
					Expect(response[0].Time.Day()).To(BeNumerically(">=", 22))
					Expect(response[0].Time.Day()).To(BeNumerically("<=", 28))
				})
			})

			Context("with W modifier (nearest weekday to 15th)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 15W * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on a weekday within 2 days of the 15th", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).NotTo(Equal(time.Saturday))
					Expect(response[0].Time.Weekday()).NotTo(Equal(time.Sunday))
					Expect(response[0].Time.Day()).To(BeNumerically(">=", 13))
					Expect(response[0].Time.Day()).To(BeNumerically("<=", 17))
				})
			})

			Context("with 1W modifier (nearest weekday to 1st)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 1W * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on a weekday day 1-3 (stays in month)", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).NotTo(Equal(time.Saturday))
					Expect(response[0].Time.Weekday()).NotTo(Equal(time.Sunday))
					Expect(response[0].Time.Day()).To(BeNumerically(">=", 1))
					Expect(response[0].Time.Day()).To(BeNumerically("<=", 3))
				})
			})

			Context("with 5L modifier (last Friday of month)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 17 * * 5L"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on a Friday where next Friday is in a different month", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(Equal(time.Friday))
					nextFriday := response[0].Time.AddDate(0, 0, 7)
					Expect(nextFriday.Month()).NotTo(Equal(response[0].Time.Month()))
				})
			})

			Context("with @yearly macro", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "@yearly"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at January 1st midnight", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Month()).To(Equal(time.January))
					Expect(response[0].Time.Day()).To(Equal(1))
					Expect(response[0].Time.Hour()).To(Equal(0))
					Expect(response[0].Time.Minute()).To(Equal(0))
				})
			})

			Context("with @monthly macro", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "@monthly"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the 1st at midnight", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Day()).To(Equal(1))
					Expect(response[0].Time.Hour()).To(Equal(0))
					Expect(response[0].Time.Minute()).To(Equal(0))
				})
			})

			Context("with @weekly macro", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "@weekly"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on Sunday at midnight", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(Equal(time.Sunday))
					Expect(response[0].Time.Hour()).To(Equal(0))
					Expect(response[0].Time.Minute()).To(Equal(0))
				})
			})

			Context("with specific month cron expression", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 0 1 6 *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version in June", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Month()).To(Equal(time.June))
				})
			})

			Context("with location configured", func() {
				BeforeEach(func() {
					loc, err := time.LoadLocation("America/New_York")
					Expect(err).ToNot(HaveOccurred())
					srcLoc := models.Location(*loc)
					source.Location = &srcLoc

					cronExpr := models.Cron{Expression: "*/5 * * * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at a cron boundary", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Minute() % 5).To(Equal(0))
					Expect(response[0].Time.Second()).To(Equal(0))
				})
			})

			Context("with range in hour field", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9-17 * * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version with hour between 9 and 17", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Hour()).To(BeNumerically(">=", 9))
					Expect(response[0].Time.Hour()).To(BeNumerically("<=", 17))
				})
			})

			Context("with list in day-of-week field", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 * * 1,3,5"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on Monday, Wednesday, or Friday", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(BeElementOf(time.Monday, time.Wednesday, time.Friday))
				})
			})
		})

		Context("when start_after is specified", func() {
			Context("when current time is after start_after", func() {
				BeforeEach(func() {
					startAfter := now.Add(-1 * time.Hour)
					source.StartAfter = (*models.StartAfter)(&startAfter)
				})

				It("outputs a version containing the current time", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
				})
			})

			Context("when current time is before start_after", func() {
				BeforeEach(func() {
					startAfter := now.Add(1 * time.Hour)
					source.StartAfter = (*models.StartAfter)(&startAfter)
				})

				It("does not output any versions", func() {
					Expect(response).To(BeEmpty())
				})
			})

			Context("when current time is before start_after but initial_version is true", func() {
				BeforeEach(func() {
					startAfter := now.Add(1 * time.Hour)
					source.StartAfter = (*models.StartAfter)(&startAfter)
					source.InitialVersion = true
				})

				It("outputs a version", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
				})
			})
		})

		Context("when initial_version is true without other configuration", func() {
			BeforeEach(func() {
				source.InitialVersion = true
			})

			It("outputs a version containing the current time", func() {
				Expect(response).To(HaveLen(1))
				Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
			})
		})

		Context("when multiple configurations are combined", func() {
			Context("with interval and matching day", func() {
				BeforeEach(func() {
					interval := models.Interval(time.Minute)
					source.Interval = &interval
					source.Days = []models.Weekday{models.Weekday(now.Weekday())}
				})

				It("outputs a version", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
				})
			})

			Context("with interval and non-matching day", func() {
				BeforeEach(func() {
					interval := models.Interval(time.Minute)
					source.Interval = &interval
					source.Days = []models.Weekday{models.Weekday(now.AddDate(0, 0, 1).Weekday())}
				})

				It("does not output any versions", func() {
					Expect(response).To(BeEmpty())
				})
			})

			Context("with time range, interval, and matching day", func() {
				BeforeEach(func() {
					start := now.Add(-1 * time.Hour)
					stop := now.Add(1 * time.Hour)
					source.Start = tod(start.Hour(), start.Minute(), 0)
					source.Stop = tod(stop.Hour(), stop.Minute(), 0)

					interval := models.Interval(time.Minute)
					source.Interval = &interval
					source.Days = []models.Weekday{models.Weekday(now.Weekday())}
				})

				It("outputs a version when all conditions match", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
				})
			})
		})
	})

	Context("with validation errors", func() {
		var source models.Source
		var response models.CheckResponse
		var cmdErr error

		JustBeforeEach(func() {
			command := resource.CheckCommand{}
			response, cmdErr = command.Run(models.CheckRequest{
				Source:  source,
				Version: models.Version{},
			})
		})

		Context("when start is provided without stop", func() {
			BeforeEach(func() {
				source = models.Source{}
				start := models.NewTimeOfDay(time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC))
				source.Start = &start
			})

			It("returns a validation error", func() {
				Expect(cmdErr).To(HaveOccurred())
				Expect(response).To(BeNil())
			})
		})

		Context("when stop is provided without start", func() {
			BeforeEach(func() {
				source = models.Source{}
				stop := models.NewTimeOfDay(time.Date(2020, 1, 1, 17, 0, 0, 0, time.UTC))
				source.Stop = &stop
			})

			It("returns a validation error", func() {
				Expect(cmdErr).To(HaveOccurred())
				Expect(response).To(BeNil())
			})
		})
	})
})

var _ = Describe("DescribeCron", func() {
	Context("common macros", func() {
		DescribeTable("describes macros correctly",
			func(expr, expected string) {
				Expect(resource.DescribeCron(expr)).To(Equal(expected))
			},
			Entry("@yearly", "@yearly", "triggers once a year at midnight on January 1st"),
			Entry("@annually", "@annually", "triggers once a year at midnight on January 1st"),
			Entry("@monthly", "@monthly", "triggers at midnight on the 1st of every month"),
			Entry("@weekly", "@weekly", "triggers at midnight every Sunday"),
			Entry("@daily", "@daily", "triggers once a day at midnight"),
			Entry("@midnight", "@midnight", "triggers once a day at midnight"),
			Entry("@hourly", "@hourly", "triggers at the start of every hour"),
		)
	})

	Context("time intervals", func() {
		DescribeTable("describes intervals correctly",
			func(expr, expected string) {
				Expect(resource.DescribeCron(expr)).To(Equal(expected))
			},
			Entry("every 5 minutes", "*/5 * * * *", "triggers every 5 minutes"),
			Entry("every 2 hours at minute 0", "0 */2 * * *", "triggers every 2 hours at minute 0"),
			Entry("every 3 hours at minute 30", "30 */3 * * *", "triggers every 3 hours at minute 30"),
		)
	})

	Context("specific times", func() {
		DescribeTable("describes specific times correctly",
			func(expr, expected string) {
				Expect(resource.DescribeCron(expr)).To(Equal(expected))
			},
			Entry("09:30", "30 9 * * *", "triggers at 09:30"),
			Entry("00:00", "0 0 * * *", "triggers at 00:00"),
			Entry("minute 30 of every hour", "30 * * * *", "triggers at minute 30 of every hour"),
			Entry("during hour 9", "* 9 * * *", "triggers during hour 9"),
		)
	})

	Context("days of week", func() {
		DescribeTable("describes days correctly",
			func(expr, expected string) {
				Expect(resource.DescribeCron(expr)).To(Equal(expected))
			},
			Entry("Monday (1)", "0 9 * * 1", "triggers on Monday, at 09:00"),
			Entry("Sunday (0)", "0 9 * * 0", "triggers on Sunday, at 09:00"),
			Entry("Friday (FRI)", "0 9 * * FRI", "triggers on Friday, at 09:00"),
		)
	})

	Context("months", func() {
		DescribeTable("describes months correctly",
			func(expr, expected string) {
				Expect(resource.DescribeCron(expr)).To(Equal(expected))
			},
			Entry("January", "0 9 * 1 *", "triggers in January, at 09:00"),
			Entry("June", "0 9 * 6 *", "triggers in June, at 09:00"),
			Entry("December", "0 9 * 12 *", "triggers in December, at 09:00"),
		)
	})

	Context("day-of-month intervals", func() {
		It("warns about back-to-back triggers when step lands on 31", func() {
			result := resource.DescribeCron("0 0 */2 * *")
			Expect(result).To(ContainSubstring("every 2 days from 1st of month"))
			Expect(result).To(ContainSubstring("31st then 1st = back-to-back triggers"))
		})

		It("does not warn when step does not land on 31", func() {
			result := resource.DescribeCron("0 0 */7 * *")
			Expect(result).To(ContainSubstring("every 7 days from 1st of month"))
			Expect(result).NotTo(ContainSubstring("back-to-back"))
		})
	})

	Context("warnings", func() {
		It("warns about day 31", func() {
			result := resource.DescribeCron("0 0 31 * *")
			Expect(result).To(ContainSubstring("only triggers in months with 31 days"))
		})

		It("warns about day 30 skipping February", func() {
			result := resource.DescribeCron("0 0 30 * *")
			Expect(result).To(ContainSubstring("skips February"))
		})

		It("warns about day 29 and leap years", func() {
			result := resource.DescribeCron("0 0 29 * *")
			Expect(result).To(ContainSubstring("only triggers in leap years for February"))
		})

		It("warns about DOM + DOW OR logic", func() {
			result := resource.DescribeCron("0 0 15 * 1")
			Expect(result).To(ContainSubstring("OR logic"))
		})

		It("warns about DST for hours 1-3", func() {
			result := resource.DescribeCron("0 2 * * *")
			Expect(result).To(ContainSubstring("DST"))
		})

		It("does not warn about DST for hour 0 or 4+", func() {
			Expect(resource.DescribeCron("0 0 * * *")).NotTo(ContainSubstring("DST"))
			Expect(resource.DescribeCron("0 4 * * *")).NotTo(ContainSubstring("DST"))
		})
	})

	Context("modifiers", func() {
		It("describes L (last day of month)", func() {
			Expect(resource.DescribeCron("0 9 L * *")).To(Equal("triggers on the last day of the month, at 09:00"))
		})

		It("describes W (nearest weekday)", func() {
			Expect(resource.DescribeCron("0 9 15W * *")).To(Equal("triggers on the nearest weekday to the 15th, at 09:00"))
		})

		It("warns about 31W in short months", func() {
			result := resource.DescribeCron("0 9 31W * *")
			Expect(result).To(ContainSubstring("nearest weekday to the 31st"))
			Expect(result).To(ContainSubstring("only triggers in months with 31 days"))
		})

		It("describes 5L (last Friday)", func() {
			Expect(resource.DescribeCron("0 17 * * 5L")).To(Equal("triggers on last Friday of the month, at 17:00"))
		})

		It("describes 1#2 (second Monday)", func() {
			Expect(resource.DescribeCron("0 7 * * 1#2")).To(Equal("triggers on 2nd Monday of the month, at 07:00"))
		})

		It("warns about 5th occurrence", func() {
			result := resource.DescribeCron("0 9 * * 2#5")
			Expect(result).To(ContainSubstring("5th Tuesday of the month"))
			Expect(result).To(ContainSubstring("5th occurrence only exists in some months"))
		})
	})

	Context("edge cases", func() {
		It("returns raw expression for invalid field count", func() {
			Expect(resource.DescribeCron("* * *")).To(Equal("schedule: * * *"))
		})

		It("returns raw expression for all wildcards", func() {
			Expect(resource.DescribeCron("* * * * *")).To(Equal("schedule: * * * * *"))
		})

		It("describes DOW ranges 0-2", func() {
			Expect(resource.DescribeCron("0 9 * * 0-2")).To(Equal("triggers on Sunday through Tuesday, at 09:00"))
		})

		It("describes DOW ranges 1-5", func() {
			Expect(resource.DescribeCron("0 9 * * 1-5")).To(Equal("triggers on Monday through Friday, at 09:00"))
		})
	})
})

func tod(hours, minutes, offset int) *models.TimeOfDay {
	loc := time.UTC
	if offset != 0 {
		loc = time.FixedZone("UnitTest", 60*60*offset)
	}
	now := time.Now()
	tod := models.NewTimeOfDay(time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, loc))
	return &tod
}
