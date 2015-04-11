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
	var checkCmd *exec.Cmd

	Describe("ParseTime", func() {
		It("can parse many formats", func() {
			expectedTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

			formats := []string{
				"1:00 PM UTC",
				"1PM UTC",
				"1 PM UTC",
				"13:00 UTC",
				"1300 UTC",
			}

			for _, format := range formats {
				By("working with " + format)
				parsedTime, err := ParseTime(format)

				Ω(err).ShouldNot(HaveOccurred())
				Ω(parsedTime.Equal(expectedTime)).Should(BeTrue())
			}
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
			Ω(err).ShouldNot(HaveOccurred())

			session, err = gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())
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
				request.Source.Stop = "3:04 PM MST"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("invalid start time"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with an invalid stop", func() {
			BeforeEach(func() {
				request.Source.Start = "3:04 PM MST"
				request.Source.Stop = "not-a-time"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("invalid stop time"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with a missing stop", func() {
			BeforeEach(func() {
				request.Source.Start = "3:04 PM MST"
			})

			It("returns an error", func() {
				Eventually(session.Err).Should(gbytes.Say("empty stop time!"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("with a missing start", func() {
			BeforeEach(func() {
				request.Source.Stop = "3:04 PM MST"
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
			Ω(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when a time range is specified", func() {
			now := time.Now()

			Context("when we are in the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(-6 * time.Hour)
					stop := now.Add(6 * time.Hour)
					timeLayout := "3:04 PM MST"

					request.Source.Start = start.Format(timeLayout)
					request.Source.Stop = stop.Format(timeLayout)
				})

				Context("when no version is given", func() {
					It("outputs a version containing the current time", func() {
						Ω(response).Should(HaveLen(1))
						Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("when a version is given", func() {
					Context("when the resource has already triggered with in the current time range", func() {
						BeforeEach(func() {
							request.Version.Time = now.Add(-6 * time.Hour)
						})

						It("does not output any versions", func() {
							Ω(response).Should(BeEmpty())
						})
					})

					Context("when the resource was triggered yesterday near the end of the time frame", func() {
						BeforeEach(func() {
							request.Version.Time = now.Add(-18 * time.Hour)
						})

						It("outputs a version containing the current time", func() {
							Ω(response).Should(HaveLen(1))
							Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when the resource was triggered yesterday in the current time frame", func() {
						BeforeEach(func() {
							request.Version.Time = now.Add(-24 * time.Hour)
						})

						It("outputs a version containing the current time", func() {
							Ω(response).Should(HaveLen(1))
							Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
						})
					})

					Context("when an interval is specified", func() {
						BeforeEach(func() {
							request.Source.Interval = "1m"
						})

						Context("when no version is given", func() {
							It("outputs a version containing the current time", func() {
								Ω(response).Should(HaveLen(1))
								Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
							})
						})

						Context("when a version is given", func() {
							Context("with its time within the interval", func() {
								BeforeEach(func() {
									request.Version.Time = time.Now()
								})

								It("does not output any versions", func() {
									Ω(response).Should(BeEmpty())
								})
							})

							Context("with its time one interval ago", func() {
								BeforeEach(func() {
									request.Version.Time = time.Now().Add(-1 * time.Minute)
								})

								It("outputs a version containing the current time", func() {
									Ω(response).Should(HaveLen(1))
									Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
								})
							})

							Context("with its time N intervals ago", func() {
								BeforeEach(func() {
									request.Version.Time = time.Now().Add(-5 * time.Minute)
								})

								It("outputs a version containing the current time", func() {
									Ω(response).Should(HaveLen(1))
									Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
								})
							})
						})
					})
				})
			})

			Context("when we out of the specified time range", func() {
				BeforeEach(func() {
					start := now.Add(6 * time.Hour)
					stop := now.Add(7 * time.Hour)
					timeLayout := "3:04 PM MST"

					request.Source.Start = start.Format(timeLayout)
					request.Source.Stop = stop.Format(timeLayout)
				})

				Context("when no version is given", func() {
					It("does not output any versions", func() {
						Ω(response).Should(BeEmpty())
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
					Ω(response).Should(HaveLen(1))
					Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
				})
			})

			Context("when a version is given", func() {
				Context("with its time within the interval", func() {
					BeforeEach(func() {
						request.Version.Time = time.Now()
					})

					It("does not output any versions", func() {
						Ω(response).Should(BeEmpty())
					})
				})

				Context("with its time one interval ago", func() {
					BeforeEach(func() {
						request.Version.Time = time.Now().Add(-1 * time.Minute)
					})

					It("outputs a version containing the current time", func() {
						Ω(response).Should(HaveLen(1))
						Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
					})
				})

				Context("with its time N intervals ago", func() {
					BeforeEach(func() {
						request.Version.Time = time.Now().Add(-5 * time.Minute)
					})

					It("outputs a version containing the current time", func() {
						Ω(response).Should(HaveLen(1))
						Ω(response[0].Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
					})
				})
			})
		})
	})
})
