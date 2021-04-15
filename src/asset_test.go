package cc_core

import (
	"fmt"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/fabric-cc-core/v2/src/konst"
	. "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("asset* method family tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("testdata/assets", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("Call to CC method assetCreate", func() {
		Context("When using private data destination", func() {
			text := "hello world"
			key := "text"
			commentReq := rmap.NewFromMap(map[string]interface{}{key: text})

			It("Should autogenerate ID", func() {
				comment := tctx.Rmap("assetCreate", "mockcomment", commentReq.Bytes(), -1, "")
				Expect(comment.Mapa).To(HaveKeyWithValue(key, text))
				id := MustGetID(comment)

				comment = tctx.Rmap("assetGet", "mockcomment", id, false, "")
				Expect(comment.Mapa).To(HaveKeyWithValue(key, text))
				Expect(comment.Mapa).To(HaveKeyWithValue(konst.AssetIdKey, id))
			})

			It("Should use ID from client", func() {
				uuid := "0d5fb738-6511-4109-bd13-61dd1a33bcc5"
				comment := tctx.Rmap("assetCreate", "mockcomment", commentReq.Bytes(), -1, uuid)
				Expect(comment.Mapa).To(HaveKeyWithValue(key, text))
				Expect(comment.Mapa).To(HaveKeyWithValue(konst.AssetIdKey, uuid))

				comment = tctx.Rmap("assetGet", "mockcomment", uuid, false, "")
				Expect(comment.Mapa).To(HaveKeyWithValue(key, text))
				Expect(comment.Mapa).To(HaveKeyWithValue(konst.AssetIdKey, uuid))
			})

			It("Should return error on duplicate ID", func() {
				uuid := "aa0ad033-1cee-4508-bbc3-c6c53703d6f0"
				tctx.Ok("assetCreate", "mockcomment", commentReq.Bytes(), -1, uuid)
				tctx.Error("private data key already exists: MOCKCOMMENT"+uuid, "assetCreate", "mockcomment", commentReq.Bytes(), -1, uuid)
			})
		})

		Context("When using state destination", func() {
			description := "hello incident"
			key := "description"
			incidentReq := rmap.NewFromMap(map[string]interface{}{key: description})

			It("Should autogenerate ID", func() {
				incident := tctx.Rmap("assetCreate", "mockincident", incidentReq.Bytes(), -1, "")
				Expect(incident.Mapa).To(HaveKeyWithValue(key, description))
				id := MustGetID(incident)

				incident = tctx.Rmap("assetGet", "mockincident", id, false, "")
				Expect(incident.Mapa).To(HaveKeyWithValue(key, description))
				Expect(incident.Mapa).To(HaveKeyWithValue(konst.AssetIdKey, id))
			})

			It("Should use ID from client", func() {
				uuid := "0d5fb738-6511-4109-bd13-61dd1a33bcc5"
				incident := tctx.Rmap("assetCreate", "mockincident", incidentReq.Bytes(), -1, uuid)
				Expect(incident.Mapa).To(HaveKeyWithValue(key, description))
				Expect(incident.Mapa).To(HaveKeyWithValue(konst.AssetIdKey, uuid))

				incident = tctx.Rmap("assetGet", "mockincident", uuid, false, "")
				Expect(incident.Mapa).To(HaveKeyWithValue(key, description))
				Expect(incident.Mapa).To(HaveKeyWithValue(konst.AssetIdKey, uuid))
			})

			It("Should return error on duplicate ID", func() {
				uuid := "aa0ad033-1cee-4508-bbc3-c6c53703d6f0"
				tctx.Ok("assetCreate", "mockincident", incidentReq.Bytes(), -1, uuid)
				tctx.Error("reg.PutAsset() failed: putRmapToState() failed: state key already exists: MOCKINCIDENT"+uuid, "assetCreate", "mockincident", incidentReq.Bytes(), -1, uuid)
			})
		})

		Context("When referencing more than cacheSize assets", func() {
			It("Should validate references properly", func() {
				parentID := tctx.Rmap("assetCreate", "mockworknoteparent", rmap.NewEmpty().Bytes(), -1, "").MustGetString("uuid")

				childData := rmap.NewFromMap(map[string]interface{}{
					"text":   "Hi",
					"entity": "mockworknoteparent:" + parentID,
				})

				for i := 0; i <= engine.CacheSize; i++ {
					tctx.Ok("assetCreate", "mockworknote", childData.Bytes(), -1, "")
				}
			})
		})

		Context("When referencing other assets", func() {
			user1 := rmap.NewFromMap(map[string]interface{}{
				"name":    "John",
				"surname": "Doe",
			})
			user1uuid := ""

			user2 := rmap.NewFromMap(map[string]interface{}{
				"name":    "Lenny",
				"surname": "Kravitz",
			})
			user2uuid := ""

			BeforeEach(func() {
				user1uuid = MustGetID(tctx.Rmap("assetCreate", "mockuser", user1.Bytes(), -1, ""))
				user2uuid = MustGetID(tctx.Rmap("assetCreate", "mockuser", user2.Bytes(), -1, ""))
			})

			It("Should return error on invalid singular reference", func() {
				incidentReq := rmap.NewFromMap(map[string]interface{}{"assigned_to": "WRONG", "description": "ABCD"})
				tctx.Error("Referenced asset 'mockuser' with ID 'wrong' not found", "assetCreate", "mockincident", incidentReq.Bytes(), -1, "")
			})

			It("Should create singular reference and resolve it", func() {
				incidentReq := rmap.NewFromMap(map[string]interface{}{"assigned_to": user1uuid, "description": "ABCD"})
				incidentID := MustGetID(tctx.Rmap("assetCreate", "mockincident", incidentReq.Bytes(), -1, ""))

				resolvedIncident := tctx.Rmap("assetGet", "mockincident", incidentID, true, "")
				Expect(resolvedIncident.Mapa).To(HaveKey("assigned_to"))
				resolvedName := resolvedIncident.MustGetJPtrString("/assigned_to/name")
				resolvedSurname := resolvedIncident.MustGetJPtrString("/assigned_to/surname")
				Expect(resolvedName).To(Equal(user1.Mapa["name"]))
				Expect(resolvedSurname).To(Equal(user1.Mapa["surname"]))
			})

			It("Should return error on invalid array of references", func() {
				incidentReq := rmap.NewFromMap(map[string]interface{}{"additional_assignees": []interface{}{"very", "wrong"}, "description": "ABCD"})
				tctx.Error("Referenced asset 'mockuser' with ID 'very' not found (currently resolved asset name: mockincident, uuid: f78147d1-2c0e-d793-3500-e7076428db7d),Referenced asset 'mockuser' with ID 'wrong' not found (currently resolved asset name: mockincident, uuid: f78147d1-2c0e-d793-3500-e7076428db7d)", "assetCreate", "mockincident", incidentReq.Bytes(), -1, "f78147d1-2c0e-d793-3500-e7076428db7d")
			})

			It("Should create array of references and resolve it", func() {
				incidentReq := rmap.NewFromMap(map[string]interface{}{"additional_assignees": []interface{}{user1uuid, user2uuid}, "description": "ABCD"})
				incidentID := MustGetID(tctx.Rmap("assetCreate", "mockincident", incidentReq.Bytes(), -1, ""))

				resolvedIncident := tctx.Rmap("assetGet", "mockincident", incidentID, true, "")
				Expect(resolvedIncident.Mapa).To(HaveKey("additional_assignees"))
				resolvedName := resolvedIncident.MustGetJPtrString("/additional_assignees/0/name")
				resolvedSurname := resolvedIncident.MustGetJPtrString("/additional_assignees/0/surname")
				Expect(resolvedName).To(Equal(user1.Mapa["name"]))
				Expect(resolvedSurname).To(Equal(user1.Mapa["surname"]))

				resolvedName = resolvedIncident.MustGetJPtrString("/additional_assignees/1/name")
				resolvedSurname = resolvedIncident.MustGetJPtrString("/additional_assignees/1/surname")
				Expect(resolvedName).To(Equal(user2.Mapa["name"]))
				Expect(resolvedSurname).To(Equal(user2.Mapa["surname"]))
			})

			It("Should return error on invalid ref on nested level", func() {
				reqNestedRef := rmap.NewFromMap(map[string]interface{}{
					"expenses": []map[string]interface{}{{
						"user": "XyZ",
					}},
				})
				tctx.Error("Referenced asset 'mockuser' with ID 'xyz' not found", "assetCreate", "mocknestedref", reqNestedRef.Bytes(), -1, "")
			})

			It("Should create and resolve reference on nested level", func() {
				reqNestedRef := rmap.NewFromMap(map[string]interface{}{
					"expenses": []map[string]interface{}{{
						"user": user1uuid,
					}},
				})
				id := MustGetID(tctx.Rmap("assetCreate", "mocknestedref", reqNestedRef.Bytes(), -1, ""))

				resolved := tctx.Rmap("assetGet", "mocknestedref", id, true, "")
				Expect(resolved.Mapa).To(HaveKey("expenses"))

				resolvedUser := resolved.MustGetJPtrRmap("/expenses/0/user")
				Expect(resolvedUser.Mapa).To(HaveKeyWithValue("name", user1.Mapa["name"]))
				Expect(resolvedUser.Mapa).To(HaveKeyWithValue("surname", user1.Mapa["surname"]))
			})

			It("Should return error if ref array contains duplicate", func() {
				req := rmap.NewFromMap(map[string]interface{}{
					"description":          "lul",
					"additional_assignees": []string{user1uuid, user1uuid},
				})
				tctx.Error(`reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident`, "assetCreate", "mockincident", req.Bytes(), -1, "")
			})

			It("Should return error if entityref is pointing to invalid asset", func() {
				commentReq := rmap.NewFromMap(map[string]interface{}{
					"text":   "Hello comment",
					"entity": "MOCKINCIDENT:ABCDEF",
				})
				tctx.Error("Referenced asset 'mockincident' with ID 'abcdef' not found", "assetCreate", "mockcomment", commentReq.Bytes(), -1, "")
			})

			It("Should create and resolve entityref", func() {
				incidentReq := rmap.NewFromMap(map[string]interface{}{"description": "hello incident"})
				incidentID := MustGetID(tctx.Rmap("assetCreate", "mockincident", incidentReq.Bytes(), -1, ""))
				commentReq := rmap.NewFromMap(map[string]interface{}{
					"text":   "Hello comment",
					"entity": fmt.Sprintf("MOCKINCIDENT:%s", incidentID),
				})
				commentID := MustGetID(tctx.Rmap("assetCreate", "mockcomment", commentReq.Bytes(), -1, ""))

				comment := tctx.Rmap("assetGet", "mockcomment", commentID, true, "")
				resolvedIncident := comment.MustGetJPtrRmap("/entity").Mapa
				Expect(resolvedIncident).To(HaveKeyWithValue("description", incidentReq.MustGetJPtrString("/description")))
			})
		})

		Context("When using pre draft-07 JSONSchema asset and compatibility enabled", func() {
			It("Should transparently allow schema to work", func() {
				tctx.Ok("assetCreate", "mocklegacyschema", rmap.NewFromMap(map[string]interface{}{"price": "0.91"}), -1, "")
			})
		})

		It("Should return error if attempting to set service keys", func() {
			req := rmap.NewFromMap(map[string]interface{}{
				konst.AssetVersionKey: -5,
				konst.AssetDocTypeKey: "INCAdenT",
				"description":         "xyz",
			})
			tctx.Error("patch contains service key(s)", "assetCreate", "mockincident", req.Bytes(), -1, "")
		})

		It("Should work with empty input", func() {
			// schema error, NOT unmarshal error, working as intended
			tctx.Error("reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident", "assetCreate", "mockincident", "", -1, "")
		})
	})

	Describe("Call to CC method assetCreateDirect", func() {
		id := "9bfb0402-a092-a9a2-7542-8bf37cb29a2e"
		name := "mockblogicfail"

		Context("When referencing other assets", func() {
			It("Should create new asset without checking reference", func() {
				incidentReq := rmap.NewFromMap(map[string]interface{}{"assigned_to": "WRONG", "description": "ABCD"})
				tctx.Ok("assetCreateDirect", "mockincident", incidentReq.Bytes(), -1, "")
			})
		})

		It("Should not execute any business logic", func() {
			tctx.Error("business logic created fail", "assetCreate", name, rmap.NewEmpty().Bytes(), -1, id) // assetCreate throws
			tctx.Ok("assetCreateDirect", name, rmap.NewEmpty().Bytes(), -1, id)                             // assetCreateDirect does not
		})
	})

	Describe("Call to CC method assetMigrate", func() {
		// regItemV1 contains mandatory description field (already initialized in CC)
		regItemV1 := rmap.MustNewFromYAMLFile("testdata/assets/mockincident.yaml")

		// regItemV2 does not contain description field anymore, but it contains mandatory short_description
		regItemV2 := regItemV1.Copy()
		regItemV2.MustDeleteJPtr("/schema/properties/description")
		regItemV2.MustSetJPtr("/schema/properties/short_description", map[string]interface{}{
			"description": "Short description of the Incident",
			"type":        "string",
		})
		regItemV2.MustSetJPtr("/schema/required", []interface{}{"short_description"})

		BeforeEach(func() {
			// add second version of mockincident to CC
			init := rmap.NewEmpty()
			init.MustSetJPtr("/registries", rmap.NewEmpty())
			init.MustSetJPtr("/registries/mockincident", regItemV2)
			tctx.InitOk(init.Bytes())
		})

		It("Should return error if attempting to set service key(s)", func() {
			reqV1 := rmap.NewFromMap(map[string]interface{}{"description": "foobar"})
			v1ID := MustGetID(tctx.Rmap("assetCreate", "mockincident", reqV1.Bytes(), 1, ""))

			req := rmap.NewFromMap(map[string]interface{}{
				konst.AssetIdKey:      "iWillBreakIt",
				konst.AssetVersionKey: 1337,
				konst.AssetDocTypeKey: "DOKTAJP",
				"description":         "incosh",
			})
			tctx.Error("patch contains service key(s)", "assetMigrate", "mockincident", v1ID, req.Bytes(), 2)
		})

		It("Should allow to migrate assets between versions", func() {
			// attempt to create asset of version 2 with invalid data for schema
			reqV1 := rmap.NewFromMap(map[string]interface{}{"description": "foobar"})
			tctx.Error("reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident", "assetCreate", "mockincident", reqV1.Bytes(), 2, "")
			tctx.Error("reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident", "assetCreate", "mockincident", reqV1.Bytes(), -1, "")

			// attempt to create asset of version 1 with invalid data for schema
			reqV2 := rmap.NewFromMap(map[string]interface{}{"short_description": "foobar"})
			tctx.Error("reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident", "assetCreate", "mockincident", reqV2.Bytes(), 1, "")

			// create asset of version 1
			v1ID := MustGetID(tctx.Rmap("assetCreate", "mockincident", reqV1.Bytes(), 1, ""))
			Expect(MustGetVersion(tctx.Rmap("assetGet", "mockincident", v1ID, false, ""))).To(Equal(1))

			// create asset of version 2
			v2ID := MustGetID(tctx.Rmap("assetCreate", "mockincident", reqV2.Bytes(), 2, ""))
			Expect(MustGetVersion(tctx.Rmap("assetGet", "mockincident", v2ID, false, ""))).To(Equal(2))

			// attempt to migrate from/to same version should return error
			tctx.Error("unable to migrate to the same version of asset", "assetMigrate", "mockincident", v1ID, rmap.NewEmpty().Bytes(), 1)

			// attempt to migrate 1 -> 2 with empty patch must fail on schema validation
			tctx.Error("reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident", "assetMigrate", "mockincident", v1ID, rmap.NewEmpty().Bytes(), 2)

			// attempt to migrate 2 -> 1 with empty patch must fail on schema validation
			tctx.Error("reg.PutAsset() failed: asset.ValidateSchema() failed on assetName: mockincident", "assetMigrate", "mockincident", v2ID, rmap.NewEmpty().Bytes(), 1)

			// migrate 1 -> 2 (-1 is synonymous for latest), remove description, add short_description
			migrateV1V2 := rmap.NewFromMap(map[string]interface{}{
				"description":       nil,
				"short_description": "foobar very migrated",
			})
			migratedV2 := tctx.Rmap("assetMigrate", "mockincident", v1ID, migrateV1V2.Bytes(), -1)
			Expect(MustGetVersion(migratedV2)).To(Equal(2))
			Expect(migratedV2.Mapa).To(Not(HaveKey("description")))
			Expect(migratedV2.Mapa).To(HaveKeyWithValue("short_description", migrateV1V2.Mapa["short_description"]))

			// migrate 2 -> 1, remove short_description, add description
			migrateV2V1 := rmap.NewFromMap(map[string]interface{}{
				"description":       "foobar very migrated",
				"short_description": nil,
			})
			migratedV1 := tctx.Rmap("assetMigrate", "mockincident", v2ID, migrateV2V1.Bytes(), 1)
			Expect(MustGetVersion(migratedV1)).To(Equal(1))
			Expect(migratedV1.Mapa).To(Not(HaveKey("short_description")))
			Expect(migratedV1.Mapa).To(HaveKeyWithValue("description", migrateV2V1.Mapa["description"]))
		})
	})

	Describe("Call to CC method assetQuery", func() {
		Context("Without any assets", func() {
			It("Should return empty result", func() {
				response := tctx.RmapNoResult("assetQuery", "mockincident", rmap.NewEmpty().Bytes(), false)
				Expect(response.Mapa).To(HaveKey("result"))
				result := response.MustGetIterable("result")
				Expect(result).To(HaveLen(0))
			})
		})

		Context("When more assets in than pageSize is present", func() {
			BeforeEach(func() {
				for i := 0; i < 15; i++ {
					tctx.Ok("assetCreate", "mockincident", rmap.NewFromMap(map[string]interface{}{"description": "mockIncident"}).Bytes(), -1, "")
				}
			})

			It("Should return pages with at most pageSize result - state", func() {
				response := tctx.RmapNoResult("assetQuery", "mockincident", "", false)
				bookmark := response.MustGetString("bookmark")

				// first page, pageSize results
				Expect(response.MustGetIterable("result")).To(HaveLen(konst.PageSize))

				//second page, rest of results (5)
				response = tctx.RmapNoResult("assetQuery", "mockincident", rmap.NewFromMap(map[string]interface{}{"bookmark": bookmark}), false)
				Expect(response.MustGetIterable("result")).To(HaveLen(5))
			})

			// does not work for private data ! no API support as of Fabric 1.4
		})

		Context("When assets are stored in state", func() {
			var usr rmap.Rmap
			BeforeEach(func() {
				usr = tctx.Rmap("assetCreate", "mockuser", rmap.NewFromMap(map[string]interface{}{"name": "John", "surname": "Doe"}).Bytes(), -1, "")
				tctx.Ok("assetCreate", "mockincident", rmap.NewFromMap(map[string]interface{}{"description": "incident1", "assigned_to": MustGetID(usr)}).Bytes(), -1, "")
				tctx.Ok("assetCreate", "mockincident", rmap.NewFromMap(map[string]interface{}{"description": "incident2", "assigned_to": MustGetID(usr)}).Bytes(), -1, "")
			})

			It("Should return all with empty selector", func() {
				response := tctx.RmapNoResult("assetQuery", "mockincident", rmap.NewEmpty().Bytes(), false)
				Expect(response.Mapa).To(HaveKey("result"))
				result := response.MustGetIterable("result")
				Expect(result).To(HaveLen(2))
			})

			It("should work with empty input", func() {
				tctx.Ok("assetQuery", "mockincident", "", false)
			})

			It("Should return only assets matching selector", func() {
				query := rmap.NewFromMap(map[string]interface{}{"selector": map[string]interface{}{"description": "incident1"}})
				response := tctx.RmapNoResult("assetQuery", "mockincident", query.Bytes(), false)
				Expect(response.Mapa).To(HaveKey("result"))
				result := response.MustGetIterable("result")
				Expect(result).To(HaveLen(1))
				incident := response.MustGetJPtrRmap("/result/0")
				Expect(incident.Mapa).To(HaveKeyWithValue("description", "incident1"))
			})

			It("Should override wrong docType in selector", func() {
				query := rmap.NewFromMap(map[string]interface{}{"selector": map[string]interface{}{"docType": "XXX"}})
				response := tctx.RmapNoResult("assetQuery", "mockincident", query.Bytes(), false)
				Expect(response.Mapa).To(HaveKey("result"))
				result := response.MustGetIterable("result")
				Expect(result).To(HaveLen(2))
			})

			It("Should return only assets with some fields if client requests it", func() {
				query := rmap.NewFromMap(map[string]interface{}{
					"fields": []interface{}{"description"},
				})
				response := tctx.RmapNoResult("assetQuery", "mockincident", query.Bytes(), false)
				Expect(response.Mapa).To(HaveKey("result"))
				result := response.MustGetIterable("result")
				Expect(result).To(HaveLen(2))

				for _, resI := range result {
					incident := rmap.MustNewFromInterface(resI).Mapa
					Expect(incident).To(HaveKey("description"))
					Expect(incident).To(Not(HaveKey("assigned_to")))
				}
			})
		})

		Context("When assets are stored in private data", func() {
			BeforeEach(func() {
				tctx.Ok("assetCreate", "mockcomment", rmap.NewFromMap(map[string]interface{}{"text": "comment1"}).Bytes(), -1, "")
				tctx.Ok("assetCreate", "mockcomment", rmap.NewFromMap(map[string]interface{}{"text": "comment2"}).Bytes(), -1, "")
			})

			It("Should return all with empty selector", func() {
				response := tctx.RmapNoResult("assetQuery", "mockcomment", rmap.NewEmpty().Bytes(), false)
				Expect(response.Mapa).To(HaveKey("result"))
				result := response.MustGetIterable("result")
				Expect(result).To(HaveLen(2))
			})
		})

		It("Should execute BeforeQuery business logic before querying", func() {
			tctx.Error("business logic created fail", "assetQuery", "mockblogicfail", rmap.NewEmpty().Bytes(), false)
		})

		It("Should return error if unexpected key(s) are contained in query", func() {
			query := rmap.NewFromMap(map[string]interface{}{
				// legit keys
				"selector": rmap.NewEmpty().Mapa,
				"fields":   []string{"bflm", "psvz"},
				"bookmark": "moje zalozcicka",
				"limit":    1337,
				"sort":     rmap.NewEmpty().Mapa,

				// invalid keys
				"email": "abc",
				"zzz":   "xxx",
			})

			tctx.Error("unexpected key(s) in query: email,zzz", "assetQuery", "incident", query.Bytes(), false)
		})
	})

	Describe("Call to CC method assetUpdate", func() {
		var incident rmap.Rmap
		name := "mockincident"

		BeforeEach(func() {
			incidentReq := rmap.NewFromMap(map[string]interface{}{"description": "hello incident"})
			incident = tctx.Rmap("assetCreate", name, incidentReq.Bytes(), -1, "")
		})

		It("Should return error if attempting to set service keys", func() {
			req := rmap.NewFromMap(map[string]interface{}{
				konst.AssetIdKey:      "XXX",
				konst.AssetVersionKey: -9999,
				konst.AssetDocTypeKey: "ZZZ",
				"description":         "bah",
			})
			tctx.Error("patch contains service key(s)", "assetUpdate", name, MustGetID(incident), req.Bytes())
		})

		It("Should return error if attempting to update with empty patch", func() {
			tctx.Error("patch is empty", "assetUpdate", name, MustGetID(incident), rmap.NewEmpty().Bytes())
		})

		It("Should update", func() {
			key := "description"
			value := "this was changed"
			req := rmap.NewFromMap(map[string]interface{}{
				key: value,
			})
			updated := tctx.Rmap("assetUpdate", name, MustGetID(incident), req.Bytes())
			Expect(updated.Mapa).To(HaveKeyWithValue(key, value))
		})
	})

	Describe("Call to CC method assetUpdateDirect", func() {
		Context("When not referencing other assets", func() {
			id := "9bfb0402-a092-a9a2-7542-8bf37cb29a2e"
			name := "mockblogicfail"
			patch := rmap.NewFromMap(map[string]interface{}{"text": "halabala"})

			BeforeEach(func() {
				tctx.Ok("assetCreateDirect", name, rmap.NewEmpty().Bytes(), -1, id)
			})

			It("Should not execute any business logic", func() {
				tctx.Error("business logic created fail", "assetUpdate", name, id, patch.Bytes()) // assetUpdate throws
				tctx.Ok("assetUpdateDirect", name, id, patch.Bytes())                             // assetUpdateDirect does not
			})

			It("Should be protected", func() {
				tctx.SetActor("ordinaryUser")
				tctx.Error("permission denied", "assetUpdateDirect", name, id, patch.Bytes())
			})
		})

		Context("When referencing other assets", func() {
			id := ""
			user1 := rmap.NewFromMap(map[string]interface{}{
				"name":    "John",
				"surname": "Doe",
			})
			user1uuid := ""

			BeforeEach(func() {
				user1uuid = MustGetID(tctx.Rmap("assetCreate", "mockuser", user1.Bytes(), -1, ""))
				id = MustGetID(tctx.Rmap("assetCreate", "mockincident", rmap.NewFromMap(map[string]interface{}{"description": "ABCD", "assigned_to": user1uuid}).Bytes(), -1, ""))
			})

			It("Should update asset without checking reference", func() {
				patch := rmap.NewFromMap(map[string]interface{}{"assigned_to": "xxx"})

				tctx.Error("Referenced asset", "assetUpdate", "mockincident", id, patch.Bytes()) // assetUpdate throws
				tctx.Ok("assetUpdateDirect", "mockincident", id, patch.Bytes())                  // assetUpdateDirect does not
			})
		})
	})

	Describe("Call to CC method assetDelete", func() {
		var incident rmap.Rmap
		name := "mockincident"

		BeforeEach(func() {
			incidentReq := rmap.NewFromMap(map[string]interface{}{"description": "hello incident"})
			incident = tctx.Rmap("assetCreate", name, incidentReq.Bytes(), -1, "")
		})

		It("Should delete asset", func() {
			tctx.Ok("assetDelete", name, MustGetID(incident))
			tctx.Error("state entry not found: MOCKINCIDENT"+MustGetID(incident), "assetGet", name, MustGetID(incident), false, "")
		})
	})

	Describe("Call to CC method assetDeleteDirect", func() {
		id := "9bfb0402-a092-a9a2-7542-8bf37cb29a2e"
		name := "mockblogicfail"

		BeforeEach(func() {
			tctx.Ok("assetCreateDirect", name, rmap.NewEmpty().Bytes(), -1, id)
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "assetDeleteDirect", name, id)
		})

		It("Should not execute any business logic", func() {
			tctx.Error("business logic created fail", "assetDelete", name, id)                              // assetDelete throws
			tctx.Ok("assetDeleteDirect", name, id)                                                          // assetDeleteDirect does not
			tctx.Error("state entry not found: "+strings.ToUpper(name)+id, "assetGet", name, id, false, "") // asset was deleted
		})
	})

	Describe("Call to CC method assetGet", func() {
		var request rmap.Rmap
		name := "mockrequest"

		BeforeEach(func() {
			requestReq := rmap.NewFromMap(map[string]interface{}{"number": "1234"})
			request = tctx.Rmap("assetCreate", name, requestReq.Bytes(), -1, "")
		})

		It("Should pass data parameter to blogic", func() {
			data := rmap.NewFromMap(map[string]interface{}{
				"magic": "value",
			})
			tctx.Error("magic value found", "assetGet", name, MustGetID(request), false, data.Bytes())
		})

		Context("When using private data destination", func() {
			It("Should return error if asset does not exist", func() {
				tctx.Error("state entry not found: MOCKREQUESTsome_id", "assetGet", name, "some_id", false, "")
			})
		})

		Context("When referencing other assets", func() {
			It("Should recursively resolve if allowed in whitelist", func() {
				level3id := MustGetID(tctx.Rmap("assetCreate", "mocklevel3", rmap.NewFromMap(map[string]interface{}{"text": "hello"}), -1, ""))
				level2id := MustGetID(tctx.Rmap("assetCreate", "mocklevel2", rmap.NewFromMap(map[string]interface{}{"level3": level3id}), -1, ""))
				level1id := MustGetID(tctx.Rmap("assetCreate", "mocklevel1", rmap.NewFromMap(map[string]interface{}{"level2": level2id}), -1, ""))
				level1 := tctx.Rmap("assetGet", "mocklevel1", level1id, true, rmap.NewEmpty())
				Expect(level1.MustGetJPtrString("/level2/level3/text")).To(Equal("hello"))
			})

			It("Should not resolve blacklisted references", func() {
				blacklistedid := MustGetID(tctx.Rmap("assetCreate", "mockblacklisted", rmap.NewFromMap(map[string]interface{}{"text": "abc"}), -1, ""))
				refid := MustGetID(tctx.Rmap("assetCreate", "mockrefblacklist", rmap.NewFromMap(map[string]interface{}{"blacklisted": blacklistedid}), -1, ""))
				asset := tctx.Rmap("assetGet", "mockrefblacklist", refid, true, rmap.NewEmpty())
				Expect(asset.Mapa).To(HaveKeyWithValue("blacklisted", blacklistedid)) // ref not resolved
			})

			It("Should not resolve field blacklisted references", func() {
				refTargetID := MustGetID(tctx.Rmap("assetCreate", "mockincident", rmap.NewFromMap(map[string]interface{}{"description": "abc"}), -1, ""))

				assetData := rmap.NewFromMap(map[string]interface{}{
					"blacklisted": refTargetID,
					"nested": map[string]interface{}{
						"blacklisted_nest": refTargetID,
					},
				})

				resolvingID := MustGetID(tctx.Rmap("assetCreate", "mockreffieldblacklist", assetData.Bytes(), -1, ""))
				asset := tctx.Rmap("assetGet", "mockreffieldblacklist", resolvingID, true, rmap.NewEmpty())
				Expect(asset.Mapa).To(HaveKeyWithValue("blacklisted", refTargetID))
				nestedVal := asset.MustGetJPtrString("/nested/blacklisted_nest")
				Expect(nestedVal).To(Equal(refTargetID))
			})

		})
	})

	Describe("Call to CC method assetGetDirect", func() {
		id := "9bfb0402-a092-a9a2-7542-8bf37cb29a2e"
		name := "mockblogicfail"

		BeforeEach(func() {
			tctx.Ok("assetCreateDirect", name, rmap.NewEmpty().Bytes(), -1, id)
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "assetGetDirect", name, id, false, "")
		})

		It("Should not execute any business logic", func() {
			tctx.Error("business logic created fail", "assetGet", name, id, false, "") // assetGet returns error
			tctx.Ok("assetGetDirect", name, id, false, "")                             // assetGetDirect does not
		})
	})

	Describe("Call to CC method assetQueryDirect", func() {
		id := "9bfb0402-a092-a9a2-7542-8bf37cb29a2e"
		name := "mockblogicfail"

		BeforeEach(func() {
			tctx.Ok("assetCreateDirect", name, rmap.NewEmpty().Bytes(), -1, id)
		})

		It("Should be protected", func() {
			tctx.SetActor("ordinaryUser")
			tctx.Error("permission denied", "assetQueryDirect", name, rmap.NewEmpty().Bytes(), false)
		})

		It("Should not execute any business logic", func() {
			tctx.SetActor("superUser")
			tctx.Error("business logic created fail", "assetQuery", name, rmap.NewEmpty().Bytes(), false) // assetQuery throws
			tctx.Ok("assetQueryDirect", name, rmap.NewEmpty().Bytes(), false)                             // assetQueryDirect does not
		})
	})

	Describe("API method registry.PutAsset", func() {
		uuid := "54079897-e85e-47a3-aca6-3443c1808984"

		Context("When working with asset in state", func() {
			It("Should return error if isCreate param is false and asset does not exist", func() {
				tctx.Error("attempt to update non-existent state key: MOCKSTATE"+uuid, "functionInvoke", "MockStateInvalidCreate", rmap.NewFromMap(map[string]interface{}{"uuid": uuid}))
			})

			It("Should return error if isCreate param is true and asset does exists", func() {
				tctx.Ok("assetCreate", "mockstate", rmap.NewEmpty().Bytes(), -1, uuid)
				tctx.Error("state key already exists: MOCKSTATE"+uuid, "functionInvoke", "MockStateInvalidUpdate", rmap.NewFromMap(map[string]interface{}{"uuid": uuid}))
			})
		})

		Context("When working with asset in private data", func() {
			It("Should return error if isCreate param is false and asset does not exist", func() {
				tctx.Error("attempt to update non-existent private data key: MOCKPD"+uuid, "functionInvoke", "MockPDInvalidCreate", rmap.NewFromMap(map[string]interface{}{"uuid": uuid}))
			})

			It("Should return error if isCreate param is true and asset does not exist", func() {
				tctx.Ok("assetCreate", "mockpd", rmap.NewEmpty().Bytes(), -1, uuid)
				tctx.Error("private data key already exists: MOCKPD"+uuid, "functionInvoke", "MockPDInvalidUpdate", rmap.NewFromMap(map[string]interface{}{"uuid": uuid}))
			})
		})
	})
})
