.PHONY: setup test lint run demo

LEFTHOOK := github.com/evilmartians/lefthook/v2@v2.2.2
ARCHLINT := github.com/fe3dback/go-arch-lint@v1.19.1

# Not a go.mod tool: lefthook needs a newer Go than this module declares.
setup:
	go install $(LEFTHOOK)
	$(shell go env GOPATH)/bin/lefthook install

# Runs lgtm with what .env names: make run ARGS='"how does x work?"' asks a
# question, ARGS=intro runs a command. The file is not in git, and nothing
# but this target reads it.
ARGS ?=
run:
	set -a; [ ! -f .env ] || . ./.env; set +a; go run ./cmd/lgtm $(ARGS)

# Records docs/demo.gif, the picture at the top of the README, with vhs:
# a question asked of a clone of go-chi/chi. It asks a real model, so lgtm
# has to be configured.
demo:
	go install ./cmd/lgtm
	[ -d /tmp/lgtm-demo/chi ] || git clone -q --depth 1 https://github.com/go-chi/chi /tmp/lgtm-demo/chi
	vhs docs/demo.tape

test:
	go test ./...
	go test -tags gallery ./...

lint:
	go vet ./...
	go vet -tags gallery ./...
	go run $(ARCHLINT) check
