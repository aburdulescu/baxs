dev: clean build vet test

build:
	go build

vet:
	go vet
	golangci-lint run

test:
	go test -cover -coverprofile cov.prof ./...

clean:
	go clean

coverage: test
	go tool cover -html cov.prof -o cov.html

release: dev clean
	CGO_ENABLED=0 go build -ldflags "-s -w"
	strip -s baxs
