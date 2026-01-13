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
	var (
		now time.Time
	)

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
				var prev time.Time

				Context("when the resource has already triggered on the current day", func() {
					BeforeEach(func() {
						prev = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, now.Second(), now.Nanosecond(), now.Location())
						version.Time = prev
					})

					It("outputs a supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", prev.Unix(), 1))
					})
				})

				Context("when the resource was triggered yesterday", func() {
					BeforeEach(func() {
						prev = now.Add(-24 * time.Hour)
						version.Time = prev
					})

					It("outputs a version containing the current time and supplied version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", prev.Unix(), 1))
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

				Context("when a version is given", func() {
					var prev time.Time

					Context("when the resource has already triggered with in the current time range", func() {
						BeforeEach(func() {
							prev = now.Add(-30 * time.Minute)
							version.Time = prev
						})

						It("outputs a supplied version", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", prev.Unix(), 1))
						})
					})

					Context("when the resource was triggered yesterday near the end of the time frame", func() {
						BeforeEach(func() {
							prev = now.Add(-23 * time.Hour)
							version.Time = prev
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", prev.Unix(), 1))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the resource was triggered last year near the end of the time frame", func() {
						BeforeEach(func() {
							prev = now.AddDate(-1, 0, 0)
							version.Time = prev
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", prev.Unix(), 1))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the resource was triggered yesterday in the current time frame", func() {
						BeforeEach(func() {
							prev = now.Add(-24 * time.Hour)
							version.Time = prev
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", prev.Unix(), 1))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
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

					Context("when a version is given", func() {
						var prev time.Time

						Context("when the interval has not elapsed", func() {
							BeforeEach(func() {
								prev = now
								version.Time = prev
							})

							It("outputs only the supplied version", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
							})
						})

						Context("when the interval has elapsed", func() {
							BeforeEach(func() {
								prev = now.Add(-1 * time.Minute)
								version.Time = prev
							})

							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("with its time N intervals ago", func() {
							BeforeEach(func() {
								prev = now.Add(-5 * time.Minute)
								version.Time = prev
							})

							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})
					})
				})

				Context("when the current day is specified", func() {
					BeforeEach(func() {
						source.Days = []models.Weekday{
							models.Weekday(now.Weekday()),
							models.Weekday(now.AddDate(0, 0, 2).Weekday()),
						}
					})

					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when we are out of the specified day", func() {
					BeforeEach(func() {
						source.Days = []models.Weekday{
							models.Weekday(now.AddDate(0, 0, 1).Weekday()),
							models.Weekday(now.AddDate(0, 0, 2).Weekday()),
						}
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

				Context("when no version is given", func() {
					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})

				Context("when an interval is given", func() {
					BeforeEach(func() {
						start := now.Add(6 * time.Hour)
						stop := now.Add(7 * time.Hour)

						source.Start = tod(start.Hour(), start.Minute(), 0)
						source.Stop = tod(stop.Hour(), stop.Minute(), 0)

						interval := models.Interval(time.Minute)
						source.Interval = &interval
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})

			Context("with a location configured", func() {
				var loc *time.Location

				BeforeEach(func() {
					var err error
					loc, err = time.LoadLocation("America/Indiana/Indianapolis")
					Expect(err).ToNot(HaveOccurred())

					srcLoc := models.Location(*loc)
					source.Location = &srcLoc

					now = now.In(loc)
				})

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

					Context("when a version is given", func() {
						var prev time.Time

						Context("when the resource has already triggered with in the current time range", func() {
							BeforeEach(func() {
								prev = now.Add(-30 * time.Minute)
								version.Time = prev
							})

							It("outputs a supplied version", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
							})
						})

						Context("when the resource was triggered yesterday near the end of the time frame", func() {
							BeforeEach(func() {
								prev = now.Add(-23 * time.Hour)
								version.Time = prev
							})

							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("when the resource was triggered yesterday in the current time frame", func() {
							BeforeEach(func() {
								prev = now.AddDate(0, 0, -1)
								version.Time = prev
							})

							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
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

						Context("when a version is given", func() {
							var prev time.Time

							Context("with its time within the interval", func() {
								BeforeEach(func() {
									prev = now
									version.Time = prev
								})

								It("outputs the given version", func() {
									Expect(response).To(HaveLen(1))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								})
							})

							Context("with its time one interval ago", func() {
								BeforeEach(func() {
									prev = now.Add(-1 * time.Minute)
									version.Time = prev
								})

								It("outputs a version containing the current time and supplied version", func() {
									Expect(response).To(HaveLen(2))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
									Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
								})
							})

							Context("with its time N intervals ago", func() {
								BeforeEach(func() {
									prev = now.Add(-5 * time.Minute)
									version.Time = prev
								})

								It("outputs a version containing the current time and supplied version", func() {
									Expect(response).To(HaveLen(2))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
									Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
								})
							})
						})
					})

					Context("when the current day is specified", func() {
						BeforeEach(func() {
							source.Days = []models.Weekday{
								models.Weekday(now.Weekday()),
								models.Weekday(now.AddDate(0, 0, 2).Weekday()),
							}
						})

						It("outputs a version containing the current time", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when we are out of the specified day", func() {
						BeforeEach(func() {
							source.Days = []models.Weekday{
								models.Weekday(now.AddDate(0, 0, 1).Weekday()),
								models.Weekday(now.AddDate(0, 0, 2).Weekday()),
							}
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

					Context("when no version is given", func() {
						It("does not output any versions", func() {
							Expect(response).To(BeEmpty())
						})
					})

					Context("when an interval is given", func() {
						BeforeEach(func() {
							start := now.Add(6 * time.Hour)
							stop := now.Add(7 * time.Hour)

							source.Start = tod(start.Hour(), start.Minute(), 0)
							source.Stop = tod(stop.Hour(), stop.Minute(), 0)

							interval := models.Interval(time.Minute)
							source.Interval = &interval
						})

						It("does not output any versions", func() {
							Expect(response).To(BeEmpty())
						})
					})
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

			Context("when a version is given", func() {
				var prev time.Time

				Context("with its time within the interval", func() {
					BeforeEach(func() {
						prev = now
						version.Time = prev
					})

					It("outputs a supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
					})
				})

				Context("with its time one interval ago", func() {
					BeforeEach(func() {
						prev = now.Add(-1 * time.Minute)
						version.Time = prev
					})

					It("outputs a version containing the current time and supplied version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("with its time N intervals ago", func() {
					BeforeEach(func() {
						prev = now.Add(-5 * time.Minute)
						version.Time = prev
					})

					It("outputs a version containing the current time and supplied version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})

			Context("with a longer interval", func() {
				BeforeEach(func() {
					interval := models.Interval(time.Hour)
					source.Interval = &interval
				})

				Context("when a version is given within the hour interval", func() {
					BeforeEach(func() {
						version.Time = now.Add(-30 * time.Minute)
					})

					It("outputs only the supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
					})
				})

				Context("when a version is given beyond the hour interval", func() {
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

				It("outputs a version at the most recent cron boundary", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Minute() % 5).To(Equal(0))
					Expect(response[0].Time.Second()).To(Equal(0))
					Expect(response[0].Time.Unix()).To(BeNumerically("<=", time.Now().Unix()))
				})
			})

			Context("when a version is given", func() {
				var prev time.Time

				Context("and next cron time has passed", func() {
					BeforeEach(func() {
						cronExpr := models.Cron{Expression: "* * * * *"}
						source.Cron = &cronExpr
						prev = now.Add(-2 * time.Minute)
						version.Time = prev
					})

					It("outputs both previous version and new version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
						Expect(response[1].Time.Second()).To(Equal(0))
					})
				})

				Context("and next cron time has not passed", func() {
					BeforeEach(func() {
						futureMinute := (now.Minute() + 30) % 60
						cronExpr := models.Cron{Expression: fmt.Sprintf("%d * * * *", futureMinute)}
						source.Cron = &cronExpr
						prev = now
						version.Time = prev
					})

					It("outputs only the previous version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
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
					versionTime := response[0].Time
					nextDay := versionTime.AddDate(0, 0, 1)
					Expect(nextDay.Day()).To(Equal(1))
				})
			})

			Context("with # modifier (nth weekday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 7 * * 1#2"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the second Monday of a month", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).To(Equal(time.Monday))
					Expect(versionTime.Day()).To(BeNumerically(">=", 8))
					Expect(versionTime.Day()).To(BeNumerically("<=", 14))
				})
			})

			Context("with W modifier (nearest weekday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 15W * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on a weekday near the 15th", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).NotTo(Equal(time.Saturday))
					Expect(versionTime.Weekday()).NotTo(Equal(time.Sunday))
					Expect(versionTime.Day()).To(BeNumerically(">=", 13))
					Expect(versionTime.Day()).To(BeNumerically("<=", 17))
				})
			})

			Context("with 5L modifier (last Friday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 17 * * 5L"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the last Friday of a month", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).To(Equal(time.Friday))
					nextFriday := versionTime.AddDate(0, 0, 7)
					Expect(nextFriday.Month()).NotTo(Equal(versionTime.Month()))
				})
			})

			Context("with 0L modifier (last Sunday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 10 * * 0L"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the last Sunday of a month", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).To(Equal(time.Sunday))
					nextSunday := versionTime.AddDate(0, 0, 7)
					Expect(nextSunday.Month()).NotTo(Equal(versionTime.Month()))
				})
			})

			Context("with 1#1 modifier (first Monday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 8 * * 1#1"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the first Monday of a month", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).To(Equal(time.Monday))
					Expect(versionTime.Day()).To(BeNumerically(">=", 1))
					Expect(versionTime.Day()).To(BeNumerically("<=", 7))
				})
			})

			Context("with 0#3 modifier (third Sunday)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 * * 0#3"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on the third Sunday of a month", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).To(Equal(time.Sunday))
					Expect(versionTime.Day()).To(BeNumerically(">=", 15))
					Expect(versionTime.Day()).To(BeNumerically("<=", 21))
				})
			})

			Context("with 1W modifier (nearest weekday to 1st)", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 1W * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on a weekday near the 1st", func() {
					Expect(response).To(HaveLen(1))
					versionTime := response[0].Time
					Expect(versionTime.Weekday()).NotTo(Equal(time.Saturday))
					Expect(versionTime.Weekday()).NotTo(Equal(time.Sunday))
					Expect(versionTime.Day()).To(BeNumerically(">=", 1))
					Expect(versionTime.Day()).To(BeNumerically("<=", 3))
				})
			})

			Context("with hourly cron expression", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 * * * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at the hour boundary", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Minute()).To(Equal(0))
					Expect(response[0].Time.Second()).To(Equal(0))
				})
			})

			Context("with daily cron expression", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 0 * * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at midnight", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Hour()).To(Equal(0))
					Expect(response[0].Time.Minute()).To(Equal(0))
				})
			})

			Context("with specific day-of-week cron expression", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 * * 1"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version on a Monday", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(Equal(time.Monday))
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

				It("outputs a version on the 1st of a month at midnight", func() {
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

				It("outputs a version on a Sunday at midnight", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Weekday()).To(Equal(time.Sunday))
					Expect(response[0].Time.Hour()).To(Equal(0))
					Expect(response[0].Time.Minute()).To(Equal(0))
				})
			})

			Context("with @daily macro", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "@daily"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at midnight", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Hour()).To(Equal(0))
					Expect(response[0].Time.Minute()).To(Equal(0))
				})
			})

			Context("with @hourly macro", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "@hourly"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("outputs a version at minute 0", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Minute()).To(Equal(0))
					Expect(response[0].Time.Second()).To(Equal(0))
				})
			})

			Context("with location configured", func() {
				var loc *time.Location

				BeforeEach(func() {
					var err error
					loc, err = time.LoadLocation("America/New_York")
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

			Context("cron version time is at boundary not check time", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "*/10 * * * *"}
					source.Cron = &cronExpr
					version.Time = now.Add(-15 * time.Minute)
				})

				It("outputs new version at 10-minute boundary", func() {
					Expect(response).To(HaveLen(2))
					Expect(response[1].Time.Minute() % 10).To(Equal(0))
					Expect(response[1].Time.Second()).To(Equal(0))
					Expect(response[1].Time.Nanosecond()).To(Equal(0))
				})
			})

			Context("with range in hour field", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9-17 * * *"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("accepts the range expression", func() {
					Expect(response).To(HaveLen(1))
					hour := response[0].Time.Hour()
					Expect(hour).To(BeNumerically(">=", 9))
					Expect(hour).To(BeNumerically("<=", 17))
				})
			})

			Context("with list in day-of-week field", func() {
				BeforeEach(func() {
					cronExpr := models.Cron{Expression: "0 9 * * 1,3,5"}
					source.Cron = &cronExpr
					source.InitialVersion = true
				})

				It("accepts the list expression", func() {
					Expect(response).To(HaveLen(1))
					weekday := response[0].Time.Weekday()
					Expect(weekday).To(BeElementOf(time.Monday, time.Wednesday, time.Friday))
				})
			})
		})

		Context("when start_after is specified", func() {
			Context("when no version is provided", func() {
				Context("and the current time is after start_after", func() {
					BeforeEach(func() {
						startAfter := now.Add(-1 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
					})

					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
					})
				})

				Context("and the current time is before start_after", func() {
					BeforeEach(func() {
						startAfter := now.Add(1 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})

			Context("when a version is provided", func() {
				var previousTime time.Time

				Context("when the current time is after start_after and a previous version exists", func() {
					BeforeEach(func() {
						previousTime = now.Add(-24 * time.Hour)
						version.Time = previousTime
						startAfter := now.Add(-25 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
					})

					It("outputs both the previous and current versions", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(previousTime.Unix()))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})

			Context("when initial_version is specified", func() {
				Context("when the current time is before start_after and initial_version is true", func() {
					BeforeEach(func() {
						startAfter := now.Add(1 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
						source.InitialVersion = true
					})

					It("outputs a single version containing the initial version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
					})
				})

				Context("when the current time is after start_after and initial_version is true", func() {
					BeforeEach(func() {
						startAfter := now.Add(-1 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
						source.InitialVersion = true
					})

					It("outputs a single version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
					})
				})

				Context("when the current time is before start_after and initial_version is false", func() {
					BeforeEach(func() {
						startAfter := now.Add(1 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
						source.InitialVersion = false
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})

				Context("when the current time is after start_after and initial_version is false", func() {
					BeforeEach(func() {
						startAfter := now.Add(-1 * time.Hour)
						source.StartAfter = (*models.StartAfter)(&startAfter)
						source.InitialVersion = false
					})

					It("outputs a single version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
					})
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
			Context("with interval and days", func() {
				BeforeEach(func() {
					interval := models.Interval(time.Minute)
					source.Interval = &interval
					source.Days = []models.Weekday{
						models.Weekday(now.Weekday()),
					}
				})

				It("outputs a version when on the correct day", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
				})
			})

			Context("with interval and days not matching", func() {
				BeforeEach(func() {
					interval := models.Interval(time.Minute)
					source.Interval = &interval
					source.Days = []models.Weekday{
						models.Weekday(now.AddDate(0, 0, 1).Weekday()),
					}
				})

				It("does not output any versions", func() {
					Expect(response).To(BeEmpty())
				})
			})

			Context("with time range and interval and days", func() {
				BeforeEach(func() {
					start := now.Add(-1 * time.Hour)
					stop := now.Add(1 * time.Hour)
					source.Start = tod(start.Hour(), start.Minute(), 0)
					source.Stop = tod(stop.Hour(), stop.Minute(), 0)

					interval := models.Interval(time.Minute)
					source.Interval = &interval

					source.Days = []models.Weekday{
						models.Weekday(now.Weekday()),
					}
				})

				It("outputs a version when all conditions match", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", now.Unix(), 1))
				})
			})

			Context("with location and interval", func() {
				BeforeEach(func() {
					loc, err := time.LoadLocation("Europe/London")
					Expect(err).ToNot(HaveOccurred())

					srcLoc := models.Location(*loc)
					source.Location = &srcLoc

					interval := models.Interval(time.Minute)
					source.Interval = &interval
				})

				It("outputs a version", func() {
					Expect(response).To(HaveLen(1))
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
		It("describes @yearly", func() {
			Expect(resource.DescribeCron("@yearly")).To(Equal("triggers once a year at midnight on January 1st"))
		})

		It("describes @annually", func() {
			Expect(resource.DescribeCron("@annually")).To(Equal("triggers once a year at midnight on January 1st"))
		})

		It("describes @monthly", func() {
			Expect(resource.DescribeCron("@monthly")).To(Equal("triggers at midnight on the 1st of every month"))
		})

		It("describes @weekly", func() {
			Expect(resource.DescribeCron("@weekly")).To(Equal("triggers at midnight every Sunday"))
		})

		It("describes @daily", func() {
			Expect(resource.DescribeCron("@daily")).To(Equal("triggers once a day at midnight"))
		})

		It("describes @midnight", func() {
			Expect(resource.DescribeCron("@midnight")).To(Equal("triggers once a day at midnight"))
		})

		It("describes @hourly", func() {
			Expect(resource.DescribeCron("@hourly")).To(Equal("triggers at the start of every hour"))
		})
	})

	Context("minute intervals", func() {
		It("describes every 5 minutes", func() {
			Expect(resource.DescribeCron("*/5 * * * *")).To(Equal("triggers every 5 minutes"))
		})

		It("describes every 15 minutes", func() {
			Expect(resource.DescribeCron("*/15 * * * *")).To(Equal("triggers every 15 minutes"))
		})

		It("describes every 1 minute", func() {
			Expect(resource.DescribeCron("*/1 * * * *")).To(Equal("triggers every 1 minutes"))
		})
	})

	Context("hour intervals", func() {
		It("describes every 2 hours at minute 0", func() {
			Expect(resource.DescribeCron("0 */2 * * *")).To(Equal("triggers every 2 hours at minute 0"))
		})

		It("describes every 3 hours at minute 30", func() {
			Expect(resource.DescribeCron("30 */3 * * *")).To(Equal("triggers every 3 hours at minute 30"))
		})

		It("describes every 6 hours at minute 15", func() {
			Expect(resource.DescribeCron("15 */6 * * *")).To(Equal("triggers every 6 hours at minute 15"))
		})
	})

	Context("day-of-month intervals", func() {
		It("describes every 2 days with back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */2 * *")).To(Equal("triggers every 2 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers"))
		})

		It("describes every 3 days with back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */3 * *")).To(Equal("triggers every 3 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers"))
		})

		It("describes every 5 days with back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */5 * *")).To(Equal("triggers every 5 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers"))
		})

		It("describes every 7 days without back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */7 * *")).To(Equal("triggers every 7 days from 1st of month, at 00:00"))
		})

		It("describes every 6 days with back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */6 * *")).To(Equal("triggers every 6 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers"))
		})

		It("describes every 10 days with back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */10 * *")).To(Equal("triggers every 10 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers"))
		})

		It("describes every 15 days with back-to-back warning", func() {
			Expect(resource.DescribeCron("0 0 */15 * *")).To(Equal("triggers every 15 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers"))
		})
	})

	Context("specific times", func() {
		It("describes 09:30", func() {
			Expect(resource.DescribeCron("30 9 * * *")).To(Equal("triggers at 09:30"))
		})

		It("describes 00:00", func() {
			Expect(resource.DescribeCron("0 0 * * *")).To(Equal("triggers at 00:00"))
		})

		It("describes 23:59", func() {
			Expect(resource.DescribeCron("59 23 * * *")).To(Equal("triggers at 23:59"))
		})

		It("describes 12:00", func() {
			Expect(resource.DescribeCron("0 12 * * *")).To(Equal("triggers at 12:00"))
		})
	})

	Context("specific days of week", func() {
		It("describes Monday", func() {
			Expect(resource.DescribeCron("0 9 * * 1")).To(Equal("triggers on Monday, at 09:00"))
		})

		It("describes Sunday with 0", func() {
			Expect(resource.DescribeCron("0 9 * * 0")).To(Equal("triggers on Sunday, at 09:00"))
		})

		It("describes Wednesday", func() {
			Expect(resource.DescribeCron("0 9 * * 3")).To(Equal("triggers on Wednesday, at 09:00"))
		})

		It("describes Saturday", func() {
			Expect(resource.DescribeCron("0 9 * * 6")).To(Equal("triggers on Saturday, at 09:00"))
		})

		It("describes Monday with MON", func() {
			Expect(resource.DescribeCron("0 9 * * MON")).To(Equal("triggers on Monday, at 09:00"))
		})

		It("describes Friday with FRI", func() {
			Expect(resource.DescribeCron("0 9 * * FRI")).To(Equal("triggers on Friday, at 09:00"))
		})

		It("describes Sunday with SUN", func() {
			Expect(resource.DescribeCron("0 9 * * SUN")).To(Equal("triggers on Sunday, at 09:00"))
		})
	})

	Context("specific days of month", func() {
		It("describes day 15", func() {
			Expect(resource.DescribeCron("0 9 15 * *")).To(Equal("triggers on day 15 of the month, at 09:00"))
		})

		It("describes day 1", func() {
			Expect(resource.DescribeCron("0 9 1 * *")).To(Equal("triggers on day 1 of the month, at 09:00"))
		})
	})

	Context("specific months", func() {
		It("describes January", func() {
			Expect(resource.DescribeCron("0 9 * 1 *")).To(Equal("triggers in January, at 09:00"))
		})

		It("describes June", func() {
			Expect(resource.DescribeCron("0 9 * 6 *")).To(Equal("triggers in June, at 09:00"))
		})

		It("describes December", func() {
			Expect(resource.DescribeCron("0 9 * 12 *")).To(Equal("triggers in December, at 09:00"))
		})

		It("handles unknown month gracefully", func() {
			Expect(resource.DescribeCron("0 9 * 13 *")).To(Equal("triggers in month 13, at 09:00"))
		})
	})

	Context("warnings", func() {
		Context("day-of-month OR day-of-week logic", func() {
			It("warns when both DOM and DOW are specified", func() {
				result := resource.DescribeCron("0 0 15 * 1")
				Expect(result).To(ContainSubstring("note: day-of-month AND day-of-week uses OR logic, not AND"))
			})
		})

		Context("short month warnings", func() {
			It("warns about day 31", func() {
				result := resource.DescribeCron("0 0 31 * *")
				Expect(result).To(ContainSubstring("note: only triggers in months with 31 days"))
			})

			It("warns about day 30 skipping February", func() {
				result := resource.DescribeCron("0 0 30 * *")
				Expect(result).To(ContainSubstring("note: skips February"))
			})

			It("warns about day 29 and leap years", func() {
				result := resource.DescribeCron("0 0 29 * *")
				Expect(result).To(ContainSubstring("note: only triggers in leap years for February"))
			})
		})

		Context("DST warnings", func() {
			It("warns about hour 1", func() {
				result := resource.DescribeCron("0 1 * * *")
				Expect(result).To(ContainSubstring("note: may skip or double-trigger during DST transitions"))
			})

			It("warns about hour 2", func() {
				result := resource.DescribeCron("0 2 * * *")
				Expect(result).To(ContainSubstring("note: may skip or double-trigger during DST transitions"))
			})

			It("warns about hour 3", func() {
				result := resource.DescribeCron("0 3 * * *")
				Expect(result).To(ContainSubstring("note: may skip or double-trigger during DST transitions"))
			})

			It("does not warn about hour 4", func() {
				result := resource.DescribeCron("0 4 * * *")
				Expect(result).NotTo(ContainSubstring("DST"))
			})

			It("does not warn about hour 0", func() {
				result := resource.DescribeCron("0 0 * * *")
				Expect(result).NotTo(ContainSubstring("DST"))
			})

			It("does not warn about hour steps", func() {
				result := resource.DescribeCron("0 */2 * * *")
				Expect(result).NotTo(ContainSubstring("DST"))
			})

			It("does not warn about hour lists", func() {
				result := resource.DescribeCron("0 1,2,3 * * *")
				Expect(result).NotTo(ContainSubstring("DST"))
			})
		})
	})

	Context("complex expressions", func() {
		It("describes day 15 of month with Monday", func() {
			result := resource.DescribeCron("0 0 15 * 1")
			Expect(result).To(ContainSubstring("on Monday"))
			Expect(result).To(ContainSubstring("on day 15 of the month"))
			Expect(result).To(ContainSubstring("OR logic"))
		})

		It("describes specific time in January", func() {
			result := resource.DescribeCron("30 14 * 1 *")
			Expect(result).To(Equal("triggers in January, at 14:30"))
		})

		It("describes Monday in March at noon", func() {
			result := resource.DescribeCron("0 12 * 3 1")
			Expect(result).To(ContainSubstring("on Monday"))
			Expect(result).To(ContainSubstring("in March"))
			Expect(result).To(ContainSubstring("at 12:00"))
		})
	})

	Context("edge cases", func() {
		It("handles wildcard hour with specific minute", func() {
			Expect(resource.DescribeCron("30 * * * *")).To(Equal("triggers at minute 30 of every hour"))
		})

		It("handles specific hour with wildcard minute", func() {
			Expect(resource.DescribeCron("* 9 * * *")).To(Equal("triggers during hour 9"))
		})

		It("returns raw expression for invalid field count", func() {
			Expect(resource.DescribeCron("* * *")).To(Equal("schedule: * * *"))
		})

		It("returns raw expression for too many fields", func() {
			Expect(resource.DescribeCron("* * * * * *")).To(Equal("schedule: * * * * * *"))
		})

		It("handles question mark in DOM", func() {
			Expect(resource.DescribeCron("0 9 ? * 1")).To(Equal("triggers on Monday, at 09:00"))
		})

		It("handles question mark in DOW", func() {
			Expect(resource.DescribeCron("0 9 15 * ?")).To(Equal("triggers on day 15 of the month, at 09:00"))
		})

		It("handles all wildcards", func() {
			Expect(resource.DescribeCron("* * * * *")).To(Equal("schedule: * * * * *"))
		})

		It("preserves unknown DOW values", func() {
			Expect(resource.DescribeCron("0 9 * * 1-5")).To(Equal("triggers on 1-5, at 09:00"))
		})
	})

	Context("describeTime edge cases", func() {
		It("handles minute step with non-zero minute", func() {
			Expect(resource.DescribeCron("*/30 * * * *")).To(Equal("triggers every 30 minutes"))
		})

		It("handles hour step with zero minute", func() {
			Expect(resource.DescribeCron("0 */4 * * *")).To(Equal("triggers every 4 hours at minute 0"))
		})
	})

	Context("cron modifiers", func() {
		Context("L modifier in day-of-month", func() {
			It("describes L as last day of month", func() {
				Expect(resource.DescribeCron("0 9 L * *")).To(Equal("triggers on the last day of the month, at 09:00"))
			})
		})

		Context("W modifier in day-of-month", func() {
			It("describes 15W as nearest weekday to 15th", func() {
				Expect(resource.DescribeCron("0 9 15W * *")).To(Equal("triggers on the nearest weekday to the 15th, at 09:00"))
			})

			It("describes 1W as nearest weekday to 1st", func() {
				Expect(resource.DescribeCron("0 9 1W * *")).To(Equal("triggers on the nearest weekday to the 1st, at 09:00"))
			})

			It("describes 2W as nearest weekday to 2nd", func() {
				Expect(resource.DescribeCron("0 9 2W * *")).To(Equal("triggers on the nearest weekday to the 2nd, at 09:00"))
			})

			It("describes 3W as nearest weekday to 3rd", func() {
				Expect(resource.DescribeCron("0 9 3W * *")).To(Equal("triggers on the nearest weekday to the 3rd, at 09:00"))
			})

			It("describes 4W as nearest weekday to 4th", func() {
				Expect(resource.DescribeCron("0 9 4W * *")).To(Equal("triggers on the nearest weekday to the 4th, at 09:00"))
			})

			It("warns about 31W in short months", func() {
				result := resource.DescribeCron("0 9 31W * *")
				Expect(result).To(ContainSubstring("on the nearest weekday to the 31st"))
				Expect(result).To(ContainSubstring("note: only triggers in months with 31 days"))
			})

			It("warns about 30W skipping February", func() {
				result := resource.DescribeCron("0 9 30W * *")
				Expect(result).To(ContainSubstring("on the nearest weekday to the 30th"))
				Expect(result).To(ContainSubstring("note: skips February"))
			})

			It("warns about 29W and leap years", func() {
				result := resource.DescribeCron("0 9 29W * *")
				Expect(result).To(ContainSubstring("on the nearest weekday to the 29th"))
				Expect(result).To(ContainSubstring("note: only triggers in leap years for February"))
			})
		})

		Context("L modifier in day-of-week", func() {
			It("describes 5L as last Friday", func() {
				Expect(resource.DescribeCron("0 17 * * 5L")).To(Equal("triggers on last Friday of the month, at 17:00"))
			})

			It("describes 0L as last Sunday", func() {
				Expect(resource.DescribeCron("0 10 * * 0L")).To(Equal("triggers on last Sunday of the month, at 10:00"))
			})

			It("describes 1L as last Monday", func() {
				Expect(resource.DescribeCron("0 9 * * 1L")).To(Equal("triggers on last Monday of the month, at 09:00"))
			})
		})

		Context("# modifier in day-of-week", func() {
			It("describes 1#2 as second Monday", func() {
				Expect(resource.DescribeCron("0 7 * * 1#2")).To(Equal("triggers on 2nd Monday of the month, at 07:00"))
			})

			It("describes 0#1 as first Sunday", func() {
				Expect(resource.DescribeCron("0 9 * * 0#1")).To(Equal("triggers on 1st Sunday of the month, at 09:00"))
			})

			It("describes 5#3 as third Friday", func() {
				Expect(resource.DescribeCron("0 9 * * 5#3")).To(Equal("triggers on 3rd Friday of the month, at 09:00"))
			})

			It("describes 0#3 as third Sunday ", func() {
				Expect(resource.DescribeCron("0 9 * * 0#3")).To(Equal("triggers on 3rd Sunday of the month, at 09:00"))
			})

			It("describes 4#4 as fourth Thursday", func() {
				Expect(resource.DescribeCron("0 9 * * 4#4")).To(Equal("triggers on 4th Thursday of the month, at 09:00"))
			})

			It("warns about 5th occurrence", func() {
				result := resource.DescribeCron("0 9 * * 2#5")
				Expect(result).To(ContainSubstring("5th Tuesday of the month"))
				Expect(result).To(ContainSubstring("note: 5th occurrence only exists in some months"))
			})

			It("warns about 5th Sunday", func() {
				result := resource.DescribeCron("0 9 * * 0#5")
				Expect(result).To(ContainSubstring("5th Sunday of the month"))
				Expect(result).To(ContainSubstring("note: 5th occurrence only exists in some months"))
			})
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
