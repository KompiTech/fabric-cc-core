package testdata

import (
	mockblogicfail2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_blogic/mockblogicfail"
	mockdataafterresolve2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_blogic/mockdataafterresolve"
	mockpaginate2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_blogic/mockpaginate"
	mockrequest2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_blogic/mockrequest"
	mocktimelog2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_blogic/mocktimelog"
	mockworknote2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_blogic/mockworknote"
	mockerror2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_flogic/mockerror"
	mockfunc2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata/mock_flogic/mockfunc"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/blogic/reusable"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/rmap"
)

// funcs here are used to configure engine in main() or in tests
// since this is a library, only tests are used
func getBusinessLogicPolicy() BusinessExecutor {
	bexec := NewBusinessExecutor()

	bexec.SetPolicy(FuncKey{Name: "mockdataafterresolve", Version: 1}, map[Stage][]BusinessPolicyMember{
		AfterResolve: {
			mockdataafterresolve2.TestPassing,
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
			mockrequest2.AfterGetCheck,
		},
		FirstCreate: {
			EnforceCreate,
		},
		BeforeCreate: {
			mockrequest2.UniqueNumber,
		},
		FirstUpdate: {
			EnforceUpdate,
		},
		BeforeUpdate: {
			mockrequest2.ImmutableFields,
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
			mocktimelog2.AttachToMockIncident,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockblogicfail", Version: 1}, map[Stage][]BusinessPolicyMember{
		FirstCreate: {
			mockblogicfail2.Fail,
		},
		AfterQuery: {
			mockblogicfail2.Fail,
		},
		AfterCreate: {
			mockblogicfail2.Fail,
		},
		PatchCreate: {
			mockblogicfail2.Fail,
		},
		PatchUpdate: {
			mockblogicfail2.Fail,
		},
		BeforeUpdate: {
			mockblogicfail2.Fail,
		},
		AfterUpdate: {
			mockblogicfail2.Fail,
		},
		AfterGet: {
			mockblogicfail2.Fail,
		},
		BeforeDelete: {
			mockblogicfail2.Fail,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockblogicfail", Version: -1}, map[Stage][]BusinessPolicyMember{
		BeforeQuery: {
			mockblogicfail2.Fail,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockpaginate", Version: 1}, map[Stage][]BusinessPolicyMember{
		BeforeCreate: {
			mockpaginate2.Paginate,
		},
	})

	bexec.SetPolicy(FuncKey{Name: "mockworknote", Version: 1}, map[Stage][]BusinessPolicyMember{
		BeforeCreate: {
			mockworknote2.AttachToEntity,
		},
	})

	return *bexec
}

func getFunctionLogicPolicy() FunctionExecutor {
	fexec := NewFunctionExecutor()

	fexec.SetPolicy("MockFunc", []FunctionPolicyMember{
		mockfunc2.MockFunc,
	})

	fexec.SetPolicy("MockStateInvalidCreate", []FunctionPolicyMember{
		mockerror2.MockStateInvalidCreate,
	})

	fexec.SetPolicy("MockStateInvalidUpdate", []FunctionPolicyMember{
		mockerror2.MockStateInvalidUpdate,
	})

	fexec.SetPolicy("MockPDInvalidCreate", []FunctionPolicyMember{
		mockerror2.MockPDInvalidCreate,
	})

	fexec.SetPolicy("MockPDInvalidUpdate", []FunctionPolicyMember{
		mockerror2.MockPDInvalidUpdate,
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
		BusinessExecutor:              getBusinessLogicPolicy(),
		FunctionExecutor:              getFunctionLogicPolicy(),
		RecursiveResolveWhitelist:     getRecursiveResolveWhitelist(),
		ResolveBlacklist:              getResolveBlacklist(),
		ResolveFieldsBlacklist:        getResolveFieldsBlacklist(),
		SchemaDefinitionCompatibility: "definitions",
	}
}
