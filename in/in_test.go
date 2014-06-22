package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/concourse/time-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("In", func() {
	var tmpdir string
	var destination string

	var inCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "in-destination")
		Ω(err).ShouldNot(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		inCmd = exec.Command(inPath, destination)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.InRequest
		var response models.InResponse

		BeforeEach(func() {
			request = models.InRequest{
				Version: models.Version{Time: time.Now()},
				Source:  models.Source{Interval: "1s"},
			}

			response = models.InResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := inCmd.StdinPipe()
			Ω(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(inCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("reports the current time as metadata", func() {
			Ω(response.Version.Time.Unix()).Should(BeNumerically("~", time.Now().Unix(), 1))
		})

		It("writes the requested version and source to the destination", func() {
			input, err := os.Open(filepath.Join(destination, "input"))
			Ω(err).ShouldNot(HaveOccurred())

			var requested models.InRequest
			err = json.NewDecoder(input).Decode(&requested)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(requested).Should(Equal(request))
		})
	})
})
