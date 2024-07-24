export GOARCH := amd64
export GOOS := linux
export CGO_ENABLED := 0

.PHONI: lint
lint:
	golangci-lint run --timeout=3m

.PHONI: build
build:
	go build -buildvcs=false -a -o server ./cmd/server

.PHONI: clean-test
clean-test:
	go clean -testcache

.PHONI: test-service
test-service:
	go test -count=1 -failfast -v ./internal/service

.PHONI: test
test: clean-test test-service