package cc_core

import (
	"strings"

	. "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("tracing tests", func() { // TODO on Fabric 2.4.3
	var tctx *TestContext
	var tracingInfo rmap.Rmap
	var tracingString string

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("testdata/assets", "").Bytes())
		tctx.RegisterAllActors()
	})

	Context("When last arg is JSON with tracing magic", func() {
		BeforeEach(func() {
			tracingInfo = rmap.NewFromMap(map[string]interface{}{
				"trace": true,
				"key1":  "val1",
				"key2": map[string]interface{}{
					"nest": true,
				},
			})

			tracingCopy := tracingInfo.Copy()
			delete(tracingCopy.Mapa, "trace")
			tracingString = tracingCopy.String()
		})

		It("Should add tracing info", func() {
			// invalid method name
			tctx.Error(tracingString, "XXX", tracingInfo.Bytes())

			// method internal error
			tctx.Error(tracingString, "assetGet", "", "", false, "", tracingInfo.Bytes())
		})
	})

	Context("When last arg is JSON without tracing magic", func() {
		BeforeEach(func() {
			tracingInfo = rmap.NewFromMap(map[string]interface{}{
				"key1": "val1",
				"key2": map[string]interface{}{
					"nest": true,
				},
			})

			tracingString = tracingInfo.String()
		})

		// test works, but arguments are repeated in error message TODO
		XIt("Should not add tracing info", func() {
			// invalid method name
			errmsg := tctx.GetCC().From(tctx.GetCurrentActor()).Invoke("XXX", tracingInfo.Bytes())
			Expect(strings.Contains(errmsg.Message, tracingString)).To(BeFalse())

			// invalid arg count
			errmsg = tctx.GetCC().From(tctx.GetCurrentActor()).Invoke("assetGet", tracingInfo.Bytes())
			Expect(strings.Contains(errmsg.Message, tracingString)).To(BeFalse())

			// method internal error
			errmsg = tctx.GetCC().From(tctx.GetCurrentActor()).Invoke("assetGet", "", "", false, "", tracingInfo.Bytes())
			Expect(strings.Contains(errmsg.Message, tracingString)).To(BeFalse())
		})
	})
})
