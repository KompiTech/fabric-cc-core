package konst

import (
	"fmt"
	"strings"

	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func HasServiceKeys(r rmap.Rmap) bool {
	if r.Exists(AssetDocTypeKey) || r.Exists(AssetVersionKey) || r.Exists(AssetIdKey) || r.Exists(AssetFingerprintKey) {
		return true
	}

	return false
}

func AssetGetVersion(r rmap.Rmap) (int, error) {
	return r.GetInt(AssetVersionKey)
}

func AssetGetDocType(r rmap.Rmap) (string, error) {
	val, err := r.GetString(AssetDocTypeKey)
	if err != nil {
		return "", errors.Wrapf(err, "r.GetString() failed")
	}

	return strings.ToLower(val), nil
}

func AssetGetIDKey(r rmap.Rmap) (string, error) {
	docType, err := AssetGetDocType(r)
	if err != nil {
		return "", errors.Wrapf(err, "assetGetDocType() failed")
	}

	if docType == IdentityAssetName {
		return AssetFingerprintKey, nil
	}
	return AssetIdKey, nil
}

func AssetGetID(r rmap.Rmap) (string, error) {
	idKey, err := AssetGetIDKey(r)
	if err != nil {
		return "", errors.Wrapf(err, "assetGetIDKey() failed")
	}

	value, err := r.GetString(idKey)
	if err != nil {
		return "", errors.Wrapf(err, "r.GetString() failed")
	}

	return value, nil
}

func AssetGetCasbinObject(r rmap.Rmap) (string, error) {
	docType, err := AssetGetDocType(r)
	if err != nil {
		return "", errors.Wrapf(err, "assetGetDocType() failed")
	}

	name, err := AssetGetID(r)
	if err != nil {
		return "", errors.Wrapf(err, "assetGetID() failed")
	}

	return fmt.Sprintf("/%s/%s", docType, name), nil
}

func IsAsset(r rmap.Rmap) (bool, error) {
	if r.Exists(AssetDocTypeKey) {
		docType, err := AssetGetDocType(r)
		if err != nil {
			return false, errors.Wrapf(err, "assetGetDocType() failed")
		}

		var keysToCheck []string
		if docType != IdentityAssetName {
			// anything else than identity uses standard service keys
			keysToCheck = append(keysToCheck, ServiceKeys()...)
		} else {
			// identity is also asset, but doesnt have uuid, has fingerprint
			keysToCheck = []string{AssetFingerprintKey, AssetVersionKey, AssetDocTypeKey}
		}

		for _, key := range keysToCheck {
			if !r.Exists(key) {
				return false, nil
			}
		}
		return true, nil
	}
	// no docType - cannot be an asset
	return false, nil
}
