package resource_test

import (
	"testing"
	"time"

	resource "github.com/concourse/time-resource"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = BeforeSuite(func() {
	mockNow := time.Now().UTC().AddDate(0, 0, -1)
	resource.GetCurrentTime = func() time.Time {
		return mockNow
	}
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestTimeResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Time Resource Suite")
}
