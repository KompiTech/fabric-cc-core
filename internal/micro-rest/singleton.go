package micro_rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func singletonUpsert(r *http.Request, urlPart string) ([]string, error) {
	elems := strings.Split(urlPart, "/")
	if len(elems) != 1 {
		return nil, fmt.Errorf("Invalid request")
	}
	name := elems[0]
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return []string{"singletonUpsert", name, string(bodyBytes)}, nil
}

func singletonGet(r *http.Request, urlPart string) ([]string, error) {
	elems := strings.Split(urlPart, "/")
	version := 0
	if len(elems) != 1 {
		return nil, fmt.Errorf("Invalid request")
	}
	name := elems[0]
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

	return []string{"singletonGet", name, fmt.Sprintf("%d", version)}, nil
}

func SingletonHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}

	var args []string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/singletons/"):]
	switch method := r.Method; method {
	case "POST":
		//POST /singleton/<name>
		args, err = singletonUpsert(r, urlPart)
		invoke = true
	case "GET":
		//GET /singleton/<name>
		args, err = singletonGet(r, urlPart)
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
