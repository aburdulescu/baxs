dev: clean build vet lint test

ci: clean build vet test

build:
	CGO_ENABLED=0 go build

vet:
	go vet

test:
	go test -cover -coverprofile cov.prof ./...

clean:
	go clean

lint:
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

coverage: test
	go tool cover -html cov.prof -o cov.html

release: dev clean
	CGO_ENABLED=0 go build -ldflags "-s -w"
	strip -s baxs
