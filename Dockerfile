ARG BUILDERIMAGE="golang:1.17"
ARG BASEIMAGE="gcr.io/distroless/static:nonroot"

FROM $BUILDERIMAGE as builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG LDFLAGS

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}

WORKDIR /go/src/github.com/sozercan/tagToDigest-provider

COPY . .

RUN go build -o provider main.go

FROM $BASEIMAGE

WORKDIR /

COPY --from=builder /go/src/github.com/sozercan/tagToDigest-provider .

USER 65532:65532

ENTRYPOINT ["/provider"]
