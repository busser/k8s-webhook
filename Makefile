
# Image URL to use all building/pushing image targets
VERSION ?= v1.0.0
IMG ?= padok.fr/webhook:$(VERSION)

all: help

## help: Display list of commands
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## test: Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

## build: Build manager binary
build: fmt vet
	go build -o bin/webhook main.go

## run: Run locally
run: fmt vet
	go run ./main.go

## deploy: Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: kind-load
	cd manifests && kustomize edit set image webhook=${IMG}
	kustomize build manifests | kubectl apply -f -

## undeploy: Remove controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	kustomize build manifests | kubectl delete -f -

## fmt: Run go fmt against code
fmt:
	go fmt ./...

## vet: Run go vet against code
vet:
	go vet ./...

## docker-build: Build the docker image
docker-build: test
	docker build . -t ${IMG}

## docker-push: Push the docker image
docker-push:
	docker push ${IMG}

## kind-load: Load the docker image to a local kind cluster
kind-load: docker-build
	kind load docker-image ${IMG}
