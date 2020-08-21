package resource_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestTimeResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Time Resource Suite")
}
