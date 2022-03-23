REPOSITORY ?= sozercan/tagtodigest-provider
IMG := $(REPOSITORY):latest
ARCH ?= "linux/amd64"

lint:
	golangci-lint run -v ./...

docker-build:
	docker buildx build --platform=${ARCH} -t ${IMG} . --load

docker-build-push:
	docker buildx build --platform=${ARCH} -t ${IMG} . --push
