export GOARCH := amd64
export GOOS := linux
export CGO_ENABLED := 0

.PHONI: lint
lint:
	golangci-lint run --timeout=3m

.PHONI: build
build:
	go build -buildvcs=false -a -o server ./cmd/server
