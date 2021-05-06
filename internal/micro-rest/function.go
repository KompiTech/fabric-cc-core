package micro_rest

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func FunctionHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		return
	}

	var fname string
	var err error
	var invoke bool
	urlPart := r.URL.Path[len("/api/v1/functions-"):]

	if strings.HasPrefix(urlPart, "query/") {
		invoke = false
		fname = urlPart[len("query/"):]
	} else if strings.HasPrefix(urlPart, "invoke/") {
		invoke = true
		fname = urlPart[len("invoke/"):]
	} else {
		err = errors.New("invalid request")
	}

	if err != nil {
		if _, err := fmt.Fprint(w, err.Error()); err != nil {
			log.Print(err)
		}
		log.Print(err)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		if _, err := fmt.Fprint(w, err.Error()); err != nil {
			log.Print(err)
		}
		log.Print(err)
		return
	}

	var args []string
	if invoke {
		args = []string{"functionInvoke", fname, string(bodyBytes)}
	} else {
		args = []string{"functionQuery", fname, string(bodyBytes)}
	}

	handleBackend(args, invoke, r, w)
}
