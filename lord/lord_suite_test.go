package lord_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLord(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lord Suite")
}
