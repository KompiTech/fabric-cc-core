package micro_rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func assetCreate(r *http.Request, urlPart string) ([]string, error) {
	version := 0
	elems := strings.Split(urlPart, "/")
	if len(elems) == 0 || len(elems) > 2 {
		return nil, fmt.Errorf("invalid request")
	}
	uuid := ""
	assetName := ""
	if len(elems) == 1 {
		assetName = elems[0]
	} else {
		assetName = elems[0]
		uuid = elems[1]
	}
	pVersion, pVersionExists := r.Form["version"]
	if assetName == "identity" {
		version = 1
	} else if !pVersionExists || envNotDefined(versionEnv) {
		version = -1
	} else {
		parsed, err := strconv.Atoi(pVersion[0])
		if err != nil {
			return nil, err
		}
		version = parsed
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	ret := []string{"assetCreate", assetName, string(bodyBytes), fmt.Sprintf("%d", version), uuid}

	if _, pForceExists := r.Form["force"]; pForceExists {
		ret[0] = ret[0] + "Direct"
	}

	return ret, nil
}

func assetGet(r *http.Request, urlPart string) ([]string, error) {
	uuid := ""
	assetName := ""
	elems := strings.Split(urlPart, "/")
	if len(elems) != 2 {
		return nil, fmt.Errorf("invalid request")
	}
	assetName = elems[0]
	uuid = elems[1]
	var resolve string
	_, resolveExists := r.Form["resolve"]
	if resolveExists {
		resolve = "true"
	} else {
		resolve = "false"
	}
	ret := []string{"assetGet", assetName, uuid, resolve, "{}"}

	if _, pForceExists := r.Form["force"]; pForceExists {
		ret[0] = ret[0] + "Direct"
	}

	return ret, nil
}

func assetUpdate(r *http.Request, urlPart string) ([]string, error) {
	uuid := ""
	assetName := ""
	elems := strings.Split(urlPart, "/")
	if len(elems) != 2 {
		return nil, fmt.Errorf("invalid request")
	}
	assetName = elems[0]
	uuid = elems[1]
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	ret := []string{"assetUpdate", assetName, uuid, string(bodyBytes)}

	if _, pForceExists := r.Form["force"]; pForceExists {
		ret[0] = ret[0] + "Direct"
	}

	return ret, nil
}

func assetMigrate(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	urlPart = strings.TrimPrefix(urlPart, "migrate/")
	uuid := ""
	assetName := ""
	elems := strings.Split(urlPart, "/")
	if len(elems) != 2 {
		return nil, fmt.Errorf("invalid request")
	}
	assetName = elems[0]
	uuid = elems[1]
	version := 0
	pVersion, pVersionExists := r.Form["version"]
	if !pVersionExists {
		return nil, fmt.Errorf("invalid request: version is mandatory!")
	} else {
		parsed, err := strconv.Atoi(pVersion[0])
		if err != nil {
			return nil, err
		}
		version = parsed
	}
	return []string{"assetMigrate", assetName, uuid, string(bodyBytes), fmt.Sprintf("%d", version)}, nil
}

func assetQuery(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	elems := strings.Split(urlPart, "/")
	if len(elems) > 1 {
		return nil, fmt.Errorf("invalid request")
	}
	assetName := elems[0]
	var resolve string
	_, resolveExists := r.Form["resolve"]
	if resolveExists {
		resolve = "true"
	} else {
		resolve = "false"
	}

	ret := []string{"assetQuery", assetName, string(bodyBytes), resolve}

	if _, pForceExists := r.Form["force"]; pForceExists {
		ret[0] = ret[0] + "Direct"
	}

	return ret, nil
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}

	var args []string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/assets/"):] //get URL path without /asset/ -> name of asset being used
	switch method := r.Method; method {
	case "POST":
		//POST /assets/<name> or /assets/<name>/<uuid> or /asset/migrate/<name>/<uuid>
		elems := strings.Split(urlPart, "/")
		if elems[0] != "migrate" {
			args, err = assetCreate(r, urlPart)
			invoke = true
		} else {
			args, err = assetMigrate(r, urlPart)
			invoke = true
		}
	case "GET":
		//GET /assets/<name>/<uuid>
		args, err = assetGet(r, urlPart)
		invoke = false
	case "PATCH":
		//PATCH /assets/<name>/<uuid>
		args, err = assetUpdate(r, urlPart)
		invoke = true
	case "OPTIONS":
		//OPTIONS /assets/<name>/<uuid> or /asset/<name>
		elems := strings.Split(urlPart, "/")
		if len(elems) == 0 {
			err = fmt.Errorf("invalid request")
		}
		args, err = assetQuery(r, urlPart)
		invoke = false
	}
	if err != nil {
		if _, err := fmt.Fprint(w, err.Error()); err != nil {
			log.Print(err)
		}
		log.Print(err)
		return
	}
	handleBackend(args, invoke, r, w)
}
