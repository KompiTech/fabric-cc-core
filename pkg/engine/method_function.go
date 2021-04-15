package engine

import (
	"strings"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func functionInvokeFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	input, err := ctx.ParamString(InputParam)
	if err != nil {
		return "", err
	}

	return functionBackend(ctx, name, true, input)
}

func functionQueryFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	input, err := ctx.ParamString(InputParam)
	if err != nil {
		return "", err
	}

	return functionBackend(ctx, name, false, input)
}

func functionBackend(ctx ContextInterface, name string, isInvoke bool, data string) (string, error) {
	var verb string
	if isInvoke {
		verb = FunctionInvokeVerb
	} else {
		verb = FunctionQueryVerb
	}

	reg := ctx.Get(RegistryKey).(*Registry)
	object := "/" + FunctionCasbinName + "/" + verb + "/" + name

	// query function called myAccess can be accessed without any grants
	// this is the only exception (for now)
	if !strings.EqualFold(object, "/"+FunctionCasbinName+"/query/"+MyAccessFuncName) {
		if err := enforceCustomAccess(reg, object, ExecuteAction); err != nil {
			return "", err
		}
	}

	var input rmap.Rmap

	if len(data) == 0 {
		input = rmap.NewEmpty()
	} else {
		var err error
		input, err = rmap.NewFromString(data)
		if err != nil {
			return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
		}
	}

	fexec := ctx.GetConfiguration().FunctionExecutor

	output, err := fexec.Execute(ctx, name, input)
	if err != nil {
		return "", errors.Wrap(err, "fexec.Execute() failed")
	}

	return string(output.WrappedResultBytes()), nil
}
