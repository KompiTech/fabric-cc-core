package cc_core

import (
	"github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/fabric-cc-core/v2/src/testdata"
	"github.com/KompiTech/fabric-cc-core/v2/src/testing"
)

// returns default TestContext for most tests
func getDefaultTextContext() *testing.TestContext {
	eng := testdata.GetConfiguration()
	eng.CurrentIDFunc = engine.CertSHA512IDFunc
	return testing.NewTestContext("mock", eng, nil, nil)
}
