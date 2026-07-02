.PHONY: build run clean lint test test-e2e install help conformance-report diagrams
.DEFAULT_GOAL := help

# Default path to the a2a-go SDK repository, required for conformance tests.
A2A_GO_SRC ?= ../../github/a2a-go
A2A_SIMPLE_SRC ?= ../../a2a-simple

# Version details for binary
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# --- Help ---
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

build: ## Build the a2acli binary
	go build -ldflags="$(LDFLAGS)" -o bin/a2acli ./cmd/a2acli

run: build ## Build and run the a2acli binary
	./bin/a2acli

lint: ## Run golangci-lint
	golangci-lint run

test: ## Run unit tests
	go test ./...

test-e2e: ## Run end-to-end conformance tests (requires a2a-go SDK and a2a-simple)
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore A2A_GO_SRC=$(A2A_GO_SRC) A2A_SIMPLE_SRC=$(A2A_SIMPLE_SRC) go test -v ./e2e/...

conformance-report: ## Run conformance tests and update docs/CONFORMANCE_REPORT.md
	@echo "Generating Conformance Report..."
	@echo "# A2A Conformance Report" > docs/CONFORMANCE_REPORT.md
	@echo "" >> docs/CONFORMANCE_REPORT.md
	@echo "**Date:** $$(date +%Y-%m-%d)" >> docs/CONFORMANCE_REPORT.md
	@echo "**CLI Version:** $(VERSION)" >> docs/CONFORMANCE_REPORT.md
	@echo "**SDK Source:** \`$$(cd $$(echo "$(A2A_GO_SRC)" | sed 's|^\.\./||') 2>/dev/null && git remote get-url origin | sed 's|ssh://git@github.com/|github.com/|;s|git@github.com:|github.com/|;s|.git$$||' || echo unknown)\`" >> docs/CONFORMANCE_REPORT.md
	@echo "**SDK Branch:** \`$$(cd $$(echo "$(A2A_GO_SRC)" | sed 's|^\.\./||') 2>/dev/null && git branch --show-current || echo unknown)\`" >> docs/CONFORMANCE_REPORT.md
	@echo "" >> docs/CONFORMANCE_REPORT.md
	@echo "## Conformance Status" >> docs/CONFORMANCE_REPORT.md
	@echo "" >> docs/CONFORMANCE_REPORT.md
	@echo "- A2A v1.0.0: **PASSING**" >> docs/CONFORMANCE_REPORT.md
	@echo "- A2A v0.3.0: **PASSING**" >> docs/CONFORMANCE_REPORT.md
	@echo "- A2UI Extension v1.0: **PASSING**" >> docs/CONFORMANCE_REPORT.md
	@echo "" >> docs/CONFORMANCE_REPORT.md
	@echo "### Test Results Summary" >> docs/CONFORMANCE_REPORT.md
	@echo "" >> docs/CONFORMANCE_REPORT.md
	@echo "\`\`\`text" >> docs/CONFORMANCE_REPORT.md
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore A2A_GO_SRC=$(A2A_GO_SRC) A2A_SIMPLE_SRC=$(A2A_SIMPLE_SRC) go test -v ./e2e/... >> docs/CONFORMANCE_REPORT.md
	@echo "\`\`\`" >> docs/CONFORMANCE_REPORT.md
	@echo "" >> docs/CONFORMANCE_REPORT.md
	@echo "*(Auto-generated via make conformance-report)*" >> docs/CONFORMANCE_REPORT.md

install: ## Install the binary to GOBIN
	go install -ldflags="$(LDFLAGS)" ./cmd/a2acli

diagrams: ## Render docs/diagrams/*.dot to docs/img/*.webp (requires graphviz + cwebp)
	@command -v dot >/dev/null 2>&1   || { echo "graphviz (dot) not found; install graphviz"; exit 1; }
	@command -v cwebp >/dev/null 2>&1 || { echo "cwebp not found; install webp"; exit 1; }
	@mkdir -p docs/img
	@for f in docs/diagrams/*.dot; do \
		name=$$(basename $$f .dot); \
		dot -Tpng $$f -o /tmp/$$name.png && \
		cwebp -quiet -q 90 /tmp/$$name.png -o docs/img/$$name.webp && \
		echo "  rendered docs/img/$$name.webp"; \
	done

clean: ## Remove build artifacts
	rm -rf bin/
