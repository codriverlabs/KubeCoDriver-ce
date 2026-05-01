# runAsRoot Feature Implementation Summary

## Overview
Implemented `runAsRoot` field in SecuritySpec to allow tools requiring root access (like aperf) to run with uid=0 while preserving group context from target containers for file compatibility.

## Changes Made

### 1. CRD Update (api/v1alpha1/common_types.go)
Added `RunAsRoot *bool` field to SecuritySpec:
```go
type SecuritySpec struct {
    AllowPrivileged *bool         `json:"allowPrivileged,omitempty"`
    AllowHostPID    *bool         `json:"allowHostPID,omitempty"`
    RunAsRoot       *bool         `json:"runAsRoot,omitempty"`  // NEW
    Capabilities    *Capabilities `json:"capabilities,omitempty"`
}
```

### 2. Controller Logic (internal/controller/powertool_controller.go)
Modified `createEphemeralContainerForPod()` to handle runAsRoot:

**When runAsRoot=true:**
- Sets `runAsUser: 0` (root)
- Clears `runAsNonRoot` 
- Inherits `runAsGroup` from target container (or pod if container doesn't specify)
- Preserves existing capabilities from SecuritySpec

**When runAsRoot=false or not set:**
- Normal inheritance: pod-level → container-level override
- Existing behavior unchanged

### 3. Unit Tests (internal/controller/runasroot_test.go)
Created 5 test cases:
1. runAsRoot enabled with container group
2. runAsRoot enabled with pod group  
3. runAsRoot enabled, container group overrides pod group
4. runAsRoot disabled, normal inheritance
5. runAsRoot not set, normal inheritance

## Testing

### Manual Testing with kubectl debug
Validated the approach using kubectl debug with custom profiles:

```bash
# Custom profile with runAsUser: 0
kubectl debug multi-container-test \
  --profile=general \
  --custom=examples/debug-root-profile.yaml \
  -- sleep 300
```

**Results:**
- Container runs as `uid=0(root) gid=1001` ✓
- aperf works without errors ✓
- No `privileged: true` flag ✓
- Only necessary capabilities (SYS_PTRACE, PERFMON, SYS_ADMIN) ✓

### Unit Test Results
```
ok  	github.com/codriverlabs/KubeCoDriver/internal/controller	6.388s	coverage: 72.4% of statements
```

All tests passing, coverage maintained at 72.4%.

## Usage Example

### CoDriverTool with runAsRoot
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: aperf
spec:
  name: "aperf"
  image: "localhost:32000/codriverlabs/ce/kubecodriver-aperf:v1.0.47"
  securityContext:
    runAsRoot: true  # Enable root access
    capabilities:
      add:
        - SYS_PTRACE
        - PERFMON
        - SYS_ADMIN
```

### Resulting Ephemeral Container
For a pod running as user 1001:1001, the ephemeral container will have:
```yaml
securityContext:
  runAsUser: 0      # Root for kernel access
  runAsGroup: 1001  # Inherited from target for file compatibility
  capabilities:
    add:
      - SYS_PTRACE
      - PERFMON
      - SYS_ADMIN
```

## Benefits

1. **Security**: More secure than `privileged: true` - only grants necessary capabilities
2. **Compatibility**: Preserves group context for file system access
3. **Flexibility**: Allows tools requiring root while maintaining file ownership compatibility
4. **Backward Compatible**: Existing configurations unchanged (runAsRoot defaults to nil/false)

## Files Modified
- `api/v1alpha1/common_types.go` - Added RunAsRoot field
- `internal/controller/powertool_controller.go` - Implemented runAsRoot logic
- `internal/controller/runasroot_test.go` - Added unit tests
- `config/crd/bases/kubecodriver.codriverlabs.ai_powertoolconfigs.yaml` - Generated CRD

## Next Steps
1. Update CoDriverTool examples to use runAsRoot for aperf
2. Document the feature in user-facing documentation
3. Test with real workloads in cluster
