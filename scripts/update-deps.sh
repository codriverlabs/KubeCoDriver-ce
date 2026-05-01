#!/bin/bash
# Update Go dependencies using Docker (no local Go required)

set -e

echo "Authenticating to ECR Public..."
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws

echo "Running go mod tidy using Go 1.25 container..."

mkdir -p ~/.cache/go-build ~/.cache/go-mod

docker run --rm \
  -v "$(pwd):/workspace" \
  -w /workspace \
  --user "$(id -u):$(id -g)" \
  -e HOME=/workspace \
  -e GOPATH=/workspace/.cache/go \
  -e GOCACHE=/workspace/.cache/go-build \
  public.ecr.aws/docker/library/golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 \
  go mod tidy

echo "Dependencies updated. Please review and commit go.mod and go.sum."
