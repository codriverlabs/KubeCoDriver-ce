#!/bin/bash
# Build script for kubecodriver-k8s-operator
# Usage: ./build.sh [local|ecr]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

# Validation functions
validate_tools() {
    echo "Validating required tools..."
    
    if ! command -v make &> /dev/null; then
        echo "❌ Error: 'make' is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        echo "❌ Error: 'docker' is not installed or not in PATH"
        exit 1
    fi
    
    if [ "$REGISTRY_TYPE" = "ecr" ] && ! command -v aws &> /dev/null; then
        echo "❌ Error: 'aws' CLI is not installed or not in PATH (required for ECR)"
        exit 1
    fi
    
    echo "✅ All required tools are available"
}

validate_registry_access() {
    if [ "$REGISTRY_TYPE" = "local" ]; then
        echo "Validating local registry access..."
        if ! docker info &> /dev/null; then
            echo "❌ Error: Docker daemon is not running"
            exit 1
        fi
        echo "✅ Docker daemon is running"
    elif [ "$REGISTRY_TYPE" = "ecr" ]; then
        echo "Validating ECR access..."
        if ! aws sts get-caller-identity &> /dev/null; then
            echo "❌ Error: AWS credentials not configured or invalid"
            exit 1
        fi
        echo "✅ AWS credentials are valid"
    fi
}

# Override registry type from command line
if [ $# -gt 0 ]; then
    REGISTRY_TYPE="$1"
fi

# Set image based on registry type
case "$REGISTRY_TYPE" in
    "local")
        IMAGE="$LOCAL_REGISTRY/codriverlabs/$PROJECT_NAME"
        echo "Building for local registry: $IMAGE"
        ;;
    "ecr")
        IMAGE="$ECR_REGISTRY/codriverlabs/$PROJECT_NAME"
        echo "Building for ECR: $IMAGE"
        ;;
    *)
        echo "❌ Error: Invalid registry type '$REGISTRY_TYPE'. Use 'local' or 'ecr'"
        exit 1
        ;;
esac

cd "$PROJECT_ROOT"

echo "=== Building kubecodriver-k8s-operator ==="
echo "Image: $IMAGE:$VERSION"
echo "Registry Type: $REGISTRY_TYPE"

# Validate environment
validate_tools
validate_registry_access

# Step 1: Generate code
echo "Step 1: Generating code..."
if ! make generate; then
    echo "❌ Error: Code generation failed"
    exit 1
fi

# Step 2: Build Go binary
echo "Step 2: Building Go binary..."
if ! make build; then
    echo "❌ Error: Go build failed"
    exit 1
fi
echo "✅ Go binary built successfully"

# Step 3: Build Docker image
echo "Step 3: Building Docker image..."
if ! make docker-build IMG="$IMAGE:$VERSION"; then
    echo "❌ Error: Docker build failed"
    exit 1
fi
echo "✅ Docker image built successfully"

# Step 3: Registry login (ECR only)
if [ "$REGISTRY_TYPE" = "ecr" ]; then
    echo "Step 3: Logging into ECR..."
    if ! aws ecr get-login-password --region "$ECR_REGION" | docker login --username AWS --password-stdin "$ECR_REGISTRY"; then
        echo "❌ Error: ECR login failed"
        exit 1
    fi
    echo "✅ ECR login successful"
fi

# Step 4: Push image
echo "Step 4: Pushing Docker image..."
if ! make docker-push IMG="$IMAGE:$VERSION"; then
    echo "❌ Error: Docker push failed"
    exit 1
fi
echo "✅ Docker image pushed successfully"

# Step 5: Generate manifests
echo "Step 5: Generating manifests..."
if ! make manifests; then
    echo "❌ Error: Manifest generation failed"
    exit 1
fi
echo "✅ Manifests generated successfully"

# Step 6: Generate installer
echo "Step 6: Generating installer..."
if ! make build-installer IMG="$IMAGE:$VERSION"; then
    echo "❌ Error: Installer generation failed"
    exit 1
fi
echo "✅ Installer generated successfully"

echo ""
echo "🎉 Build completed successfully!"
echo "📦 Image: $IMAGE:$VERSION"
echo "📄 Installer: dist/install.yaml"
