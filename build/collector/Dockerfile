FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

RUN go env -w GOCACHE=/gocache GOMODCACHE=/gomodcache

COPY go.mod go.sum ./
ARG GOPROXY
RUN --mount=type=cache,target=/gomodcache go mod download

COPY cmd/collector/ cmd/collector/
COPY pkg/ pkg/
COPY api/ api/
COPY internal/ internal/

RUN --mount=type=cache,target=/gomodcache --mount=type=cache,target=/gocache \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -ldflags '-extldflags "-static"' -o collector cmd/collector/main.go

FROM public.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base:latest-al23@sha256:9a5ef6e608f86c091d96775954184a657310d5379ce762364075fd4e8e9c5210
WORKDIR /
COPY --from=builder /workspace/collector .
USER 65532:65532
ENTRYPOINT ["/collector"]
