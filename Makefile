OPERATOR_SDK_RELEASE_VERSION ?= v0.8.1
OPERATOR_IMAGE ?= appsody/application-operator
OPERATOR_IMAGE_TAG ?= daily

GIT_COMMIT  ?= $(shell git rev-parse --short HEAD)

# Get source files, ignore vendor directory
SRC_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: build test deploy

setup:
	./scripts/install-operator-sdk.sh ${OPERATOR_SDK_RELEASE_VERSION}

setup-cluster:
	./scripts/setup-cluster.sh

tidy:
	go mod tidy

build:
	go install ./cmd/manager

unit-test:
	go test -v -mod=vendor -tags=unit github.com/appsody/appsody-operator/pkg/controller/appsodyapplication

test-e2e: setup
	operator-sdk test local github.com/appsody/appsody-operator/test/e2e --verbose --debug --up-local --namespace default

generate: setup
	operator-sdk generate k8s
	operator-sdk generate openapi

build-image: setup
	operator-sdk build ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}

push-image:
	docker push ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}

gofmt:
	@gofmt -s -l -w $(SRC_FILES)

golint:
	for file in $(SRC_FILES); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

clean:
	rm -rf build/_output

install:
	kubectl apply -f deploy/releases/daily/appsody-app-crd.yaml
	
deploy:
ifneq "${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}" "appsody/application-operator:daily"
	sed -i.bak -e 's!image: appsody/application-operator:daily!image: ${OPERATOR_IMAGE}:${OPERATOR_IMAGE_TAG}!' deploy/releases/daily/appsody-app-operator.yaml
endif
	sed -i.bak -e "s/APPSODY_WATCH_NAMESPACE/${WATCH_NAMESPACE}/" deploy/releases/daily/appsody-app-operator.yaml
	kubectl apply -f deploy/releases/daily/appsody-app-operator.yaml