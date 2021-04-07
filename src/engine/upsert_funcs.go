package engine

import "github.com/KompiTech/rmap"

//these functions are built-in to allow bulk insertion of registries/singletons

var upsertRegistriesFunc = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}

	if err := upsertRegistries(ctx, input); err != nil {
		return null, err
	}

	return rmap.NewFromMap(map[string]interface{}{"result": map[string]interface{}{"ok": true}}), nil
}

var upsertSingletonsFunc = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}

	if err := upsertSingletons(ctx, input); err != nil {
		return null, err
	}

	return rmap.NewFromMap(map[string]interface{}{"result": map[string]interface{}{"ok": true}}), nil
}
