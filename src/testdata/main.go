package testdata

import (
	. "github.com/KompiTech/fabric-cc-core/v2/src/blogic/reusable"
	. "github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_blogic/mockblogicfail"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_blogic/mockdataafterresolve"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_blogic/mockpaginate"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_blogic/mockrequest"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_blogic/mocktimelog"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_blogic/mockworknote"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_flogic/mockerror"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata/mock_flogic/mockfunc"
	"github.com/KompiTech/rmap"
)

// funcs here are used to configure engine in main() or in tests
// since this is a library, only tests are used
func getBusinessLogicPolicy() BusinessExecutor {
	bexec := NewBusinessExecutor()

	bexec.SetPolicy(FuncKey{Name: "mockdataafterresolve", Version: 1}, map[Stage][]BusinessPolicyMember{
		AfterResolve: {
			mockdataafterresolve.TestPassing,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockincident", Version: 1}, map[Stage][]BusinessPolicyMember{
		AfterGet: {
			EnforceRead,
		},
		FirstCreate: {
			EnforceCreate,
		},
		FirstUpdate: {
			EnforceUpdate,
		},
		AfterQuery: {
			FilterRead,
		}},
	)

	bexec.SetPolicy(FuncKey{Name: "mockrequest", Version: 1}, map[Stage][]BusinessPolicyMember{
		AfterGet: {
			EnforceRead,
			mockrequest.AfterGetCheck,
		},
		FirstCreate: {
			EnforceCreate,
		},
		BeforeCreate: {
			mockrequest.UniqueNumber,
		},
		FirstUpdate: {
			EnforceUpdate,
		},
		BeforeUpdate: {
			mockrequest.ImmutableFields,
		},
		AfterQuery: {
			FilterRead,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mocktimelog", Version: 1}, map[Stage][]BusinessPolicyMember{
		FirstCreate: {
			EnforceCreate,
		},
		AfterCreate: {
			mocktimelog.AttachToMockIncident,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockblogicfail", Version: 1}, map[Stage][]BusinessPolicyMember{
		FirstCreate: {
			mockblogicfail.Fail,
		},
		AfterQuery: {
			mockblogicfail.Fail,
		},
		AfterCreate: {
			mockblogicfail.Fail,
		},
		PatchCreate: {
			mockblogicfail.Fail,
		},
		PatchUpdate: {
			mockblogicfail.Fail,
		},
		BeforeUpdate: {
			mockblogicfail.Fail,
		},
		AfterUpdate: {
			mockblogicfail.Fail,
		},
		AfterGet: {
			mockblogicfail.Fail,
		},
		BeforeDelete: {
			mockblogicfail.Fail,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockblogicfail", Version: -1}, map[Stage][]BusinessPolicyMember{
		BeforeQuery: {
			mockblogicfail.Fail,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockpaginate", Version: 1}, map[Stage][]BusinessPolicyMember{
		BeforeCreate: {
			mockpaginate.Paginate,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockworknote", Version: 1}, map[Stage][]BusinessPolicyMember{
		BeforeCreate: {
			mockworknote.AttachToEntity,
		},
	})

	return *bexec
}

func getFunctionLogicPolicy() FunctionExecutor {
	fexec := NewFunctionExecutor()

	fexec.SetPolicy("MockFunc", []FunctionPolicyMember{
		mockfunc.MockFunc,
	})

	fexec.SetPolicy("MockStateInvalidCreate", []FunctionPolicyMember{
		mockerror.MockStateInvalidCreate,
	})

	fexec.SetPolicy("MockStateInvalidUpdate", []FunctionPolicyMember{
		mockerror.MockStateInvalidUpdate,
	})

	fexec.SetPolicy("MockPDInvalidCreate", []FunctionPolicyMember{
		mockerror.MockPDInvalidCreate,
	})

	fexec.SetPolicy("MockPDInvalidUpdate", []FunctionPolicyMember{
		mockerror.MockPDInvalidUpdate,
	})

	return *fexec
}

func getRecursiveResolveWhitelist() rmap.Rmap {
	rrw, _ := rmap.NewFromSlice([]interface{}{"mocklevel1.level2", "mocklevel2.level3"})
	return rrw
}

func getResolveBlacklist() rmap.Rmap {
	rbl, _ := rmap.NewFromSlice([]interface{}{"mockblacklisted"})
	return rbl
}

func getResolveFieldsBlacklist() rmap.Rmap {
	return rmap.NewFromMap(map[string]interface{}{
		"mockreffieldblacklist": map[string]interface{}{
			"blacklisted":             struct{}{},
			"nested.blacklisted_nest": struct{}{},
		},
	})
}

func GetConfiguration() Configuration {
	return Configuration{
		BusinessExecutor:          getBusinessLogicPolicy(),
		FunctionExecutor:          getFunctionLogicPolicy(),
		RecursiveResolveWhitelist: getRecursiveResolveWhitelist(),
		ResolveBlacklist:          getResolveBlacklist(),
		ResolveFieldsBlacklist:    getResolveFieldsBlacklist(),
	}
}
