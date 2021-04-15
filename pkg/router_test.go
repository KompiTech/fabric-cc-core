package cc_core

import (
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("router tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("Invalid method name", func() {
		It("Should return error - prefix", func() {
			tctx.Error("Invalid function blablaAssetGet passed", "blablaAssetGet")
		})

		It("Should return error - middle of string", func() {
			tctx.Error("Invalid function assetXXXXGet passed", "assetXXXXGet")
		})

		It("Should return error - suffix", func() {
			tctx.Error("Invalid function assetGetXXXXX passed", "assetGetXXXXX")
		})
	})

	Describe("Valid function name", func() {
		It("Should be case insensitive on first letter", func() {
			resp1 := tctx.Rmap("identityMe", false)
			resp2 := tctx.Rmap("IdentityMe", false)
			Expect(resp1.String()).To(Equal(resp2.String()))
		})
	})
})
