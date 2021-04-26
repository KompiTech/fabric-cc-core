package micro_rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func registryGet(r *http.Request, urlPart string) ([]string, error) {
	//regName, versionS := parseURL(urlPart)
	version := 0
	elems := strings.Split(urlPart, "/")
	if len(elems) != 1 {
		return nil, fmt.Errorf("invalid request")
	}
	regName := elems[0]
	pVersion, pVersionExists := r.Form["version"]
	if !pVersionExists {
		version = -1
	} else {
		parsed, err := strconv.Atoi(pVersion[0])
		if err != nil {
			return nil, err
		}
		version = parsed
	}
	return []string{"registryGet", regName, fmt.Sprintf("%d", version)}, nil
}

func registryUpsert(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	elems := strings.Split(urlPart, "/")
	if len(elems) != 1 {
		return nil, fmt.Errorf("invalid request")
	}
	assetName := elems[0]
	return []string{"registryUpsert", assetName, string(bodyBytes)}, nil
}

func registryList(r *http.Request, urlPart string) ([]string, error) {
	if len(urlPart) > 0 {
		return nil, fmt.Errorf("invalid request")
	}
	return []string{"registryList"}, nil
}

func RegistryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}

	var args []string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/registries/"):]
	switch method := r.Method; method {
	case "GET":
		//GET /registry/<name> or /registry
		if urlPart != "" {
			args, err = registryGet(r, urlPart)
		} else {
			args, err = registryList(r, urlPart)
		}
		invoke = false
	case "POST":
		//POST /registry/<name>
		args, err = registryUpsert(r, urlPart)
		invoke = true
	}
	if err != nil {
		if _, err := fmt.Fprint(w, err.Error()); err != nil {
			log.Print(err)
		}
		log.Print(err)
	} else {
		handleBackend(args, invoke, r, w)
	}
}
