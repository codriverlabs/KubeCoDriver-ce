# E2E Test Automation

## Fully Automated Workflow

The `run-tests.sh` script provides **complete end-to-end automation** with **full isolation**:

```bash
./test/e2e-kind/run-tests.sh
```

### Complete Isolation

**Cluster:** `kubecodriver-e2e-<commit-hash>`  
**Image:** `kubecodriver-controller:e2e-<commit-hash>`

Both cluster and image use the same commit hash for complete isolation.

### What Gets Automated

```
┌─────────────────────────────────────────────────────────────┐
│                    Automated E2E Workflow                   │
└─────────────────────────────────────────────────────────────┘

Step 1: Cluster Setup (setup-cluster.sh)
├── Install Kind (if needed)
├── Install kubectl (if needed)
├── Setup container runtime (Docker/Podman)
├── Create Kind cluster (commit-based name)
├── Setup networking (CNI, ingress)
├── Setup storage (default storage class)
├── Setup RBAC (service accounts, roles)
└── Verify cluster health

Step 2: Build & Load Images
├── Build controller image (make docker-build)
│   Image: kubecodriver-controller:e2e-<commit-hash>
├── Load image into Kind cluster
└── Verify image availability

Step 3: Deploy KubeCoDriver Components
├── Install CRDs (CoDriverJob, CoDriverTool)
├── Deploy controller (deployment, RBAC)
├── Wait for controller ready (300s timeout)
└── Verify deployment status

Step 4: Run Tests
├── Execute test suite (Ginkgo)
├── Collect test results
└── Generate reports

Step 5: Cleanup (teardown-cluster.sh)
├── Collect artifacts (logs, resources)
├── Delete CoDriverJobs
├── Delete cluster
└── Prune container images
```

## Usage Examples

### Run All Tests (Full Automation)
```bash
./test/e2e-kind/run-tests.sh
```

**What happens:**
1. Creates cluster: `kubecodriver-e2e-<commit-hash>`
2. Builds controller image
3. Deploys KubeCoDriver stack
4. Runs all test phases
5. Cleans up everything

### Run Specific Phase
```bash
TEST_PHASE=phase1 ./test/e2e-kind/run-tests.sh
```

### Keep Cluster for Debugging
```bash
KEEP_CLUSTER=true ./test/e2e-kind/run-tests.sh
```

### Custom Cluster Name
```bash
CLUSTER_NAME=my-test ./test/e2e-kind/run-tests.sh
```

## CI/CD Integration

### GitHub Actions (Fully Automated)

```yaml
name: E2E Tests

on:
  pull_request:
  push:
    branches: [main]

jobs:
  e2e-kind:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - uses: actions/setup-go@v5
      with:
        go-version: '1.25.3'
    
    - name: Run E2E Tests
      run: ./test/e2e-kind/run-tests.sh
    
    - name: Upload Artifacts
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: e2e-artifacts
        path: test-artifacts-*/
```

**No manual steps required!** The script handles everything.

## Build Process

### Controller Image Build

The script automatically:
1. Checks if image exists
2. Builds if missing: `make docker-build IMG=kubecodriver-controller:e2e`
3. Loads into Kind: `kind load docker-image kubecodriver-controller:e2e`
4. Verifies availability

### Build Cache

To speed up repeated runs:
```bash
# Pre-build image
make docker-build IMG=kubecodriver-controller:e2e

# Run tests (will skip build)
./test/e2e-kind/run-tests.sh
```

### Force Rebuild
```bash
# Remove cached image
docker rmi kubecodriver-controller:e2e

# Run tests (will rebuild)
./test/e2e-kind/run-tests.sh
```

## Deployment Process

### Automated Deployment Steps

1. **CRD Installation**
   ```bash
   kubectl apply -f config/crd/bases/
   ```

2. **Controller Deployment**
   ```bash
   kubectl apply -f test/e2e-kind/manifests/kubecodriver-controller.yaml
   ```

3. **Wait for Ready**
   ```bash
   kubectl wait --for=condition=available --timeout=300s \
       deployment/kubecodriver-controller-manager -n kubecodriver-system
   ```

### Deployment Manifest

The script uses: `test/e2e-kind/manifests/kubecodriver-controller.yaml`

Includes:
- Namespace (kubecodriver-system)
- ServiceAccount
- ClusterRole
- ClusterRoleBinding
- Deployment

## Test Execution

### Automated Test Run

```bash
go test -v -tags=e2ekind \
    -ginkgo.v \
    -ginkgo.show-node-events \
    -timeout=30m \
    ./test/e2e-kind/...
```

### Phase Selection

```bash
# Phase 1: Ephemeral Containers
TEST_PHASE=phase1 ./test/e2e-kind/run-tests.sh

# Phase 2: Real Workloads
TEST_PHASE=phase2 ./test/e2e-kind/run-tests.sh

# All phases
TEST_PHASE=all ./test/e2e-kind/run-tests.sh
```

## Cleanup

### Automatic Cleanup

After tests complete:
```bash
# Cluster is deleted automatically
# Images remain for potential reuse
```

### Manual Cleanup

```bash
# Cleanup specific commit
docker rmi kubecodriver-controller:e2e-<commit-hash>
kind delete cluster --name kubecodriver-e2e-<commit-hash>

# Cleanup all E2E resources
./test/e2e-kind/cleanup-old-images.sh
```

### CI/CD Cleanup

```bash
# Cleanup images after tests
CLEANUP_IMAGES=true ./test/e2e-kind/cluster/teardown-cluster.sh
``` Process

### Automatic Cleanup

On test completion (success or failure):
1. Collect artifacts (logs, resources, events)
2. Delete CoDriverJobs
3. Delete cluster
4. Prune container images

### Manual Cleanup

```bash
# Cleanup specific cluster
kind delete cluster --name kubecodriver-e2e-<commit>

# Cleanup all KubeCoDriver clusters
kind get clusters | grep "^kubecodriver-e2e-" | xargs -I {} kind delete cluster --name {}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CLUSTER_NAME` | `kubecodriver-e2e-<commit>` | Cluster name |
| `KEEP_CLUSTER` | `false` | Keep cluster after tests |
| `TEST_PHASE` | `all` | Test phase to run |
| `TEST_TIMEOUT` | `30m` | Test timeout |
| `GITHUB_SHA` | `git rev-parse HEAD` | Commit hash |

## Troubleshooting

### Build Failures

```bash
# Check Docker/Podman
docker version

# Check Go version
go version

# Manual build
make docker-build IMG=kubecodriver-controller:e2e
```

### Deployment Failures

```bash
# Check cluster
kubectl get nodes

# Check controller logs
kubectl logs -n kubecodriver-system -l app=kubecodriver-controller

# Check events
kubectl get events -n kubecodriver-system
```

### Test Failures

```bash
# Keep cluster for debugging
KEEP_CLUSTER=true ./test/e2e-kind/run-tests.sh

# Inspect resources
kubectl get powertools -A
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name> -c <container-name>
```

## Performance

### Typical Execution Times

- **Cluster Setup:** ~2 minutes
- **Image Build:** ~3 minutes (first time), ~0s (cached)
- **Deployment:** ~30 seconds
- **Phase 1 Tests:** ~5 minutes
- **Cleanup:** ~30 seconds

**Total:** ~10 minutes (first run), ~7 minutes (cached)

### Optimization Tips

1. **Pre-build images** before running tests
2. **Reuse cluster** with `KEEP_CLUSTER=true`
3. **Run specific phases** instead of all
4. **Use local registry** for faster image loading

## Summary

✅ **Fully Automated**
- No manual steps required
- Build → Deploy → Test → Cleanup

✅ **CI/CD Ready**
- Single command execution
- Artifact collection
- Exit code propagation

✅ **Developer Friendly**
- Keep cluster for debugging
- Phase selection
- Custom cluster names

✅ **Production Ready**
- Commit-based naming
- Parallel execution support
- Comprehensive error handling
