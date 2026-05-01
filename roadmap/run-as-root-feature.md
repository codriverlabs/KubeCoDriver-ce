# Run As Root Feature - Implementation Guide

## Status: 🔴 Not Started

**Priority**: High  
**Estimated Effort**: 1 hour  
**Target Version**: v1.0.52

## Problem Statement

Some profiling tools (aperf, perf, bpftrace) require root access to access kernel interfaces and perf events. Currently, ephemeral containers inherit the target pod's user context, which prevents these tools from working when target pods run as non-root.

**Evidence**: See `roadmap/use-cases/aperf-root-requirement.md`

## Proposed Solution

Add `runAsRoot` field to `SecuritySpec` to allow tools to explicitly request root access while maintaining group and filesystem context from target pod.

## Implementation

### Phase 1: API Extension (15 min)

**File**: `api/v1alpha1/common_types.go`

```go
type SecuritySpec struct {
    AllowPrivileged *bool         `json:"allowPrivileged,omitempty"`
    AllowHostPID    *bool         `json:"allowHostPID,omitempty"`
    Capabilities    *Capabilities `json:"capabilities,omitempty"`
    RunAsRoot       *bool         `json:"runAsRoot,omitempty"`  // NEW
}
```

**Commands**:
```bash
make generate
make manifests
```

### Phase 2: Update buildSecurityContext (10 min)

**File**: `internal/controller/powertool_controller.go`

```go
func (r *CoDriverJobReconciler) buildSecurityContext(securitySpec kubecodriverv1alpha1.SecuritySpec) *corev1.SecurityContext {
    securityContext := &corev1.SecurityContext{}

    if securitySpec.AllowPrivileged != nil {
        securityContext.Privileged = securitySpec.AllowPrivileged
    }

    // NEW: Handle runAsRoot
    if securitySpec.RunAsRoot != nil && *securitySpec.RunAsRoot {
        rootUser := int64(0)
        securityContext.RunAsUser = &rootUser
        nonRoot := false
        securityContext.RunAsNonRoot = &nonRoot
    }

    if securitySpec.Capabilities != nil {
        // ... existing code ...
    }

    return securityContext
}
```

### Phase 3: Update Inheritance Logic (15 min)

**File**: `internal/controller/powertool_controller.go`

Modify `createEphemeralContainerForPod()`:

```go
// Build base security context from toolConfig
securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

// Check if tool requires root
requiresRoot := toolConfig.Spec.SecurityContext.RunAsRoot != nil && *toolConfig.Spec.SecurityContext.RunAsRoot

if requiresRoot {
    // Tool requires root - set runAsUser to 0
    rootUser := int64(0)
    securityContext.RunAsUser = &rootUser
    nonRoot := false
    securityContext.RunAsNonRoot = &nonRoot
    logger.Info("Tool requires root access, setting runAsUser to 0")
    
    // Still inherit group and fsGroup from target
    if pod.Spec.SecurityContext != nil {
        if pod.Spec.SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
            logger.Info("Inherited runAsGroup from pod (keeping with root user)", "group", *pod.Spec.SecurityContext.RunAsGroup)
        }
    }
    
    // Override with target container's group if available
    if targetContainer != nil && targetContainer.SecurityContext != nil {
        if targetContainer.SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
            logger.Info("Inherited runAsGroup from target container (keeping with root user)", 
                "container", targetContainer.Name, 
                "group", *targetContainer.SecurityContext.RunAsGroup)
        }
    }
} else {
    // Normal inheritance (existing code)
    if pod.Spec.SecurityContext != nil {
        if pod.Spec.SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = pod.Spec.SecurityContext.RunAsUser
        }
        if pod.Spec.SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
        }
        if pod.Spec.SecurityContext.RunAsNonRoot != nil {
            securityContext.RunAsNonRoot = pod.Spec.SecurityContext.RunAsNonRoot
        }
    }

    // Override with target container's security context
    if targetContainer != nil && targetContainer.SecurityContext != nil {
        if targetContainer.SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = targetContainer.SecurityContext.RunAsUser
        }
        if targetContainer.SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
        }
        if targetContainer.SecurityContext.RunAsNonRoot != nil {
            securityContext.RunAsNonRoot = targetContainer.SecurityContext.RunAsNonRoot
        }
    }
}
```

### Phase 4: Update CoDriverTool (5 min)

**File**: `examples/aperf/powertoolconfig-aperf.yaml`

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: aperf-config
  namespace: kubecodriver-system
spec:
  name: "aperf"
  image: "localhost:32000/codriverlabs/ce/kubecodriver-aperf:v1.0.47"
  securityContext:
    runAsRoot: true  # NEW - aperf requires root
    allowPrivileged: false
    capabilities:
      add:
        - SYS_PTRACE
        - PERFMON
        - SYS_ADMIN
      drop:
        - ALL
```

### Phase 5: Add Unit Tests (15 min)

**File**: `internal/controller/run_as_root_test.go` (new)

```go
func TestBuildSecurityContext_RunAsRoot(t *testing.T) {
    r := &CoDriverJobReconciler{}
    
    runAsRoot := true
    spec := kubecodriverv1alpha1.SecuritySpec{
        RunAsRoot: &runAsRoot,
    }
    
    ctx := r.buildSecurityContext(spec)
    
    assert.NotNil(t, ctx.RunAsUser)
    assert.Equal(t, int64(0), *ctx.RunAsUser)
    assert.NotNil(t, ctx.RunAsNonRoot)
    assert.False(t, *ctx.RunAsNonRoot)
}

func TestSecurityContextInheritance_WithRunAsRoot(t *testing.T) {
    // Test that runAsRoot overrides user but keeps group
    // ... implementation ...
}
```

## Inheritance Rules Summary

| Field | runAsRoot: true | runAsRoot: false/unset |
|-------|----------------|------------------------|
| `runAsUser` | **0 (root)** - OVERRIDE | Inherited from target - KEEP |
| `runAsGroup` | Inherited from target - **KEEP** | Inherited from target - KEEP |
| `runAsNonRoot` | **false** - OVERRIDE | Inherited from target - KEEP |
| `fsGroup` | Inherited from pod - **KEEP** | Inherited from pod - KEEP |

**Key Principle**: When `runAsRoot: true`, only override user ID to 0, keep all other context from target.

## Security Implications

### Why This is Safe

1. **Explicit Opt-In**: Tools must explicitly declare `runAsRoot: true`
2. **Group Preservation**: Maintains group context for file access
3. **Capabilities Control**: Still uses capability-based restrictions
4. **RBAC Enforcement**: CoDriverTool creation can be restricted
5. **Namespace Restrictions**: `allowedNamespaces` still applies
6. **Audit Trail**: All root executions logged

### Why This is Necessary

1. **Kernel Access**: Many profiling tools need kernel interface access
2. **Perf Events**: Require root or very specific capabilities
3. **eBPF Programs**: Need root to load and attach
4. **System Tracing**: Requires elevated privileges

## Implementation Checklist

- [ ] Add `RunAsRoot` field to `SecuritySpec`
- [ ] Run `make generate` and `make manifests`
- [ ] Update `buildSecurityContext()` function
- [ ] Update `createEphemeralContainerForPod()` with conditional logic
- [ ] Add unit tests for `runAsRoot` behavior
- [ ] Update aperf CoDriverTool
- [ ] Test aperf with non-root pod
- [ ] Add security documentation
- [ ] Update examples

## Success Criteria

- [ ] aperf works with non-root target pods
- [ ] Ephemeral container runs as root (uid=0)
- [ ] Group context preserved from target
- [ ] Unit tests pass
- [ ] No regression in non-root tools
- [ ] Documentation updated

## Timeline

- **Phase 1**: 15 minutes (API extension)
- **Phase 2**: 10 minutes (buildSecurityContext)
- **Phase 3**: 15 minutes (inheritance logic)
- **Phase 4**: 5 minutes (update config)
- **Phase 5**: 15 minutes (unit tests)

**Total**: ~1 hour

## Related Documents

- [aperf Root Requirement Analysis](use-cases/aperf-root-requirement.md)
- [Security Context Inheritance](non-root-security-context-fix.md)
- [Container Selection](container-selection-fix.md)
