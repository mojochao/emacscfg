# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Build the controller by default.
.DEFAULT_GOAL := help

# Set app identity.
APP ?= $(shell basename "$(CURDIR)")
VERSION ?= $(shell cat VERSION)
PACKAGE ?= github.com/mojochao/${APP}

# Configure image identity
IMAGE_NAME ?= mojochao/${APP}
IMAGE_REPO ?= docker.io
IMAGE_TAG  ?= v${VERSION}
IMAGE ?= ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}

# Set Docker image build configuration.
DOCKERFILE ?= Dockerfile

##@ Info targets

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
#
# See https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters for more
# info on the usage of ANSI control characters for terminal formatting.
#
# See http://linuxcommand.org/lc3_adv_awk.php for more info on the awk command.

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: vars
vars: ## Show environment variables used by this Makefile
	@echo "APP:         $(APP)"
	@echo "PACKAGE:     $(PACKAGE)"
	@echo "VERSION:     $(VERSION)"
	@echo "IMAGE_NAME:  $(IMAGE_NAME)"
	@echo "IMAGE_TAG:   $(IMAGE_TAG)"
	@echo "IMAGE_REPO:  $(IMAGE_REPO)"
	@echo "IMAGE:       $(IMAGE)"

##@ Application targets

.PHONY: run
run: ## Run the application
	@echo 'running $(APP)'
	go run main.go

.PHONY: build
build: ## Build the application
	@echo 'building $(APP)'
	go build  -ldflags "-X $(PACKAGE)/app.version=$(VERSION)" -o $(APP) .

.PHONY: lint
lint: ## Lint the application
	@echo 'linting $(APP)'
	go vet ./...

.PHONY: test
test: ## Run all tests
	@echo 'testing $(APP)'
	go test -v ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo 'cleaning $(APP)'
	rm -f $(APP)

##@ Image targets

.PHONY: image-build
image-build: ## Build the container image
	@echo 'building $(IMAGE)'
	DOCKER_BUILDKIT=1 docker build -t $(IMAGE) .

.PHONY: image-run
image-run: ## Run the container image
	@echo 'running $(IMAGE)'
	docker run --rm $(IMAGE)

.PHONY: image-push
image-push: ## Push the container image
	@echo 'pushing $(IMAGE)'
	docker push $(IMAGE)
