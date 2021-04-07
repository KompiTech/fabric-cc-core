package engine

import (
	. "github.com/KompiTech/fabric-cc-core/v2/src/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func changelogListFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)
	if err := enforceCustomAccess(reg, "/"+ChangelogCasbinObject, ReadAction); err != nil {
		return "", err
	}

	cl, err := NewChangelog(ctx)
	if err != nil {
		return "", errors.Wrap(err, "NewChangelog() failed")
	}

	items, err := cl.List()
	if err != nil {
		return "", err
	}

	mapItems := rmap.ConvertSliceToMaps(items)

	output := rmap.NewEmpty()
	output.Mapa[OutputResultKey] = mapItems

	return string(output.Bytes()), nil
}

func changelogGetFrontend(ctx ContextInterface) (string, error) {
	number, err := ctx.ParamInt(NumberParam)
	if err != nil {
		return "", err
	}

	reg := ctx.Get(RegistryKey).(*Registry)
	if err := enforceCustomAccess(reg, "/"+ChangelogCasbinObject, ReadAction); err != nil {
		return "", err
	}

	cl, err := NewChangelog(ctx)
	if err != nil {
		return "", errors.Wrap(err, "NewChangelog() failed")
	}

	item, err := cl.Get(number)
	if err != nil {
		return "", errors.Wrap(err, "cl.Get() failed")
	}

	return string(item.WrappedResultBytes()), nil
}
