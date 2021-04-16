TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=scalr
BIN_NAME=terraform-provider-scalr
BUILD_ENV=CGO_ENABLED=0
USER_PLUGIN_DIR_LINUX=${HOME}/.terraform.d/plugins/scalr.com/scalr/scalr/1.0.0/linux_amd64
CURRENT_VERSION=$(shell BRANCH=`git branch --show-current`; if [[ $$BRANCH == "develop" ]]; then git describe --tags --abbrev=0; else echo $$BRANCH; fi)

default: build

build:
	$(BUILD_ENV) go build -ldflags='-X github.com/scalr/terraform-provider-scalr/version.ProviderVersion=$(CURRENT_VERSION)'

build-linux:
	env $(BUILD_ENV) GOOS=linux GOARCH=amd64 go build

install-linux-user: build-linux
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

