package main

import (
	"crypto/x509"
	"errors"
	"log"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/src/blogic/reusable"
	. "github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/rmap"
)

// example of business logic that uses singleton for configuration
var CheckBannedWords = func(ctx ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.NewEmpty()

	// fetch singleton from State
	// use version -1 to get the latest
	bwSing, _, err := ctx.GetRegistry().GetSingleton("banned_words", -1)
	if err != nil {
		return null, err
	}

	// singletons are always wrapped in "value" extra key
	// rmap allows easily to iterate over lists of different types
	// you can also use JSON pointer to access nested objects in one call
	// if the JSON pointer path does not exist or has invalid type, error is returned
	bwList, err := bwSing.GetIterableStringJPtr("/value/banned_words")
	if err != nil {
		return null, err
	}

	// fetch the name from asset, postPatch parameter will contain asset to be created when part of Create stage
	// or it will contain asset after applying the patch when part of Update stage
	// this allows to reuse this method in both cases
	name, err := postPatch.GetString("name")
	if err != nil {
		return null, err
	}

	nameWords := strings.Split(name, " ")

	// do a naive case-insensitive search for banned word(s)
	for _, bw := range bwList {
		for _, word := range nameWords {
			if strings.ToLower(bw) == strings.ToLower(word) {
				// by returning error, the transaction is rolled back and error is returned to client
				return null, errors.New("name contains banned word(s)")
			}
		}
	}

	// when there are no banned words, we return the same asset as entered this function
	return postPatch, nil
}

// example of function, that returns number of books written by each author name
var AuthorStats = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}

	// define query, we want to iterate through all available Book assets, so no selector is necessary
	// we can use "fields" key to restrict returned keys to save memory (we are only interested in author's name)
	// more info can be found in CouchDB docs
	bookQuery := rmap.NewFromMap(map[string]interface{}{
		"fields": []interface{}{"authors"},
	})

	// you can execute rich CouchDB queries in two ways:
	// .QueryAssets() -> returns list directly as array
	// .GetQueryIterator() -> returns iterator, preferred way, because the result set is not stored in RAM
	// fabric has additional limitation -> when using paginated query, the transaction can by only read-only
	// in this case, the TX is always read only, so it is not an issue
	// but always use empty bookmark and -1 pageSize to not use pagination
	// second returned value is bookmark, which is not used
	iter, _, err := ctx.GetRegistry().GetQueryIterator("book", bookQuery, "", -1)
	if err != nil {
		return null, err
	}

	// remember to close iterator
	defer (func() {
		_ = iter.Close()
	})()

	// prepare the output. Key will be author name, value will be number of books written
	result := rmap.NewEmpty()

	for {
		// load next asset from iterator. We also want to have book resolved, so we can get Author name in one go
		nextBook, err := iter.Next(true)
		if err != nil {
			return null, err
		}

		if nextBook == nil {
			// end of iterator
			break
		}

		// Book can have multiple Authors, we need to iterate over "authors" array. Since we did resolve=true in iter.Next() call, it will contain Author asset, instead of its key.
		authorIter, err := nextBook.GetIterable("authors")
		if err != nil {
			return null, err
		}

		for _, authorI := range authorIter {
			author, err := rmap.NewFromInterface(authorI)
			if err != nil {
				return null, err
			}

			authorName, err := author.GetString("name")
			if err != nil {
				return null, err
			}

			// when author key does not exists in result, set it to 1 (first encountered Book from this Author)
			// when author key exists, increment it
			// you can always use .Mapa to directly access underlying map
			// we dont use first returned value, because it is interface{} and not comfortable to use
			_, authorExists := result.Mapa[authorName]
			if !authorExists {
				result.Mapa[authorName] = 1 // initialize key to 1
			} else {
				// for increment, we use rmap.GetInt() to handle that key and its type are OK, instead of handling interface{}
				oldCounter, err := result.GetInt(authorName)
				if err != nil {
					return null, err
				}

				result.Mapa[authorName] = oldCounter + 1 // increment and overwrite
			}
		}
	}

	// all Books processed, return result
	return result, nil
}

func GetConfiguration() Configuration {
	// when you want to customize business logic, you define a policy
	bexec := NewBusinessExecutor()

	// each policy is valid for some asset name and version. Version -1 means any version and is useful for development.
	// for production chaincode, you shold always use some concrete version - this will allow you to update the chaincode with newer policies for newer versions, while leaving the old intact
	bexec.SetPolicy(FuncKey{Name: "author", Version: -1}, StageMembers{
		// each policy can contain multiple stages, which define, when are the functions in stage executed
		// functions in stage are executed in order, and must have this BusinessPolicyMember signature:
		// params:
		//   ctx - standard context
		//   prePatch - the original asset loaded from storage. If operation is not update, this is nil
		//   postPatch - the asset from previous function or storage if this function is first
		// return value:
		//   rmap.Rmap - the asset to be sent to next function
		//   error - if this is not nil, function chain is immediately terminated and error is returned

		BeforeCreate: { // this means that functions are executed before creating a new asset instance
			reusable.EnforceCreate, // this function checks, if identity has create grant on asset, otherwise returns error
			CheckBannedWords,       // example business logic that works with singleton
		},
		BeforeUpdate: { // functions are executed before existing asset is updated
			reusable.EnforceUpdate, // checks the update grant on asset
			CheckBannedWords,
		},
		BeforeDelete: { // functions are executed before existing asset is deleted
			reusable.Deny, // this function always returns errors, disabling the functionality
		},
	})

	bexec.SetPolicy(FuncKey{Name: "book", Version: -1}, StageMembers{
		BeforeCreate: {
			reusable.EnforceCreate,
		},
		BeforeUpdate: {
			reusable.EnforceUpdate,
		},
		BeforeDelete: {
			reusable.Deny,
		},
	})

	// you can use custom FunctionExecutor to define your custom functions
	fexec := NewFunctionExecutor()

	// we are defining authorStats function, that will use the respective implementation
	fexec.SetPolicy("authorStats", []FunctionPolicyMember{AuthorStats})

	// the Configuration struct fully describes a cc-core chaincode
	return Configuration{
		BusinessExecutor:          *bexec,
		FunctionExecutor:          *fexec,
		RecursiveResolveWhitelist: rmap.NewEmpty(),
		ResolveBlacklist:          rmap.NewEmpty(),
		ResolveFieldsBlacklist:    rmap.NewEmpty(),
		CurrentIDFunc:             certSubjectCNIDFunc,
		PreviousIDFunc:            nil,
	}
}

// this func is required to get client identity string from cert
var certSubjectCNIDFunc = func(cert *x509.Certificate) (string, error) {
	return cert.Subject.CommonName, nil
}

func main() {
	// create contractapi.ContractChaincode from our custom Configuration
	cc, err := NewChaincode(GetConfiguration())
	if err != nil {
		log.Panicf("Error creating library chaincode: %v", err)
	}

	// start serving requests
	if err := cc.Start(); err != nil {
		log.Panicf("Error starting library chaincode: %v", err)
	}
}
