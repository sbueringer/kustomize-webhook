
# Define Docker related variables.
IMG ?= docker.io/sbueringer/kustomize-webhook
TAG ?= latest

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image for controller-manager
	docker build --pull . -t $(IMG):$(TAG)

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(IMG):$(TAG)
