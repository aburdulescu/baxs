all: build vet test

build:
	go build
	cd conf && go build

vet:
	go vet
	cd conf && go vet

test:
	cd conf && go test -cover

fuzz:
	cd conf && go test -run=xxx -fuzz=. -fuzztime=20s
