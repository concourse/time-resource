package resource_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	resource "github.com/concourse/time-resource"
	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		tmpdir      string
		destination string

		source   models.Source
		version  models.Version
		response models.InResponse

		err error
	)

	BeforeEach(func() {
		tmpdir, err = ioutil.TempDir("", "in-destination")
		Expect(err).NotTo(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		source = models.Source{}
		version = models.Version{}
	})

	JustBeforeEach(func() {
		command := resource.InCommand{}
		response, err = command.Run(destination, models.InRequest{
			Source:  source,
			Version: version,
		})
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		BeforeEach(func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("reports the latest valid time as the version", func() {
			expectedTime := resource.Offset(lord.TimeLord{}, resource.GetCurrentTime())
			Expect(response.Version.Time.Unix()).To(Equal(expectedTime.Unix()))
		})

		Context("when a location is specified", func() {
			BeforeEach(func() {
				loc, err := time.LoadLocation("America/Indiana/Indianapolis")
				Expect(err).ToNot(HaveOccurred())

				srcLoc := models.Location(*loc)
				source.Location = &srcLoc
			})

			It("reports specified location's current time(offset: -0400) as the version", func() {
				// An example of response.Version.Time.String() is
				// 2019-04-03 14:53:10.951241 -0400 EDT
				contained := strings.Contains(response.Version.Time.String(), "-0400")
				Expect(contained).To(BeTrue())
			})
		})

		Context("when the request is for a specific time", func() {
			BeforeEach(func() {
				version.Time = resource.GetCurrentTime()
			})

			It("offsets the requested version's time", func() {
				expectedTime := resource.Offset(lord.TimeLord{}, version.Time)
				Expect(response.Version.Time.Unix()).To(Equal(expectedTime.Unix()))
			})

			It("writes the requested version and source to the destination", func() {
				input, err := os.Open(filepath.Join(destination, "input"))
				Expect(err).NotTo(HaveOccurred())

				var requested models.InRequest
				err = json.NewDecoder(input).Decode(&requested)
				Expect(err).NotTo(HaveOccurred())

				Expect(requested.Version.Time.Unix()).To(Equal(version.Time.Unix()))
				Expect(requested.Source).To(Equal(source))
			})

			Context("when the requested version came from a predictable generator", func() {
				BeforeEach(func() {
					version.Time = resource.Offset(lord.TimeLord{}, version.Time)
				})

				It("reports the version's time as the version", func() {
					Expect(response.Version.Time.UnixNano()).To(Equal(version.Time.UnixNano()))
				})
			})

			Context("when the requested version is in a different location", func() {
				BeforeEach(func() {
					loc, err := time.LoadLocation("America/Indiana/Indianapolis")
					Expect(err).ToNot(HaveOccurred())

					srcLoc := models.Location(*loc)
					source.Location = &srcLoc
				})

				It("reports source's location(offset: -0400) as the version", func() {
					// An example of response.Version.Time.String() is
					// 2019-04-03 14:53:10.951241 -0400 EDT
					contained := strings.Contains(response.Version.Time.String(), "-0400")
					Expect(contained).To(BeTrue())
				})
			})
		})
	})
})
