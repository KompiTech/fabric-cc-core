package engine

import (
	"fmt"
	"sort"

	. "github.com/KompiTech/fabric-cc-core/v2/src/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

// getSUQuery returns "const map" of query to select SUs
func getSUQuery() map[string]interface{} {
	return map[string]interface{}{
		QuerySelectorKey: map[string]interface{}{
			AssetDocTypeKey: IdentityAssetKeyPrefix,
			"roles": map[string]interface{}{
				"$elemMatch": map[string]interface{}{
					"$eq": SuperuserRoleUUID,
				},
			},
		},
	}
}

// initChaincode is the standard initialization method for chaincode
// it handles:
// setting of initManager variable if no initManagers are present
// upserting of assets and singletons (new version is created only if it differs from latest by hash)
func initChaincode(ctx ContextInterface) error {
	// check for old LVM key, refuse to init in this case
	lvmBytes, err := ctx.Stub().GetState(LatestVersionMapKey)
	if err != nil {
		return errors.Wrap(err, "ctx.Stub().GetState() failed")
	}

	if len(lvmBytes) > 0 {
		return fmt.Errorf("found incompatible cc-core 1.x.x state key, refusing to init()")
	}

	var input rmap.Rmap
	inputBytes, err := ctx.ParamString(InputParam)
	if err != nil {
		return err
	}

	if len(inputBytes) == 0 {
		// empty input, assume empty Rmap
		input = rmap.NewEmpty()
	} else {
		// some input, must be JSON
		var err error
		input, err = rmap.NewFromString(inputBytes)
		if err != nil {
			return errors.Wrap(err, "rmap.NewFromBytes() failed")
		}

		// if input was not empty JSON, it must match schema
		if !input.IsEmpty() {
			if err := input.ValidateSchemaBytes([]byte(InstantiateJSONSchema)); err != nil {
				return errors.Wrap(err, "input.ValidateSchemaBytes() failed")
			}
		}
	}

	if err := bootstrapSuperUser(ctx, input); err != nil {
		return errors.Wrap(err, "bootstrapSuperUser() failed")
	}

	if err := upsertRegistries(ctx, input); err != nil {
		return errors.Wrap(err, "upsertRegistries() failed")
	}

	if err := upsertSingletons(ctx, input); err != nil {
		return errors.Wrap(err, "upsertSingletons() failed")
	}

	return nil
}

// bootstrapSuperUser prepares data required to bootstrap first initManager (when he calls identityAddMe)
func bootstrapSuperUser(ctx ContextInterface, input rmap.Rmap) error {
	reg := ctx.Get(RegistryKey).(*Registry)

	superUsers, _, err := reg.QueryAssets(IdentityAssetName, rmap.NewFromMap(getSUQuery()), "", false, false, PageSize)
	if err != nil {
		return errors.Wrap(err, "reg.QueryAssets() failed")
	}

	if len(superUsers) == 0 {
		// if no superUsers are present, first init_manager must be bootstrapped, param is mandatory
		if !input.Exists(InitSuperuserKey) {
			return errors.New(InitSuperuserKey + " is mandatory when no superusers are present")
		}

		initManagerFP, err := input.GetString(InitSuperuserKey)
		if err != nil {
			return errors.Wrap(err, "input.GetString() failed")
		}

		// store as JSON object, not raw bytes
		// use the same key as input param name
		initManagerObj := rmap.NewFromMap(map[string]interface{}{
			InitSuperuserKey: initManagerFP,
		})

		// save init_manager fingerprint to state because he still didn't log in
		if err := ctx.Stub().PutState(InitSuperuserStateKey, initManagerObj.Bytes()); err != nil {
			return errors.Wrap(err, "ctx.Stub().PutState() failed")
		}
	} else if input.Exists(InitSuperuserKey) {
		// if some SU already exists, but argument was provided, log info that it was ignored
		ctx.Logger().Warningf("%s argument to Init() was ignored, because some super users already exist", InitSuperuserKey)
	}

	return nil
}

// upsertRegistries upserts everything from registries key when initing CC
func upsertRegistries(ctx ContextInterface, input rmap.Rmap) error {
	reg := ctx.GetRegistry()

	if input.Exists(InitRegistriesKey) {
		// registries key is present, upsert everything
		registries, err := input.GetRmap(InitRegistriesKey)
		if err != nil {
			return errors.Wrap(err, "input.GetRmap() failed")
		}

		buis := make([]bulkItem, 0, len(registries.Mapa))

		// iteration through keys must be deterministic on all peers -> sort keys before iterating
		keys := make([]string, 0, len(registries.Mapa))
		for k, _ := range registries.Mapa {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, assetName := range keys {
			regItemI := registries.Mapa[assetName]
			regItem, err := rmap.NewFromInterface(regItemI)
			if err != nil {
				return errors.Wrap(err, "rmap.NewFromInterface() failed")
			}

			buis = append(buis, bulkItem{
				Name:  assetName,
				Value: regItem,
			})
		}

		if err := reg.BulkUpsertItems(buis); err != nil {
			return errors.Wrap(err, "reg.BulkUpsertItems() failed")
		}
	}

	return nil
}

// upsertSingletons upserts everything from singletons key when initing CC
func upsertSingletons(ctx ContextInterface, input rmap.Rmap) error {
	reg := ctx.Get(RegistryKey).(*Registry)

	if input.Exists(InitSingletonsKey) {
		singletons, err := input.GetRmap(InitSingletonsKey)
		if err != nil {
			return errors.Wrap(err, "input.GetRmap() failed")
		}

		buis := make([]bulkItem, 0, len(singletons.Mapa))

		// iteration through keys must be deterministic on all peers -> sort keys before iterating
		keys := make([]string, 0, len(singletons.Mapa))
		for k, _ := range singletons.Mapa {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, assetName := range keys {
			singletonI := singletons.Mapa[assetName]
			singleton, err := rmap.NewFromInterface(singletonI)
			if err != nil {
				return errors.Wrap(err, "rmap.NewFromInterface() failed")
			}

			buis = append(buis, bulkItem{
				Name:  assetName,
				Value: singleton,
			})
		}

		if err := reg.BulkUpsertSingletons(buis); err != nil {
			return errors.Wrap(err, "reg.BulkUpsertSingletons() failed")
		}
	}

	return nil
}
