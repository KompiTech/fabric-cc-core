package cc_core

import (
	testdata2 "github.com/KompiTech/fabric-cc-core/v2/internal/testdata"
	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/fabric-cc-core/v2/pkg/testing"
)

// returns default TestContext for most tests
func getDefaultTextContext() *testing.TestContext {
	eng := testdata2.GetConfiguration()
	eng.CurrentIDFunc = engine.CertSHA512IDFunc
	return testing.NewTestContext("mock", eng, nil, nil)
}
