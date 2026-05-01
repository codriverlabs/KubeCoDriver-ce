#!/bin/bash
set -e

# KubeCoDriver CE Sync Script (Rsync Strategy)
# Synchronizes a tagged release from the private repo to KubeCoDriver-ce.
# Usage: ./scripts/sync-to-ce.sh <tag-name> [ce-repo-path]
# Example: ./scripts/sync-to-ce.sh v1.1.0
# Example: ./scripts/sync-to-ce.sh v1.1.0 /path/to/KubeCoDriver-ce

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEFAULT_CE_REPO="$(dirname "$PROJECT_ROOT")/KubeCoDriver-ce"
TEMP_SYNC_DIR="/tmp/kubecodriver-sync-$$"

cd "$PROJECT_ROOT"

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
    echo "Usage: $0 <tag-name> [ce-repo-path]"
    echo "Example: $0 v1.1.0"
    echo "Example: $0 v1.1.0 /path/to/KubeCoDriver-ce"
    exit 1
fi

TAG_NAME="$1"
CE_REPO="${2:-$DEFAULT_CE_REPO}"

echo "Tag: $TAG_NAME"
echo "CE repo: $CE_REPO"

# Verify tag exists
if ! git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    echo "Error: Tag '$TAG_NAME' does not exist"
    exit 1
fi

# Verify CE repo exists
if [ ! -d "$CE_REPO/.git" ]; then
    echo "Error: CE repository not found at $CE_REPO"
    exit 1
fi

echo "Syncing tag '$TAG_NAME' to CE repository..."

# Create temporary directory
mkdir -p "$TEMP_SYNC_DIR"
trap "rm -rf '$TEMP_SYNC_DIR'" EXIT

# Extract tag to temporary directory
echo "Extracting tag to temporary directory..."
git archive --format=tar "$TAG_NAME" | tar -x -C "$TEMP_SYNC_DIR"

# Sync with rsync (delete files not in source, exclude .git and .kiro)
echo "Syncing to CE repository with rsync..."
rsync -av --delete --checksum \
    --exclude='.git/' \
    --exclude='.kiro*' \
    --exclude='.agents/' \
    "$TEMP_SYNC_DIR/" "$CE_REPO/"

# Commit changes in CE repo
cd "$CE_REPO"
echo "Committing changes in CE repository..."

if git diff --quiet && git diff --cached --quiet; then
    echo "No changes to commit"
else
    git add .
    git commit -m "Sync from private repo $TAG_NAME"

    echo "Pushing changes to CE repository..."
    git push

    echo "✅ Successfully synced $TAG_NAME to CE repository"
fi

echo "Sync completed!"
