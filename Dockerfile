# Build the manager binary
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.26.5 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

RUN go env -w GOCACHE=/gocache GOMODCACHE=/gomodcache

COPY go.mod go.sum ./
ARG GOPROXY
RUN --mount=type=cache,target=/gomodcache go mod download

COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/
COPY pkg/ pkg/

RUN --mount=type=cache,target=/gomodcache --mount=type=cache,target=/gocache \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -ldflags '-extldflags "-static"' -o manager cmd/main.go

# Use EKS minimal base image (nonroot, AL2023-based, ~27MB)
FROM public.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base:latest-al23@sha256:20fc8188bde3c75f3f98c7e1f159c3c154d22a4ef9dcdb26aaf847488ec4b6a9
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
