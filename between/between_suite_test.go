package between_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBetween(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Between Suite")
}
