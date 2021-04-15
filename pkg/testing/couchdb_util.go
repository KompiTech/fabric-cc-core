package testing

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// run execs command with args, waits for completion and returns *cmd.Exec along with stdout, stderr
// any non exec.ExitError errors results in panic
func run(name string, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	log.Printf("Running: %s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Fatal(err)
		}
	}

	return cmd, &stdout, &stderr
}

func runCouch() {
	runArgs := []string{"run", "-d", "--rm",
		"-p", CouchDBExternalPort + ":" + CouchDBInternalPort,
		"-e", "COUCHDB_USER=" + username, "-e", "COUCHDB_PASSWORD=" + password,
		"--name", CouchDBContainerName,
		"--mount", "type=tmpfs,destination=" + CouchDBDataDir,
		"--ulimit", "nofile=999999:999999",
		CouchDBContainerImage,
	}

	errMsg := fmt.Sprintf("Unable to build %s image, try building it yourself \ncd `go list -f '{{.Dir}}' github.com/KompiTech/fabric-cc-core/v2/pkg`/testing_docker && docker build . -t couchdb_tmpfs && cd -", CouchDBContainerImage)

	cmd, _, serr := run("docker", runArgs...)
	if !cmd.ProcessState.Success() {
		// docker create command failed
		log.Print("docker run stderr: " + serr.String())
		log.Print("attempting to build image")

		args := []string{"list", "-f", "{{.Dir}}", "github.com/KompiTech/fabric-cc-core/v2/pkg/"}
		var sout *bytes.Buffer
		cmd, sout, serr = run("go", args...)
		if !cmd.ProcessState.Success() {
			log.Printf("unable to run go %s, stderr: %s", strings.Join(args, " "), serr.String())
			log.Fatal(errMsg)
		}

		dockerFilePath := strings.TrimSuffix(sout.String(), "\n")
		dockerFilePath = path.Join(dockerFilePath, "testing_docker")

		oldPwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		if err := os.Chdir(dockerFilePath); err != nil {
			panic(err)
		}

		args = []string{"build", ".", "-t", CouchDBContainerImage}
		cmd, _, serr = run("docker", args...)
		if !cmd.ProcessState.Success() {
			log.Printf("unable to run docker %s, stderr: %s", strings.Join(args, " "), serr.String())
			log.Fatal(errMsg)
		}

		if err := os.Chdir(oldPwd); err != nil {
			panic(err)
		}

		cmd, _, serr = run("docker", runArgs...)
		if !cmd.ProcessState.Success() {
			log.Printf("unable to run docker %s, stderr: %s", strings.Join(args, " "), serr.String())
			log.Fatal(errMsg)
		}
	}
}

func isCouchRunning() bool {
	cmd, sout, serr := run("docker", "ps", "-q", "--filter", "name="+CouchDBContainerName)
	if !cmd.ProcessState.Success() {
		log.Fatalf("docker ps failed, stderr: " + serr.String())
	}

	// if there is some output, container is running
	if sout.Len() > 0 {
		return true
	}

	return false
}

func removeCouch() {
	_, _, _ = run("docker", "rm", "-f", CouchDBContainerName)
}

func waitForCouch() {
	log.Print("Waiting for CouchDB to become ready...")
	maxIters := 600 // default 600 * 100ms = 60 seconds
	iter := 0
	stepMs := 100
	reportEveryIter := 10
	step := time.Duration(stepMs) * time.Millisecond
	client := http.Client{
		Timeout: step,
	}

	req, err := http.NewRequest("GET", CouchDBAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

	// wait for couchdb to return 200
	for {
		if resp, err := client.Do(req); err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(step)
		iter++
		if iter == maxIters {
			log.Fatal("Waited for CouchDB for too long. Something is wrong. Check docker status and try again.")
		} else if (iter > 0) && (iter%reportEveryIter == 0) {
			log.Printf("CouchDB still not ready in %d ms, waiting...", int64(iter)*int64(stepMs))
		}
	}
	log.Printf("CouchDB became ready in %d ms", int64(iter)*int64(stepMs))
}

// initSysDBs initializes system DBs, so couch wont spam log file
func initSysDBs() {
	client := http.Client{}

	req, err := http.NewRequest("PUT", CouchDBAddress+"/_users", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)
	_, _ = client.Do(req)

	req, err = http.NewRequest("PUT", CouchDBAddress+"/_replicator", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)
	_, _ = client.Do(req)

	req, err = http.NewRequest("POST", CouchDBAddress+"/_global_changes", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)
	_, _ = client.Do(req)
}

// InitializeCouchDBContainer prepares CouchDB running in docker for use by this mock. It should be called only once before test suite, because it is quite time intensive
func InitializeCouchDBContainer() {
	if !isCouchRunning() {
		log.Print("CouchDB container is not running, removing it...")
		removeCouch()
		log.Print("Creating new CouchDB container")
		runCouch()
	}
	waitForCouch()
	initSysDBs()
}
