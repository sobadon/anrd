.PHONY: test
test:
	go test -v ./...

.PHONY: gen-mock
gen-mock:
	mkdir -p ./testdata/mock
	rm -r ./testdata/mock
	go generate ./...

.PHONY: build
build:
	go build -trimpath ./cmd/anrd
