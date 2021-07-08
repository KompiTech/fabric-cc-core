test ::
	go test -v ./pkg/...

coverage :: test
	go tool cover -html=cover.out
