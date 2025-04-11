package resource_test

import (
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
		})

		Context("when a cron expression is specified", func() {
			var (
				cronExpr string
				cronObj  models.Cron
			)

			BeforeEach(func() {
				cronExpr = "*/1 * * * *" // every minute
				cronObj = models.Cron{Expression: cronExpr}
				source.Cron = &cronObj
			})

			Context("when no version is given", func() {
				BeforeEach(func() {
					version = models.Version{}
				})

				It("does not output any versions", func() {
					Expect(response).To(BeEmpty())
				})
			})

			Context("when a version is given and cron condition is met", func() {
				var prev time.Time

				BeforeEach(func() {
					// Set previous time to 1 minute ago to ensure the cron expression triggers
					prev = now.Add(-1 * time.Minute)
					version.Time = prev
				})

				It("outputs both previous version and new version", func() {
					Expect(response).To(HaveLen(2))
					Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
					Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})

			Context("when a version is given but cron condition is not yet met", func() {
				BeforeEach(func() {
					// Set expression to run only at the 30th minute of each hour
					cronExpr = "30 * * * *"
					cronObj = models.Cron{Expression: cronExpr}
					source.Cron = &cronObj

					// Set previous time to now to ensure the cron expression doesn't trigger
					// (unless by chance we're exactly at the 30th minute during the test)
					prev := now
					version.Time = prev

					// Ensure we're not at the 30th minute
					if now.Minute() == 30 {
						Skip("Skipping test as current time would cause cron expression to trigger")
					}
				})

				It("outputs only the previous version", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
				})
			})

			Context("with specific days of the week", func() {
				BeforeEach(func() {
					// Set cron to run only on Mondays at midnight
					cronExpr = "0 0 * * 1"
					cronObj = models.Cron{Expression: cronExpr}
					source.Cron = &cronObj

					// Set previous time to be over a day ago
					prev := now.Add(-25 * time.Hour)
					version.Time = prev

					// Skip test if today is not Monday
					if now.Weekday() != time.Monday {
						Skip("Skipping test as today is not Monday")
					}
				})

				It("triggers if current day matches cron pattern", func() {
					// Only expect a new version if it's Monday and near midnight
					if now.Hour() == 0 && now.Minute() < 5 {
						Expect(response).To(HaveLen(2))
					} else {
						Expect(response).To(HaveLen(1))
					}
				})
			})

			Context("with a location configured", func() {
				var loc *time.Location

				BeforeEach(func() {
					var err error
					loc, err = time.LoadLocation("America/New_York")
					Expect(err).ToNot(HaveOccurred())

					srcLoc := models.Location(*loc)
					source.Location = &srcLoc

					// Set cron to run every hour at the 0th minute
					cronExpr = "0 * * * *"
					cronObj = models.Cron{Expression: cronExpr}
					source.Cron = &cronObj

					// Set previous version to 61 minutes ago to ensure we're past the interval
					prev := now.Add(-61 * time.Minute)
					version.Time = prev

					// Skip if we're not at the 0th minute
					if now.In(loc).Minute() != 0 {
						Skip("Skipping test as current time in location would not trigger cron")
					}
				})

				It("evaluates the cron expression in the specified timezone", func() {
					Expect(response).To(HaveLen(2))
					Expect(response[0].Time.Unix()).To(Equal(version.Time.Unix()))
					Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
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
