TEST?=$$(go list ./... |grep -v 'vendor')
PLATFORM?=$$(terraform -version -json | jq .platform -r)
PKG_NAME=scalr
BUILD_ENV=CGO_ENABLED=0
TAG=$(shell PAGER= git tag --points-at HEAD)
BRANCH=$(subst /,-,$(shell git branch --show-current))
VERSION=$(if $(VER),$(VER),$(if $(TAG),$(TAG),$(BRANCH)))
USER_PLUGIN_DIR=${HOME}/.terraform.d/plugins/registry.scalr.io/scalr/scalr/$(VERSION)/$(PLATFORM)
BIN_NAME := terraform-provider-scalr_$(VERSION)
ARGS=-ldflags='-X github.com/scalr/terraform-provider-scalr/version.ProviderVersion=$(TAG) -X github.com/scalr/terraform-provider-scalr/version.Branch=$(BRANCH)'

default: build

build:
	@echo "Building version $(VERSION)"
	$(BUILD_ENV) go build -o $(BIN_NAME) $(ARGS)

test:
	echo $(TEST) | \
		$(BUILD_ENV) xargs -t -n4  go test $(TESTARGS) -timeout=30s -parallel=4


install: build
	@echo "Installing version $(VERSION) for $(PLATFORM)"
	mkdir -p $(USER_PLUGIN_DIR); cp $(BIN_NAME) $(USER_PLUGIN_DIR)

testacc:
	TF_ACC=1 go test -race $(TEST) -v $(TESTARGS) -timeout 45m  -covermode atomic -coverprofile=covprofile

# Run selected acceptance tests against the test environment in .scalr-testenv.env:
#   make testacc-one RUN='^TestAccScalrTag_basic$$'
testacc-one:
	@test -n "$(RUN)" || { echo "Set RUN to a test name regexp, e.g. make testacc-one RUN='^TestAccScalrTag_basic\$$'"; exit 1; }
	@test -f .scalr-testenv.env || { echo ".scalr-testenv.env not found: copy .scalr-testenv.env.sample and fill it in"; exit 1; }
	@set -a; . ./.scalr-testenv.env; set +a; \
	if [ -z "$$SCALR_HOSTNAME" ] || [ -z "$$SCALR_TOKEN" ]; then \
		echo "SCALR_HOSTNAME and SCALR_TOKEN must be set in .scalr-testenv.env"; exit 1; \
	fi; \
	case "$$SCALR_HOSTNAME" in \
		*://*) echo "SCALR_HOSTNAME must be a hostname without a scheme"; exit 1 ;; \
		scalr.io|*.scalr.io) echo "Refusing to run acceptance tests against production ($$SCALR_HOSTNAME)"; exit 1 ;; \
	esac; \
	TF_ACC=1 go test ./internal/provider -run '$(value RUN)' -count=1 -v -timeout 30m $(TESTARGS)

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	golangci-lint fmt

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

.PHONY: build test testacc testacc-one vet fmt test-compile
