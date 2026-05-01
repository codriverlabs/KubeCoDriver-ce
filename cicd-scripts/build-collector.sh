#!/bin/bash
# Build script for collector
# Usage: ./build-collector.sh [local|ecr]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

# Override registry type from command line
if [ $# -gt 0 ]; then
    REGISTRY_TYPE="$1"
fi

# Set image based on registry type
case "$REGISTRY_TYPE" in
    "local")
        IMAGE="$LOCAL_REGISTRY/codriverlabs/kubecodriver-collector"
        ;;
    "ecr")
        IMAGE="$ECR_REGISTRY/codriverlabs/kubecodriver-collector"
        ;;
    *)
        echo "❌ Error: Invalid registry type '$REGISTRY_TYPE'. Use 'local' or 'ecr'"
        exit 1
        ;;
esac

cd "$PROJECT_ROOT"

echo "=== Building kubecodriver-collector ==="
echo "Image: $IMAGE:$VERSION"
echo "Registry Type: $REGISTRY_TYPE"

# Validate Docker
if ! docker info &> /dev/null; then
    echo "❌ Error: Docker daemon is not running"
    exit 1
fi

# Validate Dockerfile exists
if [ ! -f "build/collector/Dockerfile" ]; then
    echo "❌ Error: Collector Dockerfile not found at build/collector/Dockerfile"
    exit 1
fi

# Step 1: Build collector binary
echo "Step 1: Building collector binary..."
if ! make build-collector; then
    echo "❌ Error: Collector binary build failed"
    exit 1
fi
echo "✅ Collector binary built successfully"

# ECR login if needed
if [ "$REGISTRY_TYPE" = "ecr" ]; then
    if ! command -v aws &> /dev/null; then
        echo "❌ Error: 'aws' CLI is not installed"
        exit 1
    fi
    
    echo "Logging into ECR..."
    if ! aws ecr get-login-password --region "$ECR_REGION" | docker login --username AWS --password-stdin "$ECR_REGISTRY"; then
        echo "❌ Error: ECR login failed"
        exit 1
    fi
    echo "✅ ECR login successful"
fi

# Build Docker image
echo "Step 2: Building Docker image..."
if ! docker build -f build/collector/Dockerfile -t "$IMAGE:$VERSION" .; then
    echo "❌ Error: Docker build failed for collector"
    exit 1
fi
echo "✅ Docker image built successfully"

# Push image
echo "Step 3: Pushing Docker image..."
if ! docker push "$IMAGE:$VERSION"; then
    echo "❌ Error: Docker push failed"
    exit 1
fi
echo "✅ Docker image pushed successfully"

echo ""
echo "🎉 Collector built successfully!"
echo "📦 Image: $IMAGE:$VERSION"
