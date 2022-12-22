all: build vet test

build:
	go build

vet:
	go vet

test:
	go test -v -cover ./...
