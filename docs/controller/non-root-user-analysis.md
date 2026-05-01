# Non-Root User and Security Context Analysis

## Current Situation

### What the Controller Does Now

The controller creates ephemeral containers with security context from `CoDriverTool`:

```go
ec := &corev1.EphemeralContainer{
    EphemeralContainerCommon: corev1.EphemeralContainerCommon{
        Name:            containerName,
        Image:           toolConfig.Spec.Image,
        SecurityContext: r.buildSecurityContext(toolConfig.Spec.SecurityContext),
    },
}
```

### Current SecuritySpec Support

The `SecuritySpec` currently supports:
- `AllowPrivileged` - Run as privileged container
- `AllowHostPID` - Access host PID namespace (not currently used in buildSecurityContext)
- `Capabilities.Add` - Add Linux capabilities
- `Capabilities.Drop` - Drop Linux capabilities

### What's Missing

The controller **does NOT** currently handle:
- `RunAsUser` - Specify user ID
- `RunAsGroup` - Specify group ID
- `RunAsNonRoot` - Enforce non-root execution
- `FSGroup` - Set filesystem group ownership

## The Problem

When a target pod runs with:
```yaml
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  runAsNonRoot: true
```

And the ephemeral container is created **without** matching security context:

### Kubernetes Behavior

According to Kubernetes documentation on ephemeral containers:

1. **Ephemeral containers inherit some pod-level security context** but NOT container-level security context
2. **Pod Security Context** (PodSecurityContext) applies to ALL containers including ephemeral
3. **Container Security Context** overrides pod-level settings for that specific container

### What Happens in Your Scenario

**Scenario**: Target pod runs as user 1000, group 1000

**Current Behavior**:
- ✅ Pod-level `securityContext.fsGroup: 1000` → Ephemeral container WILL inherit this
- ✅ Pod-level `securityContext.runAsUser: 1000` → Ephemeral container WILL inherit this
- ✅ Pod-level `securityContext.runAsGroup: 1000` → Ephemeral container WILL inherit this
- ❌ Container-level security context → Ephemeral container will NOT inherit

**Result**: If the target pod has **pod-level** security context, the ephemeral container will run with the same user/group. If it's **container-level** only, the ephemeral container may run as root.

## Potential Issues

### 1. Permission Issues
If the ephemeral container runs as a different user than the target container:
- Cannot access target container's files
- Cannot attach to target container's processes (for tools like `strace`, `perf`)
- Cannot read `/proc/<pid>` information

### 2. Security Policy Violations
If Pod Security Standards (PSS) or Pod Security Policies (PSP) enforce non-root:
- Ephemeral container creation may be **rejected** if it tries to run as root
- Cluster admission controllers may block the operation

### 3. Profiling Tool Requirements
Many profiling tools need:
- **Same user context** to attach to processes
- **Elevated capabilities** (SYS_PTRACE, SYS_ADMIN) to profile
- **Access to /proc filesystem** with correct permissions

## Recommended Solution

### Option 1: Inherit Target Container Security Context (Recommended)

Modify the controller to **copy security context from the target container**:

```go
func (r *CoDriverJobReconciler) createEphemeralContainerForPod(ctx context.Context, powerTool *toev1alpha1.CoDriverJob, toolConfig *toev1alpha1.CoDriverTool, pod corev1.Pod, containerName string) error {
    // ... existing code ...
    
    // Get target container's security context
    var targetSecurityContext *corev1.SecurityContext
    if len(pod.Spec.Containers) > 0 {
        targetSecurityContext = pod.Spec.Containers[0].SecurityContext
    }
    
    // Build security context - merge target + toolConfig
    securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)
    
    // Inherit user/group from target container if not specified in toolConfig
    if targetSecurityContext != nil {
        if securityContext.RunAsUser == nil && targetSecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = targetSecurityContext.RunAsUser
        }
        if securityContext.RunAsGroup == nil && targetSecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = targetSecurityContext.RunAsGroup
        }
        if securityContext.RunAsNonRoot == nil && targetSecurityContext.RunAsNonRoot != nil {
            securityContext.RunAsNonRoot = targetSecurityContext.RunAsNonRoot
        }
    }
    
    ec := &corev1.EphemeralContainer{
        EphemeralContainerCommon: corev1.EphemeralContainerCommon{
            Name:            containerName,
            Image:           toolConfig.Spec.Image,
            SecurityContext: securityContext,
        },
    }
    // ... rest of code ...
}
```

### Option 2: Extend SecuritySpec

Add user/group fields to `SecuritySpec`:

```go
type SecuritySpec struct {
    AllowPrivileged *bool         `json:"allowPrivileged,omitempty"`
    AllowHostPID    *bool         `json:"allowHostPID,omitempty"`
    Capabilities    *Capabilities `json:"capabilities,omitempty"`
    RunAsUser       *int64        `json:"runAsUser,omitempty"`       // NEW
    RunAsGroup      *int64        `json:"runAsGroup,omitempty"`      // NEW
    RunAsNonRoot    *bool         `json:"runAsNonRoot,omitempty"`    // NEW
}
```

Then update `buildSecurityContext`:

```go
func (r *CoDriverJobReconciler) buildSecurityContext(securitySpec toev1alpha1.SecuritySpec) *corev1.SecurityContext {
    securityContext := &corev1.SecurityContext{}

    if securitySpec.AllowPrivileged != nil {
        securityContext.Privileged = securitySpec.AllowPrivileged
    }

    if securitySpec.RunAsUser != nil {
        securityContext.RunAsUser = securitySpec.RunAsUser
    }

    if securitySpec.RunAsGroup != nil {
        securityContext.RunAsGroup = securitySpec.RunAsGroup
    }

    if securitySpec.RunAsNonRoot != nil {
        securityContext.RunAsNonRoot = securitySpec.RunAsNonRoot
    }

    // ... existing capabilities code ...

    return securityContext
}
```

### Option 3: Hybrid Approach (Best)

Combine both options:
1. Extend `SecuritySpec` to support user/group settings
2. Auto-inherit from target container if not explicitly set in `CoDriverTool`
3. Allow override via `CoDriverTool` when needed

## Testing Recommendations

Create test scenarios:
1. Target pod with pod-level security context
2. Target pod with container-level security context
3. Target pod with PSS restricted enforcement
4. Profiling tools that need process attachment (strace, perf)

## Conclusion

**Current Status**: ⚠️ **Partially Working**
- Works if target pod has **pod-level** security context
- May fail if target pod has **container-level** security context only
- May fail with strict Pod Security Standards

**Recommendation**: Implement **Option 3 (Hybrid Approach)** to ensure compatibility with all scenarios.
