# TASK-02: Critical Kustomize & Config Manifest Fixes

**Severity**: 🔴 CRITICAL — All deployed K8s resources get wrong names/labels
**Files**: 10

## Changes

### 1. `config/default/kustomization.yaml` — Line 9
```
OLD: namePrefix: toe-
NEW: namePrefix: kubecodriver-
```

### 2. `config/manager/manager.yaml` — Lines 8, 14, 22
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 3. `config/rbac/role_binding.yaml` — Line 5
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 4. `config/rbac/leader_election_role_binding.yaml` — Line 5
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 5. `config/rbac/service_account.yaml` — Line 5
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 6. `config/rbac/leader_election_role.yaml` — Line 6
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 7. `config/rbac/kustomization.yaml` — Line 24
```
OLD: # not used by the toe itself. You can comment the following lines
NEW: # not used by the operator itself. You can comment the following lines
```

### 8. `config/prometheus/monitor.yaml` — Lines 7, 27
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 9. `config/default/metrics_service.yaml` — Lines 6, 18
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

### 10. `config/network-policy/allow-metrics-traffic.yaml` — Lines 8, 16
```
OLD: app.kubernetes.io/name: toe
NEW: app.kubernetes.io/name: kubecodriver
```

## Validation
```bash
grep -rn 'toe' config/ --include='*.yaml' --include='*.yml'
# Expected: no output
```
