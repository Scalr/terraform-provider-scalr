TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=scalr
BUILD_ENV=CGO_ENABLED=0
TAG=$(shell PAGER= git tag --points-at HEAD)
BRANCH=$(subst /,-,$(shell git branch --show-current))
VERSION=$(if $(TAG),$(TAG),$(BRANCH))
USER_PLUGIN_DIR_LINUX=${HOME}/.terraform.d/plugins/scalr.io/scalr/scalr/$(VERSION)/linux_amd64
BIN_NAME := terraform-provider-scalr_$(VERSION)
ARGS=-ldflags='-X github.com/scalr/terraform-provider-scalr/version.ProviderVersion=$(TAG) -X github.com/scalr/terraform-provider-scalr/version.Branch=$(BRANCH)'

default: build

build:
	@echo "Building version $(VERSION)"
	$(BUILD_ENV) go build -o $(BIN_NAME) $(ARGS)

build-linux:
	@echo "Building version $(VERSION) for linux"
	env $(BUILD_ENV) GOOS=linux GOARCH=amd64 go build -o $(BIN_NAME) $(ARGS)

install-linux-user: build-linux
	@echo "Installing version $(VERSION) for linux"
	mkdir -p $(USER_PLUGIN_DIR_LINUX); cp $(BIN_NAME) $(USER_PLUGIN_DIR_LINUX)

test:
	echo $(TEST) | \
		$(BUILD_ENV) xargs -t -n4  go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	$(BUILD_ENV) TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 15m  -covermode atomic -coverprofile=covprofile

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)
.PHONY: build build-linux test testacc vet fmt test-compile

