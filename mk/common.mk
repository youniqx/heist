
GCI = $(shell pwd)/bin/gci
bin/gci:
	$(call go-get-tool,$(GCI),github.com/daixiang0/gci@v0.2.9)

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell pwd)/bin v1.43.0

GOFUMPT = $(shell pwd)/bin/gofumpt
bin/gofumpt:
	$(call go-get-tool,$(GOFUMPT),mvdan.cc/gofumpt)

# Format code using gofumpt
format_code: bin/gofumpt
	$(GOFUMPT) -l -w .

validate_format_code: bin/gofumpt
	@echo "gofumpt -d ."
	@GOFUMPT_DIFF=$$($(GOFUMPT) -d .) && \
	if [ -n "$$GOFUMPT_DIFF" ]; then \
	  echo "$$GOFUMPT_DIFF"; \
	  exit 1; \
	fi

# Format imports using GCI
format_imports: bin/gci
	$(GCI) -w .

validate_format_imports: bin/gci
	@echo "gci -d ."
	@GCI_DIFF=$$($(GCI) -d .) && \
	GCI_DIFF_LINES=$$(echo -n "$${GCI_DIFF}" | egrep -v 'skip file .+\.go since no import' | wc -l | sed 's# ##g') && \
	if [ "$$GCI_DIFF_LINES" != "0" ]; then \
	  echo "$$GCI_DIFF"; \
	  exit 1; \
	fi

# Validate formatting of imports and code
validate_format: validate_format_imports validate_format_code

# Format imports and code
format: format_imports format_code

# Format imports and code
fmt: format

# Lints all go files, error on findings
lint: validate_format bin/golangci-lint
	$(GOLANGCI_LINT) run --config .golangci.yaml

# Reformat and lint code, fix as much as possible automatically, error for the rest
fix: format bin/golangci-lint
	$(GOLANGCI_LINT) run --fix --config .golangci.yaml

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell pwd)
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
cd $(PROJECT_DIR)/hack/tools ;\
echo "Downloading $(2)" ;\
echo "GOBIN=$(PROJECT_DIR)/bin go install $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef
