# Complete Build Isolation

## Commit-Based Naming

Both clusters and images use commit-based naming for **complete isolation**:

```
Commit: a1b2c3d4

Cluster: kubecodriver-e2e-a1b2c3d4
Image:   kubecodriver-controller:e2e-a1b2c3d4
```

## Benefits

### ✅ Parallel Testing
Run multiple test suites simultaneously without conflicts:

```bash
# Terminal 1 - PR #123 (commit: abc123)
./test/e2e-kind/run-tests.sh
# Creates: kubecodriver-e2e-abc123
# Builds:  kubecodriver-controller:e2e-abc123

# Terminal 2 - PR #456 (commit: def456)
./test/e2e-kind/run-tests.sh
# Creates: kubecodriver-e2e-def456
# Builds:  kubecodriver-controller:e2e-def456

# No conflicts!
```

### ✅ Build Caching
Each commit gets its own image - no cache conflicts:

```bash
# First run - builds image
GITHUB_SHA=abc123 ./test/e2e-kind/run-tests.sh
# Builds: kubecodriver-controller:e2e-abc123

# Second run - reuses image
GITHUB_SHA=abc123 ./test/e2e-kind/run-tests.sh
# Reuses: kubecodriver-controller:e2e-abc123 (no rebuild)

# Different commit - new image
GITHUB_SHA=def456 ./test/e2e-kind/run-tests.sh
# Builds: kubecodriver-controller:e2e-def456 (isolated)
```

### ✅ CI/CD Safety
Multiple CI jobs can run on same runner:

```yaml
# GitHub Actions - multiple PRs
jobs:
  e2e-pr-123:
    runs-on: ubuntu-latest
    steps:
    - run: ./test/e2e-kind/run-tests.sh
    # Uses: kubecodriver-e2e-<pr-123-commit>
    # Image: kubecodriver-controller:e2e-<pr-123-commit>

  e2e-pr-456:
    runs-on: ubuntu-latest
    steps:
    - run: ./test/e2e-kind/run-tests.sh
    # Uses: kubecodriver-e2e-<pr-456-commit>
    # Image: kubecodriver-controller:e2e-<pr-456-commit>
```

### ✅ Debugging
Keep specific test environment for investigation:

```bash
# Run tests with specific commit
GITHUB_SHA=abc123 KEEP_CLUSTER=true ./test/e2e-kind/run-tests.sh

# Cluster and image remain:
# - kubecodriver-e2e-abc123
# - kubecodriver-controller:e2e-abc123

# Investigate
kubectl --context kind-kubecodriver-e2e-abc123 get pods -A

# Cleanup when done
kind delete cluster --name kubecodriver-e2e-abc123
docker rmi kubecodriver-controller:e2e-abc123
```

## How It Works

### 1. Commit Hash Detection

```bash
# From GitHub Actions
COMMIT_HASH="${GITHUB_SHA}"

# From local git
COMMIT_HASH="$(git rev-parse --short HEAD)"

# Fallback
COMMIT_HASH="local"
```

### 2. Resource Naming

```bash
CLUSTER_NAME="kubecodriver-e2e-${COMMIT_HASH}"
IMAGE_TAG="e2e-${COMMIT_HASH}"
IMAGE_NAME="kubecodriver-controller:${IMAGE_TAG}"
```

### 3. Build Process

```bash
# Check if image exists
if ! docker images | grep -q "kubecodriver-controller.*${IMAGE_TAG}"; then
    # Build with commit-specific tag
    make docker-build IMG="$IMAGE_NAME"
fi

# Load into commit-specific cluster
kind load docker-image "$IMAGE_NAME" --name "$CLUSTER_NAME"
```

### 4. Deployment

```bash
# Deploy with commit-specific image
cat manifests/kubecodriver-controller.yaml | \
    sed "s|image: kubecodriver-controller:e2e|image: $IMAGE_NAME|g" | \
    kubectl apply -f -
```

## Resource Management

### List Resources

```bash
# List all E2E clusters
kind get clusters | grep "^kubecodriver-e2e-"

# List all E2E images
docker images | grep "kubecodriver-controller.*e2e-"
```

### Cleanup Strategies

#### 1. Automatic (Default)
```bash
# Cluster deleted after tests
# Images kept for reuse
./test/e2e-kind/run-tests.sh
```

#### 2. Full Cleanup
```bash
# Delete cluster and images
CLEANUP_IMAGES=true ./test/e2e-kind/run-tests.sh
```

#### 3. Interactive Cleanup
```bash
# Cleanup all old resources
./test/e2e-kind/cleanup-old-images.sh
```

#### 4. Manual Cleanup
```bash
# Specific commit
kind delete cluster --name kubecodriver-e2e-abc123
docker rmi kubecodriver-controller:e2e-abc123

# All E2E resources
kind get clusters | grep "^kubecodriver-e2e-" | xargs -I {} kind delete cluster --name {}
docker images | grep "kubecodriver-controller.*e2e-" | awk '{print $1":"$2}' | xargs docker rmi
```

## Examples

### Local Development

```bash
# Commit: a1b2c3d4
./test/e2e-kind/run-tests.sh

# Creates:
# - Cluster: kubecodriver-e2e-a1b2c3d
# - Image: kubecodriver-controller:e2e-a1b2c3d

# Make code changes, commit: b2c3d4e5
./test/e2e-kind/run-tests.sh

# Creates:
# - Cluster: kubecodriver-e2e-b2c3d4e
# - Image: kubecodriver-controller:e2e-b2c3d4e

# Both environments isolated!
```

### CI/CD Pipeline

```bash
# PR #123 - Commit: abc123
GITHUB_SHA=abc123 ./test/e2e-kind/run-tests.sh
# Cluster: kubecodriver-e2e-abc123
# Image: kubecodriver-controller:e2e-abc123

# PR #456 - Commit: def456 (runs simultaneously)
GITHUB_SHA=def456 ./test/e2e-kind/run-tests.sh
# Cluster: kubecodriver-e2e-def456
# Image: kubecodriver-controller:e2e-def456

# No interference!
```

### Debugging Failed Tests

```bash
# Keep environment for investigation
GITHUB_SHA=abc123 KEEP_CLUSTER=true ./test/e2e-kind/run-tests.sh

# Test fails - environment preserved
# Cluster: kubecodriver-e2e-abc123
# Image: kubecodriver-controller:e2e-abc123

# Investigate
kubectl --context kind-kubecodriver-e2e-abc123 get all -A
kubectl --context kind-kubecodriver-e2e-abc123 logs -n kubecodriver-system -l app=kubecodriver-controller

# Fix code, test again with same environment
GITHUB_SHA=abc123 ./test/e2e-kind/run-tests.sh
# Reuses image: kubecodriver-controller:e2e-abc123 (no rebuild)

# Cleanup when done
kind delete cluster --name kubecodriver-e2e-abc123
docker rmi kubecodriver-controller:e2e-abc123
```

## Comparison

### Before (Shared Resources)
```
Cluster: kubecodriver-e2e (shared)
Image: kubecodriver-controller:e2e (shared)

Problems:
❌ Parallel tests conflict
❌ Image cache conflicts
❌ Hard to debug specific commits
❌ CI/CD race conditions
```

### After (Isolated Resources)
```
Cluster: kubecodriver-e2e-<commit> (isolated)
Image: kubecodriver-controller:e2e-<commit> (isolated)

Benefits:
✅ Parallel tests work
✅ Build caching per commit
✅ Easy debugging
✅ CI/CD safe
✅ Complete isolation
```

## Summary

**Complete Isolation Achieved:**
- ✅ Cluster: `kubecodriver-e2e-<commit-hash>`
- ✅ Image: `kubecodriver-controller:e2e-<commit-hash>`
- ✅ Parallel testing supported
- ✅ Build caching per commit
- ✅ CI/CD safe
- ✅ Easy cleanup

**Zero Conflicts - Fully Isolated Builds!**
