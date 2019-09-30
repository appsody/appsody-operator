OPERATOR_SDK_RELEASE_VERSION ?= v0.8.1
OPERATOR_IMAGE ?= appsody/application-operator
OPERATOR_IMAGE_TAG ?= daily

WATCH_NAMESPACE ?= default

GIT_COMMIT  ?= $(shell git rev-parse --short HEAD)

# Get source files, ignore vendor directory
SRC_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.DEFAULT_GOAL := help

.PHONY: help setup setup-cluster tidy build unit-test test-e2e generate build-image push-image gofmt golint clean install deploy

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: ## Ensure Operator SDK is installed
	./scripts/install-operator-sdk.sh ${OPERATOR_SDK_RELEASE_VERSION}

setup-cluster: ## Install `oc` and starts OpenShift on Docker
	./scripts/setup-cluster.sh

tidy: ## Clean up Go modules by adding missing and removing unused modules
	go mod tidy

build: ## Compile the operator
	go install ./cmd/manager

unit-test: ## Run unit tests
	go test -v -mod=vendor -tags=unit github.com/appsody/appsody-operator/pkg/...

login-oc-registry:
	./scripts/setup-e2e.sh

build-oc-image: setup
	operator-sdk build 172.30.1.1:5000/myproject/appsody-operator:${OPERATOR_IMAGE_TAG}

push-oc-registry:
	docker push 172.30.1.1:5000/myproject/appsody-operator:${OPERATOR_IMAGE_TAG}

restart-docker:
	./scripts/restart-docker.sh

test-e2e: setup ## Run end-to-end tests
	operator-sdk test local github.com/appsody/appsody-operator/test/e2e --namespace myproject --image 172.30.1.1:5000/myproject/appsody-operator:${OPERATOR_IMAGE_TAG}

generate: setup ## Invoke `k8s` and `openapi` generators
	operator-sdk generate k8s
	operator-sdk generate openapi

build-image: setup ## Build operator Docker image and tag with "${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}"
	operator-sdk build ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}

push-image: ## Push operator image
	docker push ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}

gofmt: ## Format the Go code with `gofmt`
	@gofmt -s -l -w $(SRC_FILES)

golint: ## Run linter on operator code
	for file in $(SRC_FILES); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

clean: ## Clean binary artifacts
	rm -rf build/_output

install: ## Installs operator CRD in the daily directory
	kubectl apply -f deploy/releases/daily/appsody-app-crd.yaml

deploy: ## Deploys operator across cluster and watches ${WATCH_NAMESPACE} namespace. If ${WATCH_NAMESPACE} is not specified, it defaults to `default` namespace
ifneq "${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}" "appsody/application-operator:daily"
	sed -i.bak -e 's!image: appsody/application-operator:daily!image: ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}!' deploy/releases/daily/appsody-app-operator.yaml
endif
	sed -i.bak -e "s/APPSODY_WATCH_NAMESPACE/${WATCH_NAMESPACE}/" deploy/releases/daily/appsody-app-operator.yaml
	kubectl apply -f deploy/releases/daily/appsody-app-operator.yaml

# namespaces, service account permissions