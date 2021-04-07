package cc_core

import (
	"github.com/KompiTech/fabric-cc-core/v2/src/konst"
	. "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Builtin function tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("testdata/assets", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("API call to functionExecute query myAccess", func() {
		It("Should be accessible to anyone", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Ok("functionQuery", "myAccess", rmap.NewEmpty().Bytes())
		})

		It("Should not be accessible as invoke", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "functionInvoke", "myAccess", rmap.NewEmpty().Bytes())
		})

		It("Should list all available permissions for SU", func() {
			myAccess := tctx.Rmap("functionQuery", "myAccess", rmap.NewEmpty().Bytes())
			allAssets := []string{"mockblacklisted", "mockdataafterresolve", "mockpaginate", "mockpd", "mockrefdata", "mockuser", "mockrefblacklist", "mockrequest", "mocklevel1", "mockincident", "mocklevel3", "mocknestedref", "mocktimelog", "mockblogicfail", "mockstate", "mockcomment", "mocklevel2", "mockreffieldblacklist", "mockworknote", "mockworknoteparent"}
			allFuncs := []string{"MockStateInvalidUpdate", "MockPDInvalidCreate", "MockPDInvalidUpdate", "myAccess", "identityAccess", "MockFunc", "MockStateInvalidCreate", "upsertRegistries", "upsertSingletons"}

			Expect(myAccess.Mapa).To(HaveKey("assets_create"))
			Expect(myAccess.Mapa["assets_create"]).To(ConsistOf(allAssets))
			Expect(myAccess.Mapa).To(HaveKey("assets_read"))
			Expect(myAccess.Mapa["assets_read"]).To(ConsistOf(allAssets))
			Expect(myAccess.Mapa).To(HaveKey("assets_update"))
			Expect(myAccess.Mapa["assets_update"]).To(ConsistOf(allAssets))
			Expect(myAccess.Mapa).To(HaveKey("assets_delete"))
			Expect(myAccess.Mapa["assets_delete"]).To(ConsistOf(allAssets))
			Expect(myAccess.Mapa).To(HaveKey("functions_query"))
			Expect(myAccess.Mapa["functions_query"]).To(ConsistOf(allFuncs))
			Expect(myAccess.Mapa).To(HaveKey("functions_invoke"))
			Expect(myAccess.Mapa["functions_invoke"]).To(ConsistOf(allFuncs))
		})

		Context("When some custom grants are present", func() {
			BeforeEach(func() {
				// create role
				roleMap := map[string]interface{}{
					"name": "customRole",
					"grants": []map[string]interface{}{{
						"object": "/user/*",
						"action": "view_sensitive",
					}, {
						"object": "/incident/particular",
						"action": "view_sensitive",
					}, {
						"object": "/incident/*",
						"action": "very_special_action",
					}},
				}
				role := tctx.JSON("roleCreate", rmap.NewFromMap(roleMap).Bytes(), "")

				// add role to identity
				identityMap := map[string]interface{}{
					"roles": []string{role[konst.AssetIdKey].(string)},
				}
				tctx.Ok("identityUpdate", tctx.GetActorFingerprint("ordinaryUser"), rmap.NewFromMap(identityMap).Bytes())
				tctx.SetActor("ordinaryUser")
			})

			It("Should return custom grants", func() {
				myAccess := tctx.Rmap("functionQuery", "myAccess", rmap.NewEmpty().Bytes())
				Expect(myAccess.Mapa).To(HaveKey("custom_grants"))

				cg := myAccess.MustGetRmap("custom_grants")
				Expect(cg.Mapa).To(HaveKeyWithValue("very_special_action", []interface{}{"/incident/*"}))
				Expect(cg.Mapa).To(HaveKeyWithValue("view_sensitive", []interface{}{"/incident/particular", "/user/*"}))
			})

			It("Should return static grants for SU", func() {
				tctx.SetActor("superUser")

				myAccess := tctx.Rmap("functionQuery", "myAccess", rmap.NewEmpty().Bytes())
				Expect(myAccess.Mapa).To(HaveKey("custom_grants"))
				cg := myAccess.MustGetRmap("custom_grants")
				Expect(cg.Mapa).To(HaveKeyWithValue("view_sensitive", []interface{}{"/user/*"}))
			})
		})
	})

	Describe("API call to functionExecute query identityAccess", func() {
		BeforeEach(func() {
			// create role
			roleMap := map[string]interface{}{
				"name": "new role",
				"grants": []map[string]interface{}{{
					"object": "/mockrequest/*",
					"action": "create",
				}, {
					"object": "/mockincident/*",
					"action": "read",
				}},
			}
			role := tctx.JSON("roleCreate", rmap.NewFromMap(roleMap).Bytes(), "")

			// add role to identity
			identityMap := map[string]interface{}{
				"roles": []string{role[konst.AssetIdKey].(string)},
			}
			tctx.Ok("identityUpdate", tctx.GetActorFingerprint("ordinaryUser"), rmap.NewFromMap(identityMap).Bytes())
		})

		It("Should list all available permissions for specific identity", func() {
			args := map[string]interface{}{
				"identity": tctx.GetActorFingerprint("ordinaryUser"), // fingerprint
			}
			access := tctx.Rmap("functionQuery", "identityAccess", rmap.NewFromMap(args).Bytes())
			Expect(access.Mapa).To(HaveKey("assets_create"))
			Expect(access.Mapa["assets_create"]).To(ConsistOf("mockrequest"))
			Expect(access.Mapa).To(HaveKey("assets_read"))
			Expect(access.Mapa["assets_read"]).To(ConsistOf("mockincident"))
			Expect(access.Mapa).NotTo(HaveKey("assets_update"))
		})
	})
})
