package cc_core

import (
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("kompiguard", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("../internal/testdata/assets", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("when defining objects", func() {
		It("Should allow to use * as object name", func() {
			// create assets of different kinds
			incidentUUID := MustGetID(tctx.Rmap("assetCreate", "mockincident", rmap.NewFromMap(map[string]interface{}{"description": "ahoj"}).Bytes(), -1, ""))
			requestUUID := MustGetID(tctx.Rmap("assetCreate", "mockrequest", rmap.NewFromMap(map[string]interface{}{"number": "1234"}).Bytes(), -1, ""))

			// create role with read on *
			role := rmap.NewFromMap(map[string]interface{}{
				"name": "All reader",
				"grants": []map[string]interface{}{{
					"object": "*",
					"action": "read",
				}},
			})
			roleUUID := MustGetID(tctx.Rmap("assetCreate", "role", role.Bytes(), -1, ""))

			// grant role to ordinaryUser
			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), rmap.NewFromMap(map[string]interface{}{"roles": []string{roleUUID}}).Bytes())

			// ordinaryUser must have rights to read both asset kinds
			tctx.SetActor("ordinaryUser")
			tctx.Ok("assetGet", "mockincident", incidentUUID, false, rmap.NewEmpty().Bytes())
			tctx.Ok("assetGet", "mockrequest", requestUUID, false, rmap.NewEmpty().Bytes())
		})

		It("Should allow description to grants", func() {
			role := rmap.NewFromMap(map[string]interface{}{
				"name": "All reader",
				"grants": []map[string]interface{}{{
					"description": "Read all objects",
					"object": "*",
					"action": "read",
				}},
			})
			tctx.Ok("assetCreate", "role", role.Bytes(), -1, "")
		})
	})

	Describe("when controlling access to functions", func() {
		It("Should work with both lowercase and uppercase object name", func() {
			tctx.SetActor("ordinaryUser")
			// no permission, cannot execute
			tctx.Error("permission denied, sub: "+tctx.GetActorFingerprint("ordinaryUser")+", obj: /function/invoke/MockFunc, act: execute", "functionInvoke", "MockFunc", rmap.NewEmpty().Bytes())

			// give permission to with object in lowercase form (old format)
			roleReq := rmap.NewFromMap(map[string]interface{}{
				"name": "Function executer",
				"grants": []map[string]interface{}{{
					"object": "/function/invoke/mockfunc",
					"action": "execute",
				}},
			})

			tctx.SetActor("superUser")
			roleID := MustGetID(tctx.Rmap("assetCreate", "role", roleReq.Bytes(), -1, ""))
			grantReq := rmap.NewFromMap(map[string]interface{}{"roles": []string{roleID}})
			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), grantReq.Bytes())

			// ordinaryUser must be able to execute now
			tctx.SetActor("ordinaryUser")
			tctx.Ok("functionInvoke", "MockFunc", rmap.NewEmpty().Bytes())

			// give permission with object in mixed case form
			roleReq = rmap.NewFromMap(map[string]interface{}{
				"name": "Function executer",
				"grants": []map[string]interface{}{{
					"object": "/function/invoke/MockFunc",
					"action": "execute",
				}},
			})
			tctx.SetActor("superUser")
			tctx.Ok("assetUpdate", "role", roleID, roleReq.Bytes())

			// ordinaryUser must be able to execute now
			tctx.SetActor("ordinaryUser")
			tctx.Ok("functionInvoke", "MockFunc", rmap.NewEmpty().Bytes())
		})
	})
})
