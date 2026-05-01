# Container Selection Fix - Implementation Guide

## Status: 🔴 Not Started

**Priority**: Medium-High  
**Estimated Effort**: 1.5 hours  
**Target Version**: v1.0.52

## Problem Statement

The `Container` field in `TargetSpec` is defined but not used by the controller. This causes issues with multi-container pods where the wrong container's security context may be inherited.

**Current Behavior**:
- Always uses first container (`pod.Spec.Containers[0]`)
- Works for single-container pods
- Fails for multi-container pods

**Expected Behavior**:
- Use specified container from `spec.targets.container`
- Fallback to first container if not specified
- Inherit security context from correct container

## Implementation Steps

### Step 1: Add Helper Function (15 min)

**File**: `internal/controller/powertool_controller.go`

Add function to get target container:

```go
// getTargetContainer returns the target container from the pod
// If targetContainerName is specified, it finds that container
// Otherwise, it returns the first container
func (r *CoDriverJobReconciler) getTargetContainer(pod corev1.Pod, targetContainerName *string) *corev1.Container {
    // If no container specified, use first container
    if targetContainerName == nil || *targetContainerName == "" {
        if len(pod.Spec.Containers) > 0 {
            return &pod.Spec.Containers[0]
        }
        return nil
    }
    
    // Find the specified container
    for i := range pod.Spec.Containers {
        if pod.Spec.Containers[i].Name == *targetContainerName {
            return &pod.Spec.Containers[i]
        }
    }
    
    // Container not found, fallback to first
    if len(pod.Spec.Containers) > 0 {
        return &pod.Spec.Containers[0]
    }
    return nil
}
```

### Step 2: Update Security Context Inheritance (20 min)

**File**: `internal/controller/powertool_controller.go`

Modify `createEphemeralContainerForPod()`:

```go
func (r *CoDriverJobReconciler) createEphemeralContainerForPod(ctx context.Context, powerTool *kubecodriverv1alpha1.CoDriverJob, toolConfig *kubecodriverv1alpha1.CoDriverTool, pod corev1.Pod, containerName string) error {
    logger := log.FromContext(ctx)

    // Get target container
    targetContainer := r.getTargetContainer(pod, powerTool.Spec.Targets.Container)
    if targetContainer != nil {
        logger.Info("Target container identified", "container", targetContainer.Name)
    }

    // Build environment variables
    envVars := r.buildCoDriverJobEnvVars(powerTool, pod)

    // ... existing collector code ...

    // Build base security context from toolConfig
    securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

    // Inherit from pod-level security context first
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

    // Override with target container's security context if available
    if targetContainer != nil && targetContainer.SecurityContext != nil {
        if targetContainer.SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = targetContainer.SecurityContext.RunAsUser
            logger.Info("Inherited runAsUser from target container", 
                "container", targetContainer.Name, 
                "user", *targetContainer.SecurityContext.RunAsUser)
        }
        if targetContainer.SecurityContext.RunAsGroup != nil {
            securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
            logger.Info("Inherited runAsGroup from target container", 
                "container", targetContainer.Name, 
                "group", *targetContainer.SecurityContext.RunAsGroup)
        }
        if targetContainer.SecurityContext.RunAsNonRoot != nil {
            securityContext.RunAsNonRoot = targetContainer.SecurityContext.RunAsNonRoot
            logger.Info("Inherited runAsNonRoot from target container", 
                "container", targetContainer.Name, 
                "nonRoot", *targetContainer.SecurityContext.RunAsNonRoot)
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

### Step 3: Add Target Container to Environment Variables (10 min)

**File**: `internal/controller/powertool_controller.go`

Update `buildCoDriverJobEnvVars()`:

```go
func (r *CoDriverJobReconciler) buildCoDriverJobEnvVars(job *kubecodriverv1alpha1.CoDriverJob, targetPod corev1.Pod) []corev1.EnvVar {
    // Extract matching labels
    matchingLabels := r.extractMatchingLabels(job.Spec.Targets.LabelSelector, targetPod.Labels)

    // Determine target container name
    targetContainerName := "default"
    if job.Spec.Targets.Container != nil && *job.Spec.Targets.Container != "" {
        targetContainerName = *job.Spec.Targets.Container
    } else if len(targetPod.Spec.Containers) > 0 {
        targetContainerName = targetPod.Spec.Containers[0].Name
    }

    envVars := []corev1.EnvVar{
        {Name: "PROFILER_TOOL", Value: job.Spec.Tool.Name},
        {Name: "PROFILER_DURATION", Value: job.Spec.Tool.Duration},
        {Name: "TARGET_POD_NAME", Value: targetPod.Name},
        {Name: "TARGET_NAMESPACE", Value: targetPod.Namespace},
        {Name: "TARGET_CONTAINER_NAME", Value: targetContainerName},  // NEW
        {Name: "POD_MATCHING_LABELS", Value: matchingLabels},
        {Name: "OUTPUT_MODE", Value: job.Spec.Output.Mode},
    }

    // ... rest of existing code ...
}
```

### Step 4: Add Unit Tests (30 min)

**File**: `internal/controller/container_selection_test.go` (new)

```go
package controller

import (
    "testing"
    "github.com/stretchr/testify/assert"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetTargetContainer_SingleContainer(t *testing.T) {
    r := &CoDriverJobReconciler{}
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "main"},
            },
        },
    }
    
    // No container specified - should return first
    container := r.getTargetContainer(pod, nil)
    assert.NotNil(t, container)
    assert.Equal(t, "main", container.Name)
}

func TestGetTargetContainer_MultiContainer_Specified(t *testing.T) {
    r := &CoDriverJobReconciler{}
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "sidecar"},
                {Name: "main-app"},
            },
        },
    }
    
    targetName := "main-app"
    container := r.getTargetContainer(pod, &targetName)
    assert.NotNil(t, container)
    assert.Equal(t, "main-app", container.Name)
}

func TestGetTargetContainer_MultiContainer_NotSpecified(t *testing.T) {
    r := &CoDriverJobReconciler{}
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "sidecar"},
                {Name: "main-app"},
            },
        },
    }
    
    // No container specified - should return first
    container := r.getTargetContainer(pod, nil)
    assert.NotNil(t, container)
    assert.Equal(t, "sidecar", container.Name)
}

func TestGetTargetContainer_NotFound_FallbackToFirst(t *testing.T) {
    r := &CoDriverJobReconciler{}
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "sidecar"},
                {Name: "main-app"},
            },
        },
    }
    
    targetName := "nonexistent"
    container := r.getTargetContainer(pod, &targetName)
    assert.NotNil(t, container)
    assert.Equal(t, "sidecar", container.Name) // Falls back to first
}

func TestGetTargetContainer_EmptyPod(t *testing.T) {
    r := &CoDriverJobReconciler{}
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{},
        },
    }
    
    container := r.getTargetContainer(pod, nil)
    assert.Nil(t, container)
}

func TestSecurityContextInheritance_TargetContainer(t *testing.T) {
    user1 := int64(2000)
    user2 := int64(1001)
    
    pod := corev1.Pod{
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name: "sidecar",
                    SecurityContext: &corev1.SecurityContext{
                        RunAsUser: &user1,
                    },
                },
                {
                    Name: "main-app",
                    SecurityContext: &corev1.SecurityContext{
                        RunAsUser: &user2,
                    },
                },
            },
        },
    }
    
    r := &CoDriverJobReconciler{}
    targetName := "main-app"
    container := r.getTargetContainer(pod, &targetName)
    
    assert.NotNil(t, container)
    assert.NotNil(t, container.SecurityContext)
    assert.NotNil(t, container.SecurityContext.RunAsUser)
    assert.Equal(t, int64(1001), *container.SecurityContext.RunAsUser)
}
```

### Step 5: Add E2E Test (15 min)

**File**: `examples/test-multicontainer-pod.yaml` (new)

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: multi-container-test
  namespace: kubecodriver-test
  labels:
    app: "multi-test"
spec:
  containers:
    - name: sidecar
      image: busybox
      command: ["sleep", "400000"]
      securityContext:
        runAsUser: 2000
        runAsGroup: 2000
    - name: main-app
      image: busybox
      command: ["sleep", "400000"]
      securityContext:
        runAsUser: 1001
        runAsGroup: 1001
```

**File**: `examples/test-multicontainer-powertool.yaml` (new)

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: aperf-multicontainer-test
  namespace: kubecodriver-test
spec:
  targets:
    labelSelector:
      matchLabels:
        app: multi-test
    container: "main-app"  # Target specific container
  tool:
    name: "aperf"
    duration: "10s"
  output:
    mode: "ephemeral"
```

## Implementation Checklist

### Step 1: Helper Function
- [ ] Add `getTargetContainer()` function
- [ ] Handle nil/empty container name
- [ ] Handle container not found (fallback to first)
- [ ] Handle empty pod

### Step 2: Security Context Inheritance
- [ ] Get target container using helper
- [ ] Inherit from pod-level first
- [ ] Override with target container's security context
- [ ] Add logging for inherited values

### Step 3: Environment Variables
- [ ] Add `TARGET_CONTAINER_NAME` to env vars
- [ ] Use specified container name or fallback to first

### Step 4: Unit Tests
- [ ] Test single container pod
- [ ] Test multi-container with specified target
- [ ] Test multi-container without specified target
- [ ] Test container not found (fallback)
- [ ] Test empty pod
- [ ] Test security context inheritance from target container

### Step 5: E2E Test
- [ ] Create multi-container test pod
- [ ] Create CoDriverJob targeting specific container
- [ ] Verify ephemeral container uses correct security context
- [ ] Verify TARGET_CONTAINER_NAME env var is set

## Success Criteria

1. ✅ `getTargetContainer()` correctly identifies target container
2. ✅ Security context inherited from specified container
3. ✅ Fallback to first container when not specified
4. ✅ `TARGET_CONTAINER_NAME` env var set correctly
5. ✅ Unit tests pass
6. ✅ E2E test with multi-container pod passes
7. ✅ No regression in single-container pods

## Timeline

- **Step 1**: 15 minutes (helper function)
- **Step 2**: 20 minutes (security context inheritance)
- **Step 3**: 10 minutes (environment variables)
- **Step 4**: 30 minutes (unit tests)
- **Step 5**: 15 minutes (e2e test)

**Total**: ~1.5 hours

## Related Files

- `api/v1alpha1/common_types.go` - TargetSpec definition
- `internal/controller/powertool_controller.go` - Main implementation
- `docs/container-selection-logic.md` - Analysis document
- `roadmap/non-root-security-context-fix.md` - Related fix

## Notes

- This fix is independent of the non-root security context fix
- Can be implemented separately or combined
- Fully backward compatible
- Single-container pods behavior unchanged
