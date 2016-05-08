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

	Describe("ParseTime", func() {
		Context("when numeric time zone offset is set", func() {
			var formats []string

			BeforeEach(func() {
				formats = []string{
					"1:00 PM -0800",
					"1PM -0800",
					"1 PM -0800",
					"13:00 -0800",
					"1300 -0800",
				}
			})

			It("can parse all format combinations", func() {
				expectedTime := time.Date(0, 1, 1, 21, 0, 0, 0, time.UTC)

				loc, _ := time.LoadLocation("UTC") // note: IANA TZ loc conflicts with offset
				for _, format := range formats {
					By("working with time " + format)
					parsedTime, err := ParseTime(format, loc)

					Expect(err).NotTo(HaveOccurred())
					Expect(parsedTime).To(Equal(expectedTime))
				}
			})
		})

		Context("when numeric time zone offset is not set", func() {
			var formats []string

			BeforeEach(func() {
				formats = []string{
					"1:00 PM",
					"1PM",
					"1 PM",
					"13:00",
					"1300",
				}
			})

			It("can parse all format combinations", func() {
				expectedTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

				loc, _ := time.LoadLocation("UTC")
				for _, format := range formats {
					By("working with time " + format)
					parsedTime, err := ParseTime(format, loc)
					Expect(err).NotTo(HaveOccurred())
					Expect(parsedTime).To(Equal(expectedTime))
				}
			})
		})
	})

	Describe("ParseWeekday", func() {
		It("can parse a weekday", func() {
			parsedWeekdays, err := ParseWeekdays([]string{"Monday", "Tuesday"})

			Expect(err).NotTo(HaveOccurred())
			Expect(parsedWeekdays).To(Equal([]time.Weekday{time.Monday, time.Tuesday}))
		})

		It("raise error if weekday can't be parsed", func() {
			_, err := ParseWeekdays([]string{"Foo", "Tuesday"})

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("IsInDays", func() {
		It("returns true if current day is in dayslist", func() {
			daysList := []time.Weekday{
				now.Weekday(),
				now.Add(24 * time.Hour).Weekday(),
			}

			Expect(IsInDays(now, daysList)).To(BeTrue())
		})

		It("return true if list is empty", func() {
			Expect(IsInDays(now, nil)).To(BeTrue())
		})

		It("returns false if not in list", func() {
			daysList := []time.Weekday{
				now.Add(24 * time.Hour).Weekday(),
				now.Add(48 * time.Hour).Weekday(),
			}

			Expect(IsInDays(now, daysList)).To(BeFalse())
		})
	})

	BeforeEach(func() {
		checkCmd = exec.Command(checkPath)
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
				Eventually(session.Err).Should(gbytes.Say("one of 'interval' or 'between' must be specified"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with an invalid start", func() {
			BeforeEach(func() {
				request.Source.Start = "not-a-time"
				request.Source.Stop = "3:04 PM -0700"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("invalid start time"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with an invalid stop", func() {
			BeforeEach(func() {
				request.Source.Start = "3:04 PM -0700"
				request.Source.Stop = "not-a-time"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("invalid stop time"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with a missing stop", func() {
			BeforeEach(func() {
				request.Source.Start = "3:04 PM -0700"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("empty stop time!"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with a missing start", func() {
			BeforeEach(func() {
				request.Source.Stop = "3:04 PM -0700"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("empty start time!"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with an invalid interval ", func() {
			BeforeEach(func() {
				request.Source.Interval = "not-an-interval"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("invalid interval"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with an invalid day ", func() {
			BeforeEach(func() {
				request.Source.Days = []string{"Foo", "Bar"}
				request.Source.Interval = "1m"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("invalid day 'Foo'"))
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
					timeLayout := "3:04 PM -0700"

					request.Source.Start = start.Format(timeLayout)
					request.Source.Stop = stop.Format(timeLayout)
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
							request.Source.Interval = "1m"
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
						request.Source.Days = []string{
							now.Add(24 * time.Hour).Weekday().String(),
							now.Add(48 * time.Hour).Weekday().String(),
						}
						request.Source.Days = []string{
							now.Weekday().String(),
							now.Add(48 * time.Hour).Weekday().String()}
					})

					It("outputs a version containing the current time", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when we are out of the specified day", func() {
					BeforeEach(func() {
						request.Source.Days = []string{
							now.Add(24 * time.Hour).Weekday().String(),
							now.Add(48 * time.Hour).Weekday().String(),
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
					timeLayout := "3:04 PM -0700"

					request.Source.Start = start.Format(timeLayout)
					request.Source.Stop = stop.Format(timeLayout)
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
						timeLayout := "3:04 PM -0700"

						request.Source.Start = start.Format(timeLayout)
						request.Source.Stop = stop.Format(timeLayout)

						request.Source.Interval = "1m"
					})

					It("does not output any versions", func() {
						Expect(response).To(BeEmpty())
					})
				})
			})
		})

		Context("when an interval is specified", func() {
			BeforeEach(func() {
				request.Source.Interval = "1m"
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
