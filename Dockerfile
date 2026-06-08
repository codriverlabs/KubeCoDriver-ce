# Build the manager binary
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.26.4@sha256:68cb6d68bed024785b69195b89af7ac7a444f27791435f98647edff595aa0479 AS builder
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
FROM public.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base:latest-al23@sha256:cc6a42d7466110f8445a5fd7acd32461367b5cf1067a5ac7760dcb9f4f0ae507
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
