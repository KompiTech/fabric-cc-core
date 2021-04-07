package cc_core

import (
	"testing"

	testing2 "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// this file is the entrypoint for all Ginkgo tests
func TestAll(t *testing.T) {
	testing2.InitializeCouchDBContainer()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cc-core integration tests")
}
