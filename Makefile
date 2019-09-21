
# Define Docker related variables.
IMG ?= docker.pkg.github.com/sbueringer/kustomize-webhook/webhook
TAG ?= latest
RELEASE_TAG ?= $(shell git describe --abbrev=0 2>/dev/null)

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image for controller-manager
	docker build --pull . -t $(IMG):$(TAG)

.PHONY: docker-push
docker-push: ## Push the docker image
	docker tag $(IMG):$(TAG) $(IMG):$(RELEASE_TAG)
	docker push $(IMG):$(RELEASE_TAG)
