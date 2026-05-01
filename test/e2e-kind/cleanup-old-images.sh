#!/bin/bash
set -euo pipefail

# Cleanup old E2E images and clusters

echo "🧹 KubeCoDriver E2E Cleanup Utility"
echo ""

# Cleanup old clusters
echo "📋 Finding old E2E clusters..."
OLD_CLUSTERS=$(kind get clusters 2>/dev/null | grep "^kubecodriver-e2e-" || true)

if [ -n "$OLD_CLUSTERS" ]; then
    echo "Found clusters:"
    echo "$OLD_CLUSTERS"
    echo ""
    read -p "Delete all KubeCoDriver E2E clusters? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "$OLD_CLUSTERS" | xargs -I {} kind delete cluster --name {}
        echo "✅ Clusters deleted"
    fi
else
    echo "No KubeCoDriver E2E clusters found"
fi

echo ""

# Cleanup old images
echo "📋 Finding old E2E images..."
OLD_IMAGES=$(docker images | grep "kubecodriver-controller.*e2e-" | awk '{print $1":"$2}' || true)

if [ -n "$OLD_IMAGES" ]; then
    echo "Found images:"
    echo "$OLD_IMAGES"
    echo ""
    read -p "Delete all KubeCoDriver E2E images? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "$OLD_IMAGES" | xargs docker rmi -f
        echo "✅ Images deleted"
    fi
else
    echo "No KubeCoDriver E2E images found"
fi

echo ""
echo "🎉 Cleanup complete!"
