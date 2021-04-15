package engine

import (
	"strings"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func registryGetFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	name = strings.ToLower(name)

	version, err := ctx.ParamInt(VersionParam)
	if err != nil {
		return "", err
	}

	if err := enforceCustomAccess(reg, "/"+RegistryCasbinObject+"/"+name, ReadAction); err != nil {
		return "", err
	}

	item, version, err := reg.GetItem(name, version)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetItem() failed")
	}

	// decorate item with version, name
	item.Mapa[RegistryItemVersionKey] = version
	item.Mapa[RegistryItemNameKey] = name

	return string(item.WrappedResultBytes()), nil
}

func registryListFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)

	if err := enforceCustomAccess(reg, "/"+RegistryCasbinObject+"/*", ReadAction); err != nil {
		return "", err
	}

	list, err := reg.ListItems()
	if err != nil {
		return "", errors.Wrap(err, "reg.ListItems() failed")
	}

	output := rmap.NewFromMap(map[string]interface{}{
		OutputResultKey: list,
	})

	// do not use method to wrap result, we did it ourselves
	return output.String(), nil
}

func registryUpsertFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)

	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	data, err := ctx.ParamBytes(DataParam)
	if err != nil {
		return "", err
	}

	name = strings.ToLower(name)

	if err := enforceCustomAccess(reg, "/"+RegistryCasbinObject+"/"+name, UpsertAction); err != nil {
		return "", err
	}

	itemToUpsert, err := rmap.NewFromBytes(data)
	if err != nil {
		return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
	}

	item, version, err := reg.UpsertItem(itemToUpsert, name)
	if err != nil {
		return "", errors.Wrap(err, "reg.UpsertItem() failed")
	}

	// decorate item with version, name
	item.Mapa[RegistryItemVersionKey] = version
	item.Mapa[RegistryItemNameKey] = name

	return string(item.WrappedResultBytes()), nil
}
