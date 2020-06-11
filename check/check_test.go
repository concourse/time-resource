package main_test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/concourse/time-resource/models"
)

var _ = Describe("Check", func() {
	var (
		checkCmd *exec.Cmd
		now      time.Time
	)

	BeforeEach(func() {
		now = time.Now().UTC()
	})

	BeforeEach(func() {
		checkCmd = exec.Command(checkPath)
	})

	Context("when executed", func() {
		var source map[string]interface{}
		var version *models.Version
		var response models.CheckResponse

		BeforeEach(func() {
			source = map[string]interface{}{}
			version = nil
			response = models.CheckResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := checkCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(map[string]interface{}{
				"source":  source,
				"version": version,
			})
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a time range is specified", func() {
			Context("when we are in the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(-1 * time.Hour)
					stop := now.Add(1 * time.Hour)

					source["start"] = tod(start.Hour(), start.Minute(), 0)
					source["stop"] = tod(stop.Hour(), stop.Minute(), 0)
				})

				Context("when no version is given", func() {
					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when the range is given in another timezone", func() {
					BeforeEach(func() {
						loc := time.FixedZone("LaLaLand", -(60 * 60 * 4))

						start := now.Add(-1 * time.Hour).In(loc)
						stop := now.Add(1 * time.Hour).In(loc)

						source["start"] = tod(start.Hour(), start.Minute(), -4)
						source["stop"] = tod(stop.Hour(), stop.Minute(), -4)
					})

					Context("when no version is given", func() {
						It("outputs a version containing the current time", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})
				})

				Context("when a version is given", func() {
					var prev time.Time

					Context("when the resource has already triggered with in the current time range", func() {
						BeforeEach(func() {
							prev = now.Add(-30 * time.Minute)
							version = &models.Version{Time: prev}
						})

						It("outputs a supplied version", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
						})
					})

					Context("when the resource was triggered yesterday at the end of the time frame", func() {
						BeforeEach(func() {
							prev = now.Add(-23 * time.Hour)
							version = &models.Version{Time: prev}
						})

						It("outputs a version containing the current time", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the resource was triggered last year near the end of the time frame", func() {
						BeforeEach(func() {
							prev = now.AddDate(-1, 0, 0)
							version = &models.Version{Time: prev}
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the resource was triggered yesterday in the current time frame", func() {
						BeforeEach(func() {
							prev = now.Add(-24 * time.Hour)
							version = &models.Version{Time: prev}
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("from a predictable implementation", func() {
						Context("when the resource has already triggered with in the current time range", func() {
							BeforeEach(func() {
								prev = now.Add(-30 * time.Minute).Truncate(time.Minute)
								version = &models.Version{Time: prev}
							})

							It("outputs a supplied version", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
							})
						})

						Context("when the resource was triggered yesterday at the end of the time frame", func() {
							BeforeEach(func() {
								prev = now.Add(-23 * time.Hour).Truncate(time.Minute)
								version = &models.Version{Time: prev}
							})
							It("outputs a version containing the current time", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("when the resource was triggered last year near the end of the time frame", func() {
							BeforeEach(func() {
								prev = now.AddDate(-1, 0, 0).Truncate(time.Minute)
								version = &models.Version{Time: prev}
							})
							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("when the resource was triggered yesterday in the current time frame", func() {
							BeforeEach(func() {
								prev = now.AddDate(0, 0, -1).Truncate(time.Minute)
								version = &models.Version{Time: prev}
							})

							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})
					})
				})

				Context("when an interval is specified", func() {
					var currentInterval time.Time

					BeforeEach(func() {
						source["interval"] = "1m"
						currentInterval = now.Truncate(time.Minute)
					})

					Context("when no version is given", func() {
						It("outputs a version containing the current interval", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(Equal(currentInterval.Unix()))
						})
					})

					Context("when a version is given", func() {
						var prev time.Time

						Context("when the interval has not elapsed", func() {
							BeforeEach(func() {
								prev = now
								version = &models.Version{Time: prev}
							})

							It("outputs no versions", func() {
								Expect(response).To(HaveLen(0))
							})
						})

						Context("when the interval has elapsed", func() {
							BeforeEach(func() {
								prev = now.Add(-1 * time.Minute)
								version = &models.Version{Time: prev}
							})

							It("outputs a version containing the current interval", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(Equal(currentInterval.Unix()))
							})
						})

						Context("with its time N intervals ago", func() {
							N_INTERVALS := 5

							BeforeEach(func() {
								prev = now.Add(-1 * time.Duration(N_INTERVALS) * time.Minute)
								version = &models.Version{Time: prev}
							})

							It("outputs N new versions (including the current interval) but not the supplied version", func() {
								Expect(response).To(HaveLen(N_INTERVALS))
								Expect(response[0].Time.Unix()).To(Not(Equal(prev.Unix())))
								Expect(response[N_INTERVALS-1].Time.Unix()).To(Equal(currentInterval.Unix()))
							})
						})

						Context("from a predictable implementation", func() {
							Context("when the interval has not elapsed", func() {
								BeforeEach(func() {
									prev = currentInterval
									version = &models.Version{Time: prev}
								})

								It("outputs a supplied version", func() {
									Expect(response).To(HaveLen(1))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								})
							})

							Context("when the interval has elapsed", func() {
								BeforeEach(func() {
									prev = currentInterval.Add(-1 * time.Minute)
									version = &models.Version{Time: prev}
								})

								It("outputs a version containing the current interval and supplied version", func() {
									Expect(response).To(HaveLen(2))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
									Expect(response[1].Time.Unix()).To(Equal(currentInterval.Unix()))
								})
							})

							Context("with its time N intervals ago", func() {
								N_INTERVALS := 5

								BeforeEach(func() {
									prev = currentInterval.Add(-1 * time.Duration(N_INTERVALS) * time.Minute)
									version = &models.Version{Time: prev}
								})

								It("outputs N new versions (including the current interval) and supplied version", func() {
									Expect(response).To(HaveLen(N_INTERVALS + 1))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
									Expect(response[N_INTERVALS].Time.Unix()).To(Equal(currentInterval.Unix()))
								})
							})
						})
					})
				})

				Context("when the current day is specified", func() {
					BeforeEach(func() {
						source["days"] = []string{
							now.Weekday().String(),
							now.AddDate(0, 0, 2).Weekday().String(),
						}
					})

					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when we are out of the specified day", func() {
					BeforeEach(func() {
						source["days"] = []string{
							now.AddDate(0, 0, 1).Weekday().String(),
							now.AddDate(0, 0, 2).Weekday().String(),
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

					source["start"] = tod(start.Hour(), start.Minute(), 0)
					source["stop"] = tod(stop.Hour(), stop.Minute(), 0)
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

						source["start"] = tod(start.Hour(), start.Minute(), 0)
						source["stop"] = tod(stop.Hour(), stop.Minute(), 0)
						source["interval"] = "1m"
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

					source["location"] = loc.String()

					now = now.In(loc)
				})

				Context("when we are in the specified time range", func() {
					BeforeEach(func() {
						start := now.Add(-1 * time.Hour)
						stop := now.Add(1 * time.Hour)

						source["start"] = tod(start.Hour(), start.Minute(), 0)
						source["stop"] = tod(stop.Hour(), stop.Minute(), 0)
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
								version = &models.Version{Time: prev}
							})

							It("outputs a supplied version", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
							})
						})

						Context("when the resource was triggered yesterday at the end of the time frame", func() {
							BeforeEach(func() {
								prev = now.Add(-23 * time.Hour)
								version = &models.Version{Time: prev}
							})

							It("outputs a version containing the current time", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("when the resource was triggered yesterday in the current time frame", func() {
							BeforeEach(func() {
								prev = now.AddDate(0, 0, -1)
								version = &models.Version{Time: prev}
							})

							It("outputs a version containing the current time and supplied version", func() {
								Expect(response).To(HaveLen(2))
								Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("from a predictable implementation", func() {
							BeforeEach(func() {
								now = now.Truncate(time.Minute)
							})

							Context("when the resource has already triggered with in the current time range", func() {
								BeforeEach(func() {
									prev = now.Add(-30 * time.Minute)
									version = &models.Version{Time: prev}
								})

								It("outputs a supplied version", func() {
									Expect(response).To(HaveLen(1))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
								})
							})

							Context("when the resource was triggered yesterday at the end of the time frame", func() {
								BeforeEach(func() {
									prev = now.Add(-23 * time.Hour)
									version = &models.Version{Time: prev}
								})

								It("outputs a version containing the current time", func() {
									Expect(response).To(HaveLen(1))
									Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
								})
							})

							Context("when the resource was triggered yesterday in the current time frame", func() {
								BeforeEach(func() {
									prev = now.AddDate(0, 0, -1)
									version = &models.Version{Time: prev}
								})

								It("outputs a version containing the current time and supplied version", func() {
									Expect(response).To(HaveLen(2))
									Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
									Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
								})
							})
						})
					})

					Context("when an interval is specified", func() {
						var currentInterval time.Time

						BeforeEach(func() {
							source["interval"] = "1m"
							currentInterval = now.Truncate(time.Minute)
						})

						Context("when no version is given", func() {
							It("outputs a version containing the current interval", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(Equal(currentInterval.Unix()))
							})
						})

						Context("when a version is given", func() {
							var prev time.Time

							Context("with its time within the interval", func() {
								BeforeEach(func() {
									prev = now
									version = &models.Version{Time: prev}
								})

								It("output no versions", func() {
									Expect(response).To(HaveLen(0))
								})
							})

							Context("with its time one interval ago", func() {
								BeforeEach(func() {
									prev = now.Add(-1 * time.Minute)
									version = &models.Version{Time: prev}
								})

								It("outputs a version containing the current interval", func() {
									Expect(response).To(HaveLen(1))
									Expect(response[0].Time.Unix()).To(Equal(currentInterval.Unix()))
								})
							})

							Context("with its time N intervals ago", func() {
								N_INTERVALS := 5

								BeforeEach(func() {
									prev = now.Add(-1 * time.Duration(N_INTERVALS) * time.Minute)
									version = &models.Version{Time: prev}
								})

								It("outputs N new versions (including the current interval) but not the supplied version", func() {
									Expect(response).To(HaveLen(N_INTERVALS))
									Expect(response[0].Time.Unix()).To(Not(Equal(prev.Unix())))
									Expect(response[N_INTERVALS-1].Time.Unix()).To(Equal(currentInterval.Unix()))
								})
							})

							Context("from a predictable implementation", func() {
								Context("with its time within the interval", func() {
									BeforeEach(func() {
										prev = currentInterval
										version = &models.Version{Time: prev}
									})

									It("outputs the given version", func() {
										Expect(response).To(HaveLen(1))
										Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
									})
								})

								Context("with its time one interval ago", func() {
									BeforeEach(func() {
										prev = currentInterval.Add(-1 * time.Minute)
										version = &models.Version{Time: prev}
									})

									It("outputs a version containing the current time and supplied version", func() {
										Expect(response).To(HaveLen(2))
										Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
										Expect(response[1].Time.Unix()).To(Equal(currentInterval.Unix()))
									})
								})

								Context("with its time N intervals ago", func() {
									N_INTERVALS := 5

									BeforeEach(func() {
										prev = currentInterval.Add(-1 * time.Duration(N_INTERVALS) * time.Minute)
										version = &models.Version{Time: prev}
									})

									It("outputs N new versions (including the current interval) and supplied version", func() {
										Expect(response).To(HaveLen(N_INTERVALS + 1))
										Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
										Expect(response[N_INTERVALS].Time.Unix()).To(Equal(currentInterval.Unix()))
									})
								})
							})
						})
					})

					Context("when the current day is specified", func() {
						BeforeEach(func() {
							source["days"] = []string{
								now.Weekday().String(),
								now.AddDate(0, 0, 2).Weekday().String(),
							}
						})

						It("outputs a version containing the current time", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when we are out of the specified day", func() {
						BeforeEach(func() {
							source["days"] = []string{
								now.AddDate(0, 0, 1).Weekday().String(),
								now.AddDate(0, 0, 2).Weekday().String(),
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

						source["start"] = tod(start.Hour(), start.Minute(), 0)
						source["stop"] = tod(stop.Hour(), stop.Minute(), 0)
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

							source["start"] = tod(start.Hour(), start.Minute(), 0)
							source["stop"] = tod(stop.Hour(), stop.Minute(), 0)

							source["interval"] = "1m"
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
				source["interval"] = "1m"
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
						version = &models.Version{Time: prev}
					})

					It("outputs a supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
					})
				})

				Context("with its time one interval ago", func() {
					BeforeEach(func() {
						prev = now.Add(-1 * time.Minute)
						version = &models.Version{Time: prev}
					})

					It("outputs a version containing the current time and supplied version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("with its time N intervals ago", func() {
					N_INTERVALS := 5

					BeforeEach(func() {
						prev = now.Add(-time.Duration(N_INTERVALS) * time.Minute)
						version = &models.Version{Time: prev}
					})

					It("outputs N new versions (including the current time) and supplied version", func() {
						Expect(response).To(HaveLen(N_INTERVALS + 1))
						Expect(response[0].Time.Unix()).To(Equal(prev.Unix()))
						Expect(response[N_INTERVALS].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})
		})
	})

	Context("with invalid inputs", func() {
		var source map[string]interface{}
		var version map[string]string
		var session *gexec.Session

		BeforeEach(func() {
			source = map[string]interface{}{}
			version = nil
		})

		JustBeforeEach(func() {
			stdin, err := checkCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err = gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(map[string]interface{}{
				"source":  source,
				"version": version,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("with a missing everything", func() {
			It("returns an error", func() {
				<-session.Exited

				Expect(session.Err).To(gbytes.Say("must configure either 'interval' or 'start' and 'stop'"))
				Expect(session.ExitCode()).To(Equal(1))
			})
		})

		Context("with a missing stop", func() {
			BeforeEach(func() {
				source["start"] = tod(3, 4, -7)
			})

			It("returns an error", func() {
				<-session.Exited

				Expect(session.Err).To(gbytes.Say("must configure 'stop' if 'start' is set"))
				Expect(session.ExitCode()).To(Equal(1))
			})
		})

		Context("with a missing start", func() {
			BeforeEach(func() {
				source["stop"] = tod(3, 4, -7)
			})

			It("returns an error", func() {
				<-session.Exited

				Expect(session.Err).To(gbytes.Say("must configure 'start' if 'stop' is set"))
				Expect(session.ExitCode()).To(Equal(1))
			})
		})
	})
})

func tod(hours, minutes, offset int) string {
	var o string
	if offset < 0 {
		o = fmt.Sprintf(" -%02d00", -offset)
	} else if offset > 0 {
		o = fmt.Sprintf(" +%02d00", offset)
	}

	return fmt.Sprintf("%d:%02d%s", hours, minutes, o)
}
