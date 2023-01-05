dev: clean build vet test

build:
	go build

vet:
	go vet

test:
	go test -cover ./...

clean:
	go clean

release: dev clean
	CGO_ENABLED=0 go build -ldflags "-s -w"
	strip -s baxs
