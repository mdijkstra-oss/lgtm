.PHONY: setup test lint

LEFTHOOK := github.com/evilmartians/lefthook/v2@v2.2.2
ARCHLINT := github.com/fe3dback/go-arch-lint@v1.19.1

# Not a go.mod tool: lefthook needs a newer Go than this module declares.
setup:
	go install $(LEFTHOOK)
	$(shell go env GOPATH)/bin/lefthook install

test:
	go test ./...
	go test -tags gallery ./...

lint:
	go vet ./...
	go vet -tags gallery ./...
	go run $(ARCHLINT) check
