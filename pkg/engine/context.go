package engine

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Context struct {
	contractapi.TransactionContext
	store    map[string]interface{}
	confFunc func() Configuration
}

func (ctx *Context) SetConfFunc(fnc func() Configuration) {
	ctx.confFunc = fnc
}

func (ctx *Context) GetConfiguration() Configuration {
	return ctx.confFunc()
}

func (ctx *Context) GetRegistry() *Registry {
	return ctx.Get(konst.RegistryKey).(*Registry)
}

func (ctx *Context) Get(key string) interface{} {
	return ctx.store[key]
}

func (ctx *Context) Set(key string, value interface{}) {
	if ctx.store == nil {
		ctx.store = map[string]interface{}{}
	}

	ctx.store[key] = value
}

func (ctx *Context) Stub() shim.ChaincodeStubInterface {
	return ctx.GetStub()
}

func (ctx *Context) Time() (time.Time, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(txTimestamp.GetSeconds(), int64(txTimestamp.GetNanos())), nil
}

// cckit provided arguments in form of Arg<Type>(string name)
// since we are not using router anymore, code needs to know about argument names to be able to have this functionality
// this is related code not existent in current impl anymore
// https://github.com/KompiTech/fabric-cc-core/v2/src/blob/6ba28e33e1f7ec6d81f561a97c6a788f25971259/engine/router.go#L50
func (ctx *Context) getArgNames() []string {
	argInfo := map[string][]string{
		"init":                 {"input"},
		"assetCreate":          {"name", "data", "version", "id"},
		"assetCreateDirect":    {"name", "data", "version", "id"},
		"assetDelete":          {"name", "id"},
		"assetDeleteDirect":    {"name", "id"},
		"assetGet":             {"name", "id", "resolve", "data"},
		"assetGetDirect":       {"name", "id", "resolve"},
		"assetHistory":         {"name", "id"},
		"assetMigrate":         {"name", "id", "patch", "version"},
		"assetUpdate":          {"name", "id", "patch"},
		"assetUpdateDirect":    {"name", "id", "patch"},
		"assetQuery":           {"name", "query", "resolve"},
		"assetQueryDirect":     {"name", "query", "resolve"},
		"changelogGet":         {"number"},
		"changelogList":        {},
		"functionInvoke":       {"name", "input"},
		"functionQuery":        {"name", "input"},
		"identityAddMe":        {"input"},
		"identityGet":          {"fingerprint", "resolve", "data"},
		"identityMe":           {"resolve"},
		"identityUpdate":       {"fingerprint", "patch"},
		"identityQuery":        {"query", "resolve"},
		"identityCreateDirect": {"data", "id"},
		"identityUpdateDirect": {"id", "patch"},
		"registryGet":          {"name", "version"},
		"registryUpsert":       {"name", "data"},
		"registryList":         {},
		"roleGet":              {"id", "data"},
		"roleCreate":           {"data", "id"},
		"roleUpdate":           {"id", "patch"},
		"roleQuery":            {"query"},
		"singletonGet":         {"name", "version"},
		"singletonUpsert":      {"name", "data"},
		"singletonList":        {},
	}
	method, _ := ctx.Stub().GetFunctionAndParameters()
	namedArgs := argInfo[method]
	if namedArgs == nil {
		// older cc-core used always first lowercase letter in method name
		lowerMethod := strings.ToLower(string(method[0])) + method[1:]
		namedArgs = argInfo[lowerMethod]

		if namedArgs == nil {
			return nil
		}
	}
	return namedArgs
}

func (ctx *Context) paramByName(name string) ([]byte, error) {
	argNames := ctx.getArgNames()
	if argNames == nil {
		// this should never happen, router will not allow it
		return nil, errors.New("unknown param: " + name)
	}

	for argIdx, namedArg := range argNames {
		if namedArg == name {
			args := ctx.Stub().GetArgs()
			if len(args) <= argIdx+1 {
				return nil, errors.New("param value for: " + name + " is not present")
			}

			return args[argIdx+1], nil // first arg is always function name
		}
	}

	// again, should not ever happen
	return nil, errors.New("unknown param: " + name)
}

func (ctx *Context) Params() map[string]interface{} {
	return ctx.mapArgs()
}

func (ctx *Context) Param(name string) (interface{}, error) {
	return ctx.paramByName(name)
}

func (ctx *Context) ParamString(name string) (string, error) {
	val, err := ctx.paramByName(name)
	if err != nil {
		return "", err
	}

	return string(val), nil
}

func (ctx *Context) ParamBytes(name string) ([]byte, error) {
	return ctx.paramByName(name)
}

func (ctx *Context) ParamInt(name string) (int, error) {
	valB, err := ctx.paramByName(name)
	if err != nil {
		return -1, err
	}

	val, err := strconv.Atoi(string(valB))
	if err != nil {
		return -1, err
	}

	return val, nil
}

func (ctx *Context) ParamBool(name string) (bool, error) {
	valS, err := ctx.ParamString(name)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(valS)
}

func (ctx *Context) mapArgs() map[string]interface{} {
	argNames := ctx.getArgNames()
	args := ctx.GetStub().GetArgs()
	outMap := make(map[string]interface{}, len(argNames))

	for argIdx, argName := range ctx.getArgNames() {
		outMap[argName] = args[argIdx+1] // first arg is always function name
	}

	return outMap
}

func (ctx *Context) Logger() Logger {
	return NewLogger()
}
