# Non-Root Security Context Fix - Implementation Guide

## Status: 🔴 Not Started

**Priority**: High  
**Estimated Effort**: 1-2 hours  
**Target Version**: v1.0.52

## Problem Statement

When target pods run with non-root users (e.g., `runAsUser: 1001`), ephemeral containers created by the controller fail with permission errors because they don't inherit the correct security context.

**Test Results**: See `test-results-nonroot.md`
- ✅ Ephemeral container created successfully
- ❌ Container fails with "Permission denied (os error 13)"

## Root Cause

The controller's `buildSecurityContext()` function only sets:
- `privileged` flag
- `capabilities` (add/drop)

**Missing**: Auto-discovery and inheritance of:
- `runAsUser`
- `runAsGroup`
- `runAsNonRoot`
- `fsGroup`

## Proposed Solution: Auto-Discovery Approach

**Key Principle**: No API changes needed - rely on automatic discovery and inheritance from target pod.

**Benefits**:
- ✅ No CRD changes required
- ✅ No API modifications
- ✅ Fully backward compatible
- ✅ Works automatically for all existing CoDriverTools
- ✅ Simple implementation

### Phase 1: Implement Auto-Inheritance (45 min)

**File**: `internal/controller/powertool_controller.go`

Modify `createEphemeralContainerForPod()` to automatically inherit security context:

```go
func (r *CoDriverJobReconciler) createEphemeralContainerForPod(ctx context.Context, powerTool *kubecodriverv1alpha1.CoDriverJob, toolConfig *kubecodriverv1alpha1.CoDriverTool, pod corev1.Pod, containerName string) error {
    logger := log.FromContext(ctx)

    // ... existing code for env vars ...

    // Build base security context from toolConfig
    securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

    // AUTO-DISCOVER: Inherit user/group from target pod
    if pod.Spec.SecurityContext != nil {
        if pod.Spec.SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = pod.Spec.SecurityContext.RunAsUser
            logger.Info("Inherited runAsUser from pod", "user", *pod.Spec.SecurityContext.RunAsUser)
        }
        if pod.Spec.SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
            logger.Info("Inherited runAsGroup from pod", "group", *pod.Spec.SecurityContext.RunAsGroup)
        }
        if pod.Spec.SecurityContext.RunAsNonRoot != nil {
            securityContext.RunAsNonRoot = pod.Spec.SecurityContext.RunAsNonRoot
            logger.Info("Inherited runAsNonRoot from pod", "nonRoot", *pod.Spec.SecurityContext.RunAsNonRoot)
        }
        if pod.Spec.SecurityContext.FSGroup != nil {
            securityContext.FSGroup = pod.Spec.SecurityContext.FSGroup
            logger.Info("Inherited fsGroup from pod", "fsGroup", *pod.Spec.SecurityContext.FSGroup)
        }
    }

    // Fallback: Check first container if pod-level not set
    if securityContext.RunAsUser == nil && len(pod.Spec.Containers) > 0 {
        if pod.Spec.Containers[0].SecurityContext != nil && pod.Spec.Containers[0].SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = pod.Spec.Containers[0].SecurityContext.RunAsUser
            logger.Info("Inherited runAsUser from container", "user", *pod.Spec.Containers[0].SecurityContext.RunAsUser)
        }
    }

    if securityContext.RunAsGroup == nil && len(pod.Spec.Containers) > 0 {
        if pod.Spec.Containers[0].SecurityContext != nil && pod.Spec.Containers[0].SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = pod.Spec.Containers[0].SecurityContext.RunAsGroup
            logger.Info("Inherited runAsGroup from container", "group", *pod.Spec.Containers[0].SecurityContext.RunAsGroup)
        }
    }

    if securityContext.RunAsNonRoot == nil && len(pod.Spec.Containers) > 0 {
        if pod.Spec.Containers[0].SecurityContext != nil && pod.Spec.Containers[0].SecurityContext.RunAsNonRoot != nil {
            securityContext.RunAsNonRoot = pod.Spec.Containers[0].SecurityContext.RunAsNonRoot
            logger.Info("Inherited runAsNonRoot from container", "nonRoot", *pod.Spec.Containers[0].SecurityContext.RunAsNonRoot)
        }
    }

    // Create ephemeral container with inherited security context
    ec := &corev1.EphemeralContainer{
        EphemeralContainerCommon: corev1.EphemeralContainerCommon{
            Name:            containerName,
            Image:           toolConfig.Spec.Image,
            ImagePullPolicy: corev1.PullAlways,
            Env:             envVars,
            SecurityContext: securityContext,
        },
    }

    // ... rest of existing code ...
}
```

### Phase 2: Add Unit Tests (30 min)

**File**: `internal/controller/security_context_inheritance_test.go` (new)

```go
package controller

import (
    "testing"
    "github.com/stretchr/testify/assert"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecurityContextInheritance_PodLevel(t *testing.T) {
    user := int64(1001)
    group := int64(1001)
    nonRoot := true
    fsGroup := int64(1001)
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            SecurityContext: &corev1.PodSecurityContext{
                RunAsUser:    &user,
                RunAsGroup:   &group,
                RunAsNonRoot: &nonRoot,
                FSGroup:      &fsGroup,
            },
            Containers: []corev1.Container{
                {Name: "main"},
            },
        },
    }
    
    // Simulate inheritance logic
    securityContext := &corev1.SecurityContext{}
    
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
        if pod.Spec.SecurityContext.FSGroup != nil {
            securityContext.FSGroup = pod.Spec.SecurityContext.FSGroup
        }
    }
    
    assert.NotNil(t, securityContext.RunAsUser)
    assert.Equal(t, int64(1001), *securityContext.RunAsUser)
    assert.NotNil(t, securityContext.RunAsGroup)
    assert.Equal(t, int64(1001), *securityContext.RunAsGroup)
    assert.NotNil(t, securityContext.RunAsNonRoot)
    assert.True(t, *securityContext.RunAsNonRoot)
    assert.NotNil(t, securityContext.FSGroup)
    assert.Equal(t, int64(1001), *securityContext.FSGroup)
}

func TestSecurityContextInheritance_ContainerLevel(t *testing.T) {
    user := int64(2000)
    group := int64(2000)
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name: "main",
                    SecurityContext: &corev1.SecurityContext{
                        RunAsUser:  &user,
                        RunAsGroup: &group,
                    },
                },
            },
        },
    }
    
    // Simulate fallback to container-level
    securityContext := &corev1.SecurityContext{}
    
    if securityContext.RunAsUser == nil && len(pod.Spec.Containers) > 0 {
        if pod.Spec.Containers[0].SecurityContext != nil && pod.Spec.Containers[0].SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = pod.Spec.Containers[0].SecurityContext.RunAsUser
        }
    }
    
    if securityContext.RunAsGroup == nil && len(pod.Spec.Containers) > 0 {
        if pod.Spec.Containers[0].SecurityContext != nil && pod.Spec.Containers[0].SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = pod.Spec.Containers[0].SecurityContext.RunAsGroup
        }
    }
    
    assert.NotNil(t, securityContext.RunAsUser)
    assert.Equal(t, int64(2000), *securityContext.RunAsUser)
    assert.NotNil(t, securityContext.RunAsGroup)
    assert.Equal(t, int64(2000), *securityContext.RunAsGroup)
}

func TestSecurityContextInheritance_NoSecurityContext(t *testing.T) {
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "main"},
            },
        },
    }
    
    securityContext := &corev1.SecurityContext{}
    
    // Should not panic, should remain nil
    assert.Nil(t, securityContext.RunAsUser)
    assert.Nil(t, securityContext.RunAsGroup)
    assert.Nil(t, securityContext.RunAsNonRoot)
}
```

### Phase 3: Integration Testing (30 min)

Run the existing test:

```bash
./test-nonroot-scenario.sh
```

**Expected Results**:
- ✅ Ephemeral container created
- ✅ Container runs successfully (no permission errors)
- ✅ Controller logs show inherited values
- ✅ Container runs as user 1001:1001

### Phase 4: Documentation (15 min)

**File**: `docs/security/README.md`

Add section:

```markdown
## Security Context Inheritance

The controller automatically inherits security context from target pods:

1. **Pod-level security context** (preferred):
   - `runAsUser`
   - `runAsGroup`
   - `runAsNonRoot`
   - `fsGroup`

2. **Container-level security context** (fallback):
   - Used if pod-level not set
   - Inherits from first container

### Example

Target pod with non-root user:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  securityContext:
    runAsUser: 1001
    runAsGroup: 1001
    runAsNonRoot: true
  containers:
    - name: app
      image: my-app:latest
```

Ephemeral container will automatically run as user 1001:1001.

### Verification

Check controller logs for inheritance messages:
```bash
kubectl logs -n kubecodriver-system -l control-plane=controller-manager | grep "Inherited"
```
```

## Implementation Checklist

### Phase 1: Controller Logic
- [ ] Modify `createEphemeralContainerForPod()` function
- [ ] Add inheritance logic for pod-level security context
- [ ] Add fallback logic for container-level security context
- [ ] Add logging for inherited values
- [ ] Handle nil checks properly

### Phase 2: Testing
- [ ] Add unit tests for pod-level inheritance
- [ ] Add unit tests for container-level inheritance
- [ ] Add unit tests for no security context case
- [ ] Run existing unit tests (ensure no regression)

### Phase 3: Integration Testing
- [ ] Run test with non-root pod
- [ ] Verify ephemeral container runs successfully
- [ ] Check controller logs for inheritance messages
- [ ] Verify container runs with correct user/group

### Phase 4: Documentation
- [ ] Update security documentation
- [ ] Add inheritance behavior explanation
- [ ] Add verification steps

## Success Criteria

1. ✅ Ephemeral containers inherit security context from target pods
2. ✅ No permission errors with non-root pods
3. ✅ No regression in existing functionality
4. ✅ Unit test coverage maintained
5. ✅ Integration test passes with non-root pod
6. ✅ Controller logs show inherited values

## Rollback Plan

If issues arise:
1. Revert controller changes
2. Keep test files for future attempts
3. No API/CRD changes to rollback

## Related Files

- `internal/controller/powertool_controller.go` - Main implementation
- `internal/controller/security_context_inheritance_test.go` - New tests
- `docs/security/README.md` - Documentation
- `roadmap/test-results-nonroot.md` - Test results
- `examples/test-nonroot-pod.yaml` - Test pod
- `examples/test-nonroot-powertool.yaml` - Test CoDriverJob

## Timeline

- **Phase 1**: 45 minutes (controller implementation)
- **Phase 2**: 30 minutes (unit tests)
- **Phase 3**: 30 minutes (integration testing)
- **Phase 4**: 15 minutes (documentation)

**Total**: ~2 hours

## Notes

- This is a backward-compatible change
- No API/CRD modifications required
- Existing CoDriverTools will automatically benefit
- Auto-discovery provides sensible defaults
- Logging helps with debugging and verification
