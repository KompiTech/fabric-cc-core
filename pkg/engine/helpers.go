package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/kompiguard"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/pkg/errors"
)

// this file contains different pieces of repeated code for use in engine package only
// everything here should be private (lowercase first letter)!
func newRmapFromDestination(ctx ContextInterface, docType, key, destination string, failOnNotFound bool) (rmap.Rmap, error) {
	if destination == StateDestinationValue {
		return newRmapFromState(ctx, key, failOnNotFound)
	}
	return newRmapFromPrivateData(ctx, docType, key, failOnNotFound)
}

func newRmapFromPrivateData(ctx ContextInterface, collectionName, key string, failOnNotFound bool) (rmap.Rmap, error) {
	collectionName = strings.ToUpper(collectionName)
	assetBytes, err := ctx.Stub().GetPrivateData(collectionName, key)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "r.ctx.Stub().GetPrivateData() failed")
	}

	if len(assetBytes) == 0 {
		if failOnNotFound {
			return rmap.Rmap{}, ErrorNotFound(fmt.Sprintf("private data entry not found, collection: %s, key: %s", collectionName, strings.Replace(key, "\x00", "", -1)))
		}
		return rmap.NewEmpty(), nil
	}

	return rmap.NewFromBytes(assetBytes)
}

// Helper to create Rmap from State
// Do not want to make rmap dependent on cckit
func newRmapFromState(ctx ContextInterface, key string, failOnNotFound bool) (rmap.Rmap, error) {
	assetBytes, err := ctx.Stub().GetState(key)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "r.ctx.Stub().GetState() failed")
	}

	if len(assetBytes) == 0 {
		if failOnNotFound {
			return rmap.Rmap{}, ErrorNotFound(fmt.Sprintf("state entry not found: %s", strings.Replace(key, "\x00", "", -1)))
		}
		return rmap.NewEmpty(), nil
	}

	return rmap.NewFromBytes(assetBytes)
}

func putRmapToPrivateData(ctx ContextInterface, collectionName, key string, isCreate bool, rm rmap.Rmap) error {
	collectionName = strings.ToUpper(collectionName)
	// get existing data. If key does not exists, length is 0
	existingData, err := ctx.Stub().GetPrivateData(collectionName, key)
	if err != nil {
		return errors.Wrap(err, "ctx.Stub().GetPrivateData() failed")
	}

	if isCreate && len(existingData) != 0 {
		// when creating, it is error if key already exists
		return ErrorConflict(fmt.Sprintf("private data key already exists: %s", strings.Replace(key, ZeroByte, "", -1)))
	} else if !isCreate && len(existingData) == 0 {
		// when updating, it is error if key does not exists
		return ErrorConflict(fmt.Sprintf("attempt to update non-existent private data key: %s", strings.Replace(key, ZeroByte, "", -1)))
	}

	return ctx.Stub().PutPrivateData(collectionName, key, rm.Bytes())
}

func putRmapToState(ctx ContextInterface, key string, isCreate bool, rm rmap.Rmap) error {
	// get existing data. If key does not exists, length is 0
	existingData, err := ctx.Stub().GetState(key)
	if err != nil {
		return errors.Wrap(err, "ctx.Stub().GetState() failed")
	}

	if isCreate && len(existingData) != 0 {
		// when creating, it is error if key already exists
		return ErrorConflict(fmt.Sprintf("state key already exists: %s", strings.Replace(key, ZeroByte, "", -1)))
	} else if !isCreate && len(existingData) == 0 {
		// when updating, it is error if key does not exists
		return ErrorConflict(fmt.Sprintf("attempt to update non-existent state key: %s", strings.Replace(key, ZeroByte, "", -1)))
	}

	return ctx.Stub().PutState(key, rm.Bytes())
}

// enforceCustomAccess loads identity for this and checks, if identity can do action on object
// this is used for methods where there is no related asset for inferring object
func enforceCustomAccess(reg *Registry, object, action string) error {
	thisIdentity, err := reg.GetThisIdentityResolved()
	if err != nil {
		return errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
	}

	subject, err := AssetGetID(thisIdentity)
	if err != nil {
		return errors.Wrap(err, "konst.AssetGetID(thisIdentity) failed")
	}

	kmpg, err := kompiguard.New()
	if err != nil {
		return errors.Wrap(err, "kompiguard.New() failed")
	}

	if err := kmpg.LoadRoles(thisIdentity); err != nil {
		return errors.Wrap(err, "kmpg.LoadRoles() failed")
	}

	granted, reason, err := kmpg.EnforceCustom(object, subject, action, nil)
	if err != nil {
		return errors.Wrap(err, "kompiguard.New().EnforceCustom() failed")
	}

	if !granted {
		return ErrorForbidden(reason)
	}

	return nil
}

// enforceAssetAccess loads identity for this and enforces standard action for some asset
func enforceAssetAccess(reg *Registry, asset rmap.Rmap, action string) error {
	thisIdentity, err := reg.GetThisIdentityResolved()
	if err != nil {
		return errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
	}

	kmpg, err := kompiguard.New()
	if err != nil {
		return errors.Wrap(err, "kompiguard.New() failed")
	}

	granted, reason, err := kmpg.EnforceAsset(asset, thisIdentity, action)
	if err != nil {
		return errors.Wrap(err, "kompiguard.New().EnforceAsset() failed")
	}

	if !granted {
		return ErrorForbidden(reason)
	}

	return nil
}

func GetMyFingerprint(ctx ContextInterface) (string, error) {
	myCert, err := cid.GetX509Certificate(ctx.Stub())
	if err != nil {
		return "", errors.Wrap(err, "cid.GetX509Certificate() failed")
	}

	//call function instead of using hardcoded SHA512 as before
	return ctx.GetConfiguration().CurrentIDFunc(myCert)
}

func GetMyOrgName(ctx ContextInterface) (string, error) {
	myCert, err := cid.GetX509Certificate(ctx.Stub())
	if err != nil {
		return "", err
	}

	//testing certs have issuer CN identical to subject CN ({user_id}.{org_id}.kompitech.com)
	//real certs have issuer CN: {org_id}.kompitech.com, subject CN: {user_id}.{org_id}.kompitech.com
	//this code will always return {org_id}.kompitech.com in both cases
	fields := strings.Split(myCert.Issuer.CommonName, ".")

	if len(fields) == 3 { // {org_id}.kompitech.com , no transformation needed
		return myCert.Issuer.CommonName, nil
	} else if len(fields) == 4 { // {user_id}.{org_id}.kompitech.com , remove {user_id}
		return strings.Join(fields[1:], "."), nil
	} else {
		return "", fmt.Errorf("unexpected cert.Issuer.CommonName: %s", myCert.Issuer.CommonName)
	}
}

func MakeUUID(now time.Time) (string, error) {
	rand.Seed(now.UnixNano()) // seed RNG with this TX time, this will make all peers to generate the same UUID

	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])), nil
}

// GetUUIDGeneratingFunc returns func that returns unique UUID and returned func can be called repeatedly in one transaction
func GetUUIDGeneratingFunc() (f func(ctx ContextInterface) (string, error)) {
	var i int

	return func(ctx ContextInterface) (string, error) {
		now, err := ctx.Time()
		if err != nil {
			return "", err
		}

		uuid, err := MakeUUID(now.Add(time.Duration(i) * time.Nanosecond))
		if err != nil {
			return "", err
		}

		i++

		return uuid, nil
	}
}
