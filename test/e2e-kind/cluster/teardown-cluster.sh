#!/bin/bash
set -euo pipefail

# KubeCoDriver Kind E2E Cluster Teardown Script
# This script cleans up the Kind cluster and associated resources

CLUSTER_NAME="${CLUSTER_NAME:-kubecodriver-e2e}"
CLEANUP_IMAGES="${CLEANUP_IMAGES:-false}"

echo "🧹 Tearing down Kind cluster for KubeCoDriver E2E testing..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Collect logs and artifacts before cleanup
collect_artifacts() {
    echo "📋 Collecting test artifacts..."
    
    local artifact_dir="test-artifacts-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$artifact_dir"
    
    # Collect cluster info
    kubectl cluster-info dump > "$artifact_dir/cluster-info.yaml" 2>/dev/null || true
    
    # Collect KubeCoDriver resources
    kubectl get codriverjobs -A -o yaml > "$artifact_dir/codriverjobs.yaml" 2>/dev/null || true
    kubectl get codrivertools -A -o yaml > "$artifact_dir/codrivertools.yaml" 2>/dev/null || true
    
    # Collect pod logs
    kubectl logs -n kubecodriver-system -l app=kubecodriver-controller --tail=1000 > "$artifact_dir/controller-logs.txt" 2>/dev/null || true
    kubectl logs -n kubecodriver-system -l app=kubecodriver-collector --tail=1000 > "$artifact_dir/collector-logs.txt" 2>/dev/null || true
    
    # Collect events
    kubectl get events -A --sort-by='.lastTimestamp' > "$artifact_dir/events.txt" 2>/dev/null || true
    
    # Collect node information
    kubectl describe nodes > "$artifact_dir/nodes.txt" 2>/dev/null || true
    
    echo "✅ Artifacts collected in: $artifact_dir"
}

# Clean up KubeCoDriver resources
cleanup_kubecodriver_resources() {
    echo "🗑️ Cleaning up KubeCoDriver resources..."
    
    # Delete CoDriverJobs
    kubectl delete codriverjobs --all -A --timeout=60s || true
    
    # Delete CoDriverTools
    kubectl delete codrivertools --all -A --timeout=60s || true
    
    # Delete KubeCoDriver namespace
    kubectl delete namespace kubecodriver-system --timeout=60s || true
    
    echo "✅ KubeCoDriver resources cleaned up"
}

# Delete Kind cluster
delete_cluster() {
    if ! command_exists kind; then
        echo "⚠️ Kind not found, skipping cluster deletion"
        return 0
    fi
    
    if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
        echo "🗑️ Deleting Kind cluster: $CLUSTER_NAME"
        kind delete cluster --name "$CLUSTER_NAME"
        echo "✅ Cluster deleted successfully"
    else
        echo "ℹ️ Cluster $CLUSTER_NAME not found, nothing to delete"
    fi
}

# Clean up container resources
cleanup_container_resources() {
    echo "🧹 Cleaning up container resources..."
    
    # Remove commit-specific images if requested
    if [ "$CLEANUP_IMAGES" = "true" ]; then
        echo "🗑️ Removing commit-specific images..."
        if command_exists docker; then
            docker images | grep "kubecodriver-controller.*e2e-" | awk '{print $1":"$2}' | xargs -r docker rmi -f || true
        elif command_exists podman; then
            podman images | grep "kubecodriver-controller.*e2e-" | awk '{print $1":"$2}' | xargs -r podman rmi -f || true
        fi
    fi
    
    # Remove dangling images
    if command_exists docker; then
        docker image prune -f || true
        docker container prune -f || true
    elif command_exists podman; then
        podman image prune -f || true
        podman container prune -f || true
    fi
    
    echo "✅ Container resources cleaned up"
}

# Clean up temporary files
cleanup_temp_files() {
    echo "🗑️ Cleaning up temporary files..."
    
    # Remove temporary kubeconfig files
    rm -f /tmp/kubeconfig-* || true
    
    # Remove temporary manifests
    rm -rf /tmp/kubecodriver-e2e-* || true
    
    echo "✅ Temporary files cleaned up"
}

# Main execution
main() {
    echo "🎯 KubeCoDriver Kind E2E Cluster Teardown"
    echo "Cluster Name: $CLUSTER_NAME"
    echo ""
    
    # Check if cluster exists before collecting artifacts
    if command_exists kind && kind get clusters | grep -q "^$CLUSTER_NAME$"; then
        # Set kubectl context
        kubectl config use-context "kind-$CLUSTER_NAME" || true
        
        # Collect artifacts before cleanup
        collect_artifacts
        
        # Clean up KubeCoDriver resources
        cleanup_kubecodriver_resources
    else
        echo "ℹ️ Cluster $CLUSTER_NAME not found, skipping resource cleanup"
    fi
    
    # Delete cluster
    delete_cluster
    
    # Clean up container resources
    cleanup_container_resources
    
    # Clean up temporary files
    cleanup_temp_files
    
    echo ""
    echo "🎉 Teardown complete!"
    echo ""
    echo "Summary:"
    echo "  ✅ Test artifacts collected"
    echo "  ✅ KubeCoDriver resources cleaned up"
    echo "  ✅ Kind cluster deleted"
    echo "  ✅ Container resources pruned"
    echo "  ✅ Temporary files removed"
}

# Handle script interruption
trap 'echo "⚠️ Script interrupted, performing cleanup..."; main; exit 1' INT TERM

# Execute main function
main "$@"
