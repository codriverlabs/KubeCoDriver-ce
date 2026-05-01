#!/bin/bash

set -e

# Configuration
AWS_ACCOUNT_ID=""
AWS_REGION=""
IMAGE_VERSION=""

# Images to sync
IMAGES=(
    "kubecodriver-controller"
    "kubecodriver-collector"
    "kubecodriver-aperf"
    "kubecodriver-tcpdump"
    "kubecodriver-chaos"
)

usage() {
    echo "Usage: $0 -a ACCOUNT_ID -r REGION --image-version VERSION"
    echo "Options:"
    echo "  -a, --account-id      AWS Account ID (required)"
    echo "  -r, --region          AWS Region (required)"
    echo "      --image-version   Image version (required)"
    echo "  -h, --help            Show this help"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--account-id)
            AWS_ACCOUNT_ID="$2"
            shift 2
            ;;
        -r|--region)
            AWS_REGION="$2"
            shift 2
            ;;
        -v|--image-version)
            IMAGE_VERSION="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

if [[ -z "$AWS_ACCOUNT_ID" || -z "$AWS_REGION" || -z "$IMAGE_VERSION" ]]; then
    echo "Error: All parameters are required"
    usage
fi

ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

echo "Logging into ECR..."
aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$ECR_REGISTRY"

for image in "${IMAGES[@]}"; do
    SOURCE_IMAGE="ghcr.io/codriverlabs/ce/${image}:${IMAGE_VERSION}"
    TARGET_IMAGE="${ECR_REGISTRY}/codriverlabs/ce/${image}:${IMAGE_VERSION}"
    
    echo "Syncing $SOURCE_IMAGE -> $TARGET_IMAGE"
    
    # Create repository if it doesn't exist
    aws ecr describe-repositories --repository-names "codriverlabs/ce/${image}" --region "$AWS_REGION" >/dev/null 2>&1 || \
        aws ecr create-repository --repository-name "codriverlabs/ce/${image}" --region "$AWS_REGION" >/dev/null
    
    # Pull both architectures with specific platform tags
    docker pull --platform linux/amd64 "$SOURCE_IMAGE"
    docker tag "$SOURCE_IMAGE" "${TARGET_IMAGE}-amd64"
    
    docker pull --platform linux/arm64 "$SOURCE_IMAGE"  
    docker tag "$SOURCE_IMAGE" "${TARGET_IMAGE}-arm64"
    
    # Push temporary tags
    docker push "${TARGET_IMAGE}-amd64"
    docker push "${TARGET_IMAGE}-arm64"
    
    # Remove existing manifest if it exists
    docker manifest rm "$TARGET_IMAGE" 2>/dev/null || true
    
    # Create manifest list with the main tag
    docker manifest create "$TARGET_IMAGE" "${TARGET_IMAGE}-amd64" "${TARGET_IMAGE}-arm64"
    docker manifest annotate "$TARGET_IMAGE" "${TARGET_IMAGE}-amd64" --arch amd64
    docker manifest annotate "$TARGET_IMAGE" "${TARGET_IMAGE}-arm64" --arch arm64
    docker manifest push "$TARGET_IMAGE"
    
    # Clean up architecture-specific tags and source images
    docker rmi "${TARGET_IMAGE}-amd64" "${TARGET_IMAGE}-arm64" "$SOURCE_IMAGE" 2>/dev/null || true
    
    echo "✓ Synced $image"
done

echo "All images synced successfully!"

# Final cleanup - remove any remaining images and dangling images
echo "Cleaning up local images..."
docker image prune -f
docker rmi $(docker images -f "dangling=true" -q) 2>/dev/null || true
