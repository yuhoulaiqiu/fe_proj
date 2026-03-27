.PHONY: run test build fmt

run:
	go run ./cmd/community-help-hub-server

test:
	go test ./...

build:
	go build -o bin/community-help-hub-server ./cmd/community-help-hub-server

fmt:
	gofmt -w $$(find . -name '*.go' -print)
