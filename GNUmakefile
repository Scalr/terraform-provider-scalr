TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PLATFORM?=$$(terraform -version -json | jq .platform -r)
PKG_NAME=scalr
BUILD_ENV=CGO_ENABLED=0
TAG=$(shell PAGER= git tag --points-at HEAD)
BRANCH=$(subst /,-,$(shell git branch --show-current))
VERSION=$(if $(VER),$(VER),$(if $(TAG),$(TAG),$(BRANCH)))
USER_PLUGIN_DIR=${HOME}/.terraform.d/plugins/registry.scalr.io/scalr/scalr/$(VERSION)/$(PLATFORM)
BIN_NAME := terraform-provider-scalr_$(VERSION)
ARGS=-ldflags='-X github.com/scalr/terraform-provider-scalr/version.ProviderVersion=$(TAG) -X github.com/scalr/terraform-provider-scalr/version.Branch=$(BRANCH)'
UPSTREAM_COMMIT_DESCRIPTION="Scalr terraform provider acceptance tests"
UPSTREAM_COMMIT_TARGET_URL = "https://github.com/Scalr/terraform-provider-scalr/actions/runs/$(run_id)"

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

notify-upstream:
	curl -X POST \
	-H "Accept: application/vnd.github.v3+json" \
	-H "Authorization: token $(org_admin_token)" \
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

resource:
	@echo "Skaffolding resource $(name)..."
	@cd skaff && go run cmd/main.go -type=resource -name=$(name)

datasource:
	@echo "Skaffolding datasource $(name)..."
	@cd skaff && go run cmd/main.go -type=data_source -name=$(name)

.PHONY: build test testacc vet fmt test-compile notify-upstream resource datasource
