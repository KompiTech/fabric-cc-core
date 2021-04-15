package mockfunc

import (
	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/rmap"
)

var MockFunc = func(ctx engine.ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	return rmap.NewFromMap(map[string]interface{}{"result": "true"}), nil
}
