.PHONY: help test lint build check

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-8s %s\n", $$1, $$2}'

test: ## Run the Go tests (goldens, auth, admin, static, activate)
	cd server && go test -race ./...

lint: ## gofmt, go vet and shellcheck
	@test -z "$$(gofmt -l server)" || { gofmt -l server; exit 1; }
	cd server && go vet ./...
	shellcheck -S warning -x cli/dumps

build: ## Build dumpsd for this machine into bin/
	mkdir -p bin && cd server && CGO_ENABLED=0 go build -trimpath -o ../bin/dumpsd .

check: lint test ## Everything CI runs
