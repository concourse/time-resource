package models_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
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
})
