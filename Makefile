.PHONY: dev-backend dev-frontend frontend-install frontend-check frontend-build web-generate build

dev-backend:
	GOWORK=off go run ./cmd/vm-system serve --listen 127.0.0.1:3210

dev-frontend:
	pnpm -C ui dev

frontend-install:
	pnpm -C ui install --frozen-lockfile

frontend-check:
	pnpm -C ui check

frontend-build:
	pnpm -C ui run build

web-generate:
	go generate ./internal/web

build:
	go generate ./internal/web
	GOWORK=off go build -tags embed -o vm-system ./cmd/vm-system

.PHONY: bump-go-go-golems
bump-go-go-golems:
	@deps="$$(awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $$2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $$1 }' go.mod | sort -u)"; \
	if [ -z "$$deps" ]; then \
		echo "No github.com/go-go-golems dependencies in go.mod"; \
	else \
		echo "Bumping go-go-golems dependencies:"; \
		echo "$$deps"; \
		for dep in $$deps; do GOWORK=off go get "$${dep}@latest"; done; \
	fi
	GOWORK=off go mod tidy

GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_LINT_TOOL_VERSION ?= v1.3.5
GLAZED_LINT_FLAGS ?= -glazedclilint.allow-paths=cmd/vm-system/
GLAZED_LINT_DIRS ?= ./cmd/... ./pkg/... ./internal/...

.PHONY: glazed-lint-build glazed-lint

 glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_LINT_TOOL_VERSION)"; \
	GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_LINT_TOOL_VERSION)

glazed-lint: glazed-lint-build
	GOWORK=off go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)

.PHONY: logcopter-generate
logcopter-generate:
	GOWORK=off go tool logcopter-gen -include-main -area-prefix go-go-golems.vm-system -strip-prefix github.com/go-go-golems/vm-system ./cmd/... ./internal/... ./pkg/...

.PHONY: logcopter-check
logcopter-check:
	GOWORK=off go tool logcopter-gen -include-main -area-prefix go-go-golems.vm-system -strip-prefix github.com/go-go-golems/vm-system -check ./cmd/... ./internal/... ./pkg/...
