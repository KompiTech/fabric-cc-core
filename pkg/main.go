package cc_core

import (
	"testing"

	testing2 "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// just a dummy main, so go doesnt complain about no non-test files
func main() {
	var t *testing.T

	testing2.InitializeCouchDBContainer()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cc-core integration tests")
}
