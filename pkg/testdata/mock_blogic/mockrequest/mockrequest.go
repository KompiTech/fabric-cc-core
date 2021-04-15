package mockrequest

import (
	"fmt"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// these are required for tests to pass

// AfterGetCheck is for checking parameter passing to AfterGet stage
// if it detects magic value, it returns error
var AfterGetCheck = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	if prePatch == nil {
		return rmap.Rmap{}, errors.New("invalid prePatch value: nil")
	}

	if prePatch.Exists("magic") {
		value, err := prePatch.GetJPtrString("/magic")
		if err != nil {
			return rmap.Rmap{}, errors.Wrap(err, "prePatch.GetJPtrString() failed")
		}

		if value == "value" {
			return rmap.Rmap{}, errors.New("magic value found")
		}
	}

	return postPatch, nil
}

// ImmutableFields enforces fields that cannot be changed after creation
var ImmutableFields = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	for _, fieldName := range []string{"number"} {
		if prePatch.Mapa[fieldName] != postPatch.Mapa[fieldName] {
			return rmap.Rmap{}, fmt.Errorf("field: %s is not mutable", fieldName)
		}
	}
	return postPatch, nil
}

// UniqueNumber enforces unique number in State
var UniqueNumber = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	numberI, exists := postPatch.Mapa["number"]
	if !exists {
		// no number, do nothing
		return postPatch, nil
	}

	//var number string
	number := numberI.(string)

	docType := strings.ToUpper(postPatch.Mapa[konst.AssetDocTypeKey].(string))
	key := "number"

	err := keyMustBeUnique(ctx, docType, key, number)
	if err != nil {
		return rmap.Rmap{}, err
	}

	return postPatch, nil
}

// keyMustBeUnique returns true if value stored in key is not already used, otherwise return false
func keyMustBeUnique(ctx engine.ContextInterface, docType string, key, value string) error {
	docType = strings.ToUpper(docType)
	queryMap := map[string]interface{}{
		"selector": map[string]interface{}{
			konst.AssetDocTypeKey: docType,
			key:                   value,
		},
	}

	queryBytes, err := json.Marshal(queryMap)
	if err != nil {
		return fmt.Errorf("could not determine number uniqueness: %v", err)
	}

	iter, err := ctx.Stub().GetQueryResult(string(queryBytes))
	if err != nil {
		return fmt.Errorf("could not determine number uniqueness: %v", err)
	}

	for iter.HasNext() {
		return fmt.Errorf("%s %s: %s is not unique", strings.Title(strings.ToLower(docType)), key, value)
	}

	return nil
}
