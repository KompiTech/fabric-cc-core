package cc_core

import (
	. "github.com/KompiTech/fabric-cc-core/v2/src/testing"
	"github.com/KompiTech/rmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("upsert* function tests", func() {
	var tctx *TestContext

	BeforeEach(func() {
		tctx = getDefaultTextContext()
		tctx.InitOk(tctx.GetInit("", "").Bytes())
		tctx.RegisterAllActors()
	})

	Describe("API call to functionExecute invoke upsertRegistries", func() {
		It("Should upsert registries", func() {
			names := []string{"mockincident", "mockcomment"}
			var name string

			for _, name = range names {
				tctx.Error("state entry not found:", "registryGet", name, -1)
			}

			data := rmap.NewFromMap(map[string]interface{}{
				"registries": map[string]interface{}{},
			})

			for _, name = range names {
				data.SetJPtr("/registries/"+name, rmap.MustNewFromYAMLFile("testdata/assets/"+name+".yaml"))
			}

			tctx.Ok("functionInvoke", "upsertRegistries", data.Bytes())

			for _, name = range names {
				reg := tctx.Rmap("registryGet", name, -1)

				delete(reg.Mapa, "version") // version key is dependent on runtime, so it is not contained in yaml
				yamlData := rmap.MustNewFromYAMLFile("testdata/assets/" + name + ".yaml")
				yamlData.Mapa["name"] = name // name is omitted from yaml because it is already contained in filename

				Expect(reg.Mapa).To(Equal(yamlData.Mapa))
			}
		})
	})

	Describe("API call to functionExecute invoke upsertSingletons", func() {
		It("Should upsert registries", func() {
			names := []string{"mocksingleton", "mocksingleton2"}
			var name string

			for _, name = range names {
				tctx.Error("singleton name: "+name+" not found", "singletonGet", name, -1)
			}

			data := rmap.NewFromMap(map[string]interface{}{
				"singletons": map[string]interface{}{},
			})

			for _, name = range names {
				data.SetJPtr("/singletons/"+name, rmap.MustNewFromYAMLFile("testdata/singletons/"+name+".yaml"))
			}

			tctx.Ok("functionInvoke", "upsertSingletons", data.Bytes())

			for _, name = range names {
				reg := tctx.Rmap("singletonGet", name, -1)

				delete(reg.Mapa, "version") // version key is dependent on runtime, so it is not contained in yaml
				yamlData := rmap.MustNewFromYAMLFile("testdata/singletons/" + name + ".yaml")
				yamlData.Mapa["name"] = name // name is omitted from yaml because it is already contained in filename

				Expect(reg.Mapa).To(Equal(yamlData.Mapa))
			}
		})
	})
})
