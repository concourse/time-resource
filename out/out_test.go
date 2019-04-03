package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/concourse/time-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Out", func() {
	var tmpdir string
	var source string

	var outCmd *exec.Cmd
	var now time.Time

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "out-source")
		Expect(err).NotTo(HaveOccurred())

		source = path.Join(tmpdir, "out-dir")

		outCmd = exec.Command(outPath, source)
		now = time.Now().UTC()
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var source map[string]interface{}
		var response models.OutResponse

		BeforeEach(func() {
			source = map[string]interface{}{}
			response = models.OutResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := outCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(outCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(map[string]interface{}{
				"source": source,
			})
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a location is specified", func() {
			var loc *time.Location

			BeforeEach(func() {
				var err error
				loc, err = time.LoadLocation("America/Indiana/Indianapolis")
				Expect(err).ToNot(HaveOccurred())

				source["location"] = loc.String()

				now = now.In(loc)
			})

			It("reports specified location's current time(offset: -0400) as the version", func() {
				contained := strings.Contains(response.Version.Time.String(), "-0400")
				Expect(contained).To(BeTrue())
			})
		})
		Context("when a location is not specified", func() {
			It("reports the current time(offset: 0000) as the version", func() {
				contained := strings.Contains(response.Version.Time.String(), "0000")
				Expect(contained).To(BeTrue())
			})
		})
	})
})
