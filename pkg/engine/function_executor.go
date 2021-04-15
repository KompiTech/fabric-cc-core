package engine

import (
	"fmt"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

// FunctionPolicyMember is signature for all functions in policy
type FunctionPolicyMember = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error)

// FunctionPolicy is definition of list of functions to execute for some name
type FunctionPolicy map[string][]FunctionPolicyMember

// FunctionExecutor holds configuration of callable functions
type FunctionExecutor struct {
	policy FunctionPolicy
}

func NewFunctionExecutor() *FunctionExecutor {
	return &FunctionExecutor{
		policy: FunctionPolicy{
			konst.MyAccessFuncName:         []FunctionPolicyMember{myAccessFunc},       // add myAccess built-in function
			konst.UserAccessFuncName:       []FunctionPolicyMember{identityAccessFunc}, // add userAccess built-in function
			konst.UpsertRegistriesFuncName: []FunctionPolicyMember{upsertRegistriesFunc},
			konst.UpsertSingletonsFuncName: []FunctionPolicyMember{upsertSingletonsFunc},
		},
	}
}

func (fe *FunctionExecutor) SetPolicy(key string, members []FunctionPolicyMember) {
	fe.policy[key] = members
}

func (fe FunctionExecutor) getPolicy(key string) []FunctionPolicyMember {
	funcs, exists := fe.policy[key]
	if !exists {
		return nil
	}

	return funcs
}

func (fe FunctionExecutor) Execute(ctx ContextInterface, name string, input rmap.Rmap) (rmap.Rmap, error) {
	funcs := fe.getPolicy(name)
	if len(funcs) == 0 {
		return rmap.Rmap{}, fmt.Errorf("no policy for function '%s' is defined", name)
	}

	output := rmap.NewEmpty()
	for index, fnc := range funcs {
		outputNext, err := fnc(ctx, input, output)
		if err != nil {
			return rmap.Rmap{}, errors.Wrapf(err, "execution of func #%d failed", index)
		}
		output = outputNext
	}

	return output, nil
}

// List returns list of all available functions
func (fe FunctionExecutor) List() []string {
	lst := make([]string, 0, len(fe.policy))

	for name := range fe.policy {
		lst = append(lst, name)
	}

	return lst
}
