package resource_test

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	resource "github.com/concourse/time-resource"
	"github.com/concourse/time-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out", func() {
	var (
		now time.Time

		tmpdir string

		source   models.Source
		response models.OutResponse

		err error
	)

	BeforeEach(func() {
		now = time.Now().UTC()

		tmpdir, err = ioutil.TempDir("", "out-source")
		Expect(err).NotTo(HaveOccurred())

		source = models.Source{}
	})

	JustBeforeEach(func() {
		command := resource.OutCommand{}
		response, err = command.Run(models.OutRequest{
			Source: source,
		})
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {

		JustBeforeEach(func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a location is specified", func() {

			BeforeEach(func() {
				loc, err := time.LoadLocation("America/Indiana/Indianapolis")
				Expect(err).ToNot(HaveOccurred())

				srcLoc := models.Location(*loc)
				source.Location = &srcLoc

				now = now.In(loc)
			})

			It("reports specified location's current time(offset: -0400) as the version", func() {
				// An example of response.Version.Time.String() is
				// 2019-04-03 14:53:10.951241 -0400 EDT
				contained := strings.Contains(response.Version.Time.String(), "-0400")
				Expect(contained).To(BeTrue())
			})
		})
		Context("when a location is not specified", func() {
			It("reports the current time(offset: 0000) as the version", func() {
				// An example of response.Version.Time.String() is
				// 2019-04-03 18:53:10.964705 +0000 UTC
				contained := strings.Contains(response.Version.Time.String(), "0000")
				Expect(contained).To(BeTrue())
			})
		})
	})
})
