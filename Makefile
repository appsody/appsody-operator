OPERATOR_SDK_RELEASE_VERSION ?= v0.15.2
OPERATOR_IMAGE ?= appsody/application-operator
OPERATOR_IMAGE_TAG ?= daily
OPERATOR_MUST_GATHER_TAG ?= daily-must-gather

WATCH_NAMESPACE ?= default
OPERATOR_NAMESPACE ?= ${WATCH_NAMESPACE}

GIT_COMMIT  ?= $(shell git rev-parse --short HEAD)

# Get source files, ignore vendor directory
SRC_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.DEFAULT_GOAL := help

.PHONY: help setup setup-cluster tidy build unit-test test-e2e generate build-image push-image gofmt golint clean install-crd install-rbac install-operator install-all uninstall-all

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: ## Ensure Operator SDK is installed
	./scripts/installers/install-operator-sdk.sh ${OPERATOR_SDK_RELEASE_VERSION}

setup-manifest:
	./scripts/installers/install-manifest-tool.sh

setup-minikube:
	./scripts/installers/install-minikube.sh

tidy: ## Clean up Go modules by adding missing and removing unused modules
	go mod tidy

build: ## Compile the operator
	go install ./cmd/manager

unit-test: ## Run unit tests
	go test -v -mod=vendor -tags=unit github.com/appsody/appsody-operator/pkg/...

test-e2e: setup ## Run end-to-end tests
	./scripts/e2e.sh

test-minikube: setup setup-minikube
	CLUSTER_ENV=minikube operator-sdk test local github.com/appsody/appsody-operator/test/e2e --verbose --debug --up-local --namespace ${WATCH_NAMESPACE}

test-e2e-locally: setup
	kubectl apply -f scripts/servicemonitor.crd.yaml
	CLUSTER_ENV=local operator-sdk test local github.com/appsody/appsody-operator/test/e2e --verbose --debug --up-local --namespace ${WATCH_NAMESPACE}

generate: setup ## Invoke `k8s` and `openapi` generators
	operator-sdk generate k8s
	operator-sdk generate openapi
	
	# Remove `x-kubernetes-int-or-string: true` from CRD. Cusing issues on clusters with older k8s: https://github.com/kubernetes/kubernetes/issues/83778 https://github.com/openshift/api/pull/505
	sed -i '' '/x\-kubernetes\-int\-or\-string\: true/d' deploy/crds/appsody.dev_appsodyapplications_crd.yaml
	
	cat deploy/crds/appsody.dev_appsodyapplications_crd.yaml | awk '/type: object/ {max=NR} {a[NR]=$$0} END{for (i=1;i<=NR;i++) {if (i!=max) print a[i]}}' > deploy/crds/appsody.dev_appsodyapplications_crd.yaml.tmp
	mv deploy/crds/appsody.dev_appsodyapplications_crd.yaml.tmp deploy/crds/appsody.dev_appsodyapplications_crd.yaml

build-image: setup ## Build operator Docker image and tag with "${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}"
	operator-sdk build ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}

build-multiarch-image: setup ## Build and push operator image
	./scripts/build-releases.sh -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}" --image "${OPERATOR_IMAGE}"

build-manifest: setup-manifest
	./scripts/build-manifest.sh -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}" --image "${OPERATOR_IMAGE}"

push-image: ## Push operator image
	docker push ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}

build-must-gather: setup ## Build operator Docker image and tag with "${OPERATOR_IMAGE}:${OPERATOR_MUST_GATHER_TAG}"
	docker build ./must-gather -t ${OPERATOR_IMAGE}:${OPERATOR_MUST_GATHER_TAG} 

push-must-gather: ## Push operator must gather image
	docker push ${OPERATOR_IMAGE}:${OPERATOR_MUST_GATHER_TAG}

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

install-crd: ## Installs operator CRD in the daily directory
	kubectl apply -f deploy/releases/daily/appsody-app-crd.yaml

install-rbac: ## Installs RBAC objects required for the operator to in a cluster-wide manner
	sed -i.bak -e "s/APPSODY_OPERATOR_NAMESPACE/${OPERATOR_NAMESPACE}/" deploy/releases/daily/appsody-app-cluster-rbac.yaml
	kubectl apply -f deploy/releases/daily/appsody-app-cluster-rbac.yaml

install-operator: ## Installs operator in the ${OPERATOR_NAMESPACE} namespace and watches ${WATCH_NAMESPACE} namespace. ${WATCH_NAMESPACE} defaults to `default`. ${OPERATOR_NAMESPACE} defaults to ${WATCH_NAMESPACE}
ifneq "${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}" "appsody/application-operator:daily"
	sed -i.bak -e 's!image: appsody/application-operator:daily!image: ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}!' deploy/releases/daily/appsody-app-operator.yaml
endif
	sed -i.bak -e "s/APPSODY_WATCH_NAMESPACE/${WATCH_NAMESPACE}/" deploy/releases/daily/appsody-app-operator.yaml
	kubectl apply -n ${OPERATOR_NAMESPACE} -f deploy/releases/daily/appsody-app-operator.yaml

install-all: install-crd install-rbac install-operator

uninstall-all:
	kubectl delete -n ${OPERATOR_NAMESPACE} -f deploy/releases/daily/appsody-app-operator.yaml
	kubectl delete -f deploy/releases/daily/appsody-app-cluster-rbac.yaml
	kubectl delete -f deploy/releases/daily/appsody-app-crd.yaml
