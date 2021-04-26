package mockpaginate

import (
	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/rmap"
)

// Paginate does some pagination with rich query to trigger pagination error on mock
var Paginate = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	query := rmap.NewFromMap(map[string]interface{}{
		"selector": map[string]interface{}{},
	})

	_, _, _ = ctx.GetRegistry().QueryAssets("mockpaginate", query, "", false, true, 1000)
	return postPatch, nil
}
