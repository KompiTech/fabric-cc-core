package cc_core

import (
	"testing"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// this file is the entrypoint for all Ginkgo tests
func TestAll(t *testing.T) {
	InitializeCouchDBContainer()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cc-core integration tests")
}
