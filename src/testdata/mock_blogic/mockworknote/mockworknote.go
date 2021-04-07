package mockworknote

import (
	"fmt"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/fabric-cc-core/v2/src/konst"
	"github.com/KompiTech/rmap"
)

// impl copied from itsm-cgi to test error condition

// AttachToEntity updates parent entity specified under key: entity in format <name>:<uuid>
// By convention, it expects array of references to entity that is being attach under key: <docType>s (in lower case).
// Array is appended to and parent entity is updated and saved
// Another option is singular reference. If any value is present in singular, error is produced because it would be overwritten.
// If parent defines both singular and plural form, this will return error
// If parent doesn't define singular or plural -> error
// ONLY USE THIS IF THIS IS THE FINAL MODIFICATION OF PARENT ASSET
// DO NOT READ PARENTASSET IN FOLLOWING BUSINESS LOGIC, BECAUSE YOU WILL GET OLD VERSION BEFORE TX
var AttachToEntity = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	if !postPatch.Exists("entity") {
		//no entity spec, nothing to do
		return postPatch, nil
	}

	reg := ctx.Get("registry").(*engine.Registry)

	parentAsset, err := attachToEntityImpl(ctx, postPatch)
	if err != nil {
		return rmap.Rmap{}, err
	}

	// save changes to parentAsset
	if err := reg.PutAsset(parentAsset, false); err != nil {
		return rmap.Rmap{}, err
	}

	return postPatch, nil
}

// ParseEntityField parses 'entity' field and returns name and UUID of the entity
func parseEntityField(entity string) (entityName, entityUUID string, err error) {
	fields := strings.Split(entity, ":")
	if len(fields) != 2 {
		return "", "", engine.ErrorUnprocessableEntity(fmt.Sprintf("invalid \"entity\" format: %s, it must follow schema: <name>:<uuid>", entity))
	}

	entityName = strings.ToLower(fields[0])
	entityUUID = strings.ToLower(fields[1])

	return entityName, entityUUID, nil
}

// AttachToEntityImpl looks for key "entity" in assetMap
// fetches the specified entity from State
// adds ref to assetMap to a key by its docType and looks either for singular or plural key
// returns modified map of parent asset. DOES NOT PERFORM saving to state, you must handle this yourself
func attachToEntityImpl(ctx engine.ContextInterface, assetMap rmap.Rmap) (rmap.Rmap, error) {
	entity, err := assetMap.GetString("entity")
	if err != nil {
		return rmap.Rmap{}, err
	}

	parentName, parentUUID, err := parseEntityField(entity)
	if err != nil {
		return rmap.Rmap{}, err
	}

	reg := ctx.Get("registry").(*engine.Registry)

	parentAsset, err := reg.GetAsset(parentName, parentUUID, false, true)
	if err != nil {
		return rmap.Rmap{}, err
	}

	parentVersion, err := parentAsset.GetInt(konst.AssetVersionKey)
	if err != nil {
		return rmap.Rmap{}, err
	}

	// get schema for parent
	parentRegistryItem, _, err := reg.GetItem(parentName, parentVersion)
	if err != nil {
		return rmap.Rmap{}, err
	}

	// get properties part of parent schema
	parentPropertiesRMap, err := parentRegistryItem.GetJPtrRmap("/schema/properties")
	if err != nil {
		return rmap.Rmap{}, err
	}

	parentProperties := parentPropertiesRMap.Mapa

	thisDocType, err := assetMap.GetString(konst.AssetDocTypeKey)
	if err != nil {
		return rmap.Rmap{}, err
	}

	// check parent's schema for convention keys for plural and singular forms
	singularName := strings.ToLower(thisDocType)
	pluralName := fmt.Sprintf("%ss", strings.ToLower(thisDocType))

	_, singularExists := parentProperties[singularName]
	_, pluralExists := parentProperties[pluralName]

	if !pluralExists && !singularExists {
		return rmap.Rmap{}, fmt.Errorf("AttachToEntity: parent does not have fields: %s or %s, unable to attach", singularName, pluralName)
	} else if pluralExists && singularExists {
		return rmap.Rmap{}, fmt.Errorf("AttachToEntity: parent does have both fields: %s and %s, unable to attach", singularName, pluralName)
	}

	// we assume that singular or plural field has proper type on schema and description with reference
	// TODO check schema

	if singularExists {
		value, exists := parentAsset.Mapa[singularName]
		if exists {
			return rmap.Rmap{}, fmt.Errorf("AttachToEntity: parent has field: %s with value: %s, unable to attach because value would be overwritten", singularName, value)
		}

		// set singular reference to asset that will be created
		parentAsset.Mapa[singularName] = assetMap.Mapa[konst.AssetIdKey]
	} else {
		doAppend := true
		_, exists := parentAsset.Mapa[pluralName]
		if !exists {
			// create array if it doesnt exist on parent
			parentAsset.Mapa[pluralName] = []interface{}{}
		} else {
			// if array exists, check if ref is not present already
			for _, v := range parentAsset.Mapa[pluralName].([]interface{}) {
				if v == assetMap.Mapa[konst.AssetIdKey] {
					// ref is already present, do not append
					doAppend = false
				}
			}
		}

		if doAppend {
			// append ID of this to parentAsset array
			parentAsset.Mapa[pluralName] = append(parentAsset.Mapa[pluralName].([]interface{}), assetMap.Mapa[konst.AssetIdKey])
		}
	}

	return parentAsset, nil
}
