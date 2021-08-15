# Tag-to-digest provider

tagToDigest-provider is used for mutating image tag to a digest using [crane](https://github.com/google/go-containerregistry/tree/main/cmd/crane).

> This repo is meant for testing Gatekeeper external data feature. Do not use for production.

# Installation

- Deploy Gatekeeper with external data enabled (`--enable-external-data`)

- `kubectl apply -f manifest`

- `kubectl apply -f policy/provider.yaml`
  - Update `proxyURL` if it's not `http://tagtodigest-provider.default:8090`

- `kubectl apply -f policy/assign.yaml`

# Verification

- `kubectl apply -f examples/test.yaml`

- `kubectl get deploy test-deployment -o yaml`
  - you should see digests in image
  ```
  ...
      spec:
      containers:
      - image: gcr.io/distroless/static:nonroot@sha256:c9f9b040044cc23e1088772814532d90adadfa1b86dcba17d07cb567db18dc4e
      ...
      - image: gcr.io/distroless/static:nonroot@sha256:c9f9b040044cc23e1088772814532d90adadfa1b86dcba17d07cb567db18dc4e"
  ...
  ```
