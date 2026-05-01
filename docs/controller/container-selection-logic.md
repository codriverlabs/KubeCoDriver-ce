# Container Selection Logic Analysis

## Current Implementation

### Pod Selection
✅ **Works correctly** - Uses label selector to find target pods:

```go
// Line 284-296
selector, err := metav1.LabelSelectorAsSelector(powerTool.Spec.Targets.LabelSelector)
if err := r.List(ctx, &podList, &client.ListOptions{
    Namespace:     powerTool.Namespace,
    LabelSelector: selector,
}); err != nil {
    // handle error
}
```

### Container Selection
❌ **NOT IMPLEMENTED** - The `Container` field in `TargetSpec` is defined but not used!

**API Definition** (`api/v1alpha1/common_types.go`):
```go
type TargetSpec struct {
    NamespaceSelector *NamespaceSelector    `json:"namespaceSelector,omitempty"`
    LabelSelector     *metav1.LabelSelector `json:"labelSelector"`
    Container         *string               `json:"container,omitempty"`  // ← DEFINED BUT NOT USED
}
```

**Current Behavior**:
The controller always uses the **first container** (`pod.Spec.Containers[0]`) when inheriting security context:

```go
// Line 109 in buildCoDriverJobEnvVars - uses first matching label, not container
matchingLabels := r.extractMatchingLabels(job.Spec.Targets.LabelSelector, targetPod.Labels)

// Security context inheritance (in our proposed fix) would use:
if len(pod.Spec.Containers) > 0 && pod.Spec.Containers[0].SecurityContext != nil {
    // Always uses first container ← PROBLEM
}
```

## The Issue

### Scenario 1: Single Container Pod
✅ **Works fine** - First container is the only container

```yaml
spec:
  containers:
    - name: main-app
      image: my-app:latest
```

### Scenario 2: Multi-Container Pod
❌ **Problem** - May select wrong container

```yaml
spec:
  containers:
    - name: sidecar        # ← Controller uses THIS (index 0)
      image: sidecar:latest
      securityContext:
        runAsUser: 2000
    - name: main-app       # ← User wants to profile THIS
      image: my-app:latest
      securityContext:
        runAsUser: 1001
```

**CoDriverJob spec**:
```yaml
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
    container: "main-app"  # ← This is IGNORED!
```

## Container Name Uniqueness

**Yes, container names are unique within a pod** (Kubernetes requirement):
- Each container in `pod.Spec.Containers` must have a unique name
- Each ephemeral container in `pod.Spec.EphemeralContainers` must have a unique name
- Container names cannot overlap between regular and ephemeral containers

## Required Fixes

### Fix 1: Use Target Container for Security Context Inheritance

Update the security context inheritance logic to use the specified target container:

```go
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

Then use it in `createEphemeralContainerForPod`:

```go
func (r *CoDriverJobReconciler) createEphemeralContainerForPod(ctx context.Context, powerTool *toev1alpha1.CoDriverJob, toolConfig *toev1alpha1.CoDriverTool, pod corev1.Pod, containerName string) error {
    logger := log.FromContext(ctx)

    // Get target container
    targetContainer := r.getTargetContainer(pod, powerTool.Spec.Targets.Container)
    
    // Build security context from toolConfig
    securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

    // Inherit from pod-level security context first
    if pod.Spec.SecurityContext != nil {
        if pod.Spec.SecurityContext.RunAsUser != nil {
            securityContext.RunAsUser = pod.Spec.SecurityContext.RunAsUser
        }
        // ... other pod-level fields
    }

    // Override with target container's security context if specified
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
        // ... other container-level fields
    }

    // ... rest of function
}
```

### Fix 2: Add Target Container Name to Environment Variables

The profiling tools may need to know which container to profile:

```go
func (r *CoDriverJobReconciler) buildCoDriverJobEnvVars(job *toev1alpha1.CoDriverJob, targetPod corev1.Pod) []corev1.EnvVar {
    // ... existing code ...
    
    // Add target container name
    targetContainerName := "default"
    if job.Spec.Targets.Container != nil && *job.Spec.Targets.Container != "" {
        targetContainerName = *job.Spec.Targets.Container
    } else if len(targetPod.Spec.Containers) > 0 {
        targetContainerName = targetPod.Spec.Containers[0].Name
    }
    
    envVars = append(envVars, corev1.EnvVar{
        Name:  "TARGET_CONTAINER_NAME",
        Value: targetContainerName,
    })
    
    // ... rest of function
}
```

## Priority

**Medium-High** - This affects multi-container pods:
- Single container pods work fine (most common case)
- Multi-container pods may profile wrong container
- Security context inheritance may use wrong user/group

## Implementation Order

1. ✅ Document the issue (this file)
2. 🔴 Add `getTargetContainer()` helper function
3. 🔴 Update security context inheritance to use target container
4. 🔴 Add `TARGET_CONTAINER_NAME` to environment variables
5. 🔴 Add unit tests for multi-container scenarios
6. 🔴 Update documentation

## Related Files

- `api/v1alpha1/common_types.go` - TargetSpec definition
- `internal/controller/powertool_controller.go` - Main logic
- `roadmap/non-root-security-context-fix.md` - Security context fix

## Examples

### Single Container (Current behavior works)
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: profile-single
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
    # container: not needed for single container pods
```

### Multi-Container (Needs fix)
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: profile-multi
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
    container: "main-app"  # ← Should use this container's security context
```

## Testing

Add test case for multi-container pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: multi-container-test
  labels:
    app: test-app
spec:
  containers:
    - name: sidecar
      image: busybox
      command: ["sleep", "3600"]
      securityContext:
        runAsUser: 2000
    - name: main-app
      image: busybox
      command: ["sleep", "3600"]
      securityContext:
        runAsUser: 1001
```

Expected: When targeting `main-app`, ephemeral container should run as user 1001, not 2000.
