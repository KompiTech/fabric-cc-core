package testing

import (
	"errors"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"strings"
	"time"

	examplecert2 "github.com/KompiTech/fabric-cc-core/v2/internal/examplecert"
	"github.com/KompiTech/fabric-cc-core/v2/internal/testing"
	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	. "github.com/KompiTech/rmap"
	. "github.com/onsi/gomega"

	expectcc "github.com/KompiTech/fabric-cc-core/v2/pkg/testing/expect"
)

// NewTestContext creates a new *TestContext
// if mockStub and/or CouchDBMock are nil, then they are recreated, otherwise existing instance is reused (useful for mocking chaincode update with old data still present)
func NewTestContext(name string, config engine.Configuration, mockStubIn *testing.MockStub, couchDBMockIn *testing.CouchDBMock) *TestContext {
	chaincode, err := engine.NewChaincode(config)
	if err != nil {
		panic(err)
	}

	var mockStub *testing.MockStub
	if mockStubIn == nil {
		mockStub = testing.NewMockStub(name, chaincode) // get new MockStub for testing
	} else {
		mockStub = mockStubIn   // reuse existing MockStub
		mockStub.Cc = chaincode // when reusing MockStub, always use Cc from engine to have latest changes
	}

	var couchDBMock *testing.CouchDBMock
	if couchDBMockIn == nil {
		couchDBMock = testing.NewCouchDBMock()
	} else {
		couchDBMock = couchDBMockIn
	}

	mockStub.CouchDBMock = couchDBMock

	// load actor certificates
	actors, err := testing.IdentitiesFromFiles(`SOME_MSP`, map[string]string{
		"superUser":    "s7techlab.pem",
		"ordinaryUser": "victor-nosov.pem",
		"nobodyUser":   "some-person.pem",
	}, examplecert2.Content)
	if err != nil {
		log.Fatal(err)
	}

	return &TestContext{
		cc:                mockStub,
		currentActorName:  "superUser", //superUser is default actor
		actors:            actors,
		AssetsAllowed:     map[string]struct{}{},
		SingletonsAllowed: map[string]struct{}{},
		idFunc:            config.CurrentIDFunc, // func used to get identity ID
	}
}

type TestContext struct {
	cc                *testing.MockStub
	currentActorName  string
	actors            testing.Identities
	AssetsAllowed     map[string]struct{}
	SingletonsAllowed map[string]struct{}
	idFunc            engine.IDFunc
}

func (tctx *TestContext) Wait() {
	tctx.cc.CouchDBMock.Wait()
}

func (tctx *TestContext) GetCouchDBMock() *testing.CouchDBMock {
	return tctx.cc.CouchDBMock
}

func (tctx *TestContext) GetMockStub() *testing.MockStub {
	return tctx.cc
}

func (tctx *TestContext) IsAssetAllowed(name string) bool {
	_, isAllowed := tctx.AssetsAllowed[strings.ToLower(name)]
	return isAllowed
}

func (tctx *TestContext) AllowAsset(name string) {
	tctx.AssetsAllowed[strings.ToLower(name)] = struct{}{}
}

func (tctx *TestContext) AllowAssets(names []string) {
	for _, schemaName := range names {
		tctx.AllowAsset(schemaName)
	}
}

func (tctx *TestContext) AllowSingleton(name string) {
	tctx.SingletonsAllowed[strings.ToLower(name)] = struct{}{}
}

func (tctx *TestContext) AllowSingletons(names []string) {
	for _, name := range names {
		tctx.AllowSingleton(name)
	}
}

// AddActor adds actorName with identity ref to this TestContext
func (tctx *TestContext) AddActor(actorName string, identity *testing.Identity) {
	tctx.actors[actorName] = identity
}

// MustAddActorFromPemFile
func (tctx *TestContext) MustAddActorFromPemFile(actorName, filePath string) {
	identity, err := testing.IdentityFromFile("SOME_MSP", filePath, content)
	if err != nil {
		log.Fatal(err)
	}

	tctx.AddActor(actorName, identity)
}

// ClearEvents clears all previously received events
func (tctx *TestContext) ClearEvents() {
	tctx.cc.ClearEvents()
}

// SetTime changes mock now() time to some time.Time value
func (tctx *TestContext) SetTime(value time.Time) {
	tctx.cc.SetTime(value)
}

// TravelInTime changes mock now() time to some secs in future
func (tctx *TestContext) TravelInTime(sec int64) {
	tctx.cc.TravelInTime(sec, 0)
}

// TravelInTime changes mock now() time to some secs in future
func (tctx *TestContext) TravelInTimeNs(sec int64, nsec int32) {
	tctx.cc.TravelInTime(sec, nsec)
}

// AdvanceTime increases mock time by smallest possible increment, to generate different UUID
func (tctx *TestContext) AdvanceTime() {
	tctx.cc.TravelInTime(0, 1)
}

// ResetTime restores original time handling
func (tctx *TestContext) ResetTime() {
	tctx.cc.ResetTime()
}

// GetLastEventPayload returns data for last received event
func (tctx *TestContext) GetLastEventPayload() map[string]interface{} {
	return tctx.cc.GetLastEventPayload()
}

// GetCC returns MockStub ref
func (tctx *TestContext) GetCC() *testing.MockStub {
	return tctx.cc
}

// SetChannelID sets channel ID in MockStub
func (tctx *TestContext) SetChannelID(id string) {
	tctx.cc.ChannelID = id
}

// GetCurrentActor returns ref to identity for current actor
func (tctx *TestContext) GetCurrentActor() *testing.Identity {
	return tctx.actors[tctx.currentActorName]
}

// GetCurrentActorName returns current actor's name
func (tctx *TestContext) GetCurrentActorName() string {
	return tctx.currentActorName
}

func (tctx *TestContext) getActorFingerprint(actor *testing.Identity) string {
	fp, err := tctx.idFunc(actor.Certificate)
	if err != nil {
		panic(err)
	}
	return fp
}

func (tctx *TestContext) GetActorFingerprint(actorName string) string {
	return tctx.getActorFingerprint(tctx.actors[actorName])
}

func (tctx *TestContext) GetCurrentActorFingerprint() string {
	return tctx.getActorFingerprint(tctx.GetCurrentActor())
}

func (tctx *TestContext) GetCurrentActorOrgName() string {
	return tctx.GetCurrentActor().Certificate.Issuer.CommonName
}

// InitOk executes CC init method and expects no error
func (tctx *TestContext) InitOk(iargs ...interface{}) {
	expectcc.ResponseOk(tctx.cc.From(tctx.GetCurrentActor()).Init(iargs...))
}

// InitError executes CC init method with Args and expects error with errorSubstr to be present
func (tctx *TestContext) InitError(errorSubstr string, iargs ...interface{}) {
	ResponseErrorSubstr(tctx.cc.From(tctx.GetCurrentActor()).Init(iargs...), errorSubstr)
}

// Ok invokes mock CC method with Args and expects no error. Result is thrown away
func (tctx *TestContext) Ok(funcName string, iargs ...interface{}) {
	expectcc.ResponseOk(tctx.cc.From(tctx.GetCurrentActor()).Invoke(funcName, iargs...))
}

// Errors invokes mock CC method with Args and expects error with errorSubstr to be present
func (tctx *TestContext) Error(errorSubstr string, funcName string, iargs ...interface{}) {
	ResponseErrorSubstr(tctx.cc.From(tctx.GetCurrentActor()).Invoke(funcName, iargs...), errorSubstr)
}

// Rmap invokes mock CC method with Args and returns "result" key of response as Rmap
func (tctx *TestContext) Rmap(funcName string, iargs ...interface{}) Rmap {
	rm := tctx.invoke(funcName, iargs...)
	output, err := rm.GetRmap("result")
	Expect(err).To(BeNil())
	return output
}

// RmapNoResult invokes mock CC method with Args and returns result as Rmap
func (tctx *TestContext) RmapNoResult(funcName string, iargs ...interface{}) Rmap {
	return tctx.invoke(funcName, iargs...)
}

// JSON invokes mock CC method with Args and returns "result" key of response
func (tctx *TestContext) JSON(funcName string, iargs ...interface{}) map[string]interface{} {
	return tctx.Rmap(funcName, iargs...).Mapa
}

// JSONNoResult invokes mock CC method with Args and returns result directly
func (tctx *TestContext) JSONNoResult(funcName string, iargs ...interface{}) map[string]interface{} {
	rm := tctx.invoke(funcName, iargs...)
	return rm.Mapa
}

// invoke for mock CC method invoking
func (tctx *TestContext) invoke(funcName string, iargs ...interface{}) Rmap {
	//var bytes []byte
	//data := expectcc.PayloadIs(tctx.Cc.From(tctx.GetCurrentActor()).Invoke(funcName, iargs...), &bytes).([]byte)
	response := expectcc.ResponseOk(tctx.cc.From(tctx.GetCurrentActor()).Invoke(funcName, iargs...))
	rm, err := NewFromBytes(response.Payload)
	Expect(err).To(BeNil())
	return rm
}

// SetActor sets current actor by name
func (tctx *TestContext) SetActor(actorName string) {
	if _, exists := tctx.actors[actorName]; exists {
		tctx.currentActorName = actorName
	} else {
		log.Fatalf("actorName: %s does not exist", actorName)
	}
}

// GetInit returns JSON for current actor as init_manager
// registryPath and singletonPath contains path to respective file system locations to scan and load
// if empty string is supplied to either, this section is not generated
// scanning takes into account AssetsAllowed and SingletonsAllowed (but only if they are not-empty)
func (tctx *TestContext) GetInit(registryPath, singletonPath string) Rmap {
	init := NewFromMap(map[string]interface{}{
		"init_manager": tctx.GetCurrentActorFingerprint(),
	})

	if registryPath != "" {
		if len(tctx.AssetsAllowed) == 0 {
			// no allowlist, scan everything
			init.Mapa["registries"] = ScanSomething(registryPath)
		} else {
			// allowlist, load only specified files
			regs := NewEmpty()
			for k, _ := range tctx.AssetsAllowed {
				if len(k) == 0 {
					continue // skip placeholders
				}
				regs.Mapa[k] = MustNewFromYAMLFile(path.Join(registryPath, k+".yaml")).Mapa
			}
			init.Mapa["registries"] = regs.Mapa
		}
	}

	if singletonPath != "" {
		if len(tctx.SingletonsAllowed) == 0 {
			init.Mapa["singletons"] = ScanSomething(singletonPath)
		} else {
			sings := NewEmpty()
			for k, _ := range tctx.SingletonsAllowed {
				if len(k) == 0 {
					continue // skip placeholders
				}
				sings.Mapa[k] = MustNewFromYAMLFile(path.Join(singletonPath, k+".yaml")).Mapa
			}
			init.Mapa["singletons"] = sings.Mapa
		}
	}

	return init
}

// RegisterAllActors calls identityAddMe for each actor present
func (tctx *TestContext) RegisterAllActors() {
	for actorName, _ := range tctx.actors {
		tctx.SetActor(actorName)
		tctx.Ok("identityAddMe", NewEmpty().Bytes())
	}
	tctx.SetActor("superUser")
}

// ScanSomething scans a dir for yaml files and returns a map with keys being the file names without .yaml and values being the contents of the files as a map
func ScanSomething(dir string) Rmap {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	regs := NewEmpty()

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			// skip dirs and not .yaml files
			continue
		}

		assetData, err := NewFromYAMLFile(path.Join(dir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		assetName := strings.TrimSuffix(file.Name(), ".yaml")
		regs.Mapa[assetName] = assetData.Mapa
	}

	return regs
}

func MustGetID(r Rmap) string {
	val, err := AssetGetID(r)
	if err != nil {
		panic(err)
	}

	return val
}

func MustGetVersion(r Rmap) int {
	val, err := AssetGetVersion(r)
	if err != nil {
		panic(err)
	}
	return val
}

// content is func for getting path to mock certs
func content(fixtureFile string) ([]byte, error) {
	_, curFile, _, ok := runtime.Caller(3) // skip needs to be tuned to actual stack depth where it was called
	if !ok {
		return nil, errors.New(`can't load file, error accessing runtime caller'`)
	}
	return ioutil.ReadFile(path.Dir(curFile) + "/" + fixtureFile)
}
