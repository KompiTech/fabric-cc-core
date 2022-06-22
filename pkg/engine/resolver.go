package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	. "github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

type resolver struct{}

func (r resolver) WalkReferences(ctx ContextInterface, asset Rmap, resolve bool) error {
	assetName, err := asset.GetString(konst.AssetDocTypeKey)
	if err != nil {
		return errors.Wrap(err, "asset.GetString(DocTypeKey) failed")
	}
	assetName = strings.ToLower(assetName)

	assetVersion, err := asset.GetInt(konst.AssetVersionKey)
	if err != nil {
		return errors.Wrap(err, "asset.GetInt(VersionKey) failed")
	}

	assetRegItem, _, err := ctx.Get(konst.RegistryKey).(*Registry).GetItem(assetName, assetVersion)
	if err != nil {
		return errors.Wrap(err, "reg.GetItem(assetName, assetVersion) failed")
	}

	assetSchema, err := assetRegItem.GetRmap(konst.RegistryItemSchemaKey)
	if err != nil {
		return errors.Wrap(err, "regItem.GetRmap(schema) failed")
	}

	var errs []string

	err = r.walkReferences(ctx, assetName, nil, asset, asset.Mapa, assetSchema, resolve, &errs)
	sort.Strings(errs)
	if err != nil && len(errs) > 0 {
		return errors.Wrapf(err, "r.walkReferences() failed, refs failed: %s", strings.Join(errs, ","))
	} else if err != nil && len(errs) == 0 {
		return errors.Wrap(err, "r.walkReferences() failed")
	} else if len(errs) > 0 {
		return fmt.Errorf("r.walkReferences() failed, refs failed: %s", strings.Join(errs, ","))
	}

	return nil
}

// walkReferences recursively visits all attributes on asset. For each, decide if it is a reference by looking into related schema for reference magic string in description.
// Params:
// ctx - Context
// thisAssetName - name of currently walked asset instance
// pathJPtrSlice - current search path JSONPointer, relative to root, as slice with all elements (no separators)
// root - root of asset structure, this doesn't change between invocations
// dataPtr - currently analyzed value, in first call the value must be same as root
// schema - schema structure
// resolve - mode of operation, if false, then refs are only validated, if true, refs are replaced with resolved variants
// errs - list of errors that is populated with all non-valid refs
func (r resolver) walkReferences(ctx ContextInterface, thisAssetName string, pathJPtrSlice []string, root Rmap, dataPtr interface{}, schema Rmap, resolve bool, errs *[]string) error {
	registry := ctx.Get(konst.RegistryKey).(*Registry)
	eng := ctx.GetConfiguration()

	// check if field blacklist is defined to have backward compatibility.
	// but when not resolving, the blacklist is ignored to properly check refs
	if resolve && (eng.ResolveFieldsBlacklist.Mapa != nil) {
		// check if field blacklist is defined for asset
		_, blExists := eng.ResolveFieldsBlacklist.Mapa[thisAssetName]
		if blExists {
			bl, err := eng.ResolveFieldsBlacklist.GetRmap(thisAssetName)
			if err != nil {
				return err
			}

			// check if field name is blacklisted
			blKey := strings.Join(pathJPtrSlice, ".")

			if bl.Exists(blKey) {
				// this field is blacklisted, end
				return nil
			}
		}
	}

	switch el := dataPtr.(type) {
	case map[string]interface{}:
		// nested object, engage recursion
		for k, vI := range el {
			if err := r.walkReferences(ctx, thisAssetName, append(pathJPtrSlice, k), root, vI, schema, resolve, errs); err != nil {
				return err
			}
		}
	case []interface{}:
		// nested array, engage recursion
		for i, iface := range el {
			if err := r.walkReferences(ctx, thisAssetName, append(pathJPtrSlice, strconv.Itoa(i)), root, iface, schema, resolve, errs); err != nil {
				return err
			}
		}
	case string:
		// since ref can be only contained in string, we do not have to take other types than string into account
		// search schema and find if this pathJPtrSlice is reference in this schema
		pathJPtr := konst.JPtrSeparator + strings.Join(pathJPtrSlice, konst.JPtrSeparator)

		descJPtr, err := r.getDescriptionJPtr(pathJPtr)
		if err != nil {
			return err
		}

		descExists, err := schema.ExistsJPtr(descJPtr)
		if err != nil {
			return err
		}

		// description in schema does not exist, cannot be a ref
		if !descExists {
			return nil
		}

		desc, err := schema.GetJPtrString(descJPtr)
		if err != nil {
			return errors.Wrap(err, "sch.GetDescription() failed")
		}

		// analyze schema description and element value
		// decide, if this is ref at all, or entityref
		isRef, targetName, targetUUID, err := r.analyzeRef(desc, el)
		if err != nil {
			return errors.Wrap(err, "r.analyzeRef() failed")
		}

		// no ref, done
		if !isRef {
			//this is not a reference, finished
			return nil
		}

		// check if this reference is not blacklisted
		if eng.ResolveBlacklist.Exists(targetName) {
			return nil
		}

		// asset must exist whether resolving or not
		// if an asset is going to be created in this TX (assetCreate method), then aCache was populated before calling this and references will be validated OK
		exists, err := registry.ExistsAsset(targetName, targetUUID)
		if err != nil {
			return errors.Wrap(err, "registry.ExistsAsset() failed")
		}

		if !exists {
			thisId, err := konst.AssetGetID(root)
			if err != nil {
				return err
			}

			*errs = append(*errs, fmt.Sprintf("Referenced asset '%s' with ID '%s' not found (currently resolved asset name: %s, uuid: %s)", targetName, targetUUID, thisAssetName, thisId))
			return nil // stop recursion here because this is dead end
		}

		if resolve {
			// resolve true requires actual asset, fetch it
			target, err := registry.GetAsset(targetName, targetUUID, false, true)
			if err != nil {
				return errors.Wrap(err, "registry.GetAsset() failed")
			}

			// check if assetName.fieldName is allowed in recursive resolve whitelist
			fieldName := ""
			if len(pathJPtrSlice) > 0 {
				fieldName = pathJPtrSlice[len(pathJPtrSlice)-1]
			}

			condition := eng.RecursiveResolveWhitelist.Exists(thisAssetName + "." + fieldName)

			if !condition {
				// allow transitive recursive resolve - if target points to something defined in whitelist, recurse
				key := targetName + "."
				for k := range eng.RecursiveResolveWhitelist.Mapa {
					if strings.HasPrefix(k, key) {
						condition = true
						break
					}
				}
			}

			if condition {
				// cycle protection -> wont resolve something that is the same asset as thisAssetName
				if thisAssetName == targetName {
					condition = false
				}
			}

			if condition {
				if err := r.WalkReferences(ctx, target, resolve); err != nil {
					return errors.Wrap(err, "r.WalkReferences() failed")
				}
			}

			// if data is present in params, make it accessible in blogic
			// if data is not bytes, you will get a nice traceback here
			var dataR *Rmap
			if dataI, dataExists := ctx.Params()["data"]; dataExists {
				dataB, ok := dataI.([]byte)
				if !ok {
					return fmt.Errorf("data is not []byte")
				}

				// only attempt to parse data if it actually contains something
				if len(dataB) > 0 {
					tmp, err := NewFromBytes(dataB)
					if err != nil {
						return errors.Wrap(err, "rmap.NewFromInterface(dataI) failed")
					}

					dataR = &tmp
				}
			}

			// execute business logic stage AfterResolve
			target, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, AfterResolve, dataR, target)
			if err != nil {
				return errors.Wrap(err, `ctx.GetConfiguration().BusinessExecutor.Execute(AfterResolve) failed`)
			}

			// replace the key with its resolved value
			if err := root.SetJPtr(pathJPtr, target); err != nil {
				return errors.Wrap(err, "root.SetJPtr() failed")
			}
		}
	}
	return nil
}

func (r resolver) ParseEntityField(entity string) (entityName, entityUUID string, err error) {
	fields := strings.Split(entity, ":")
	if len(fields) != 2 {
		return "", "", fmt.Errorf("invalid \"entity\" format: %s, it must follow schema: <name>:<uuid>", entity)
	}

	entityName = strings.ToLower(fields[0])
	entityUUID = strings.ToLower(fields[1])

	return entityName, entityUUID, nil
}

// analyzeRef parses:
//   reference string found in schema description
//   actual asset element value
// and returns information about reference:
// isRef - is it reference at all?
// assetName, assetUUID - referenced asset name and UUID
func (r resolver) analyzeRef(description, element string) (isRef bool, assetName, assetUUID string, err error) {
	if !strings.HasPrefix(description, konst.RefDescriptionPrefix) {
		if strings.HasPrefix(description, konst.EntityRefDescriptionPrefix) {
			// description value prefix is entityref, parse actual name and uuid from element
			assetName, assetUUID, err = r.ParseEntityField(element)
			if err != nil {
				return false, "", "", err
			}

			return true, assetName, assetUUID, nil
		}
		// description value prefix is not ref or entityref, cannot be a reference
		return false, "", "", nil
	}

	// description value prefix is ref
	// find first space after REF->{SOMETHING}
	assetNameEnd := strings.Index(description, " ")
	if assetNameEnd == -1 {
		// no space found, take whole description
		assetNameEnd = len(description)
	}

	// take "SOMETHING" from "REF->SOMETHING more text" - assetName
	// assetUUID is directly the value of element
	return true, strings.ToLower(description[len(konst.RefDescriptionPrefix):assetNameEnd]), strings.ToLower(element), nil
}

func (r resolver) getDescriptionJPtr(jptrIn string) (string, error) {
	inFields := strings.Split(jptrIn[1:], konst.JPtrSeparator) // this is immutable fields from parameter, first empty string is skipped
	inIndex := 0                                               // index of currently processed element of inFields

	// output adjusted to schema gets generated here
	outJptrFields := make([]string, 0, len(inFields))

	for inIndex < len(inFields) {
		// we assume that each part of jptr is nested object
		// only exception is numeric array index
		var isIndex bool

		// check potential index, do not care about actual value
		_, err := strconv.Atoi(inFields[inIndex])
		if err == nil {
			isIndex = true
			if inIndex == len(inFields)-1 {
				// index is last elem of jptr, end now
				break
			}
		}

		if isIndex {
			// array index gets replaced with "items" to get to actual schema part
			outJptrFields = append(outJptrFields, "items")
		} else {
			// anything other is expected to be nested object and needs "properties" and actual key added
			outJptrFields = append(outJptrFields, "properties", inFields[inIndex])
		}
		inIndex++ // processed one element of input
	}
	outJptrFields = append(outJptrFields, "description")
	return konst.JPtrSeparator + strings.Join(outJptrFields, konst.JPtrSeparator), nil
}
