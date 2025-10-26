.PHONY: help build install test clean controller dashboard docker-build docker-push deploy

BINARY_NAME=codedance-controller
DASHBOARD_BINARY=codedance-dashboard
DOCKER_IMAGE=codedance-controller
VERSION?=latest

help:
	@echo "Available targets:"
	@echo "  build          - Build the controller binary"
	@echo "  dashboard      - Build and run the dashboard"
	@echo "  install        - Install CRDs to the cluster"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  controller     - Run controller locally"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image"
	@echo "  deploy         - Deploy to Kubernetes cluster"

build:
	@echo "Building controller..."
	go build -o bin/$(BINARY_NAME) cmd/controller/main.go

build-dashboard:
	@echo "Building dashboard..."
	go build -o bin/$(DASHBOARD_BINARY) cmd/dashboard/main.go

dashboard: build-dashboard
	@echo "Starting dashboard on http://localhost:8080"
	./bin/$(DASHBOARD_BINARY) --kubeconfig=$(HOME)/.kube/config

install:
	@echo "Installing CRDs..."
	kubectl apply -f config/crd/canary_deployment.yaml
	kubectl apply -f config/rbac/

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

controller: build
	@echo "Running controller..."
	./bin/$(BINARY_NAME) --kubeconfig=$(HOME)/.kube/config

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

docker-push:
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)

deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deploy/kubernetes/

generate:
	@echo "Generating CRD code..."
	go mod tidy
	go mod vendor

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run

.DEFAULT_GOAL := help
