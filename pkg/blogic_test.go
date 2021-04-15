package cc_core

import (
	"fmt"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Business logic execution tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("testdata/assets", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("mockrequest business logic tests", func() {
		It("Should enforce unique number", func() {
			number := "REQ1337"
			requestReq := rmap.NewFromMap(map[string]interface{}{
				"number": number,
			})

			// first Request created ok
			tctx.Ok("assetCreate", "mockrequest", requestReq.Bytes(), -1, "")
			// second Request with the same number must fail because of business logic
			tctx.Error(fmt.Sprintf("Mockrequest number: %s is not unique", number), "assetCreate", "mockrequest", requestReq.Bytes(), -1, "")
		})

		It("Should enforce immutable fields", func() {
			name := "mockrequest"
			requestReq := rmap.NewFromMap(map[string]interface{}{
				"number": "REQ1337",
			})

			uuid := MustGetID(tctx.Rmap("assetCreate", name, requestReq.Bytes(), -1, ""))

			requestReq.Mapa["number"] = "REQ1338"
			tctx.Error("field: number is not mutable", "assetUpdate", name, uuid, requestReq.Bytes())
		})
	})

	Describe("mockincident, mocktimelog business logic tests", func() {
		It("Should attach mocktimelog to mockincident by business logic", func() {
			incidentReq := rmap.NewFromMap(map[string]interface{}{
				"description": "Mock incident",
			})

			incidentID := MustGetID(tctx.Rmap("assetCreate", "mockincident", incidentReq.Bytes(), -1, ""))

			timelogReq := rmap.NewFromMap(map[string]interface{}{
				"incident": incidentID,
			})

			timelogID := MustGetID(tctx.Rmap("assetCreate", "mocktimelog", timelogReq.Bytes(), -1, ""))

			incident := tctx.Rmap("assetGet", "mockincident", incidentID, true, "").Mapa
			Expect(incident).To(HaveKey("timelogs"))
			Expect(incident["timelogs"]).To(HaveLen(1))

			timelogs := incident["timelogs"].([]interface{})
			Expect(timelogs[0]).To(HaveKeyWithValue("uuid", timelogID))
			Expect(timelogs[0]).To(HaveKey("incident"))

			//assert when getting incident with resolve false -> timelog doesn't get resolved
			incident = tctx.JSON("assetGet", "mockincident", incidentID, false, "")
			Expect(incident).To(HaveKey("timelogs"))
			Expect(timelogs).To(HaveLen(1))

			timelogs = incident["timelogs"].([]interface{})
			Expect(timelogs[0]).To(Equal(timelogID))
		})
	})

	Describe("mockdataafterresolve business logic tests", func() {
		It("Must allow data parameter from assetGet to be available in AfterResolve blogic", func() {
			darID := MustGetID(tctx.Rmap("assetCreate", "mockdataafterresolve", rmap.NewFromMap(map[string]interface{}{"text": "hello"}), -1, ""))
			refID := MustGetID(tctx.Rmap("assetCreate", "mockrefdata", rmap.NewFromMap(map[string]interface{}{"ref": darID}), -1, ""))
			tctx.Error("magic value found, param passing works", "assetGet", "mockrefdata", refID, true, rmap.NewFromMap(map[string]interface{}{"magic": "value"}))
		})
	})
})
