.PHONY: build test lint clean crossval dogfood

GOPATH := $(shell go env GOPATH)
BINARY := nocrap

build:
	go build -o $(BINARY) .

test:
	go test ./... -v -count=1

test-race:
	go test ./... -v -race -count=1

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

crossval:
	go test ./crossval/ -v -count=1

dogfood: build
	go test -coverprofile=cover.out ./...
	./$(BINARY) --lang go --threshold 9 ./
