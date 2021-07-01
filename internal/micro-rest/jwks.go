package micro_rest

import (
	"io/ioutil"
	"net/http"
)

type JWKSHandler struct {
	pubkey []byte
}

func NewJWKSHandler(pubKeyPath string) (JWKSHandler, error) {
	h := JWKSHandler{}

	pubKeyData, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return h, err
	}

	h.pubkey = pubKeyData

	return h, nil
}

func (j JWKSHandler) WellKnownHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "GET":
		w.WriteHeader(200)
		w.Write(j.pubkey)
	}
}
