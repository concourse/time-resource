package models_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/concourse/time-resource/models"
)

var _ = Describe("Source", func() {
	var (
		config string

		source models.Source
		err    error
	)

	BeforeEach(func() {
		config = ""
		source = models.Source{}
		err = nil
	})

	JustBeforeEach(func() {
		err = json.Unmarshal([]byte(config), &source)
	})

	Context("a start with no stop", func() {
		BeforeEach(func() {
			config = `{ "start": "3:04" }`
		})

		It("generates a validation error", func() {
			Expect(err).ToNot(HaveOccurred())

			err = source.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("must configure 'stop' if 'start' is set"))
		})
	})

	Context("a stop with no start", func() {
		BeforeEach(func() {
			config = `{ "stop": "3:04" }`
		})

		It("generates a validation error", func() {
			Expect(err).ToNot(HaveOccurred())

			err = source.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("must configure 'start' if 'stop' is set"))
		})
	})

	Context("when the range is given in another timezone", func() {
		BeforeEach(func() {
			config = `{ "start": "3:04 -0100", "stop": "9:04 -0700" }`
		})

		It("is valid", func() {
			Expect(err).ToNot(HaveOccurred())

			err = source.Validate()
			Expect(err).ToNot(HaveOccurred())

			Expect(source.Start).ToNot(BeNil())
			Expect(source.Stop).ToNot(BeNil())

			Expect(source.Stop.Minute()).To(Equal(source.Start.Minute()))
			Expect(source.Stop.Hour()).To(Equal(source.Start.Hour() + 12))
		})
	})

	Context("when using cron expressions", func() {
		Context("with both interval and cron specified", func() {
			BeforeEach(func() {
				config = `{ "interval": "1h", "cron": "*/5 * * * *" }`
			})

			It("generates a validation error", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot configure 'interval' or 'start'/'stop' or 'days' with 'cron'"))
			})
		})

		Context("with both interval and days and cron specified", func() {
			BeforeEach(func() {
				config = `{ "interval": "1h", "days": ["Monday"], "cron": "*/5 * * * *" }`
			})

			It("generates a validation error", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot configure 'interval' or 'start'/'stop' or 'days' with 'cron'"))
			})
		})

		Context("with both interval and start/stop and cron specified", func() {
			BeforeEach(func() {
				config = `{ "interval": "1h", "start": "6AM", "stop": "7am", "cron": "*/5 * * * *" }`
			})

			It("generates a validation error", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot configure 'interval' or 'start'/'stop' or 'days' with 'cron'"))
			})
		})

		Context("with an invalid cron expression", func() {
			BeforeEach(func() {
				config = `{ "cron": "invalid expression" }`
			})

			It("generates a validation error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid cron expression"))
			})
		})

		Context("with a cron expression that runs every second (6-field)", func() {
			BeforeEach(func() {
				config = `{ "cron": "* * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression that runs every 10 seconds", func() {
			BeforeEach(func() {
				config = `{ "cron": "*/10 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression that runs every 30 seconds", func() {
			BeforeEach(func() {
				config = `{ "cron": "*/30 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression that runs at specific seconds", func() {
			BeforeEach(func() {
				config = `{ "cron": "0,15,30,45 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression using a seconds range", func() {
			BeforeEach(func() {
				config = `{ "cron": "0-30 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression that runs exactly every minute (5-field)", func() {
			BeforeEach(func() {
				config = `{ "cron": "* * * * *" }`
			})

			It("is valid (minimum acceptable frequency)", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a cron expression that runs every 5 minutes", func() {
			BeforeEach(func() {
				config = `{ "cron": "*/5 * * * *" }`
			})

			It("is valid", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a cron expression that runs hourly", func() {
			BeforeEach(func() {
				config = `{ "cron": "0 * * * *" }`
			})

			It("is valid", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a cron expression that runs every 2 hours", func() {
			BeforeEach(func() {
				config = `{ "cron": "0 */2 * * *" }`
			})

			It("is valid", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a cron expression that runs daily", func() {
			BeforeEach(func() {
				config = `{ "cron": "0 0 * * *" }`
			})

			It("is valid", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a cron expression that runs weekly", func() {
			BeforeEach(func() {
				config = `{ "cron": "@weekly" }`
			})

			It("is valid", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a cron expression using seconds field but running only once per minute", func() {
			BeforeEach(func() {
				config = `{ "cron": "0 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression that runs exactly every 60 seconds", func() {
			BeforeEach(func() {
				config = `{ "cron": "*/60 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression that runs every 90 seconds", func() {
			BeforeEach(func() {
				config = `{ "cron": "*/90 * * * * *" }`
			})

			It("generates a validation error for having seconds field", func() {
				Expect(err).ToNot(HaveOccurred())

				err = source.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cron expressions with seconds field are not supported"))
			})
		})

		Context("with a cron expression with an invalid step value in seconds", func() {
			BeforeEach(func() {
				config = `{ "cron": "*/abc * * * * *" }`
			})

			It("is caught during initial cron validation", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid cron expression"))
			})
		})
	})
})
