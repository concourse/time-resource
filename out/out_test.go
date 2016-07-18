package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "out-source")
		Expect(err).NotTo(HaveOccurred())

		source = path.Join(tmpdir, "out-dir")

		outCmd = exec.Command(outPath, source)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.OutRequest
		var response models.OutResponse

		BeforeEach(func() {
			request = models.OutRequest{
				Source: models.Source{Interval: "1s"},
			}

			response = models.OutResponse{}
		})

		JustBeforeEach(func() {
			stdin := new(bytes.Buffer)

			err := json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			outCmd.Stdin = stdin

			session, err := gexec.Start(outCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reports the current time as the version", func() {
			Expect(response.Version.Time).To(BeTemporally("~", time.Now(), time.Second))
		})
	})
})
