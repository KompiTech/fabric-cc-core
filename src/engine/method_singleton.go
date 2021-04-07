package engine

import (
	"strings"

	. "github.com/KompiTech/fabric-cc-core/v2/src/konst"

	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func singletonGetFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	version, err := ctx.ParamInt(VersionParam)
	if err != nil {
		return "", err
	}

	name = strings.ToLower(name)

	if err := enforceCustomAccess(reg, "/"+SingletonCasbinObject+"/"+name, ReadAction); err != nil {
		return "", err
	}

	item, version, err := reg.GetSingleton(name, version)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetSingleton() failed")
	}

	item.Mapa[SingletonVersionKey] = version
	item.Mapa[SingletonNameKey] = name

	return string(item.WrappedResultBytes()), nil
}

func singletonListFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)

	if err := enforceCustomAccess(reg, "/"+SingletonCasbinObject+"/*", ReadAction); err != nil {
		return "", err
	}

	names, err := reg.ListSingletons()
	if err != nil {
		return "", errors.Wrap(err, "reg.ListSingletons() failed")
	}

	output := rmap.NewFromMap(map[string]interface{}{
		OutputResultKey: names,
	})

	return output.String(), nil
}

func singletonUpsertFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	name = strings.ToLower(name)
	data, err := ctx.ParamString(DataParam)
	if err != nil {
		return "", err
	}

	if err := enforceCustomAccess(reg, "/"+SingletonCasbinObject+"/"+name, UpsertAction); err != nil {
		return "", err
	}

	var singletonToUpsert rmap.Rmap
	if len(data) == 0 {
		singletonToUpsert = rmap.NewEmpty()
	} else {
		var err error
		singletonToUpsert, err = rmap.NewFromString(data)
		if err != nil {
			return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
		}
	}

	version, err := reg.UpsertSingleton(singletonToUpsert, name)
	if err != nil {
		return "", errors.Wrap(err, "reg.UpsertSingleton() failed")
	}

	singletonToUpsert.Mapa[SingletonNameKey] = name
	singletonToUpsert.Mapa[SingletonVersionKey] = version

	return string(singletonToUpsert.WrappedResultBytes()), nil
}
