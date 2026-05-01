#!/bin/bash
# Clean script for kubecodriver-k8s-operator
# Usage: ./clean.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load configuration
source "$SCRIPT_DIR/config.env"

# Validation functions
validate_tools() {
    echo "Validating required tools..."
    
    if ! command -v kubectl &> /dev/null; then
        echo "❌ Error: 'kubectl' is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v make &> /dev/null; then
        echo "❌ Error: 'make' is not installed or not in PATH"
        exit 1
    fi
    
    echo "✅ All required tools are available"
}

validate_cluster_access() {
    echo "Validating Kubernetes cluster access..."
    
    if ! kubectl cluster-info &> /dev/null; then
        echo "❌ Error: Cannot connect to Kubernetes cluster"
        echo "   Make sure kubectl is configured and cluster is accessible"
        exit 1
    fi
    
    echo "✅ Kubernetes cluster is accessible"
}

cd "$PROJECT_ROOT"

echo "=== Cleaning kubecodriver-k8s-operator ==="

# Validate environment
validate_tools
validate_cluster_access

# Step 1: Undeploy operator
echo "Step 1: Undeploying operator..."
if ! make undeploy-controller-only; then
    echo "⚠️  Warning: Controller undeploy failed (may not exist)"
    # Fallback to full undeploy if controller-only fails
    if ! make undeploy; then
        echo "⚠️  Warning: Full undeploy also failed"
    fi
else
    echo "✅ Controller undeployed successfully"
fi

# Step 2: Uninstall CRDs
echo "Step 2: Uninstalling CRDs..."
if ! make uninstall; then
    echo "⚠️  Warning: Uninstall failed (CRDs may not exist)"
else
    echo "✅ CRDs uninstalled successfully"
fi

# Step 3: Verify cleanup
echo "Step 3: Verifying cleanup..."

# Brief pause to allow controller resources to be cleaned up
echo "  Allowing time for controller resource cleanup..."
sleep 2

echo "  Checking for remaining controller resources in namespace '$NAMESPACE'..."
if kubectl get deployment,service,configmap -n "$NAMESPACE" -l app.kubernetes.io/name=kubecodriver &> /dev/null; then
    echo "⚠️  Some controller resources may still exist:"
    kubectl get deployment,service,configmap -n "$NAMESPACE" -l app.kubernetes.io/name=kubecodriver 2>/dev/null || true
else
    echo "✅ No controller resources found in namespace"
fi

echo "  Checking for CRDs..."
if kubectl get crd | grep -q "kubecodriver.codriverlabs.ai"; then
    echo "⚠️  Some CRDs may still exist:"
    kubectl get crd | grep "kubecodriver.codriverlabs.ai" || true
else
    echo "✅ No kubecodriver CRDs found"
fi

echo ""
echo "🧹 Cleanup completed!"
echo "🔍 Manual verification: kubectl get all -n $NAMESPACE"
