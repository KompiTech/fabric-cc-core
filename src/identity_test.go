package cc_core

import (
	"fmt"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/src/konst"
	. "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("identity* method family tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
	})

	Describe("Call to CC method identityAddMe", func() {
		It("Should create superuser role and grant it to init_manager", func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			myIdentity := tctx.Rmap("identityAddMe", rmap.NewEmpty().Bytes())

			// current role has SuperUser
			Expect(myIdentity.Mapa).To(HaveKeyWithValue("is_enabled", true))
			Expect(myIdentity.Mapa).To(HaveKeyWithValue("fingerprint", tctx.GetCurrentActorFingerprint()))
			Expect(myIdentity.Mapa).To(HaveKey("roles"))
			Expect(myIdentity.Mapa["roles"]).To(ContainElement(konst.SuperuserRoleUUID))

			// SuperUser role was created
			suRole := tctx.Rmap("assetGet", "role", konst.SuperuserRoleUUID, false, "")
			Expect(suRole.Mapa).To(HaveKeyWithValue("name", "Superuser"))
		})

		It("Should allow to be called repeatedly", func() {
			tctx.Ok("identityAddMe", rmap.NewEmpty().Bytes())
			tctx.Ok("identityAddMe", rmap.NewEmpty().Bytes())
		})

		It("Should grant SU even when identity already exists", func() {
			// init_manager is SuperUser
			tctx.InitOk(tctx.GetInit("", "").Bytes())

			// ordinaryUser identity created, no SU granted
			tctx.SetActor("ordinaryUser")
			ordinaryIdentity := tctx.Rmap("identityAddMe", rmap.NewEmpty().Bytes())
			Expect(ordinaryIdentity.Mapa).To(Not(HaveKey("roles")))

			// init_manager is ordinaryUser
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			ordinaryIdentity = tctx.Rmap("identityAddMe", rmap.NewEmpty().Bytes())
			Expect(ordinaryIdentity.Mapa).To(HaveKey(konst.RolesKey))

			// ordinaryUser SU granted for existing identity
			roles := ordinaryIdentity.MustGetIterable(konst.RolesKey)
			Expect(roles).To(ConsistOf(konst.SuperuserRoleUUID))
		})
	})

	Describe("Call to CC method identityMe", func() {
		Context("Without existing identity", func() {
			It("Should return particular error message", func() {
				tctx.InitOk(tctx.GetInit("", "").Bytes())
				tctx.Error(fmt.Sprintf("cannot get identity for fingerprint: %s. Did you call identityAddMe?", tctx.GetCurrentActorFingerprint()), "identityMe", false)
			})
		})

		It("Should return information about current identity", func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.Ok("identityAddMe", rmap.NewEmpty().Bytes())

			myIdentity := tctx.Rmap("identityMe", false)
			Expect(myIdentity.Mapa).To(HaveKeyWithValue("is_enabled", true))
			Expect(myIdentity.Mapa).To(HaveKeyWithValue("fingerprint", tctx.GetCurrentActorFingerprint()))
		})

		It("Should be callable without any permissions", func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.SetActor("ordinaryUser")
			tctx.Ok("identityAddMe", rmap.NewEmpty().Bytes())

			myIdentity := tctx.Rmap("identityMe", false)
			Expect(myIdentity.Mapa).To(HaveKeyWithValue("is_enabled", true))
			Expect(myIdentity.Mapa).To(HaveKeyWithValue("fingerprint", tctx.GetCurrentActorFingerprint()))
		})
	})

	Describe("Call to CC method assetUpdate(identity, ...)", func() {
		var grantSU, grantRole, revoke, disable rmap.Rmap

		BeforeSuite(func() {
			grantSU = rmap.NewFromMap(map[string]interface{}{"roles": []string{konst.SuperuserRoleUUID}})
			revoke = rmap.NewFromMap(map[string]interface{}{"roles": []string{}})
			disable = rmap.NewFromMap(map[string]interface{}{"is_enabled": false})
		})

		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()

			newRole := rmap.NewFromMap(map[string]interface{}{
				"name": "Identity updater",
				"grants": []map[string]interface{}{{
					"object": "/identity/*",
					"action": "update",
				}},
			})

			// creates role that can update identity
			newRole = tctx.Rmap("assetCreate", "role", newRole.Bytes(), -1, "")

			grantRole = rmap.NewFromMap(map[string]interface{}{
				"roles": []string{newRole.Mapa["uuid"].(string)},
			})
			// grants this role to ordinaryUser
			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), grantRole.Bytes())
		})

		It("Should allow to grant SU to somebody else", func() {
			tctx.SetActor("nobodyUser")
			ordIdentity := tctx.Rmap("identityMe", false)
			Expect(ordIdentity.Mapa).To(Not(HaveKey("roles")))

			tctx.SetActor("superUser")
			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("nobodyUser"), grantSU.Bytes())

			tctx.SetActor("nobodyUser")
			ordIdentity = tctx.Rmap("identityMe", false)
			Expect(ordIdentity.Mapa).To(HaveKey("roles"))
			Expect(ordIdentity.Mapa["roles"]).To(ContainElement(konst.SuperuserRoleUUID))
		})

		It("Should return error if attempting to remove last superuser", func() {
			tctx.SetActor("superUser")
			tctx.Error("identityExtraValidate() failed: unable to remove last superuser role", "assetUpdate", "identity", tctx.GetCurrentActorFingerprint(), revoke.Bytes())
		})

		It("Should return error if attempting to remove last superuser by clearing all roles", func() {
			removeAllRoles := rmap.NewFromMap(map[string]interface{}{"roles": nil})
			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), removeAllRoles.Bytes())
		})

		It("Should return error if attempting to remove last superuser by clearing all roles", func() {
			removeAllRoles := rmap.NewFromMap(map[string]interface{}{"roles": nil})
			tctx.SetActor("superUser")
			tctx.Error("identityExtraValidate() failed: unable to remove last superuser role", "assetUpdate", "identity", tctx.GetCurrentActorFingerprint(), removeAllRoles.Bytes())
		})

		It("Should require permission", func() {
			tctx.SetActor("nobodyUser")
			tctx.Error("permission denied", "assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), grantSU.Bytes())
		})

		It("Should return error if somebody without SU attempts to update SU", func() {
			tctx.SetActor("ordinaryUser")
			// ordinaryUser cannot grant SU role to nobodyUser
			tctx.Error("identityExtraValidate() failed: to manage SuperUser role, you must have it granted", "identityUpdate", tctx.GetActorFingerprint("nobodyUser"), grantSU.Bytes())
			// ordinaryUser cannot revoke SU role from superUser
			tctx.Error("identityExtraValidate() failed: to manage SuperUser role, you must have it granted", "identityUpdate", tctx.GetActorFingerprint("superUser"), revoke.Bytes())
			// ordinaryUser cannot disable SU role
			tctx.Error("identityExtraValidate() failed: to manage SuperUser role, you must have it granted", "identityUpdate", tctx.GetActorFingerprint("superUser"), disable.Bytes())
			// ordinaryUser can grant non-SU role to nobodyUser because he has identityUpdate grant
			tctx.Ok("identityUpdate", tctx.GetActorFingerprint("nobodyUser"), grantRole.Bytes())
		})

		It("Should return error if setting is_enabled=false on own identity asset", func() {
			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("ordinaryUser"), grantSU.Bytes())
			tctx.Error("identityExtraValidate() failed: unable to set is_enabled to false on own identity", "assetUpdate", "identity", tctx.GetCurrentActorFingerprint(), disable.Bytes())
		})

		It("Should return error if setting is_enabled=false on own identity and it is the last SU", func() {
			tctx.Error("identityExtraValidate() failed: unable to set is_enabled to false on own identity", "assetUpdate", "identity", tctx.GetCurrentActorFingerprint(), disable.Bytes())
		})
	})

	Describe("Call to CC method assetGet(identity, ...) with existing IDENTITY", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should require read permission when reading non-own asset", func() {
			tctx.SetActor("ordinaryUser")
			msg := fmt.Sprintf("permission denied, sub: %s, obj: /identity/%s, act: read", tctx.GetCurrentActorFingerprint(), tctx.GetActorFingerprint("superUser"))
			tctx.Error(msg, "assetGet", "identity", tctx.GetActorFingerprint("superUser"), false, "")
		})

		It("Should allow access through legacy identityGet method", func() {
			tctx.Ok("assetGet", "identity", tctx.GetCurrentActorFingerprint(), false, "")
		})
	})

	Describe("Call to CC method assetQuery(identity, ...) with existing IDENTITY", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should filter returned assets", func() {
			queryAll := rmap.NewFromMap(map[string]interface{}{
				"selector": map[string]interface{}{},
			})

			// SU can see every identity
			result := tctx.RmapNoResult("assetQuery", "identity", queryAll.Bytes(), false)
			Expect(result.Mapa).To(HaveKey("result"))
			for _, identityI := range result.MustGetJPtrIterable("/result") {
				identity, err := rmap.NewFromInterface(identityI)
				Expect(err).To(BeNil())
				Expect(identity.Mapa).To(Not(HaveKey("error")))
			}

			// ordinary can see only his own identity
			tctx.SetActor("ordinaryUser")
			result = tctx.RmapNoResult("assetQuery", "identity", queryAll.Bytes(), false)
			Expect(result.Mapa).To(HaveKey("result"))
			seenOwn := false
			for _, identityI := range result.MustGetJPtrIterable("/result") {
				identity, err := rmap.NewFromInterface(identityI)
				Expect(err).To(BeNil())
				if !identity.MustExistsJPtr("/error") {
					if seenOwn {
						Fail("Seen multiple identities, error")
					}
					// no error, this must be own identity asset
					Expect(identity.Mapa).To(HaveKeyWithValue("fingerprint", tctx.GetCurrentActorFingerprint()))
					seenOwn = true
				} else {
					Expect(identity.Mapa).To(HaveKey("error"))
				}
			}
			Expect(seenOwn).To(BeTrue())

			// when granted, nobodyUser can see own + granted identity
			identityReader := rmap.NewFromMap(map[string]interface{}{
				"name": "Identity reader",
				"grants": []map[string]interface{}{
					{
						"object": fmt.Sprintf("/identity/%s", tctx.GetActorFingerprint("ordinaryUser")),
						"action": "read",
					},
				},
			})
			// SU grants creates reader role and grants it to nobodyUser
			tctx.SetActor("superUser")
			identityReader = tctx.Rmap("assetCreate", "role", identityReader.Bytes(), -1, "")
			grantIdentityReader := rmap.NewFromMap(map[string]interface{}{
				"roles": []string{identityReader.MustGetJPtrString("/uuid")},
			})

			tctx.Ok("assetUpdate", "identity", tctx.GetActorFingerprint("nobodyUser"), grantIdentityReader.Bytes())
			// nobodyUser must see own identity + granted one
			tctx.SetActor("nobodyUser")
			result = tctx.RmapNoResult("assetQuery", "identity", queryAll.Bytes(), false)
			Expect(result.Mapa).To(HaveKey("result"))
			seenOwn = false
			seenGranted := false
			for _, identityI := range result.MustGetJPtrIterable("/result") {
				identity, err := rmap.NewFromInterface(identityI)
				Expect(err).To(BeNil())
				if !identity.MustExistsJPtr("/error") {
					fp := identity.MustGetJPtrString("/fingerprint")
					if fp == tctx.GetCurrentActorFingerprint() && !seenOwn {
						seenOwn = true
					} else if fp == tctx.GetActorFingerprint("ordinaryUser") && !seenGranted {
						seenGranted = true
					} else {
						Fail("Test failed")
					}
				}
			}
			Expect(seenOwn).To(BeTrue())
			Expect(seenGranted).To(BeTrue())
		})

		It("Should allow access through legacy identityQuery method", func() {
			tctx.Ok("identityQuery", "", false)
		})
	})

	Describe("Call to CC method identityCreateDirect()", func() {
		customID := strings.Repeat("f", 128)
		identityData := rmap.NewFromMap(map[string]interface{}{
			"is_enabled": true,
		})

		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should require permission", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "identityCreateDirect", rmap.NewEmpty().Bytes(), customID)
		})

		It("Should create identity asset with id specified in CC param", func() {
			response := tctx.Rmap("identityCreateDirect", identityData.Bytes(), customID)
			Expect(response.Mapa).To(HaveKeyWithValue("is_enabled", true))
			Expect(response.Mapa).To(HaveKeyWithValue(konst.AssetFingerprintKey, customID))
		})

		It("Should create identity asset with fp specified in data", func() {
			myIdentityData := identityData.Copy()
			myIdentityData.Mapa[konst.AssetFingerprintKey] = customID

			response := tctx.Rmap("identityCreateDirect", myIdentityData.Bytes(), "")
			Expect(response.Mapa).To(HaveKeyWithValue("is_enabled", true))
			Expect(response.Mapa).To(HaveKeyWithValue(konst.AssetFingerprintKey, customID))
		})

		It("Should create identity asset with the same fp specified in data and CC param", func() {
			myIdentityData := identityData.Copy()
			myIdentityData.Mapa[konst.AssetFingerprintKey] = customID

			response := tctx.Rmap("identityCreateDirect", myIdentityData.Bytes(), customID)
			Expect(response.Mapa).To(HaveKeyWithValue("is_enabled", true))
			Expect(response.Mapa).To(HaveKeyWithValue(konst.AssetFingerprintKey, customID))
		})

		It("Should return error if id specified in data in param is not the same", func() {
			wrongID := strings.Repeat("a", 128)

			myIdentityData := identityData.Copy()
			myIdentityData.Mapa[konst.AssetFingerprintKey] = customID
			tctx.Error(fmt.Sprintf("fingerprint from CC param: %s does not match fingerprint in data: %s", wrongID, customID), "identityCreateDirect", myIdentityData.Bytes(), wrongID)
		})

		It("Should return error if no id is specified in data and CC param", func() {
			tctx.Error("identity fingerprint cannot be autogenerated, you must send one in data or in CC param", "identityCreateDirect", identityData.Bytes(), "")
		})

		It("Should return error if attempting to create using assetCreate variant", func() {
			tctx.Error("identity can only be created with identityCreateDirect method", "assetCreate", "identity", rmap.NewEmpty().Bytes(), 1, customID)
		})

		It("Should return error if using assetCreateDirect with invalid version", func() {
			tctx.Error("identity must be created with version 1", "assetCreateDirect", "identity", rmap.NewEmpty().Bytes(), 2, customID)
		})
	})

	Describe("Call to CC method identityUpdateDirect() with existing identity asset", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should require permission", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "identityUpdateDirect", tctx.GetActorFingerprint("superUser"), rmap.NewEmpty().Bytes())
		})

		It("Should not allow last SU to remove superRole", func() {
			req := rmap.NewFromMap(map[string]interface{}{
				"roles": nil,
			})
			tctx.Error("identityExtraValidate() failed: unable to remove last superuser role", "identityUpdateDirect", tctx.GetActorFingerprint("superUser"), req.Bytes())
		})
	})

	Describe("Call to CC method assetDeleteDirect(identity, ...) with existing identity asset", func() {
		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.RegisterAllActors()
		})

		It("Should delete identity asset and leave related user asset untouched", func() {
			tctx.Ok("assetDeleteDirect", "identity", tctx.GetActorFingerprint("ordinaryUser"))
			tctx.SetActor("ordinaryUser")
			// identity was deleted
			tctx.Error("cannot get identity for fingerprint: "+tctx.GetActorFingerprint("ordinaryUser")+". Did you call identityAddMe?", "identityMe", false)
		})
	})

	Describe("Call to any CC method except identityAddMe without existing identity for client", func() {
		var expectedError string

		BeforeEach(func() {
			tctx.InitOk(tctx.GetInit("", "").Bytes())
			tctx.SetActor("ordinaryUser")
			expectedError = "cannot get identity for fingerprint: " + tctx.GetCurrentActorFingerprint() + ". Did you call identityAddMe?"
		})

		Context("asset* method family", func() {
			It("Should return friendly message - assetCreate", func() {
				tctx.Error(expectedError, "assetCreate", "some_asset", rmap.NewEmpty().Bytes(), -1, "")
			})

			It("Should return friendly message - assetCreateDirect", func() {
				tctx.Error(expectedError, "assetCreateDirect", "some_asset", rmap.NewEmpty().Bytes(), -1, "")
			})

			It("Should return friendly message - assetDelete", func() {
				tctx.Error(expectedError, "assetDelete", "some_asset", "some_id")
			})

			It("Should return friendly message - assetGet", func() {
				tctx.Error(expectedError, "assetGet", "some_asset", "some_id", false, "")
			})

			It("Should return friendly message - assetHistory", func() {
				tctx.Error(expectedError, "assetHistory", "some_asset", "some_id")
			})

			It("Should return friendly message - assetMigrate", func() {
				tctx.Error(expectedError, "assetMigrate", "some_asset", "some_id", "", 1)
			})

			It("Should return friendly message - assetUpdate", func() {
				tctx.Error(expectedError, "assetUpdate", "some_asset", "some_id", rmap.NewFromMap(map[string]interface{}{"some": "change"}).Bytes())
			})

			It("Should return friendly message - assetUpdateDirect", func() {
				tctx.Error(expectedError, "assetUpdateDirect", "some_asset", "some_id", rmap.NewFromMap(map[string]interface{}{"some": "change"}).Bytes())
			})

			It("Should return friendly message - assetQuery", func() {
				tctx.Error(expectedError, "assetQuery", "some_asset", "", false)
			})
		})

		Context("changelog* method family", func() {
			It("Should return friendly message - changelogGet", func() {
				tctx.Error(expectedError, "changelogGet", 0)
			})

			It("Should return friendly message - changelogList", func() {
				tctx.Error(expectedError, "changelogList")
			})
		})

		Context("function* method family", func() {
			It("Should return friendly message - functionInvoke", func() {
				tctx.Error(expectedError, "functionInvoke", "some_func", "")
			})

			It("Should return friendly message - functionQuery", func() {
				tctx.Error(expectedError, "functionQuery", "some_func", "")
			})
		})

		Context("identity* method family", func() {
			It("Should return friendly message - identityGet", func() {
				tctx.Error(expectedError, "identityGet", "some_fp", false, "")
			})

			It("Should return friendly message - identityMe", func() {
				tctx.Error(expectedError, "identityMe", false)
			})

			It("Should return friendly message - identityUpdate", func() {
				tctx.Error(expectedError, "identityUpdate", "some_fp", rmap.NewFromMap(map[string]interface{}{"some": "change"}).Bytes())
			})

			It("Should return friendly message - identityQuery", func() {
				tctx.Error(expectedError, "identityQuery", "", false)
			})

			It("Should return friendly message - identityCreateDirect", func() {
				tctx.Error(expectedError, "identityCreateDirect", "", "some_fp")
			})

			It("Should return friendly message - identityUpdateDirect", func() {
				tctx.Error(expectedError, "identityUpdateDirect", "some_fp", "")
			})
		})

		Context("registry* method family", func() {
			It("Should return friendly message - registryGet", func() {
				tctx.Error(expectedError, "registryGet", "some_asset", -1)
			})

			It("Should return friendly message - registryUpsert", func() {
				tctx.Error(expectedError, "registryUpsert", "some_asset", rmap.NewFromMap(map[string]interface{}{"some": "change"}).Bytes())
			})

			It("Should return friendly message - registryList", func() {
				tctx.Error(expectedError, "registryList")
			})
		})

		Context("role* method family", func() {
			It("Should return friendly message - roleGet", func() {
				tctx.Error(expectedError, "roleGet", "some_id", "")
			})

			It("Should return friendly message - roleCreate", func() {
				tctx.Error(expectedError, "roleCreate", "", -1)
			})

			It("Should return friendly message - roleUpdate", func() {
				tctx.Error(expectedError, "roleUpdate", "some_id", rmap.NewFromMap(map[string]interface{}{"some": "change"}).Bytes())
			})

			It("Should return friendly message - roleQuery", func() {
				tctx.Error(expectedError, "roleQuery", "")
			})
		})

		Context("singleton* method family", func() {
			It("Should return friendly message - singletonGet", func() {
				tctx.Error(expectedError, "singletonGet", "some_asset", -1)
			})

			It("Should return friendly message - singletonUpsert", func() {
				tctx.Error(expectedError, "singletonUpsert", "some_asset", rmap.NewFromMap(map[string]interface{}{"some": "change"}).Bytes())
			})
		})
	})
})
