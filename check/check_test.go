package main_test

import (
	"encoding/json"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/concourse/time-resource/check"
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

	Describe("IsInDays", func() {
		It("returns true if current day is in dayslist", func() {
			daysList := []models.Weekday{
				models.Weekday(now.Weekday()),
				models.Weekday(now.Add(24 * time.Hour).Weekday()),
			}

			Expect(IsInDays(now, daysList)).To(BeTrue())
		})

		It("return true if list is empty", func() {
			Expect(IsInDays(now, nil)).To(BeTrue())
		})

		It("returns false if not in list", func() {
			daysList := []models.Weekday{
				models.Weekday(now.Add(24 * time.Hour).Weekday()),
				models.Weekday(now.Add(48 * time.Hour).Weekday()),
			}

			Expect(IsInDays(now, daysList)).To(BeFalse())
		})
	})

	Context("with invalid inputs", func() {
		var request models.CheckRequest
		var response models.CheckResponse
		var session *gexec.Session

		BeforeEach(func() {
			request = models.CheckRequest{}
			response = models.CheckResponse{}
		})

		JustBeforeEach(func() {
			var err error

			stdin, err := checkCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err = gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("with a missing everything", func() {
			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("must configure either 'interval' or 'start' and 'stop'"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with a missing stop", func() {
			BeforeEach(func() {
				request.Source.Start = tod(3, 4, -7)
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("must configure 'stop' if 'start' is set"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with a missing start", func() {
			BeforeEach(func() {
				request.Source.Stop = tod(3, 4, -7)
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("must configure 'start' if 'stop' is set"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})

	Context("when executed", func() {
		var request models.CheckRequest
		var response models.CheckResponse

		BeforeEach(func() {
			request = models.CheckRequest{}
			response = models.CheckResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := checkCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a time range is specified", func() {
			Context("when we are in the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(-1 * time.Hour)
					stop := now.Add(1 * time.Hour)

					request.Source.Start = tod(start.Hour(), start.Minute(), 0)
					request.Source.Stop = tod(stop.Hour(), stop.Minute(), 0)
				})

				Context("when no version is given", func() {
					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when a version is given", func() {
					Context("when the resource has already triggered with in the current time range", func() {
						BeforeEach(func() {
							request.Version.Time = now.Add(-30 * time.Minute)
						})

						It("outputs a supplied version", func() {
							Expect(response).To(HaveLen(1))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
						})
					})

					Context("when the resource was triggered yesterday near the end of the time frame", func() {
						BeforeEach(func() {
							request.Version.Time = now.Add(-23 * time.Hour)
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the resource was triggered yesterday in the current time frame", func() {
						BeforeEach(func() {
							request.Version.Time = now.Add(-24 * time.Hour)
						})

						It("outputs a version containing the current time and supplied version", func() {
							Expect(response).To(HaveLen(2))
							Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
							Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when an interval is specified", func() {
						BeforeEach(func() {
							request.Source.Interval = i(time.Minute)
						})

						Context("when no version is given", func() {
							It("outputs a version containing the current time", func() {
								Expect(response).To(HaveLen(1))
								Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("when a version is given", func() {
							Context("with its time within the interval", func() {
								BeforeEach(func() {
									request.Version.Time = time.Now()
								})

								It("outputs a supplied version", func() {
									Expect(response).To(HaveLen(1))
									Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
								})
							})

							Context("with its time one interval ago", func() {
								BeforeEach(func() {
									request.Version.Time = time.Now().Add(-1 * time.Minute)
								})

								It("outputs a version containing the current time and supplied version", func() {
									Expect(response).To(HaveLen(2))
									Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
									Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
								})
							})

							Context("with its time N intervals ago", func() {
								BeforeEach(func() {
									request.Version.Time = time.Now().Add(-5 * time.Minute)
								})

								It("outputs a version containing the current time and supplied version", func() {
									Expect(response).To(HaveLen(2))
									Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
									Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
								})
							})
						})
					})
				})

				Context("when the current day is specified", func() {
					BeforeEach(func() {
						request.Source.Days = []models.Weekday{
							models.Weekday(now.Weekday()),
							models.Weekday(now.Add(48 * time.Hour).Weekday()),
						}
					})

					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when we are out of the specified day", func() {
					BeforeEach(func() {
						request.Source.Days = []models.Weekday{
							models.Weekday(now.Add(24 * time.Hour).Weekday()),
							models.Weekday(now.Add(48 * time.Hour).Weekday()),
						}
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})

			Context("when we out of the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(6 * time.Hour)
					stop := now.Add(7 * time.Hour)

					request.Source.Start = tod(start.Hour(), start.Minute(), 0)
					request.Source.Stop = tod(stop.Hour(), stop.Minute(), 0)
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

						request.Source.Start = tod(start.Hour(), start.Minute(), 0)
						request.Source.Stop = tod(stop.Hour(), stop.Minute(), 0)

						request.Source.Interval = i(time.Minute)
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})
		})

		Context("when an interval is specified", func() {
			BeforeEach(func() {
				request.Source.Interval = i(time.Minute)
			})

			Context("when no version is given", func() {
				It("outputs a version containing the current time", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})

			Context("when a version is given", func() {
				Context("with its time within the interval", func() {
					BeforeEach(func() {
						request.Version.Time = time.Now()
					})

					It("outputs a supplied version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
					})
				})

				Context("with its time one interval ago", func() {
					BeforeEach(func() {
						request.Version.Time = time.Now().Add(-1 * time.Minute)
					})

					It("outputs a version containing the current time and supplied version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("with its time N intervals ago", func() {
					BeforeEach(func() {
						request.Version.Time = time.Now().Add(-5 * time.Minute)
					})

					It("outputs a version containing the current time and supplied version", func() {
						Expect(response).To(HaveLen(2))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", request.Version.Time.Unix(), 1))
						Expect(response[1].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})
		})
	})
})

func tod(hours, minutes, offset int) *models.TimeOfDay {
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	d += time.Duration(offset) * time.Hour
	if d < 0 {
		d += 12 * time.Hour
	}

	return (*models.TimeOfDay)(&d)
}

func i(d time.Duration) *models.Interval {
	return (*models.Interval)(&d)
}
