package cc_core

import (
	"testing"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// just a dummy main, so go doesnt complain about no non-test files
func main() {
	var t *testing.T

	InitializeCouchDBContainer()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cc-core integration tests")
}
