all: build vet test

build:
	go build
	cd conf && go build

vet:
	go vet
	cd conf && go vet

test:
	cd conf && go test -cover

bench:
	cd conf && go test -run=xxx -bench=. -benchmem

fuzz:
	cd conf && go test -run=xxx -fuzz=. -fuzztime=20s
