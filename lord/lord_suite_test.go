package lord_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLord(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lord Suite")
}
