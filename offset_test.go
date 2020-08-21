package resource_test

import (
	"os"
	"time"

	resource "github.com/concourse/time-resource"
	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testEnvironment struct {
	teamName       string
	pipelineName   string
	hashPercentile float64
}

var xsmallOffset = testEnvironment{
	teamName:       "",
	pipelineName:   "",
	hashPercentile: 0.16425462769,
}

var smallOffset = testEnvironment{
	teamName:       "concourse",
	pipelineName:   "time-resource",
	hashPercentile: 0.49905898922,
}

var largeOffset = testEnvironment{
	teamName:       "concourse",
	pipelineName:   "concourse",
	hashPercentile: 0.6995269071,
}

var xlargeOffset = testEnvironment{
	teamName:       "foo",
	pipelineName:   "bar",
	hashPercentile: 0.77989363246,
}

var _ = Describe("Offset", func() {
	originalTeam := os.Getenv(resource.BUILD_TEAM_NAME)
	originalPipeline := os.Getenv(resource.BUILD_PIPELINE_NAME)

	var (
		env testEnvironment
		now time.Time
		loc *time.Location

		tl          lord.TimeLord
		dayDuration time.Duration
		reference   time.Time

		actualOffsetTime   time.Time
		expectedOffsetTime time.Time
	)

	BeforeEach(func() {
		env = testEnvironment{}
		now = time.Now()
		actualOffsetTime = time.Time{}
		expectedOffsetTime = time.Time{}
	})

	JustBeforeEach(func() {
		os.Setenv(resource.BUILD_TEAM_NAME, env.teamName)
		os.Setenv(resource.BUILD_PIPELINE_NAME, env.pipelineName)
		actualOffsetTime = resource.Offset(tl, reference)
	})

	AfterEach(func() {
		os.Setenv(resource.BUILD_TEAM_NAME, originalTeam)
		os.Setenv(resource.BUILD_PIPELINE_NAME, originalPipeline)
	})

	validateExpectedTime := func() {
		Expect(actualOffsetTime.Unix()).To(Equal(expectedOffsetTime.Unix()))
	}

	RunIntervalAndOrRangeTests := func() {
		var rangeDuration time.Duration

		Context("when a range is not specified", func() {
			BeforeEach(func() {
				rangeDuration = dayDuration

				tl.Start = nil
				tl.Stop = nil
			})

			Context("when an interval is not specified", func() {
				BeforeEach(func() {
					tl.Interval = nil
				})

				JustBeforeEach(func() {
					expectedOffsetTime = time.Date(reference.Year(), reference.Month(), reference.Day(), 0, 0, 0, 0, loc).Add(time.Duration(rangeDuration.Minutes()*env.hashPercentile) * time.Minute)
				})

				Context("when using a team name and pipeline name that generates a very small offset", func() {
					BeforeEach(func() { env = xsmallOffset })
					It("generates an overnight version", validateExpectedTime)
				})

				Context("when using a team name and pipeline name that generates a small offset", func() {
					BeforeEach(func() { env = smallOffset })
					It("generates an morning version", validateExpectedTime)
				})

				Context("when using a team name and pipeline name that generates a large offset", func() {
					BeforeEach(func() { env = largeOffset })
					It("generates an afternoon version", validateExpectedTime)
				})

				Context("when using a team name and pipeline name that generates a very large offset", func() {
					BeforeEach(func() { env = xlargeOffset })
					It("generates an evening version", validateExpectedTime)
				})
			})

			Context("when an interval is specified", func() {
				var intervalDuration time.Duration

				BeforeEach(func() {
					tl.Interval = new(models.Interval)
				})

				Context("when the interval can't get any smaller", func() {
					BeforeEach(func() {
						intervalDuration = time.Minute
						*tl.Interval = models.Interval(intervalDuration)
					})

					for _, testEnv := range []testEnvironment{xsmallOffset, smallOffset, largeOffset, xlargeOffset} {
						Context("for the "+env.teamName+"/"+env.pipelineName+" pipeline", func() {
							BeforeEach(func() {
								env = testEnv
								expectedOffsetTime = reference.Truncate(time.Minute)
							})
							It("returns the reference time", validateExpectedTime)
						})
					}
				})

				Context("when the interval is smaller than the range", func() {
					BeforeEach(func() {
						intervalDuration = time.Hour
						*tl.Interval = models.Interval(intervalDuration)
					})

					JustBeforeEach(func() {
						expectedOffsetTime = reference.Truncate(intervalDuration).Add(time.Duration(intervalDuration.Minutes()*env.hashPercentile) * time.Minute)
					})

					Context("when using a team name and pipeline name that generates a very small offset", func() {
						BeforeEach(func() { env = xsmallOffset })
						It("generates a version at the very beginning of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a small offset", func() {
						BeforeEach(func() { env = smallOffset })
						It("generates a version toward the beginning of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a large offset", func() {
						BeforeEach(func() { env = largeOffset })
						It("generates a version toward the end of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a very large offset", func() {
						BeforeEach(func() { env = xlargeOffset })
						It("generates a version at the very end of the interval", validateExpectedTime)
					})
				})

				Context("when the interval is larger than the range", func() {
					BeforeEach(func() {
						intervalDuration = time.Hour * 168
						*tl.Interval = models.Interval(intervalDuration)
					})

					JustBeforeEach(func() {
						expectedOffsetTime = time.Date(reference.Year(), reference.Month(), reference.Day(), 0, 0, 0, 0, loc).Add(time.Duration(rangeDuration.Minutes()*env.hashPercentile) * time.Minute)
					})

					Context("when using a team name and pipeline name that generates a very small offset", func() {
						BeforeEach(func() { env = xsmallOffset })
						It("generates a version at the very beginning of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a small offset", func() {
						BeforeEach(func() { env = smallOffset })
						It("generates a version toward the beginning of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a large offset", func() {
						BeforeEach(func() { env = largeOffset })
						It("generates a version toward the end of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a very large offset", func() {
						BeforeEach(func() { env = xlargeOffset })
						It("generates a version at the very end of the interval", validateExpectedTime)
					})
				})
			})
		})

		Context("when a range is specified", func() {
			BeforeEach(func() {
				rangeDuration = 6 * time.Hour

				tlStart := reference.Truncate(rangeDuration)
				tl.Start = new(models.TimeOfDay)
				*tl.Start = models.NewTimeOfDay(tlStart)

				tlStop := tlStart.Add(rangeDuration)
				tl.Stop = new(models.TimeOfDay)
				*tl.Stop = models.NewTimeOfDay(tlStop)
			})

			Context("when an interval is not specified", func() {
				BeforeEach(func() {
					tl.Interval = nil
				})

				JustBeforeEach(func() {
					expectedOffsetTime = reference.Truncate(rangeDuration).Add(time.Duration(rangeDuration.Minutes()*env.hashPercentile) * time.Minute)
				})

				Context("when using a team name and pipeline name that generates a very small offset", func() {
					BeforeEach(func() { env = xsmallOffset })
					It("generates a version at the very beginning of the range", validateExpectedTime)
				})

				Context("when using a team name and pipeline name that generates a small offset", func() {
					BeforeEach(func() { env = smallOffset })
					It("generates a version toward the beginning of the range", validateExpectedTime)
				})

				Context("when using a team name and pipeline name that generates a large offset", func() {
					BeforeEach(func() { env = largeOffset })
					It("generates a version toward the end of the range", validateExpectedTime)
				})

				Context("when using a team name and pipeline name that generates a very large offset", func() {
					BeforeEach(func() { env = xlargeOffset })
					It("generates a version at the very end of the range", validateExpectedTime)
				})
			})

			Context("when an interval is specified", func() {
				var intervalDuration time.Duration

				BeforeEach(func() {
					tl.Interval = new(models.Interval)
				})

				Context("when the interval can't get any smaller", func() {
					BeforeEach(func() {
						intervalDuration = time.Minute
						*tl.Interval = models.Interval(intervalDuration)
					})

					for _, testEnv := range []testEnvironment{xsmallOffset, smallOffset, largeOffset, xlargeOffset} {
						Context("for the "+env.teamName+"/"+env.pipelineName+" pipeline", func() {
							BeforeEach(func() {
								env = testEnv
								expectedOffsetTime = reference.Truncate(time.Minute)
							})
							It("returns the reference time", validateExpectedTime)
						})
					}
				})

				Context("when the interval is smaller than the range", func() {
					BeforeEach(func() {
						intervalDuration = time.Hour
						*tl.Interval = models.Interval(intervalDuration)
					})

					JustBeforeEach(func() {
						expectedOffsetTime = reference.Truncate(intervalDuration).Add(time.Duration(intervalDuration.Minutes()*env.hashPercentile) * time.Minute)
					})

					Context("when using a team name and pipeline name that generates a very small offset", func() {
						BeforeEach(func() { env = xsmallOffset })
						It("generates a version at the very beginning of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a small offset", func() {
						BeforeEach(func() { env = smallOffset })
						It("generates a version toward the beginning of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a large offset", func() {
						BeforeEach(func() { env = largeOffset })
						It("generates a version toward the end of the interval", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a very large offset", func() {
						BeforeEach(func() { env = xlargeOffset })
						It("generates a version at the very end of the interval", validateExpectedTime)
					})
				})

				Context("when the interval is larger than the range", func() {
					BeforeEach(func() {
						intervalDuration = time.Hour * 168
						*tl.Interval = models.Interval(intervalDuration)
					})

					JustBeforeEach(func() {
						expectedOffsetTime = reference.Truncate(rangeDuration).Add(time.Duration(rangeDuration.Minutes()*env.hashPercentile) * time.Minute)
					})

					Context("when using a team name and pipeline name that generates a very small offset", func() {
						BeforeEach(func() { env = xsmallOffset })
						It("generates a version at the very beginning of the range", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a small offset", func() {
						BeforeEach(func() { env = smallOffset })
						It("generates a version toward the beginning of the range", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a large offset", func() {
						BeforeEach(func() { env = largeOffset })
						It("generates a version toward the end of the range", validateExpectedTime)
					})

					Context("when using a team name and pipeline name that generates a very large offset", func() {
						BeforeEach(func() { env = xlargeOffset })
						It("generates a version at the very end of the range", validateExpectedTime)
					})
				})
			})
		})
	}

	Context("when a location is not specified", func() {
		BeforeEach(func() {
			loc = time.UTC
			tl.Location = nil
			reference = time.Date(2012, time.April, 21, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), loc)
			dayDuration = 24 * time.Hour
		})

		RunIntervalAndOrRangeTests()
	})

	Context("when a location is specified", func() {
		BeforeEach(func() {
			var err error

			loc, err = time.LoadLocation("America/New_York")
			Expect(err).NotTo(HaveOccurred())

			tlLoc := models.Location(*loc)
			tl.Location = &tlLoc
		})

		Context("when referencing a 23-hour day", func() {
			BeforeEach(func() {
				reference = time.Date(2012, time.March, 11, 6, now.Minute(), now.Second(), now.Nanosecond(), loc)
				dayDuration = 23 * time.Hour
			})

			RunIntervalAndOrRangeTests()
		})

		Context("when referencing a 25-hour day", func() {
			BeforeEach(func() {
				reference = time.Date(2012, time.November, 4, 13, now.Minute(), now.Second(), now.Nanosecond(), loc)
				dayDuration = 25 * time.Hour
			})

			RunIntervalAndOrRangeTests()
		})
	})
})
