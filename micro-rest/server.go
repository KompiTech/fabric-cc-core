package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/KompiTech/rmap"
)

const (
	versionEnv           = "LOCAL_REST_VERSION_FULL"
	userHeaderName       = "X-Fabric-User"
	fabricSamplesEnvName = "FABRIC_SAMPLES"
)

var port int

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mandatoryEnv(key string) string {
	value, isDefined := os.LookupEnv(key)
	if !isDefined {
		log.Fatalf("Mandatory ENV variable: %s is not defined", key)
	}

	return value
}

func envNotDefined(key string) bool {
	_, isDefined := os.LookupEnv(key)
	return !isDefined
}

func callBackend(args []string, invoke bool) (string, int) {
	invokePayload := rmap.NewFromMap(map[string]interface{}{
		"Args": args,
	})

	peerArgs := []string{
		"chaincode",
		"XXXINVOKEORQUERYXXX", // filled in next part
		"-o",
		"localhost:7050",
		"--ordererTLSHostnameOverride",
		"orderer.example.com",
		"--tls",
		"--cafile",
		mandatoryEnv("LOCAL_CAFILE"),
		"-C",
		mandatoryEnv("LOCAL_CHANNEL_NAME"),
		"-n",
		mandatoryEnv("LOCAL_CC_NAME"),
		"--peerAddresses",
		"localhost:7051",
		"--tlsRootCertFiles",
		mandatoryEnv("LOCAL_TLSCERT_ORG1"),
		"--peerAddresses",
		"localhost:9051",
		"--tlsRootCertFiles",
		mandatoryEnv("LOCAL_TLSCERT_ORG2"),
		"-c",
		// here append call in format: {"Args": ["Method", "Arg1", ....]}
		invokePayload.String(),
	}

	if invoke {
		peerArgs[1] = "invoke"
	} else {
		peerArgs[1] = "query"
		// query wants only one peer as destination, delete second peer from args
		peerArgs = append(peerArgs[:17], peerArgs[21:]...)
	}

	log.Printf("peerArgs: %v", peerArgs)

	cmd := exec.Command("peer", peerArgs...)
	outBytes, _ := cmd.CombinedOutput()
	out := string(outBytes)
	fmt.Print(out)

	if cmd.ProcessState.ExitCode() != 0 {
		return out, cmd.ProcessState.ExitCode()
	}

	// queries are returned OK, but invoke has some extra debug stuff, remove
	if invoke {
		// strip peer debug from payload and unescape
		start := strings.Index(out, "{")                     // find JSON opening brace
		end := strings.LastIndex(out, "}") + 1               // find JSON closing brace
		out = strings.Replace(out[start:end], `\"`, `"`, -1) // unescape \"
	}

	return out, cmd.ProcessState.ExitCode()
}

func jsonizeErrorString(msg string) string {
	return rmap.NewFromMap(map[string]interface{}{"error": msg}).String()
}

// set CORE_PEER_MSPCONFIGPATH env variable with Fabric user to be used by peer binary
func prepareEnv(hdrs http.Header) error {
	var userName string
	userHdr, userHdrExists := hdrs[userHeaderName]
	if userHdrExists {
		userName = userHdr[0] // use specified header
	} else {
		userName = "Admin@org1.example.com" // fallback, if header is not sent
	}

	userFields := strings.Split(userName, "@")

	if len(userFields) != 2 {
		return errors.New("invalid " + userHeaderName + " header name, expecting user@org")
	}

	// prepare env values to be set. This is very dependent on test-network directory structure with generated crypto material
	orgFields := strings.Split(userFields[1], ".")
	orgFirstName := orgFields[0]
	localMspID := strings.ToUpper(string(orgFirstName[0])) + orgFirstName[1:] + "MSP"
	peerOrgPath := path.Join(mandatoryEnv(fabricSamplesEnvName), "test-network/organizations/peerOrganizations/", userFields[1])
	peerTLSPath := path.Join(peerOrgPath, "peers", "peer0."+userFields[1], "tls", "ca.crt") // always using peer0 of org
	mspConfigPath := path.Join(peerOrgPath, "users", userName, "msp")

	// set the ENV values, overwriting previous values. These will be used by 'peer' binary called in next step
	os.Setenv("CORE_PEER_LOCALMSPID", localMspID)
	os.Setenv("CORE_PEER_TLS_ROOTCERT_FILE", peerTLSPath)
	os.Setenv("CORE_PEER_MSPCONFIGPATH", mspConfigPath)

	return nil
}

func handleBackend(args []string, invoke bool, r *http.Request, w http.ResponseWriter) {
	log.Print(args)

	if err := prepareEnv(r.Header); err != nil {
		if _, err := fmt.Fprint(w, jsonizeErrorString(err.Error())); err != nil {
			log.Print(err)
		}
		return
	}

	beOutput, exitCode := callBackend(args, invoke)
	if exitCode == 0 {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
		// wrap error into JSON
		beOutput = jsonizeErrorString(beOutput)
	}
	log.Print(beOutput)

	if _, err := fmt.Fprint(w, beOutput); err != nil {
		log.Print(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: %s <port>", os.Args[0])
	} else {
		var err error
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Print(err)
		}
	}
	http.HandleFunc("/api/v1/identities/", identityHandler)
	http.HandleFunc("/api/v1/registries/", registryHandler)
	http.HandleFunc("/api/v1/assets/", assetHandler)
	http.HandleFunc("/api/v1/roles/", roleHandler)
	http.HandleFunc("/api/v1/functions-query/", functionHandler)
	http.HandleFunc("/api/v1/functions-invoke/", functionHandler)
	http.HandleFunc("/api/v1/singletons/", singletonHandler)
	http.HandleFunc("/api/v1/histories/", historyHandler)
	log.Printf("Listening at 0.0.0.0:%d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
