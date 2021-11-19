package cc_core

import (
	"fmt"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("role* method family tests", func() {
	var tctx *TestContext
	var roleReq rmap.Rmap

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("", "").Bytes())
		tctx.RegisterAllActors()

		roleReq = rmap.NewFromMap(map[string]interface{}{
			"name": "Testing role",
			"grants": []map[string]interface{}{{
				"object": "/incident/*",
				"action": "read",
			}},
			"overrides": []map[string]interface{}{{
				"action":  "read",
				"effect":  "allow",
				"subject": tctx.GetActorFingerprint("superUser"),
			}},
		})
	})

	Describe("Call to CC method assetGet(role, ...)", func() {
		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "assetGet", "role", konst.SuperuserRoleUUID, false, "")
			tctx.Error("permission denied", "roleGet", konst.SuperuserRoleUUID, rmap.NewEmpty().Bytes())
		})

		It("Should return data for existing role", func() {
			role := tctx.Rmap("assetGet", "role", konst.SuperuserRoleUUID, false, "").Mapa
			Expect(role).To(HaveKeyWithValue("uuid", konst.SuperuserRoleUUID))
		})
	})

	Describe("Call to CC method assetCreate(role, ...)", func() {
		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "assetCreate", "role", rmap.NewEmpty().Bytes(), -1, "")
			tctx.Error("permission denied", "roleCreate", rmap.NewEmpty().Bytes(), "")
		})

		It("Should create a new role", func() {
			role := tctx.Rmap("assetCreate", "role", roleReq.Bytes(), -1, "").Mapa
			Expect(role).To(HaveKeyWithValue("name", roleReq.MustGetJPtrString("/name")))
			grants := role["grants"].([]interface{})
			overrides := role["overrides"].([]interface{})
			Expect(role).To(HaveKeyWithValue("grants", grants))
			Expect(role).To(HaveKeyWithValue("overrides", overrides))
		})

		Context("when creating system role (is_system_role = true)", func() {
			var patch = rmap.NewFromMap(map[string]interface{}{"name": "System role", "is_system_role": true})

			It("Should require create_system action granted", func() {
				roleUUID := "8466c3d5-1ff5-dfa5-fb3a-4d74ee940eb0"
				tctx.SetActor("ordinaryUser")
				expectedErrMsg := fmt.Sprintf("permission denied, sub: %s, obj: %s, act: create_system", tctx.GetCurrentActorFingerprint(), "/role/"+roleUUID)
				tctx.Error(expectedErrMsg, "assetCreate", "role", patch.Bytes(), -1, roleUUID)
			})

			It("Should create role when create_system action is granted", func() {
				sysRoleCreator := rmap.NewFromMap(map[string]interface{}{
					"name": "Sys Role creator",
					"grants": []map[string]interface{}{{
						"object": "/role/*",
						"action": "create_system",
					}},
				})
				roleUpdaterID := MustGetID(tctx.Rmap("assetCreate", "role", sysRoleCreator.Bytes(), -1, ""))
				ordinaryUserIdentityPatch := rmap.NewFromMap(map[string]interface{}{"roles": []interface{}{roleUpdaterID}})
				tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), ordinaryUserIdentityPatch.Bytes())

				tctx.SetActor("ordinaryUser")
				sysRole := tctx.Rmap("assetCreate", "role", patch.Bytes(), -1, "")
				Expect(sysRole.Mapa).To(HaveKeyWithValue("is_system_role", true))
			})
		})
	})

	Describe("Call to CC method assetUpdate(role, ...)", func() {
		var id string

		BeforeEach(func() {
			id = MustGetID(tctx.Rmap("assetCreate", "role", roleReq.Bytes(), -1, ""))
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "assetUpdate", "role", id, roleReq.Bytes())
		})

		It("Should allow existing role to be updated", func() {
			patchReq := rmap.NewFromMap(map[string]interface{}{
				"name": "Updated role",
				"grants": []map[string]interface{}{{
					"object": "/incident/*",
					"action": "read",
				}, {
					"object": "/request/*",
					"action": "create",
				}},
			})

			role := tctx.Rmap("assetUpdate", "role", id, patchReq.Bytes()).Mapa
			Expect(role).To(HaveKeyWithValue("name", patchReq.MustGetJPtrString("/name")))
			grants := role["grants"].([]interface{})
			overrides := role["overrides"].([]interface{})
			Expect(role).To(HaveKeyWithValue("grants", grants))
			Expect(role).To(HaveKeyWithValue("overrides", overrides))
		})

		Context("when role has is_system_role bool set to true", func() {
			var (
				systemID string
				patch = rmap.NewFromMap(map[string]interface{}{
					"name": "Changed ROLE",
				})
			)

			BeforeEach(func() {
				systemRoleReqID := rmap.NewFromMap(map[string]interface{}{
					"name": "A system role",
					"is_system_role": true,
				})
				systemID = MustGetID(tctx.Rmap("assetCreate", "role", systemRoleReqID.Bytes(), -1, ""))
			})

			It("Should not allow updating system role, even if user has /role/* update granted", func() {
				roleUpdater := rmap.NewFromMap(map[string]interface{}{
					"name": "Role updater",
					"grants": []map[string]interface{}{{
						"object": "/role/*",
						"action": "update",
					}},
				})
				roleUpdaterID := MustGetID(tctx.Rmap("assetCreate", "role", roleUpdater.Bytes(), -1, ""))
				ordinaryUserIdentityPatch := rmap.NewFromMap(map[string]interface{}{"roles": []interface{}{roleUpdaterID}})
				tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), ordinaryUserIdentityPatch.Bytes())

				tctx.SetActor("ordinaryUser")
				expectedErrMsg := fmt.Sprintf("permission denied, sub: %s, obj: %s, act: update_system", tctx.GetCurrentActorFingerprint(), "/role/"+systemID)
				tctx.Error(expectedErrMsg, "assetUpdate", "role", systemID, patch.Bytes())
			})

			It("Should allow updating system role when granted update_system permission", func() {
				sysRoleUpdater := rmap.NewFromMap(map[string]interface{}{
					"name": "Role updater",
					"grants": []map[string]interface{}{{
						"object": "/role/*",
						"action": "update_system",
					}},
				})
				roleUpdaterID := MustGetID(tctx.Rmap("assetCreate", "role", sysRoleUpdater.Bytes(), -1, ""))
				ordinaryUserIdentityPatch := rmap.NewFromMap(map[string]interface{}{"roles": []interface{}{roleUpdaterID}})
				tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), ordinaryUserIdentityPatch.Bytes())

				tctx.SetActor("ordinaryUser")
				tctx.Ok("assetUpdate", "role", systemID, patch.Bytes())
			})
		})
	})

	Describe("Call to CC method assetQuery(role, ...)", func() {
		It("Should filter returned assets", func() {
			tctx.SetActor("ordinaryUser")
			response := tctx.RmapNoResult("assetQuery", "role", rmap.NewEmpty().Bytes(), true)
			result := response.MustGetJPtrIterable("/result")
			Expect(result).To(HaveLen(1)) // superUser role and created role
			Expect(response.MustGetJPtrRmap("/result/0").Mapa).To(HaveKeyWithValue("error", fmt.Sprintf("permission denied, sub: %s, obj: /role/a00a1f64-01a1-4153-b22e-35cf7026ba7e, act: read", tctx.GetCurrentActorFingerprint())))
		})
	})
})
