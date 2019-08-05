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
		Expect(err).NotTo(HaveOccurred())

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
			interval := models.Interval(time.Second)

			request = models.InRequest{
				Version: models.Version{Time: time.Now()},
				Source:  models.Source{Interval: &interval},
			}

			response = models.InResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := inCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(inCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reports the version's time as the version", func() {
			Expect(response.Version.Time.UnixNano()).To(Equal(request.Version.Time.UnixNano()))
		})

		It("writes the requested version and source to the destination", func() {
			input, err := os.Open(filepath.Join(destination, "input"))
			Expect(err).NotTo(HaveOccurred())

			var requested models.InRequest
			err = json.NewDecoder(input).Decode(&requested)
			Expect(err).NotTo(HaveOccurred())

			Expect(requested.Version.Time.Unix()).To(Equal(request.Version.Time.Unix()))
			Expect(requested.Source).To(Equal(request.Source))
		})

		Context("when the request has no time in its version", func() {
			BeforeEach(func() {
				request.Version = models.Version{}
			})

			It("reports the current time as the version", func() {
				Expect(response.Version.Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
			})
		})
		Context("when the request has a future time in its version", func() {
			duration := time.Second * 10

			BeforeEach(func() {
				request.Version.Time = time.Now().Add(duration)
			})

			It("reports the future time as the version", func() {
				Expect(response.Version.Time.Unix()).To(BeNumerically(">", time.Now().Unix(), duration))
			})

			Context("waits for the future time", func() {
				BeforeEach(func() {
					request.Params.Wait = true
				})

				It("waits until the future time has been reached", func() {
					Expect(response.Version.Time.Unix()).To(BeNumerically("~", time.Now().Unix(), 1))
				})
			})
		})
	})
})
