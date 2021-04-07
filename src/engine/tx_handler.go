package engine

import (
	"errors"
	"strconv"
	"strings"

	"github.com/KompiTech/rmap"
)

// makeInitializationFunc prepares ConfigInterface with chaincode configuration and registry
// returns in it form of function that is called before all transactions. This func must also manage checking of identity
func makeInitializationFunc(config Configuration) func(ContextInterface) error {
	confFunc := func() Configuration {
		return config
	}

	// this func is registered in Contract as BeforeTransaction handler
	return func(ctx ContextInterface) error {
		reg, err := newRegistry(ctx)
		if err != nil {
			return err
		}

		ctx.Set("registry", reg)
		ctx.SetConfFunc(confFunc)

		funcName, _ := ctx.GetStub().GetFunctionAndParameters()
		skipFuncs := []string{"IdentityAddMe", "IdentityMe", "Init"} // do not check identity existence when these funcs are called

		isSkip := false
		for _, skipFunc := range skipFuncs {
			if iCompare(skipFunc, funcName) {
				isSkip = true
				break
			}
		}

		if !isSkip {
			_, err := reg.GetThisIdentity()
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func iCompare(us1, s2 string) bool {
	if us1 == s2 {
		return true
	}

	if (strings.ToLower(string(us1[0])) + us1[1:]) == s2 {
		return true
	}

	return false
}

// We use unknownTransactionHandler as our entrypoint for routing, because this allows us more flexibility for error reporting and argument validation
// Standard Fabric contract public methods are not used
func unknownTransactionHandler(ctx ContextInterface) (string, error) {
	args := ctx.GetStub().GetArgs()
	traceEnabled := false

	tracingInfo, err := rmap.NewFromBytes(args[len(args)-1])
	if err == nil {
		// tracing info is a JSON, continue
		traceRequested, err := tracingInfo.GetBool("trace")
		if err == nil {
			// tracing key "trace" was found
			if traceRequested {
				// tracing key has boolean true value, enable
				traceEnabled = true
				delete(tracingInfo.Mapa, "trace")
			}
		}
	}

	// call correct CC method and get response
	ret, err := route(ctx)
	if err != nil && traceEnabled {
		// some CC error occured and tracing is enabled, append to error message
		err = errors.New(addTracingMessage(err.Error(), tracingInfo.String()))
	}

	// TODO logic of setting correct HTTP status must be probably done on rest server
	// this function could return peer.Response, but then all responses are wrapped in "message" key
	// breaking all tests
	return ret, err
}

func addTracingMessage(msg, tmsg string) string {
	// message could contain our error code in format |||NNN. If that is the case, we want to move the code to the end of the error message
	if len(msg) < 6 {
		return msg + tmsg
	}

	pipes := msg[len(msg)-6 : (len(msg) - 3)]
	if pipes == "|||" {
		// last 3 chars must be number
		errCode := msg[len(msg)-3:]
		_, err := strconv.Atoi(errCode)
		if err == nil {
			// all matched, construct a new error message
			fmsg := msg[:len(msg)-6] + tmsg + pipes + errCode
			return fmsg
		}
	}

	// kompitech error code is not present, append to end without modification of msg
	return msg + tmsg
}
