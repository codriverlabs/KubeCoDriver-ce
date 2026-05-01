# Testing Setup

## Overview

The KubeCoDriver project uses different testing approaches for different test types:

## Unit Tests (`internal/controller/*_test.go`)

**Location**: `internal/controller/`

**Approach**: Uses **fake Kubernetes clients** - no real cluster required

**Key Features**:
- Fast execution
- No external dependencies
- Uses `sigs.k8s.io/controller-runtime/pkg/client/fake`
- Tests individual functions and methods in isolation

**Example**:
```go
func TestCheckForConflicts(t *testing.T) {
    scheme := runtime.NewScheme()
    _ = kubecodriverv1alpha1.AddToScheme(scheme)
    
    fakeClient := fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(objs...).
        Build()
    // ... test logic
}
```

**Run**: `make test`

## Integration Tests (Ginkgo Suite)

**Location**: `internal/controller/suite_test.go` and related Ginkgo specs

**Approach**: Uses **envtest** - provides a real Kubernetes API server without full cluster

**Key Features**:
- Real API server behavior
- CRD validation
- Webhook testing
- Uses `sigs.k8s.io/controller-runtime/pkg/envtest`
- Automatically downloads and runs etcd + kube-apiserver binaries

**Setup**:
```go
testEnv = &envtest.Environment{
    CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
    ErrorIfCRDPathMissing: true,
}
cfg, err = testEnv.Start()
```

**Run**: `make test` (included in unit test run)

## E2E Tests

**Location**: `test/e2e/`

**Approach**: Uses **real Kubernetes cluster** (Kind or existing cluster)

**Key Features**:
- Full cluster integration
- Tests actual pod creation, ephemeral containers, etc.
- Requires running cluster
- Uses `ctrl.GetConfig()` to connect to cluster

**Run**: 
- `make test-e2e` - Sets up Kind cluster and runs tests
- `make setup-test-e2e` - Only sets up Kind cluster
- `make cleanup-test-e2e` - Tears down Kind cluster

## Test Execution Requirements

### Unit Tests + Integration Tests
```bash
make test
```
**Requirements**:
- ✅ No cluster needed
- ✅ envtest binaries (auto-downloaded via `make setup-envtest`)
- ✅ Go 1.25.3+

### E2E Tests
```bash
make test-e2e
```
**Requirements**:
- ✅ Kind installed
- ✅ Docker running
- ✅ Sufficient resources for Kind cluster

## Environment Variables

- `KUBEBUILDER_ASSETS`: Path to envtest binaries (set automatically by Makefile)
- `ENVTEST_K8S_VERSION`: Kubernetes version for envtest (default: 1.34)
- `KIND_CLUSTER`: Name of Kind cluster for e2e tests (default: kubecodriver-test-e2e)
- `GOTOOLCHAIN`: Go toolchain version (use: go1.25.3)

## Best Practices

1. **Unit tests** should use fake clients for speed and isolation
2. **Integration tests** (Ginkgo) should use envtest for API validation
3. **E2E tests** should use real clusters for end-to-end scenarios
4. Always run `make test` before committing to ensure unit tests pass
5. Run `make test-e2e` before major releases to validate full integration
