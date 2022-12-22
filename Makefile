all: build vet test

build:
	go build

vet:
	go vet

test:
	go test -v -cover ./...

release: vet
	go build -ldflags "-X main.versionDate=`date -u -Iseconds` -X main.versionCommit=`git rev-parse HEAD`"
