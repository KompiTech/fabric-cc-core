package mockerror

import (
	"github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/fabric-cc-core/v2/src/konst"
	"github.com/KompiTech/rmap"
)

// these funcs attempt to invoke invalid isCreate setting to test on PutAsset

func setup(reg *engine.Registry, input rmap.Rmap, name string) (rmap.Rmap, error) {
	id, err := input.GetString("uuid")
	if err != nil {
		return rmap.Rmap{}, err
	}

	asset, err := reg.MakeAsset(name, id, -1)
	if err != nil {
		return rmap.Rmap{}, err
	}

	return asset, nil
}

var MockStateInvalidCreate = func(ctx engine.ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	reg := ctx.Get(konst.RegistryKey).(*engine.Registry)

	asset, err := setup(reg, input, "mockstate")
	if err != nil {
		return rmap.Rmap{}, err
	}

	if err := reg.PutAsset(asset, false); err != nil {
		return rmap.Rmap{}, err
	}

	return rmap.Rmap{}, nil
}

var MockStateInvalidUpdate = func(ctx engine.ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	reg := ctx.Get(konst.RegistryKey).(*engine.Registry)

	asset, err := setup(reg, input, "mockstate")
	if err != nil {
		return rmap.Rmap{}, err
	}

	if err := reg.PutAsset(asset, true); err != nil {
		return rmap.Rmap{}, err
	}

	return rmap.Rmap{}, nil
}

var MockPDInvalidCreate = func(ctx engine.ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	reg := ctx.Get(konst.RegistryKey).(*engine.Registry)

	asset, err := setup(reg, input, "mockpd")
	if err != nil {
		return rmap.Rmap{}, err
	}

	if err := reg.PutAsset(asset, false); err != nil {
		return rmap.Rmap{}, err
	}

	return rmap.Rmap{}, nil
}

var MockPDInvalidUpdate = func(ctx engine.ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	reg := ctx.Get(konst.RegistryKey).(*engine.Registry)

	asset, err := setup(reg, input, "mockpd")
	if err != nil {
		return rmap.Rmap{}, err
	}

	if err := reg.PutAsset(asset, true); err != nil {
		return rmap.Rmap{}, err
	}

	return rmap.Rmap{}, nil
}
