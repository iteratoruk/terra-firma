# Terra Firma — task runner.
#
# This file IS the project's feedback loop. An agent (or human) should be able to
# discover every verification and run command here without inferring it. Keep the
# targets honest: if a command is how you check your work, it belongs here.
#
# Go is installed locally; CI pins the version (see .github/workflows/ci.yml).

# Local toolchain pin: never silently download a different Go than is installed.
export GOTOOLCHAIN := local

.DEFAULT_GOAL := help

.PHONY: help setup build test test-race cover lint fmt fmt-check imports imports-check vet run golden check clean

help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

setup: ## Install/verify dependencies (C2.1). No external deps yet beyond the stdlib.
	go mod download
	go mod verify

build: ## Compile everything; fails on any build error (C5.1 L1).
	go build ./...

test: ## Run the full test suite (C5.2). The primary correctness signal.
	go test ./...

test-race: ## Run tests under the race detector. Determinism must hold under -race.
	go test -race ./...

cover: ## Report unit-test coverage (the C5.2 metric the readiness model scores).
	go test -cover ./...
	@echo "For a detailed per-function breakdown: make cover-html"

cover-html: ## Write an HTML coverage report to coverage.html.
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "wrote coverage.html"

lint: ## Run golangci-lint (C6.1 L1). Requires golangci-lint on PATH.
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install: https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run

fmt: ## Apply gofmt to all Go files.
	gofmt -w .

fmt-check: ## Fail if any file is not gofmt-clean (used by CI).
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Not gofmt-clean:"; echo "$$unformatted"; exit 1; \
	fi

# Pinned goimports — invoked via `go run` so no PATH setup is required.
GOIMPORTS := go run golang.org/x/tools/cmd/goimports@v0.45.0

imports: ## Apply goimports (formats + organises imports).
	$(GOIMPORTS) -w .

imports-check: ## Fail if any file is not goimports-clean (used by CI).
	@unorganised=$$($(GOIMPORTS) -l .); \
	if [ -n "$$unorganised" ]; then \
		echo "Not goimports-clean:"; echo "$$unorganised"; exit 1; \
	fi

vet: ## Run go vet (cheap static analysis, part of C6.1).
	go vet ./...

run: ## Run the headless debug harness. Args: SEED, TICKS. e.g. make run SEED=7 TICKS=30
	go run ./cmd/headless -seed $(or $(SEED),1) -ticks $(or $(TICKS),20)

golden: ## Regenerate golden-file fixtures. DELIBERATE step — review the diff before committing.
	UPDATE_GOLDEN=1 go test ./engine/ -run Golden
	@echo "Golden files regenerated. Inspect 'git diff' before committing — a changed"
	@echo "golden file means the simulation's behaviour changed. Confirm that was intended."

check: fmt-check imports-check vet build test-race ## The full local gate. Run before every commit.
	@echo "All checks passed."

clean: ## Remove build/coverage artefacts.
	rm -f coverage.out coverage.html
	go clean ./...
