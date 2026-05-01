#!/bin/bash
# Deploy script for collector
# Usage: ./deploy-collector.sh [local|ecr] [--clean]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

CLEAN_DEPLOY=false
REGISTRY_TYPE="${REGISTRY_TYPE:-local}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        local|ecr)
            REGISTRY_TYPE="$1"
            shift
            ;;
        --clean)
            CLEAN_DEPLOY=true
            shift
            ;;
        *)
            echo "❌ Error: Invalid argument '$1'"
            echo "Usage: $0 [local|ecr] [--clean]"
            exit 1
            ;;
    esac
done

cd "$PROJECT_ROOT"

echo "=== Deploying KubeCoDriver Collector ==="
echo "Registry Type: $REGISTRY_TYPE"
echo "Version: $VERSION"
echo "Clean Deploy: $CLEAN_DEPLOY"

# Validate kubectl
if ! command -v kubectl &> /dev/null; then
    echo "❌ Error: 'kubectl' is not installed"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Error: Cannot connect to Kubernetes cluster"
    exit 1
fi

# Clean deployment if requested
if [ "$CLEAN_DEPLOY" = true ]; then
    echo "Cleaning existing collector deployment..."
    if ! make collector-undeploy; then
        echo "⚠️  Warning: Collector undeploy failed (may not exist)"
    else
        echo "✅ Collector cleanup completed"
    fi
fi

# Ensure controller is deployed first (contains RBAC and ServiceAccount)
echo "Ensuring controller is deployed (contains collector RBAC)..."
if ! kubectl get serviceaccount kubecodriver-collector -n "$NAMESPACE" &>/dev/null; then
    echo "❌ Error: Controller must be deployed first (contains kubecodriver-collector ServiceAccount)"
    echo "Run: make install && make deploy IMG=<your-image>"
    exit 1
fi

# Deploy collector workload (Service, PVC, Deployment)
echo "Deploying collector workload..."
if ! make collector-deploy; then
    echo "❌ Error: Failed to deploy collector"
    exit 1
fi
echo "✅ Collector deployed successfully"

# Wait for deployment to be ready
echo "Waiting for collector to be ready..."
if ! kubectl wait --for=condition=available --timeout=120s deployment/kubecodriver-collector -n "$NAMESPACE"; then
    echo "❌ Error: Collector deployment failed to become ready"
    kubectl get pods -n "$NAMESPACE" -l app=kubecodriver-collector
    exit 1
fi
echo "✅ Collector is ready"

# Show deployment status
echo ""
echo "🎉 Collector deployed successfully!"
echo "🔍 Status:"
kubectl get pods -n "$NAMESPACE" -l app=kubecodriver-collector
echo ""
echo "📝 Service endpoint:"
kubectl get service kubecodriver-collector -n "$NAMESPACE"
echo ""
echo "💾 Storage:"
kubectl get pvc profiles-pvc -n "$NAMESPACE"
