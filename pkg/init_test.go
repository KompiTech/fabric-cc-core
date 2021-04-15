package cc_core

import (
	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
	. "github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Init() method tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
	})

	Describe("Call to CC method Init()", func() {
		It("Should return error if first initializing without init_manager key", func() {
			tctx.InitError("bootstrapSuperUser() failed: init_manager is mandatory when no superusers are present", NewEmpty().Bytes())
		})

		Context("When superuser was already created", func() {
			BeforeEach(func() {
				tctx.InitOk(tctx.GetInit("", "").Bytes())
				tctx.Ok("identityAddMe", NewEmpty().Bytes())
			})

			It("Should allow repeated init with empty JSON or empty string value", func() {
				tctx.InitOk([]byte("{}"))
				tctx.InitOk([]byte(""))
			})

			It("Should ignore init_manager value", func() {
				init := NewEmpty()
				init.MustSetJPtr("/init_manager", tctx.GetActorFingerprint("nobodyUser"))
				tctx.InitOk(init.Bytes())

				// superUser still has his SU role
				me := tctx.Rmap("identityMe", true)
				Expect(me.MustGetJPtrString("/roles/0/uuid")).To(Equal(konst.SuperuserRoleUUID))

				// nobodyUser does not have SU role, even if he was set as init_manager, because it was ignored
				tctx.SetActor("nobodyUser")
				tctx.Ok("identityAddMe", NewEmpty().Bytes())
				me = tctx.Rmap("identityMe", true)
				Expect(me.Mapa).To(Not(HaveKey("roles")))
			})
		})

		Context("When old latest map for singletons and registryItems is present", func() {
			var lsm, lvm Rmap

			BeforeEach(func() {
				txId := "ca84c66c-f893-391c-590d-dec80bd0a8c0"

				tctx.GetCC().MockTransactionStart(txId)
				// emulate some old LVM/LSM contents
				lvm = NewFromMap(map[string]interface{}{
					"FOOBAR": float64(4),
					"REHOR":  float64(3),
				})
				Expect(tctx.GetCC().PutState(konst.LatestVersionMapKey, lvm.Bytes())).To(BeNil())

				lsm = NewFromMap(map[string]interface{}{
					"KOMPITECH_LIFECYCLE": float64(2),
				})
				Expect(tctx.GetCC().PutState(konst.LatestSingletonMapKey, lsm.Bytes())).To(BeNil())
				tctx.GetCC().MockTransactionEnd(txId)
			})

			It("Should refuse to initialize", func() {
				tctx.InitError("found incompatible cc-core 1.x.x state key, refusing to init()", tctx.GetInit("", "").Bytes())
			})
		})
	})
})
