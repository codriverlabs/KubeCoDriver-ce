# Kind E2E Testing Strategy

## Overview

This document outlines the strategy for implementing comprehensive E2E tests using Kind clusters for the KubeCoDriver project. It identifies reusable test code, new test cases needed, and cluster management approach.

## Cluster Naming Strategy

### Commit-Based Cluster Names
For single GitHub runner environments, use commit hash in cluster name to enable parallel test runs:

```bash
CLUSTER_NAME="kubecodriver-e2e-${GITHUB_SHA:0:8}"  # e.g., kubecodriver-e2e-a1b2c3d4
```

**Benefits:**
- Enables parallel PR testing on single runner
- Prevents cluster name conflicts
- Easy identification of test runs
- Automatic cleanup based on commit lifecycle

**Implementation:**
```bash
# In setup-cluster.sh
COMMIT_HASH="${GITHUB_SHA:-$(git rev-parse --short HEAD)}"
CLUSTER_NAME="${CLUSTER_NAME:-kubecodriver-e2e-${COMMIT_HASH}}"
```

## Reusable Test Code from Existing E2E Suite

### ✅ Directly Reusable (No Changes Needed)

**1. Test Utilities (`simple_utils.go`)**
- `CreateSimpleTestNamespace()` - Namespace creation
- `CreateSimpleMockTargetPod()` - Mock pod generation
- `CreateSimpleTestCoDriverJob()` - CoDriverJob creation
- `CreateSimpleTestCoDriverTool()` - CoDriverTool creation
- `CreateSimpleBasicCoDriverJobSpec()` - Spec generation
- `GetSimpleCoDriverJob()` / `GetSimplePod()` - Resource retrieval
- `LogSimpleCoDriverJobStatus()` - Debug logging

**2. Test Scenarios (Logic Tests)**
All controller logic tests from existing suite:
- Target pod discovery and label selectors
- Reconciliation state management
- Resource lifecycle management
- Error handling and recovery
- Multi-pod scenarios
- Performance and timing tests

**3. Validation Tests**
- CoDriverJob validation logic
- CoDriverTool validation
- Tool configuration validation
- Output mode configuration tests
- RBAC security tests (partial)

### 🔄 Needs Adaptation (Minor Changes)

**1. Suite Setup (`e2e_suite_test.go`)**
```go
// CHANGE: Remove envtest setup, use existing Kind cluster
var _ = BeforeSuite(func() {
    // REMOVE: Docker build (done by setup script)
    // REMOVE: Kind image loading (done by setup script)
    // KEEP: CertManager check and install
    
    // ADD: Connect to existing Kind cluster
    By("connecting to Kind cluster")
    config, err := ctrl.GetConfig()
    Expect(err).NotTo(HaveOccurred())
    
    // ADD: Initialize clients
    err = InitializeSimpleClients()
    Expect(err).NotTo(HaveOccurred())
})
```

**2. Wait Functions**
```go
// CHANGE: Increase timeouts for real cluster operations
func WaitForSimpleCoDriverJobPhase(powerTool *v1alpha1.CoDriverJob, expectedPhase string) {
    Eventually(func() string {
        // ... existing logic ...
    }, "120s", "2s").Should(Equal(expectedPhase))  // Was: 30s, 1s
}
```

## New Test Cases for Kind Cluster

### 🆕 Phase 1: Ephemeral Container Tests (Critical)

**These tests CANNOT run on envtest - require real kubelet**

```go
var _ = Describe("Ephemeral Container Profiling", func() {
    Context("Container Creation", func() {
        It("should create ephemeral container in target pod", func() {
            // Verify actual ephemeral container creation
            // Check container appears in pod.spec.ephemeralContainers
        })
        
        It("should execute profiling tool in ephemeral container", func() {
            // Verify tool execution
            // Check container logs for tool output
        })
        
        It("should handle ephemeral container lifecycle", func() {
            // Verify container starts, runs, completes
            // Check container status transitions
        })
    })
    
    Context("Tool Execution", func() {
        It("should run aperf tool and collect data", func() {
            // Real aperf execution
            // Verify profiling data collection
        })
        
        It("should run strace tool and collect data", func() {
            // Real strace execution
            // Verify trace data collection
        })
        
        It("should handle tool timeout correctly", func() {
            // Verify tool stops after duration
            // Check cleanup happens
        })
    })
    
    Context("Data Collection", func() {
        It("should collect profiling data in ephemeral mode", func() {
            // Verify data appears in expected location
            // Check data format and content
        })
        
        It("should handle large profiling output", func() {
            // Test with tool generating large output
            // Verify no data loss
        })
    })
})
```

### 🆕 Phase 2: Real Workload Profiling

```go
var _ = Describe("Real Workload Profiling", func() {
    Context("Web Application Profiling", func() {
        It("should profile nginx under load", func() {
            // Deploy nginx
            // Generate traffic
            // Profile with aperf
            // Verify profiling data captured
        })
        
        It("should profile application with multiple containers", func() {
            // Deploy multi-container pod
            // Profile specific container
            // Verify correct container targeted
        })
    })
    
    Context("Database Profiling", func() {
        It("should profile redis under load", func() {
            // Deploy redis
            // Generate operations
            // Profile with strace
            // Verify syscall traces captured
        })
    })
    
    Context("Batch Job Profiling", func() {
        It("should profile short-lived job", func() {
            // Deploy job
            // Profile during execution
            // Handle job completion
        })
    })
})
```

### 🆕 Phase 3: Storage Integration

```go
var _ = Describe("Storage Integration", func() {
    Context("PVC Output Mode", func() {
        It("should write profiling data to PVC", func() {
            // Create PVC
            // Run CoDriverJob with PVC mode
            // Verify data written to PVC
            // Verify data persists after pod deletion
        })
        
        It("should handle PVC storage class selection", func() {
            // Test with different storage classes
            // Verify correct provisioning
        })
        
        It("should handle PVC capacity limits", func() {
            // Test with small PVC
            // Verify behavior when full
        })
    })
    
    Context("Collector Output Mode", func() {
        It("should send data to collector service", func() {
            // Deploy collector
            // Run CoDriverJob with collector mode
            // Verify data received by collector
            // Check collector storage
        })
        
        It("should handle collector authentication", func() {
            // Test with token authentication
            // Verify secure data transfer
        })
        
        It("should retry on collector failure", func() {
            // Simulate collector unavailable
            // Verify retry logic
            // Verify eventual success
        })
    })
})
```

### 🆕 Phase 4: Multi-Node Scenarios

```go
var _ = Describe("Multi-Node Scenarios", func() {
    Context("Pod Scheduling", func() {
        It("should profile pods on different nodes", func() {
            // Deploy pods to different nodes
            // Profile all pods
            // Verify profiling works across nodes
        })
        
        It("should handle node affinity constraints", func() {
            // Deploy with node affinity
            // Verify profiling respects constraints
        })
    })
    
    Context("Network Policies", func() {
        It("should respect network policies for collector", func() {
            // Apply network policy
            // Verify collector communication
            // Test policy enforcement
        })
    })
})
```

### 🆕 Phase 5: Security and RBAC

```go
var _ = Describe("Security and RBAC", func() {
    Context("Namespace Isolation", func() {
        It("should enforce namespace boundaries", func() {
            // Create CoDriverJob in namespace A
            // Try to target pod in namespace B
            // Verify access denied
        })
        
        It("should respect CoDriverTool namespace restrictions", func() {
            // Configure allowed namespaces
            // Test enforcement
        })
    })
    
    Context("Security Context Enforcement", func() {
        It("should enforce privileged container restrictions", func() {
            // Test with PSP/PSA enabled
            // Verify security context applied
        })
        
        It("should enforce capability requirements", func() {
            // Test capability add/drop
            // Verify enforcement
        })
    })
    
    Context("Service Account Permissions", func() {
        It("should require correct RBAC permissions", func() {
            // Test with limited service account
            // Verify permission checks
        })
    })
})
```

### 🆕 Phase 6: Failure Scenarios

```go
var _ = Describe("Failure Scenarios", func() {
    Context("Pod Lifecycle Failures", func() {
        It("should handle target pod deletion during profiling", func() {
            // Start profiling
            // Delete target pod
            // Verify graceful handling
        })
        
        It("should handle target pod restart during profiling", func() {
            // Start profiling
            // Restart pod
            // Verify recovery
        })
    })
    
    Context("Controller Failures", func() {
        It("should recover from controller restart", func() {
            // Start profiling
            // Restart controller
            // Verify profiling continues
        })
    })
    
    Context("Resource Exhaustion", func() {
        It("should handle node resource exhaustion", func() {
            // Fill node resources
            // Attempt profiling
            // Verify error handling
        })
    })
})
```

## Test Organization

### Directory Structure
```
test/e2e-kind/
├── suite_test.go                    # Suite setup (adapted from e2e_suite_test.go)
├── utils.go                         # Reused from simple_utils.go
├── ephemeral_containers_test.go     # NEW: Phase 1
├── real_workloads_test.go           # NEW: Phase 2
├── storage_integration_test.go      # NEW: Phase 3
├── multi_node_test.go               # NEW: Phase 4
├── security_rbac_test.go            # NEW: Phase 5 (enhanced)
├── failure_scenarios_test.go        # NEW: Phase 6
└── fixtures/
    ├── workloads/                   # Sample applications
    ├── configs/                     # CoDriverTools
    └── scenarios/                   # Test scenarios
```

## Cluster Management Workflow

### 1. Setup Phase
```bash
#!/bin/bash
# test/e2e-kind/run-tests.sh

set -euo pipefail

# Get commit hash for cluster name
COMMIT_HASH="${GITHUB_SHA:-$(git rev-parse --short HEAD)}"
export CLUSTER_NAME="kubecodriver-e2e-${COMMIT_HASH}"

echo "🚀 Starting Kind E2E tests with cluster: $CLUSTER_NAME"

# Setup cluster
./test/e2e-kind/cluster/setup-cluster.sh

# Deploy KubeCoDriver components
kubectl apply -f config/crd/bases/
kubectl apply -f test/e2e-kind/manifests/

# Wait for deployments
kubectl wait --for=condition=available --timeout=300s \
    deployment/kubecodriver-controller -n kubecodriver-system
kubectl wait --for=condition=available --timeout=300s \
    deployment/kubecodriver-collector -n kubecodriver-system

# Run tests
go test -v -tags=e2e-kind ./test/e2e-kind/... \
    -ginkgo.v \
    -ginkgo.progress \
    -timeout=30m

# Cleanup
./test/e2e-kind/cluster/teardown-cluster.sh
```

### 2. GitHub Actions Integration
```yaml
# .github/workflows/e2e-kind.yml
name: E2E Tests (Kind)

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  e2e-kind:
    runs-on: ubuntu-latest
    timeout-minutes: 45
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25.3'
    
    - name: Install Kind
      run: |
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
        chmod +x ./kind
        sudo mv ./kind /usr/local/bin/kind
    
    - name: Run E2E Tests
      env:
        GITHUB_SHA: ${{ github.sha }}
      run: |
        ./test/e2e-kind/run-tests.sh
    
    - name: Upload Test Artifacts
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: e2e-kind-artifacts-${{ github.sha }}
        path: test-artifacts-*/
        retention-days: 7
```

### 3. Cleanup Strategy

**Automatic Cleanup:**
- After each test run (success or failure)
- On GitHub Actions workflow completion
- Timeout-based cleanup (stale clusters)

**Manual Cleanup:**
```bash
# Clean up all KubeCoDriver E2E clusters
kind get clusters | grep "^kubecodriver-e2e-" | xargs -I {} kind delete cluster --name {}
```

## Test Execution Strategy

### Local Development
```bash
# Quick test (single phase)
CLUSTER_NAME=kubecodriver-e2e-dev make test-e2e-kind-phase1

# Full test suite
CLUSTER_NAME=kubecodriver-e2e-dev make test-e2e-kind-all

# Keep cluster for debugging
CLUSTER_NAME=kubecodriver-e2e-dev KEEP_CLUSTER=true make test-e2e-kind-all
```

### CI/CD Pipeline
```bash
# Automatic cluster naming with commit hash
# Runs all phases
# Collects artifacts
# Cleans up automatically
```

## Success Criteria

### Phase 1 (Ephemeral Containers)
- ✅ 100% ephemeral container creation success
- ✅ Tool execution verified in real containers
- ✅ Data collection from ephemeral containers

### Phase 2 (Real Workloads)
- ✅ Profile 3+ different application types
- ✅ Verify profiling data quality
- ✅ Handle application lifecycle correctly

### Phase 3 (Storage)
- ✅ PVC mode fully functional
- ✅ Collector mode fully functional
- ✅ Data persistence verified

### Phase 4 (Multi-Node)
- ✅ Cross-node profiling works
- ✅ Network policies respected
- ✅ Node affinity handled

### Phase 5 (Security)
- ✅ RBAC enforcement verified
- ✅ Namespace isolation confirmed
- ✅ Security contexts applied

### Phase 6 (Failures)
- ✅ Graceful failure handling
- ✅ Recovery mechanisms work
- ✅ No resource leaks

## Timeline

- **Week 1**: Phase 1 (Ephemeral Containers) - Critical path
- **Week 2**: Phase 2 (Real Workloads) + Phase 3 (Storage)
- **Week 3**: Phase 4 (Multi-Node) + Phase 5 (Security)
- **Week 4**: Phase 6 (Failures) + CI/CD Integration
- **Week 5**: Documentation + Optimization

## Next Steps

1. ✅ Review and approve strategy
2. 🔄 Implement Phase 1 (ephemeral containers)
3. 🔄 Create test fixtures and sample workloads
4. 🔄 Integrate with CI/CD pipeline
5. 🔄 Document test scenarios and expected results
