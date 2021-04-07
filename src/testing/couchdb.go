package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/KompiTech/rmap"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

const (
	CouchDBContainerName         = "ccmock_couchdb_tmpfs" // docker container name for CouchDB mock
	CouchDBContainerImage        = "couchdb_tmpfs:latest" // docker container image name for CouchDB container
	CouchDBDataDir               = "/opt/couchdb/data"    // path to data directory for CouchDB inside container
	CouchDBInternalPort          = "5984"                 // port inside docker container which is exposed to host. Should be CouchDB default: 5984
	CouchDBExternalPort          = "8888"                 // port number which is exposed on host machine to access CouchDB API
	CouchDBAddress               = "http://localhost:" + CouchDBExternalPort
	MockChannelPrefix            = "mockchannel"
	MockChaincodeName            = "itsm"
	CouchDBMetaInfDir            = "META-INF/statedb/couchdb"         // location of state indexes
	CouchDBPrivateDataMetaInfDir = CouchDBMetaInfDir + "/collections" // location of private data indexes
	username                     = "admin"
	password                     = "password"
)

type pdKey struct {
	name string
	key  string
}

// CouchDBMock is mock for rich query functionality
type CouchDBMock struct {
	addr                string              // http address where couchdb is running
	dbName              string              // name of couchDB that is used for STATE
	dbURL               string              // addr + "/" + dbName
	sideDBs             rmap.Rmap           // map of PRIVATE_DATA DBs that are initialized
	indexesDone         bool                // flag, if indexes are initialized. We postpone creation of indexes until 1st read
	sideDBIndexesDone   rmap.Rmap           // ditto for each SideDB name
	httpClient          *http.Client        // http client for all requests
	putWg               *sync.WaitGroup     // synchronization for outgoing PUT requests
	existingStateAssets map[string]struct{} // keys for assets that are existing in state
	existingPDAssets    map[pdKey]struct{}  // keys for assets that are existing in private data
	authCookie          *http.Cookie
}

func NewCouchDBMock() *CouchDBMock {
	dbName := MockChannelPrefix + "-" + randStringBytes(10) + "_" + MockChaincodeName

	cdb := &CouchDBMock{
		addr:                CouchDBAddress,
		dbName:              dbName,
		dbURL:               CouchDBAddress + "/" + dbName,
		sideDBs:             rmap.NewEmpty(),
		sideDBIndexesDone:   rmap.NewEmpty(),
		httpClient:          &http.Client{},
		putWg:               &sync.WaitGroup{},
		existingStateAssets: map[string]struct{}{},
		existingPDAssets:    map[pdKey]struct{}{},
	}
	cdb.auth()
	cdb.createDB(dbName)
	return cdb
}

func (cdb *CouchDBMock) auth() {
	data := rmap.NewFromMap(map[string]interface{}{
		"name":     username,
		"password": password,
	})

	url := CouchDBAddress + "/_session"

	req, err := http.NewRequest("POST", url, bytes.NewReader(data.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("content-Type", "application/json")

	resp, err := cdb.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("POST %s %d, response body: %s", url, resp.StatusCode, body)
	}

	cdb.authCookie = resp.Cookies()[0]
}

func (cdb *CouchDBMock) Wait() {
	cdb.putWg.Wait()
}

func (cdb *CouchDBMock) PutState(key string, value []byte) {
	if _, exists := cdb.existingStateAssets[key]; exists {
		// this is an UPDATE of existing key, wait for PUTs to finish
		cdb.Wait()
	} else {
		// CREATE op
		cdb.existingStateAssets[key] = struct{}{}
	}

	// put data to couchDB
	//async
	cdb.putWg.Add(1)
	go cdb.putStateInWg(cdb.dbName, key, value)
}

func (cdb *CouchDBMock) PutPrivateData(name string, key string, value []byte) {
	name = cdb.getSideDBName(name)
	cdb.ensureSideDB(name)

	pdKey := pdKey{name, key}
	if _, exists := cdb.existingPDAssets[pdKey]; exists {
		cdb.Wait()
	} else {
		cdb.existingPDAssets[pdKey] = struct{}{}
	}

	// put data to couchDB
	cdb.putWg.Add(1)
	go cdb.putStateInWg(name, key, value)
}

func (cdb *CouchDBMock) getSideDBName(collection string) string {
	return strings.ToLower(cdb.dbName + "$$pcollection$" + collection)
}

func (cdb *CouchDBMock) createIndexes(dbName, metaInf string) {
	// createDB all couchDB indexes
	indexPath := filepath.Join(metaInf, "indexes")

	indexFiles, err := ioutil.ReadDir(indexPath)
	if err != nil {
		//probably cannot read the dir, skip
		return
	}

	for _, filename := range indexFiles {
		// read all indexes on the disk and create them in CouchDB
		indexData, err := ioutil.ReadFile(filepath.Join(indexPath, filename.Name()))
		if err != nil {
			log.Fatal(err)
		}

		// send requests in parallel
		cdb.putWg.Add(1)
		go cdb.makeRequestInWg("POST", "/"+dbName+"/_index", bytes.NewReader(indexData), cdb.putWg)
	}

	cdb.Wait()
}

func (cdb *CouchDBMock) createDB(dbName string) {
	// create DB (we silently assume that it doesn't exist)
	path := "/" + dbName
	resp := cdb.makeRequest("PUT", path, nil)
	defer func() { _ = resp.Body.Close() }()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == 412 {
		// DB already exists, just remove it and create again
		deleteResp := cdb.makeRequest("DELETE", path, nil)

		defer func() { _ = deleteResp.Body.Close() }()

		deleteBody, err := ioutil.ReadAll(deleteResp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if deleteResp.StatusCode != 200 {
			log.Fatalf("DELETE %s %d, response body: %s", path, deleteResp.StatusCode, deleteBody)
		}

		createResp := cdb.makeRequest("PUT", path, nil)

		defer func() { _ = createResp.Body.Close() }()

		createBody, err := ioutil.ReadAll(createResp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if createResp.StatusCode != 201 {
			log.Fatalf("PUT %s %d, response body: %s", path, createResp.StatusCode, createBody)
		}
		log.Printf("Existing database: %s re-created", dbName)
	} else if resp.StatusCode != 201 {
		log.Fatalf("PUT %s %d, response body: %s", path, resp.StatusCode, body)
	}
}

// make HTTP request to some relative path to addr of this
func (cdb *CouchDBMock) makeRequest(method, path string, body io.Reader) *http.Response {
	url := strings.Replace(cdb.addr+path, "\x00", "", -1) // if key contains zero bytes from composite key, remove them

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}

	req.AddCookie(cdb.authCookie)

	if body != nil {
		req.Header.Add("content-Type", "application/json")
	}

	resp, err := cdb.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	return resp
}

func (cdb *CouchDBMock) makeRequestInWg(method, path string, body io.Reader, wg *sync.WaitGroup) {
	defer func() { wg.Done() }()
	resp := cdb.makeRequest(method, path, body)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("PUT %s %d, response body: %s", path, resp.StatusCode, body)
	}
}

func (cdb *CouchDBMock) ensureSideDB(dbName string) {
	if !cdb.sideDBs.Exists(dbName) {
		// first access to SideDB collection, createDB it, create indexes
		cdb.createDB(dbName)
		cdb.createIndexes(dbName, filepath.Join(CouchDBPrivateDataMetaInfDir, strings.ToUpper(dbName)))

		// mark as initialized
		cdb.sideDBs.Mapa[dbName] = struct{}{}
	}
}

func (cdb *CouchDBMock) GetQueryResult(dbName string, query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata) {
	var isSideDB bool

	if dbName == "" {
		dbName = cdb.dbName
		isSideDB = false
	} else {
		dbName = cdb.getSideDBName(dbName)
		isSideDB = true
	}

	// wait for any in-flight PUT requests to finish, so DB read is consistent
	cdb.Wait()

	if !cdb.indexesDone && !isSideDB {
		// first read from new mock (STATE DB), load indexes
		cdb.createIndexes(dbName, CouchDBMetaInfDir)
		cdb.indexesDone = true
	}

	if isSideDB {
		// first read from some sideDB, check if its index is initialized
		if !cdb.sideDBIndexesDone.Exists(dbName) {
			cdb.createIndexes(dbName, CouchDBMetaInfDir)
			cdb.sideDBIndexesDone.Mapa[dbName] = struct{}{}
		}
	}

	queryR, err := rmap.NewFromBytes([]byte(query))
	if err != nil {
		log.Fatal(err)
	}

	if pageSize > 0 {
		queryR.Mapa["limit"] = pageSize
	}

	if bookmark != "" {
		queryR.Mapa["bookmark"] = bookmark
	}

	path := "/" + dbName + "/_find"
	resp := cdb.makeRequest("POST", path, bytes.NewReader(queryR.Bytes()))
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("POST %s %d response body: %s, request: %s", path, resp.StatusCode, body, queryR.String())
	}

	result, err := rmap.NewFromBytes(body)
	if err != nil {
		log.Fatal(err)
	}

	resBookmark, err := result.GetString("bookmark")
	if err != nil {
		log.Fatal(err)
	}

	docs, err := result.GetIterable("docs")
	if err != nil {
		log.Fatal(err)
	}

	results := make([][]byte, 0, len(docs))
	for _, resultI := range docs {
		doc, err := rmap.NewFromInterface(resultI)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, doc.Bytes())
	}

	resultMock := &mockIterator{
		Results: results,
	}

	return resultMock, &peer.QueryResponseMetadata{FetchedRecordsCount: int32(len(docs)), Bookmark: resBookmark}
}

func (cdb *CouchDBMock) putStateInWg(dbName string, key string, value []byte) {
	defer func() { cdb.putWg.Done() }()

	path := cdb.getPathString(dbName, key)

	//check, if document exists
	resp := cdb.makeRequest("HEAD", path, nil)
	defer func() { _ = resp.Body.Close() }()

	rev := ""
	if resp.StatusCode == 200 {
		//add the revision for CouchDB to not complain about conflicts
		//remove quotes from value
		rev = getRev(resp.Header)
	} else if resp.StatusCode == 404 {
		//do nothing, document doesnt exist, we will be inserting it
	} else {
		//everything else is an error
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("HEAD %s %d, response body: %s", path, resp.StatusCode, body)
	}

	cdb.put(dbName, key, rev, value)
}

func randStringBytes(n int) string {
	// use fully seeded rand because we want unique names
	// not using unique name incurs couchdb penalty to overwrite
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	letterBytes := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[seededRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (cdb *CouchDBMock) getPathString(dbName, key string) string {
	return "/" + dbName + "/" + key
}

func (cdb *CouchDBMock) get(dbName, key string) []byte {
	path := cdb.getPathString(dbName, key)

	resp := cdb.makeRequest("GET", path, nil)
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("GET %s %d, response body: %s", path, resp.StatusCode, body)
	}

	dataz := map[string]interface{}{}
	if err := json.Unmarshal(body, &dataz); err != nil {
		log.Fatal(err)
	}

	// CouchDB does not like these keys in input
	delete(dataz, "_id")
	delete(dataz, "_rev")

	data, err := json.Marshal(dataz)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

// if rev is "" then it attempts to create a new document
// if rev is something then it must match previous revision of updated document
// returns rev of created item
func (cdb *CouchDBMock) put(dbName, key, rev string, value []byte) string {
	path := cdb.getPathString(dbName, key)

	if rev != "" {
		// couchDB expects _rev as part of JSON
		data, err := rmap.NewFromBytes(value)
		if err != nil {
			log.Fatal(err)
		}

		data.Mapa["_rev"] = rev

		value = data.Bytes()
	}

	resp := cdb.makeRequest("PUT", path, bytes.NewReader(value))
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("PUT %s %d, response body: %s", path, resp.StatusCode, body)
	}

	return getRev(resp.Header)
}

func (cdb *CouchDBMock) delete(dbName, key string) {
	defer func() { cdb.putWg.Done() }()
	path := cdb.getPathString(dbName, key)

	//check, if document exists
	resp := cdb.makeRequest("HEAD", path, nil)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		//document does not exist, error
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("HEAD %s %d, response body: %s", path, resp.StatusCode, body)
	}

	//add the revision for CouchDB to not complain about conflicts
	//remove quotes from value
	resp = cdb.makeRequest("DELETE", path+"?rev="+getRev(resp.Header), nil)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("DELETE %s %d, response body: %s", path, resp.StatusCode, body)
	}
}

func (cdb *CouchDBMock) DelPrivateData(dbName string, key string) {
	dbName = cdb.getSideDBName(dbName)

	cdb.putWg.Add(1)
	go cdb.delete(dbName, key)
}

// getRev gets revision from headers
func getRev(hdrs http.Header) string {
	return strings.Replace(hdrs["Etag"][0], "\"", "", -1)
}
