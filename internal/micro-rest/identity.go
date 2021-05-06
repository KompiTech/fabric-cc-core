package micro_rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func identityAddMe(r *http.Request) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return []string{"identityAddMe", string(bodyBytes)}, nil
}

func identityMe(r *http.Request) ([]string, error) {
	var resolve string
	_, resolveExists := r.Form["resolve"]
	if resolveExists {
		resolve = "true"
	} else {
		resolve = "false"
	}
	return []string{"identityMe", resolve}, nil
}

func identityGet(r *http.Request, urlPart string) ([]string, error) {
	var resolve string
	_, resolveExists := r.Form["resolve"]
	if resolveExists {
		resolve = "true"
	} else {
		resolve = "false"
	}
	return []string{"identityGet", urlPart, resolve}, nil
}

func identityUpdate(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return []string{"identityUpdate", urlPart, string(bodyBytes)}, nil
}

func identityQuery(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var resolve string
	_, resolveExists := r.Form["resolve"]
	if resolveExists {
		resolve = "true"
	} else {
		resolve = "false"
	}
	return []string{"identityQuery", string(bodyBytes), resolve}, nil
}

func IdentityHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}
	var args []string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/identities/"):]
	elems := strings.Split(urlPart, "/")
	log.Printf("Method: %s", r.Method)
	switch method := r.Method; method {
	case "POST":
		//POST /identities/me
		if len(elems) != 0 {
			err = fmt.Errorf("invalid request")
		}
		if elems[0] == "me" {
			args, err = identityAddMe(r)
			invoke = true
		} else {
			err = fmt.Errorf("invalid request")
		}
	case "GET":
		// GET /identities/me or /identities/<fingerprint>
		if urlPart == "me" {
			args, err = identityMe(r)
		} else {
			args, err = identityGet(r, urlPart)
		}
		invoke = false
	case "PATCH":
		//PATCH /identities/<fingerprint>
		args, err = identityUpdate(r, urlPart)
		invoke = true
	case "OPTIONS":
		//OPTIONS /identities/
		args, err = identityQuery(r, urlPart)
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
