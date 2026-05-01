# aperf Tool Root Requirement Analysis

**Date**: 2025-11-01  
**Status**: ✅ Root Cause Identified

## Issue

The aperf profiling tool fails with "Permission denied (os error 13)" when running in ephemeral containers, even with correct security context inheritance and capabilities.

## Root Cause

**aperf requires root access** to access kernel perf events and tracing interfaces.

## Testing Evidence

### Test 1: kubectl debug with default profile
```bash
kubectl debug busybox-nonroot -n kubecodriver-test \
  --image=localhost:32000/codriverlabs/ce/kubecodriver-aperf:v1.0.47 \
  --target=main-container \
  -- aperf record --period=5 --run-name=test --profile
```
**Result**: `Error: Permission denied (os error 13)`  
**User**: `uid=1001` (inherited from pod)

### Test 2: kubectl debug with sysadmin profile
```bash
kubectl debug busybox-nonroot -n kubecodriver-test \
  --profile=sysadmin \
  --image=localhost:32000/codriverlabs/ce/kubecodriver-aperf:v1.0.47 \
  --target=main-container \
  -- aperf record --period=5 --run-name=test --profile
```
**Result**: `Error: Permission denied (os error 13)`  
**User**: `uid=1001` (inherited from pod)  
**Security Context**: `privileged: true` (but still runs as user 1001)  
**Warning**: "Non-root user is configured for the entire target Pod, and some capabilities granted by debug profile may not work. Please consider using '--custom' with a custom profile that specifies 'securityContext.runAsUser: 0'."

## Kubernetes Debug Profiles

Available profiles:
- `legacy` (default, deprecated)
- `general` - Basic debugging
- `baseline` - Minimal permissions
- `netadmin` - Network debugging
- `restricted` - Most restricted
- `sysadmin` - Sets `privileged: true`

**Key Finding**: Even `sysadmin` profile with `privileged: true` doesn't override pod-level `runAsUser`. The pod-level security context takes precedence.

## Current CoDriverTool

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
    allowPrivileged: false  # ← Prevents privileged mode
    capabilities:
      add:
        - SYS_PTRACE
        - PERFMON
        - SYS_ADMIN
      drop:
        - ALL
```

**Problem**: This configuration doesn't allow running as root, which aperf requires.

## Solution: Add runAsRoot Field

### Proposed API Extension

Add `runAsRoot` field to `SecuritySpec`:

```go
type SecuritySpec struct {
    AllowPrivileged *bool         `json:"allowPrivileged,omitempty"`
    AllowHostPID    *bool         `json:"allowHostPID,omitempty"`
    Capabilities    *Capabilities `json:"capabilities,omitempty"`
    RunAsRoot       *bool         `json:"runAsRoot,omitempty"`  // NEW
}
```

### Implementation Logic

```go
func (r *CoDriverJobReconciler) createEphemeralContainerForPod(...) error {
    // Build base security context from toolConfig
    securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

    // If tool requires root, set runAsUser to 0
    if toolConfig.Spec.SecurityContext.RunAsRoot != nil && *toolConfig.Spec.SecurityContext.RunAsRoot {
        rootUser := int64(0)
        securityContext.RunAsUser = &rootUser
        logger.Info("Tool requires root access, setting runAsUser to 0")
    } else {
        // Inherit from pod-level security context
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

        // Override with target container's security context if available
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

    // Note: fsGroup is pod-level only, always inherited from pod
    // runAsGroup should still be inherited from target pod/container even when runAsRoot=true
    
    return nil
}
```

### Updated CoDriverTool for aperf

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
    runAsRoot: true  # NEW - aperf requires root access
    allowPrivileged: false
    capabilities:
      add:
        - SYS_PTRACE
        - PERFMON
        - SYS_ADMIN
      drop:
        - ALL
```

## Inheritance Behavior with runAsRoot

### When runAsRoot: true
- `runAsUser`: Set to `0` (root) - **OVERRIDE**
- `runAsGroup`: Inherited from target pod/container - **KEEP**
- `runAsNonRoot`: Set to `false` or omitted - **OVERRIDE**
- `fsGroup`: Inherited from pod-level (if set) - **KEEP**

### When runAsRoot: false or not set
- `runAsUser`: Inherited from target pod/container - **KEEP**
- `runAsGroup`: Inherited from target pod/container - **KEEP**
- `runAsNonRoot`: Inherited from target pod/container - **KEEP**
- `fsGroup`: Inherited from pod-level (if set) - **KEEP**

## Rationale

### Why Override Only runAsUser?

1. **Root Access for Kernel Interfaces**: Tools like aperf need `uid=0` to access `/sys/kernel/debug/tracing`, perf events, and other kernel interfaces.

2. **Preserve Group Context**: Keeping the target's `runAsGroup` and `fsGroup` ensures:
   - File permissions remain compatible with target container
   - Shared volumes remain accessible
   - Group-based access controls work correctly

3. **Security Principle**: Only escalate the minimum required privilege (user ID), maintain other security boundaries.

### Example Scenario

**Target Pod**:
```yaml
securityContext:
  runAsUser: 1001
  runAsGroup: 1001
  fsGroup: 1001
```

**Ephemeral Container with runAsRoot: true**:
```yaml
securityContext:
  runAsUser: 0      # ← Set to root (OVERRIDE)
  runAsGroup: 1001  # ← Inherited from pod (KEEP)
  fsGroup: 1001     # ← Inherited from pod (KEEP)
  capabilities:
    add: [SYS_PTRACE, PERFMON, SYS_ADMIN]
```

**Result**: `uid=0(root) gid=1001 groups=1001`
- Root access for kernel interfaces ✅
- Group compatibility with target ✅
- File access to shared volumes ✅

## Security Considerations

### Risks
- Running as root increases attack surface
- Compromised profiling tool could affect host

### Mitigations
1. **Explicit Opt-In**: `runAsRoot` must be explicitly set in CoDriverTool
2. **RBAC Controls**: Limit who can create CoDriverTools with `runAsRoot: true`
3. **Namespace Restrictions**: Use `allowedNamespaces` to limit where root tools can run
4. **Audit Logging**: Log all CoDriverJob executions with root access
5. **Capabilities Still Dropped**: Even as root, only specific capabilities are added

### Recommended RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: powertool-root-tools
  namespace: kubecodriver-system
rules:
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs"]
  verbs: ["create", "update"]
  resourceNames: ["aperf-config", "perf-config"]  # Explicit allow-list
```

## Tools Requiring Root

Based on analysis, these tools likely need `runAsRoot: true`:
- ✅ **aperf** - Requires root for perf events
- ✅ **perf** - Requires root for kernel tracing
- ✅ **bpftrace** - Requires root for eBPF
- ✅ **systemtap** - Requires root for kernel probes
- ❓ **strace** - May work with SYS_PTRACE capability only
- ❓ **tcpdump** - May work with NET_ADMIN capability only

## Implementation Checklist

- [ ] Add `RunAsRoot *bool` to `SecuritySpec` in `api/v1alpha1/common_types.go`
- [ ] Update `buildSecurityContext()` to handle `runAsRoot`
- [ ] Update `createEphemeralContainerForPod()` with conditional logic
- [ ] Add unit tests for `runAsRoot` behavior
- [ ] Update aperf CoDriverTool with `runAsRoot: true`
- [ ] Add security documentation
- [ ] Add RBAC examples
- [ ] Test aperf with updated configuration

## Related Documents

- [Security Context Inheritance](../non-root-security-context-fix.md)
- [Container Selection](../container-selection-fix.md)
- [Test Results](test-results-final.md)

## Conclusion

The aperf tool failure is **not a bug** in our security context inheritance implementation. It's a **tool requirement** that needs explicit configuration via a new `runAsRoot` field in CoDriverTool.

Our implementation correctly inherits security context from target pods/containers. The solution is to extend the API to allow tools that require root access to explicitly request it, while still maintaining group and filesystem context from the target.
