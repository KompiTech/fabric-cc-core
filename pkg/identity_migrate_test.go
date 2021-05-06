package cc_core

import (
	testdata2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata"
	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("identity IDFunc migration", func() {
	It("Should be able to migrate identity assets with old fingerprint", func() {
		// this tests needs to reuse existing data between invocations, so custom tctx handling must be done
		eng := testdata2.GetConfiguration()
		eng.CurrentIDFunc = engine.CertSHA512IDFunc

		tctx := NewTestContext("mock", eng, nil, nil)
		tctx.InitOk(tctx.GetInit("", "").Bytes())
		tctx.Ok("identityAddMe", rmap.NewEmpty().Bytes()) // create identity using old IDFunc

		// simulate update to new cc instance with updated IDFunction
		// Previous FP was SHA512
		// Current FP is SHA256
		// just an example, real CC uses different impl, but still based on cert
		eng.CurrentIDFunc = engine.CertSHA256IDFunc
		var prev engine.IDFunc = engine.CertSHA512IDFunc
		eng.PreviousIDFunc = &prev
		tctx = NewTestContext("mock", eng, tctx.GetMockStub(), tctx.GetCouchDBMock())

		identityMe := tctx.RmapNoResult("identityMe", false)
		Expect(identityMe.Mapa).To(HaveKeyWithValue("can_migrate", true)) // identity can be migrated

		migrateResult := tctx.RmapNoResult("identityAddMe", []byte("{}"))
		Expect(migrateResult.Mapa).To(HaveKeyWithValue("migrated", true)) // identity was migrated

		oldIdentity := tctx.Rmap("assetGet", "identity", identityMe.MustGetString("fingerprint"), false, "")
		Expect(oldIdentity.Mapa).To(HaveKeyWithValue("is_enabled", false))

		newIdentity := tctx.Rmap("assetGet", "identity", migrateResult.MustGetJPtrString("/result/fingerprint"), false, "")
		Expect(newIdentity.Mapa).To(HaveKeyWithValue("is_enabled", true))
	})
})
