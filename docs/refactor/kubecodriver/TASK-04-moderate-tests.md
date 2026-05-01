# TASK-04: Test Infrastructure Fixes

**Severity**: 🟡 MODERATE — Tests reference old names
**Files**: 5

## Changes

### 1. `internal/controller/deletion_test.go` — Line 38
```
OLD: Finalizers: []string{"toe.run/finalizer"},
NEW: Finalizers: []string{"kubecodriver.codriverlabs.ai/finalizer"},
```

### 2. `test/e2e/simple_utils.go`
- Line 54: `GenerateName: "toe-simple-e2e-"` → `GenerateName: "kubecodriver-simple-e2e-"`
- Line 128: `Image: "ghcr.io/codriverlabs/toe-aperf:latest"` → `Image: "ghcr.io/codriverlabs/ce/kubecodriver-aperf:latest"`

### 3. `test/e2e-kind/utils.go`
- Line 24: `GenerateName: "toe-kind-e2e-"` → `GenerateName: "kubecodriver-kind-e2e-"`
- Line 94: `Image: "ghcr.io/codriverlabs/toe-aperf:latest"` → `Image: "ghcr.io/codriverlabs/ce/kubecodriver-aperf:latest"`

### 4. `pkg/collector/auth/k8s_token_test.go` — Lines 41, 75, 139
```
OLD: Name: "toe-collector",
NEW: Name: "kubecodriver-collector",
```

### 5. Verify no other test files reference `toe`
```bash
grep -rn '\btoe\b' internal/controller/*_test.go test/ pkg/**/*_test.go --include='*.go'
```

## Validation
```bash
grep -rn 'toe' test/ internal/controller/*_test.go pkg/collector/auth/*_test.go --include='*.go'
# Expected: no output
```
