package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/KompiTech/fabric-cc-core/v2/internal/micro-rest"
)

const jwkPubKeyEnv = "JWK_PUB_KEY"

func main() {
	var port int

	if len(os.Args) != 2 {
		log.Fatal("Usage: %s <port>", os.Args[0])
	} else {
		var err error
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	}

	jwkPubKeyPath, jwkDefined := os.LookupEnv(jwkPubKeyEnv)
	if jwkDefined {
		jwkh, err := micro_rest.NewJWKSHandler(jwkPubKeyPath)
		if err != nil {
			log.Fatal(err)
		}

		// when JWK pubkey is defined in env, provide well-known route also
		// useful for integration testing with other APIs and their tokens
		http.HandleFunc("/.well-known/jwks.json", jwkh.WellKnownHandler)
	}

	http.HandleFunc("/api/v1/identities/", micro_rest.IdentityHandler)
	http.HandleFunc("/api/v1/registries/", micro_rest.RegistryHandler)
	http.HandleFunc("/api/v1/assets/", micro_rest.AssetHandler)
	http.HandleFunc("/api/v1/roles/", micro_rest.RoleHandler)
	http.HandleFunc("/api/v1/functions-query/", micro_rest.FunctionHandler)
	http.HandleFunc("/api/v1/functions-invoke/", micro_rest.FunctionHandler)
	http.HandleFunc("/api/v1/singletons/", micro_rest.SingletonHandler)
	http.HandleFunc("/api/v1/histories/", micro_rest.HistoryHandler)
	log.Printf("Listening at 0.0.0.0:%d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
