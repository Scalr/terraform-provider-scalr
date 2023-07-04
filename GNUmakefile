TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PLATFORM?=$$(terraform -version -json | jq .platform -r)
PKG_NAME=scalr
BUILD_ENV=CGO_ENABLED=0
TAG=$(shell PAGER= git tag --points-at HEAD)
BRANCH=$(subst /,-,$(shell git branch --show-current))
VERSION=$(if $(VER),$(VER),$(if $(TAG),$(TAG),$(BRANCH)))
USER_PLUGIN_DIR_LINUX=${HOME}/.terraform.d/plugins/scalr.io/scalr/scalr/$(VERSION)/linux_amd64
USER_PLUGIN_DIR=${HOME}/.terraform.d/plugins/scalr.io/scalr/scalr/$(VERSION)/$(PLATFORM)
BIN_NAME := terraform-provider-scalr_$(VERSION)
ARGS=-ldflags='-X github.com/scalr/terraform-provider-scalr/version.ProviderVersion=$(TAG) -X github.com/scalr/terraform-provider-scalr/version.Branch=$(BRANCH)'
UPSTREAM_COMMIT_DESCRIPTION="Scalr terraform provider acceptance tests"
UPSTREAM_COMMIT_TARGET_URL = "https://github.com/Scalr/terraform-provider-scalr/actions/runs/$(run_id)"

default: build

build:
	@echo "Building version $(VERSION)"
	$(BUILD_ENV) go build -o $(BIN_NAME) $(ARGS)

install: build
	@echo "Installing version $(VERSION) for $(PLATFORM)"
	mkdir -p $(USER_PLUGIN_DIR); cp $(BIN_NAME) $(USER_PLUGIN_DIR)

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
	TF_ACC=1 go test -race $(TEST) -v $(TESTARGS) -timeout 15m  -covermode atomic -coverprofile=covprofile

notify-upstream:
	curl -X POST \
	-H "Accept: application/vnd.github.v3+json" \
	-H "Authorization: token $(ORG_ADMIN_TOKEN)" \
	https://api.github.com/repos/Scalr/fatmouse/statuses/$(upstream_sha) \
	-d '{"context":"downstream/provider", "state":"$(state)", "description": $(UPSTREAM_COMMIT_DESCRIPTION), "target_url": $(UPSTREAM_COMMIT_TARGET_URL)}'


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
.PHONY: build build-linux install install-linux-user test testacc vet fmt test-compile notify-upstream


# bump
