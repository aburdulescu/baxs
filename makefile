dev: clean build vet lint test

ci: clean build vet test

generate:
	go generate ./...

build:
	CGO_ENABLED=0 go build

vet:
	go vet

test:
	go test -cover -coverprofile cov.prof ./...

clean:
	go clean

lint:
	@which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

fieldalignment:
	@which fieldalignment || go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	fieldalignment -test=false ./...

coverage: test
	go tool cover -html cov.prof -o cov.html

fuzz:
	go test -v -run=xxx -fuzz=. -fuzztime=10s ./internal/baxsfile

release: dev clean
	CGO_ENABLED=0 go build -ldflags "-s -w"
	strip -s baxs

misc:
	go install golang.org/x/tools/cmd/stringer@latest
