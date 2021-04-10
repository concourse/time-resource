package resource_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

		It("writes the requested version to the destination", func() {
			input, err := os.ReadFile(filepath.Join(destination, "timestamp"))
			Expect(err).NotTo(HaveOccurred())

			givenTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", string(input))
			Expect(err).NotTo(HaveOccurred())
			Expect(givenTime.Unix()).To(Equal(version.Time.Unix()))
		})

		Context("when the request has no time in its version", func() {
			BeforeEach(func() {
				version = models.Version{}
			})

			It("reports the current time as the version", func() {
				Expect(response.Version.Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
			})
		})
	})
})
