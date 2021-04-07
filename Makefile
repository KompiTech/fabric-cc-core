export GONOSUMDB = "github.com/KompiTech/*,github.com/cucumber/godog"
export GOPRIVATE = "github.com/KompiTech/*"

test ::
	go test -v ./src/engine/...
	go test -v ./src/

coverage :: test
	go tool cover -html=cover.out
