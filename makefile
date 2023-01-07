dev: clean build vet lint test

ci: clean build vet test

build:
	CGO_ENABLED=0 go build

vet:
	which fieldalignment || go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	go vet
	go vet -vettool $(shell which fieldalignment) ./...

test:
	go test -cover -coverprofile cov.prof ./...

clean:
	go clean

lint:
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

coverage: test
	go tool cover -html cov.prof -o cov.html

fuzz:
	go test -v -run=xxx -fuzz=. -fuzztime=10s ./internal/baxsfile

release: dev clean
	CGO_ENABLED=0 go build -ldflags "-s -w"
	strip -s baxs
