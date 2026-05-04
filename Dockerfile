# Build the manager binary
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.26.2@sha256:5f3787b7f902c07c7ec4f3aa91a301a3eda8133aa32661a3b3a3a86ab3a68a36 AS builder
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
FROM public.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base:latest-al23@sha256:b9188d3b949bae3e3f22ab3d0f3e0e450aaa4faec03deac31298031596f3aa19
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
