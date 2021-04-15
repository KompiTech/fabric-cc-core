package cc_core

import (
	. "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("registry* method family tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
	})

	var assetName string

	Describe("Call to CC method Init()", func() {
		It("Should upsert any registries and create changelog items", func() {
			assetName = "mockincident"
			regItemFile := rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")

			init := tctx.GetInit("", "")
			registries := map[string]interface{}{
				assetName: regItemFile.Mapa,
			}
			init.Mapa["registries"] = registries

			tctx.InitOk(init.Bytes())
			tctx.Rmap("identityAddMe", rmap.NewEmpty().Bytes())

			// registryItem must be present
			regItem := tctx.Rmap("registryGet", assetName, -1)
			Expect(regItem.Mapa).To(HaveKeyWithValue("name", assetName))
			Expect(regItem.Mapa).To(HaveKeyWithValue("destination", regItemFile.MustGetJPtrString("/destination")))
			Expect(regItem.Mapa).To(HaveKeyWithValue("version", float64(1)))
			Expect(regItem.Mapa).To(HaveKeyWithValue("schema", regItemFile.MustGetJPtr("/schema")))

			// changelogItem must be present
			clItem := tctx.Rmap("changelogGet", 1)
			Expect(clItem.Mapa).To(HaveKey("timestamp"))
			Expect(clItem.Mapa).To(HaveKey("txid"))
			Expect(clItem.Mapa).To(HaveKey("changes"))
			Expect(clItem.Mapa["changes"]).To(HaveLen(1))
			Expect(clItem.MustGetJPtrString("/changes/0/operation")).To(Equal("create"))
			Expect(clItem.MustGetJPtrInt("/changes/0/version")).To(Equal(1))
			Expect(clItem.MustGetJPtrString("/changes/0/assetName")).To(Equal(assetName))

			// changelogItem is the only item in list
			data := tctx.RmapNoResult("changelogList")
			clist := data.MustGetJPtrIterable("/result")
			Expect(clist).To(HaveLen(1))
			clItem = rmap.MustNewFromInterface(clist[0])
			Expect(clItem.Mapa).To(HaveKey("timestamp"))
			Expect(clItem.Mapa).To(HaveKey("txid"))
			Expect(clItem.Mapa).To(HaveKey("changes"))
			Expect(clItem.Mapa["changes"]).To(HaveLen(1))
			Expect(clItem.MustGetJPtrString("/changes/0/operation")).To(Equal("create"))
			Expect(clItem.MustGetJPtrInt("/changes/0/version")).To(Equal(1))
			Expect(clItem.MustGetJPtrString("/changes/0/assetName")).To(Equal(assetName))
		})

		It("Should return error if some schema root type is not object", func() {
			regItemFile := rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")
			regItemFile.MustSetJPtr("/schema/type", "array")

			init := tctx.GetInit("", "")
			registries := map[string]interface{}{
				assetName: regItemFile.Mapa,
			}
			init.Mapa["registries"] = registries

			tctx.InitError("type on top level is not: object", init.Bytes())
		})

		Context("When one or more schemas contains errors in instantiate args", func() {
			It("Should return error and not update asset", func() {
				initData := tctx.GetInit("", "")
				initData.Mapa["registries"] = map[string]interface{}{}

				initData.SetJPtr("/registries/mockcommentinvalidschema", rmap.MustNewFromYAMLFile("testdata/mockcommentinvalidschema.yaml"))
				tctx.InitError("schema for: mockcommentinvalidschema is not a valid JSON schema", initData.Bytes())
			})
		})

		Context("When registryItem for asset already exists", func() {
			var regItemFile rmap.Rmap

			BeforeEach(func() {
				assetName = "mockincident"
				regItemFile = rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")

				init := tctx.GetInit("", "")
				registries := map[string]interface{}{
					assetName: regItemFile.Mapa,
				}
				init.Mapa["registries"] = registries

				tctx.InitOk(init.Bytes())
				tctx.Rmap("identityAddMe", rmap.NewEmpty().Bytes())
			})

			It("Should not upsert and not create changelog item if upserting identical schema", func() {
				registries := map[string]interface{}{
					assetName: regItemFile.Mapa,
				}
				init := rmap.NewEmpty()
				init.Mapa["registries"] = registries
				tctx.InitOk(init.Bytes())

				// registryItem must be present with only one version
				tctx.Error("reg.GetItem() failed: registryItem name: mockincident, version: 2 not found", "registryGet", assetName, 2)
				regItem := tctx.Rmap("registryGet", assetName, 1)
				Expect(regItem.Mapa).To(HaveKeyWithValue("name", assetName))
				Expect(regItem.Mapa).To(HaveKeyWithValue("destination", regItemFile.MustGetJPtrString("/destination")))
				Expect(regItem.Mapa).To(HaveKeyWithValue("version", float64(1)))
				Expect(regItem.Mapa).To(HaveKeyWithValue("schema", regItemFile.MustGetJPtrRmap("/schema").Mapa))

				// changelogItem must be present with only one version
				tctx.Error("cl.Get() failed: invalid changelog number", "changelogGet", 2)
				clItem := tctx.Rmap("changelogGet", 1)
				Expect(clItem.Mapa).To(HaveKey("timestamp"))
				Expect(clItem.Mapa).To(HaveKey("txid"))
				Expect(clItem.Mapa).To(HaveKey("changes"))
				Expect(clItem.Mapa["changes"]).To(HaveLen(1))
				Expect(clItem.MustGetJPtrString("/changes/0/operation")).To(Equal("create"))
				Expect(clItem.MustGetJPtrInt("/changes/0/version")).To(Equal(1))
				Expect(clItem.MustGetJPtrString("/changes/0/assetName")).To(Equal(assetName))

				// changelogItem is the only item in list
				data := tctx.RmapNoResult("changelogList")
				clist := data.MustGetJPtrIterable("/result")
				Expect(clist).To(HaveLen(1))
				Expect(clItem.Mapa).To(HaveKey("timestamp"))
				Expect(clItem.Mapa).To(HaveKey("txid"))
				Expect(clItem.Mapa).To(HaveKey("changes"))
				Expect(clItem.Mapa["changes"]).To(HaveLen(1))
				Expect(clItem.MustGetJPtrString("/changes/0/operation")).To(Equal("create"))
				Expect(clItem.MustGetJPtrInt("/changes/0/version")).To(Equal(1))
				Expect(clItem.MustGetJPtrString("/changes/0/assetName")).To(Equal(assetName))
			})

			It("Should upsert new schema version and create changelog item if changed", func() {
				oldRegItemFile := regItemFile
				updatedRegItemFile := oldRegItemFile.Copy()

				// modify the schema -> make "description" optional
				updatedRegItemFile.MustDeleteJPtr("/schema/required")

				registries := map[string]interface{}{
					assetName: updatedRegItemFile.Mapa,
				}
				init := rmap.NewEmpty()
				init.Mapa["registries"] = registries
				tctx.InitOk(init.Bytes())

				// old registryItem must be unchanged with version 1
				oldRegItem := tctx.Rmap("registryGet", assetName, 1)
				Expect(oldRegItem.Mapa).To(HaveKeyWithValue("name", assetName))
				Expect(oldRegItem.Mapa).To(HaveKeyWithValue("destination", oldRegItemFile.MustGetJPtrString("/destination")))
				Expect(oldRegItem.Mapa).To(HaveKeyWithValue("version", float64(1)))
				Expect(oldRegItem.Mapa).To(HaveKeyWithValue("schema", oldRegItemFile.MustGetJPtrRmap("/schema").Mapa))

				// new registryItem must exist with version 2
				regItem := tctx.Rmap("registryGet", assetName, 2)
				Expect(regItem.Mapa).To(HaveKeyWithValue("name", assetName))
				Expect(regItem.Mapa).To(HaveKeyWithValue("destination", updatedRegItemFile.MustGetJPtrString("/destination")))
				Expect(regItem.Mapa).To(HaveKeyWithValue("version", float64(2)))
				Expect(regItem.Mapa).To(HaveKeyWithValue("schema", updatedRegItemFile.MustGetJPtrRmap("/schema").Mapa))

				// latest can also be reached by -1 version
				latestRegItem := tctx.Rmap("registryGet", assetName, -1)
				Expect(latestRegItem.Mapa).To(HaveKeyWithValue("name", assetName))
				Expect(latestRegItem.Mapa).To(HaveKeyWithValue("destination", updatedRegItemFile.MustGetJPtrString("/destination")))
				Expect(latestRegItem.Mapa).To(HaveKeyWithValue("version", float64(2)))
				Expect(latestRegItem.Mapa).To(HaveKeyWithValue("schema", updatedRegItemFile.MustGetJPtrRmap("/schema").Mapa))
			})

			It("Should not allow destination to be changed", func() {
				regItemFile.MustSetJPtr("/destination", "private_data:FOOBAR")
				registries := map[string]interface{}{
					assetName: regItemFile.Mapa,
				}
				init := rmap.NewEmpty()
				init.Mapa["registries"] = registries
				tctx.InitError("upsertRegistries() failed: reg.BulkUpsertItems() failed: r.upsertItem() failed: unable to change destination of: mockincident, from: state, to: private_data:FOOBAR", init.Bytes())
			})
		})
	})

	Describe("Call to CC method registryUpsert", func() {
		It("Should be protected", func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "registryUpsert", "foobar", rmap.NewEmpty().Bytes())
		})

		Context("When no registryItems for some asset exists", func() {
			BeforeEach(func() {
				tctx.InitOk(tctx.GetInit("", "").Bytes())
				tctx.RegisterAllActors()
			})

			It("Should create registryItem with version 1 and produce changelog item", func() {
				assetName = "mockincident"
				regItemFile := rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")
				result := tctx.Rmap("registryUpsert", assetName, regItemFile.Bytes()).Mapa
				Expect(result).To(HaveKeyWithValue("version", float64(1)))
				Expect(result).To(HaveKeyWithValue("destination", regItemFile.MustGetJPtrString("/destination")))
				Expect(result).To(HaveKeyWithValue("schema", regItemFile.MustGetJPtrRmap("/schema").Mapa))

				clItem := tctx.Rmap("changelogGet", 1)
				Expect(clItem.Mapa).To(HaveKey("changes"))
				Expect(clItem.MustGetJPtrIterable("/changes")).To(HaveLen(1))

				change := clItem.MustGetJPtrRmap("/changes/0").Mapa
				Expect(change).To(HaveKeyWithValue("operation", "create"))
				Expect(change).To(HaveKeyWithValue("assetName", assetName))
				Expect(change).To(HaveKeyWithValue("version", float64(1)))
			})

			Context("When schema contains error", func() {
				It("Should return error", func() {
					tctx.Error("schema for: mockcommentinvalidschema is not a valid JSON schema", "registryUpsert", "mockcommentinvalidschema", rmap.MustNewFromYAMLFile("testdata/mockcommentinvalidschema.yaml").Bytes())
				})
			})
		})

		Context("When registryItem for some asset exists in multiple versions", func() {
			var v1, v2 rmap.Rmap
			assetName := "mockincident"

			BeforeEach(func() {
				tctx.InitOk(tctx.GetInit("", "").Bytes())
				tctx.RegisterAllActors()

				v1 = rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")

				tctx.Ok("registryUpsert", assetName, v1.Bytes()) // version 1

				v2 = v1.Copy()
				// version 2 adds a new string field foobar
				v2.MustSetJPtr("/schema/properties/foobar", map[string]interface{}{"type": "string"})
				tctx.Ok("registryUpsert", assetName, v2.Bytes())
			})

			Context("When schema contains error", func() {
				It("Should return error", func() {
					tctx.Error("schema for: mockincident is not a valid JSON schema", "registryUpsert", assetName, rmap.MustNewFromYAMLFile("testdata/mockcommentinvalidschema.yaml").Bytes())
				})
			})

			It("Should create registryItem with next version and produce changelog item", func() {
				// version 3 add a new string field foo
				myRegItem := v2.Copy()
				myRegItem.MustSetJPtr("/schema/properties/foo", map[string]interface{}{"type": "string"})
				result := tctx.Rmap("registryUpsert", assetName, myRegItem.Bytes()).Mapa
				Expect(result).To(HaveKeyWithValue("name", assetName))
				Expect(result).To(HaveKeyWithValue("version", float64(3)))
				Expect(result).To(HaveKeyWithValue("destination", myRegItem.MustGetJPtrString("/destination")))
				Expect(result).To(HaveKeyWithValue("schema", myRegItem.MustGetJPtrRmap("/schema").Mapa))

				clItem := tctx.Rmap("changelogGet", 3)
				Expect(clItem.Mapa).To(HaveKey("changes"))
				Expect(clItem.MustGetJPtrIterable("/changes")).To(HaveLen(1))

				change := clItem.MustGetJPtrRmap("/changes/0").Mapa
				Expect(change).To(HaveKeyWithValue("operation", "update"))
				Expect(change).To(HaveKeyWithValue("assetName", assetName))
				Expect(change).To(HaveKeyWithValue("version", float64(3)))
			})

			It("Should not create registryItem with next version and not produce changelog item if schema is identical to previous version", func() {
				result := tctx.Rmap("registryUpsert", assetName, v2.Bytes()).Mapa
				Expect(result).To(HaveKeyWithValue("name", assetName))
				Expect(result).To(HaveKeyWithValue("version", float64(2))) // still returning version 2
				Expect(result).To(HaveKeyWithValue("destination", v2.MustGetJPtrString("/destination")))
				Expect(result).To(HaveKeyWithValue("schema", v2.MustGetJPtrRmap("/schema").Mapa))

				tctx.Error("cl.Get() failed: invalid changelog number", "changelogGet", 3)
			})

			It("Should return data for particular version", func() {
				v1Result := tctx.Rmap("registryGet", assetName, 1).Mapa
				Expect(v1Result).To(HaveKeyWithValue("name", assetName))
				Expect(v1Result).To(HaveKeyWithValue("version", float64(1)))
				Expect(v1Result).To(HaveKeyWithValue("destination", v1.MustGetJPtrString("/destination")))
				Expect(v1Result).To(HaveKeyWithValue("schema", v1.MustGetJPtrRmap("/schema").Mapa))

				v2Result := tctx.Rmap("registryGet", assetName, 2).Mapa
				Expect(v2Result).To(HaveKeyWithValue("name", assetName))
				Expect(v2Result).To(HaveKeyWithValue("version", float64(2)))
				Expect(v2Result).To(HaveKeyWithValue("destination", v2.MustGetJPtrString("/destination")))
				Expect(v2Result).To(HaveKeyWithValue("schema", v2.MustGetJPtrRmap("/schema").Mapa))

				latestResult := tctx.Rmap("registryGet", assetName, -1).Mapa
				Expect(latestResult).To(HaveKeyWithValue("name", assetName))
				Expect(latestResult).To(HaveKeyWithValue("version", float64(2)))
				Expect(latestResult).To(HaveKeyWithValue("destination", v2.MustGetJPtrString("/destination")))
				Expect(latestResult).To(HaveKeyWithValue("schema", v2.MustGetJPtrRmap("/schema").Mapa))
			})
		})
	})

	Describe("Call to CC method registryGet", func() {
		BeforeEach(func() {
			init := tctx.GetInit("", "")
			regs := ScanSomething("testdata/assets")
			init.MustSetJPtr("/registries", regs.Mapa)
			tctx.InitOk(init.Bytes())
			tctx.RegisterAllActors()
		})

		It("Should return data for some registryItem", func() {
			mockincidentRI := tctx.Rmap("registryGet", "mockincident", -1)
			// remove extra keys that registryGet adds
			delete(mockincidentRI.Mapa, "name")
			delete(mockincidentRI.Mapa, "version")

			refRI := rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")
			Expect(mockincidentRI.Mapa).To(Equal(refRI.Mapa))
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "registryGet", "foobar", -1)
		})
	})

	Describe("Call to CC method registryList", func() {
		BeforeEach(func() {
			init := tctx.GetInit("", "")
			regs := ScanSomething("testdata/assets")
			init.MustSetJPtr("/registries", regs.Mapa)
			tctx.InitOk(init.Bytes())
			tctx.RegisterAllActors()
		})

		It("Should return list of available registries", func() {
			regs := tctx.RmapNoResult("registryList").MustGetJPtrIterable("/result")
			seen := rmap.NewEmpty()
			for _, nameI := range regs {
				seen.Mapa[nameI.(string)] = struct{}{}
			}

			refMap := map[string]interface{}{
				"mockuser":              struct{}{},
				"mockcomment":           struct{}{},
				"mockincident":          struct{}{},
				"mocklevel1":            struct{}{},
				"mocklevel2":            struct{}{},
				"mocklevel3":            struct{}{},
				"mocknestedref":         struct{}{},
				"mockrequest":           struct{}{},
				"mocktimelog":           struct{}{},
				"mockblogicfail":        struct{}{},
				"mockpaginate":          struct{}{},
				"mockdataafterresolve":  struct{}{},
				"mockrefdata":           struct{}{},
				"mockblacklisted":       struct{}{},
				"mockrefblacklist":      struct{}{},
				"mockpd":                struct{}{},
				"mockstate":             struct{}{},
				"mockreffieldblacklist": struct{}{},
				"mockworknote":          struct{}{},
				"mockworknoteparent":    struct{}{},
				"mocklegacyschema": 	 struct{}{},
			}
			Expect(seen.Mapa).To(Equal(refMap))
		})

		It("Should be protected", func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "registryList")
		})
	})
})
