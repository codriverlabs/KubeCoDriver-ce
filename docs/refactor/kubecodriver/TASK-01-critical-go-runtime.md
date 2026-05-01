# TASK-01: Critical Go Runtime Fixes

**Severity**: 🔴 CRITICAL — Causes runtime failures in deployed clusters
**Files**: 2

## Changes

### 1. `cmd/main.go` — Line 153
```
OLD: LeaderElectionID: "9410be53.toe.run",
NEW: LeaderElectionID: "9410be53.kubecodriver.codriverlabs.ai",
```

### 2. `pkg/collector/auth/k8s_token.go` — Line 41
```
OLD: "toe-collector", // Collector's ServiceAccount (with kustomize prefix)
NEW: "kubecodriver-collector", // Collector's ServiceAccount (with kustomize prefix)
```

## Impact
- LeaderElectionID mismatch would cause dual-leader scenarios during rolling upgrades from old to new versions.
- Wrong ServiceAccount name means token generation fails → all collector uploads fail with 401.

## Validation
```bash
grep -n 'toe' cmd/main.go pkg/collector/auth/k8s_token.go
# Expected: no output
```
