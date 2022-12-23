all: build vet test

build:
	go build

vet:
	go vet

test:
	go test -v -cover ./...

release: vet
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.versionDate=`date -u -Iseconds` -X main.versionCommit=`git rev-parse HEAD`"
