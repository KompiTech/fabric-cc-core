package cc_core

import (
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("singleton* method family tests", func() {
	var tctx *TestContext

	singletonName := "mocksingleton"
	singletonMock := rmap.MustNewFromYAMLFile("../internal/testdata/singletons/mocksingleton.yaml")

	BeforeEach(func() {
		tctx = getDefaultTextContext()
	})

	Describe("Call to CC method Init()", func() {
		Context("If singleton of the same name exists", func() {
			BeforeEach(func() {
				init := tctx.GetInit("", "")
				init.MustSetJPtr("/singletons", map[string]interface{}{
					singletonName: singletonMock.Mapa,
				})
				tctx.InitOk(init.Bytes())
				tctx.RegisterAllActors()
			})

			It("Should not create a new version if value is identical to latest", func() {
				init := rmap.NewEmpty()
				init.MustSetJPtr("/singletons", map[string]interface{}{
					singletonName: singletonMock.Mapa,
				})
				tctx.InitOk(init.Bytes())

				result := tctx.Rmap("singletonGet", singletonName, -1).Mapa
				Expect(result).To(HaveKeyWithValue("name", singletonName))
				Expect(result).To(HaveKeyWithValue("version", float64(1))) // still version 1 because no change
				Expect(result).To(HaveKeyWithValue("value", singletonMock.MustGetJPtrRmap("/value").Mapa))
			})

			It("Should create a new version if value is different from latest", func() {
				singletonMockV2 := singletonMock.Copy()
				singletonMockV2.MustDeleteJPtr("/value/arrayKey")

				init := rmap.NewEmpty()
				init.MustSetJPtr("/singletons", map[string]interface{}{
					singletonName: singletonMockV2.Mapa,
				})
				tctx.InitOk(init.Bytes())
				result := tctx.Rmap("singletonGet", singletonName, -1).Mapa
				Expect(result).To(HaveKeyWithValue("name", singletonName))
				Expect(result).To(HaveKeyWithValue("version", float64(2))) // new version 2 created
				Expect(result).To(HaveKeyWithValue("value", singletonMockV2.MustGetJPtrRmap("/value").Mapa))
			})
		})
	})

	Describe("Call to CC method singletonGet", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "singletonGet", singletonName, -1)
		})

		Context("When singleton with name exists", func() {
			BeforeEach(func() {
				init := rmap.NewEmpty()
				init.MustSetJPtr("/singletons", map[string]interface{}{
					singletonName: singletonMock.Mapa,
				})
				tctx.InitOk(init.Bytes())
			})

			It("Should return singleton data", func() {
				result := tctx.Rmap("singletonGet", singletonName, -1).Mapa

				Expect(result).To(HaveKeyWithValue("name", singletonName))
				Expect(result).To(HaveKeyWithValue("version", float64(1)))
				Expect(result).To(HaveKeyWithValue("value", singletonMock.MustGetJPtrRmap("/value").Mapa))
			})
		})
	})

	Describe("Call to CC method singletonList", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "singletonList")
		})

		Context("When no singletons are present", func() {
			It("Should return empty list", func() {
				names := tctx.RmapNoResult("singletonList").MustGetJPtrIterable("/result")
				Expect(names).To(BeEmpty())
			})
		})

		Context("When some singletons are present", func() {
			BeforeEach(func() {
				tctx.Ok("singletonUpsert", singletonName, singletonMock.Bytes())
			})

			It("Should return list of singleton names", func() {
				names := tctx.RmapNoResult("singletonList").MustGetJPtrIterable("/result")
				Expect(names).To(ConsistOf(singletonName))
			})
		})
	})

	Describe("Call to CC method singletonUpsert", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "singletonUpsert", singletonName, rmap.NewEmpty().Bytes())
		})

		Context("When no singleton with name exists", func() {
			It("Should create version 1", func() {
				result := tctx.Rmap("singletonUpsert", singletonName, singletonMock.Bytes()).Mapa
				Expect(result).To(HaveKeyWithValue("name", singletonName))
				Expect(result).To(HaveKeyWithValue("version", float64(1)))
				Expect(result).To(HaveKeyWithValue("value", singletonMock.MustGetJPtrRmap("/value").Mapa))
			})
		})

		Context("When singleton with name exists", func() {
			BeforeEach(func() {
				tctx.Ok("singletonUpsert", singletonName, singletonMock.Bytes())
			})

			It("Should not create a new version if value is identical to latest", func() {
				result := tctx.Rmap("singletonUpsert", singletonName, singletonMock.Bytes()).Mapa
				Expect(result).To(HaveKeyWithValue("name", singletonName))
				Expect(result).To(HaveKeyWithValue("version", float64(1))) // still version 1
				Expect(result).To(HaveKeyWithValue("value", singletonMock.MustGetJPtrRmap("/value").Mapa))
			})

			It("Should create a new version if value is different from latest", func() {
				singletonMockV2 := singletonMock.Copy()
				singletonMockV2.MustDeleteJPtr("/value/stringKey")

				result := tctx.Rmap("singletonUpsert", singletonName, singletonMockV2.Bytes()).Mapa
				Expect(result).To(HaveKeyWithValue("name", singletonName))
				Expect(result).To(HaveKeyWithValue("version", float64(2))) // new version 2 created
				Expect(result).To(HaveKeyWithValue("value", singletonMockV2.MustGetJPtrRmap("/value").Mapa))
			})
		})
	})
})
