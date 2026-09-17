.PHONY: run test build fmt vet

run:
	go run ./cmd/server -addr :9092

test:
	go test ./...

build:
	go build -o server ./cmd/server

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

vet:
	go vet ./...
