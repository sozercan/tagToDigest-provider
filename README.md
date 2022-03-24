# Tag-to-digest provider

tagToDigest-provider is used for mutating image tag to a digest using [crane](https://github.com/google/go-containerregistry/tree/main/cmd/crane).

> This repo is meant for testing Gatekeeper external data feature. Do not use for production.

# Installation

- Deploy Gatekeeper with external data enabled (`--enable-external-data`)

- `kubectl apply -f manifest/`

- `kubectl apply -f policy/provider.yaml`
  - Update `url` if it's not `http://tagtodigest-provider.external-data-providers:8090/mutate`

- `kubectl apply -f policy/assign.yaml`

# Verification

- `kubectl apply -f policy/examples/test.yaml --dry-run=server -ojson | jq -r '.spec.template.spec.containers[].image`

  - before:

  ```
  gcr.io/distroless/static:nonroot
  gcr.io/distroless/static:nonroot@sha256:c9f9b040044cc23e1088772814532d90adadfa1b86dcba17d07cb567db18dc4e
  busybox
  nginx:1.21.6
  quay.io/prometheus/node-exporter:v1.3.1
  mcr.microsoft.com/oss/kubernetes/pause:3.6
  upstream.azurecr.io/oss/kubernetes/pause:3.6
  public.ecr.aws/datadog/agent:latest
  public.ecr.aws/amazonlinux/amazonlinux:latest
  mcr.microsoft.com/oss/bitnami/redis:6.0.8
  ```

  - after:

  ```
  gcr.io/distroless/static:nonroot@sha256:80c956fb0836a17a565c43a4026c9c80b2013c83bea09f74fa4da195a59b7a99
  gcr.io/distroless/static:nonroot@sha256:c9f9b040044cc23e1088772814532d90adadfa1b86dcba17d07cb567db18dc4e
  busybox@sha256:caa382c432891547782ce7140fb3b7304613d3b0438834dce1cad68896ab110a
  nginx:1.21.6@sha256:85f3b7a34506d74088124917360343e00c73a1b617a2371cc59fa9fa44d89a42
  quay.io/prometheus/node-exporter:v1.3.1@sha256:f2269e73124dd0f60a7d19a2ce1264d33d08a985aed0ee6b0b89d0be470592cd
  mcr.microsoft.com/oss/kubernetes/pause:3.6@sha256:b4b669f27933146227c9180398f99d8b3100637e4a0a1ccf804f8b12f4b9b8df
  upstream.azurecr.io/oss/kubernetes/pause:3.6@sha256:1fe8b51fb6120c0504015e795b9b048f1e7a26b548ed1a7713ec20c6c69be508
  public.ecr.aws/datadog/agent:latest@sha256:2ef4ef739b3809872bc8bb959b19c0fc665d239cae306c7adec95e63deb4ab3c
  public.ecr.aws/amazonlinux/amazonlinux:latest@sha256:334ec0ec042eff13d9581120f80a46fb0861d6a4ba0e9f44e09650979ec5d2df
  mcr.microsoft.com/oss/bitnami/redis:6.0.8@sha256:9b53ae0f1cf3f7d7854584c8b7c5a96fe732c48d504331da6c00f892fdcce102
  ```
