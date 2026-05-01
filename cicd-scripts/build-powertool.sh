#!/bin/bash
# Build script for individual power-tool
# Usage: ./build-power-tool.sh <tool-name> [local|ecr]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

# Parse arguments
if [ $# -lt 1 ]; then
    echo "Usage: $0 <tool-name> [local|ecr]"
    echo "Available tools:"
    ls -1 "$PROJECT_ROOT/power-tools/" | grep -v README.md || echo "  No tools found"
    exit 1
fi

TOOL_NAME="$1"
REGISTRY_TYPE="${2:-${REGISTRY_TYPE:-local}}"

# Validate tool exists
if [ ! -d "$PROJECT_ROOT/power-tools/$TOOL_NAME" ]; then
    echo "❌ Error: Powertool '$TOOL_NAME' not found in power-tools/"
    echo "Available tools:"
    ls -1 "$PROJECT_ROOT/power-tools/" | grep -v README.md || echo "  No tools found"
    exit 1
fi

if [ ! -f "$PROJECT_ROOT/power-tools/$TOOL_NAME/Dockerfile" ]; then
    echo "❌ Error: Dockerfile not found for tool '$TOOL_NAME'"
    exit 1
fi

# Set image based on registry type
case "$REGISTRY_TYPE" in
    "local")
        IMAGE="$LOCAL_REGISTRY/codriverlabs/kubecodriver/$TOOL_NAME"
        ;;
    "ecr")
        IMAGE="$ECR_REGISTRY/codriverlabs/kubecodriver/$TOOL_NAME"
        ;;
    *)
        echo "❌ Error: Invalid registry type '$REGISTRY_TYPE'. Use 'local' or 'ecr'"
        exit 1
        ;;
esac

echo "=== Building Power Tool: $TOOL_NAME ==="
echo "Image: $IMAGE:$VERSION"
echo "Registry Type: $REGISTRY_TYPE"

# Validate Docker
if ! docker info &> /dev/null; then
    echo "❌ Error: Docker daemon is not running"
    exit 1
fi

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
echo "Building Docker image..."
if ! docker build -t "$IMAGE:$VERSION" -f "$PROJECT_ROOT/power-tools/$TOOL_NAME/Dockerfile" "$PROJECT_ROOT/power-tools/"; then
    echo "❌ Error: Docker build failed for $TOOL_NAME"
    exit 1
fi
echo "✅ Docker image built successfully"

# Push image
echo "Pushing Docker image..."
if ! docker push "$IMAGE:$VERSION"; then
    echo "❌ Error: Docker push failed"
    exit 1
fi
echo "✅ Docker image pushed successfully"

echo ""
echo "🎉 Power tool '$TOOL_NAME' built successfully!"
echo "📦 Image: $IMAGE:$VERSION"
