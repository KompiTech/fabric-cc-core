export GONOSUMDB = "github.com/KompiTech/*,github.com/cucumber/godog"
export GOPRIVATE = "github.com/KompiTech/*"

test ::
	go test -v ./pkg/engine/...
	go test -v ./pkg/

coverage :: test
	go tool cover -html=cover.out
