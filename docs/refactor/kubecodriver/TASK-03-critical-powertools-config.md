# TASK-03: Critical Power-Tools Config Fixes

**Severity**: 🔴 CRITICAL — Wrong container images pulled at runtime
**Files**: 3

## Changes

### 1. `power-tools/aperf/config/powertoolconfig-aperf.yaml`
- Line 8: `image: ghcr.io/codriverlabs/ce/toe-aperf:1.1.0` → `image: ghcr.io/codriverlabs/ce/kubecodriver-aperf:v1.1.0`
- Line 20: `- "toe-test"` → `- "kubecodriver-test"`

### 2. `power-tools/tcpdump/config/powertoolconfig-tcpdump.yaml`
- Line 9: `image: ghcr.io/codriverlabs/ce/toe-tcpdump:v1.1.0` → `image: ghcr.io/codriverlabs/ce/kubecodriver-tcpdump:v1.1.0`

### 3. `power-tools/chaos/config/powertoolconfig-chaos.yaml`
- Line 9: `image: ghcr.io/codriverlabs/ce/toe-chaos:1.1.0` → `image: ghcr.io/codriverlabs/ce/kubecodriver-chaos:v1.1.0`

## Validation
```bash
grep -rn 'toe' power-tools/ --include='*.yaml'
# Expected: no output
```
