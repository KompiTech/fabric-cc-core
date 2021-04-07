package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func historyGet(r *http.Request, urlPart string) ([]string, error) {
	uuid := ""
	assetName := ""
	elems := strings.Split(urlPart, "/")
	if len(elems) != 2 {
		return nil, fmt.Errorf("invalid request")
	}
	assetName = elems[0]
	uuid = elems[1]
	return []string{"assetHistory", assetName, uuid}, nil
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}

	var args []string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/histories/"):] //get URL path without /asset/ -> name of asset being used
	switch method := r.Method; method {
	case "GET":
		//GET /histories/<name>/<uuid>
		args, err = historyGet(r, urlPart)
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
