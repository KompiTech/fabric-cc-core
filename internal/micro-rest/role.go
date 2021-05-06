package micro_rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func roleCreate(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	uuid := ""
	elems := strings.Split(urlPart, "/")
	if len(elems) > 1 {
		return nil, fmt.Errorf("invalid request")
	} else if len(elems) == 1 {
		uuid = elems[0]
	}
	return []string{"roleCreate", string(bodyBytes), uuid}, nil
}

func roleGet(r *http.Request, urlPart string) ([]string, error) {
	elems := strings.Split(urlPart, "/")
	if len(elems) != 1 {
		return nil, fmt.Errorf("Invalid request")
	}
	var resolve string
	_, resolveExists := r.Form["resolve"]
	if resolveExists {
		resolve = "true"
	} else {
		resolve = "false"
	}
	return []string{"roleGet", elems[0], resolve}, nil
}

func roleUpdate(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	uuid := ""
	elems := strings.Split(urlPart, "/")
	if len(elems) != 1 {
		return nil, fmt.Errorf("Invalid request")
	}
	return []string{"roleUpdate", uuid, string(bodyBytes)}, nil
}

func roleQuery(r *http.Request, urlPart string) ([]string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	elems := strings.Split(urlPart, "/")
	if len(elems) != 0 {
		return nil, fmt.Errorf("Invalid request")
	}
	return []string{"roleQuery", string(bodyBytes)}, nil
}

func RoleHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}

	var args []string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/roles/"):]
	switch method := r.Method; method {
	case "POST":
		//POST /role or /role/<uuid>
		args, err = roleCreate(r, urlPart)
		invoke = true
	case "GET":
		//GET /role/<uuid>
		args, err = roleGet(r, urlPart)
		invoke = false
	case "PATCH":
		//PATCH /role/<uuid>
		args, err = roleUpdate(r, urlPart)
		invoke = true
	case "OPTIONS":
		//OPTIONS /role
		args, err = roleQuery(r, urlPart)
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
