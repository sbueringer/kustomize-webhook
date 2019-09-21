
# Define Docker related variables.
# REGISTRY ?= docker.io/sbueringer
REGISTRY ?= reg-dhc.app.corpintra.net/sbuerin
IMAGE_NAME ?= kustomize-webhook
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME)
TAG ?= latest

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image for controller-manager
	docker build --pull . -t $(CONTROLLER_IMG):$(TAG)

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG):$(TAG)

## --------------------------------------
## Kubernetes
## --------------------------------------

.PHONY: k8s-deploy
k8s-deploy: ## Deploy the webhook to Kubernetes
	k apply -f deploy/webhook.yaml