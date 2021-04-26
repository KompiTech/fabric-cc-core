package cc_core

import (
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("mockstub tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("../internal/testdata/assets", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("Pagination in rich query in RW TX", func() {
		It("Mockstub must return error", func() {
			req := rmap.NewFromMap(map[string]interface{}{
				"description": "hello world",
			})

			tctx.Error("Transaction has already performed a paginated query. Writes are not allowed", "assetCreate", "mockpaginate", req.Bytes(), -1, "")
		})
	})
})
