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

		version = models.Version{Time: time.Now()}

		interval := models.Interval(time.Second)
		source = models.Source{Interval: &interval}

		response = models.InResponse{}
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

		It("reports the version's time as the version", func() {
			Expect(response.Version.Time.UnixNano()).To(Equal(version.Time.UnixNano()))
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

		Context("when the request has no time in its version", func() {
			BeforeEach(func() {
				version = models.Version{}
			})

			It("reports the current time as the version", func() {
				Expect(response.Version.Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
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
		})
	})
})
