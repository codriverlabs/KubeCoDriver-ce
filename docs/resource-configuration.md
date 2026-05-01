# Resource Configuration for Ephemeral Containers

## Overview

CoDriverTool now supports configuring resource requests and limits for ephemeral containers. This allows administrators to control the CPU and memory resources allocated to profiling and chaos engineering tools.

## Configuration

### CoDriverTool Spec

Add a `resources` field to your CoDriverTool:

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: chaos-config
  namespace: kubecodriver-system
spec:
  name: "chaos"
  image: ghcr.io/codriverlabs/kubecodriver-chaos:v1.0.47
  securityContext:
    capabilities:
      add:
        - SYS_PTRACE
        - KILL
  resources:
    requests:
      memory: "64Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "1000m"
```

## Resource Specification

### Structure

```yaml
resources:
  requests:
    cpu: "<cpu-value>"
    memory: "<memory-value>"
  limits:
    cpu: "<cpu-value>"
    memory: "<memory-value>"
```

### Fields

- `requests` (optional): Minimum resources guaranteed for the container
  - `cpu` (optional): CPU request (e.g., "100m", "0.5", "1")
  - `memory` (optional): Memory request (e.g., "64Mi", "128Mi", "1Gi")

- `limits` (optional): Maximum resources the container can use
  - `cpu` (optional): CPU limit (e.g., "1000m", "1", "2")
  - `memory` (optional): Memory limit (e.g., "512Mi", "1Gi", "2Gi")

### Valid Resource Formats

**CPU:**
- Millicores: "100m", "500m", "1000m"
- Cores: "0.1", "0.5", "1", "2"

**Memory:**
- Mebibytes: "64Mi", "128Mi", "512Mi"
- Gibibytes: "1Gi", "2Gi"
- Megabytes: "64M", "128M", "512M"
- Gigabytes: "1G", "2G"

## Examples

### Minimal Resources (Lightweight Tools)

```yaml
resources:
  requests:
    memory: "32Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

### Standard Resources (Most Tools)

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "1000m"
```

### High Resources (CPU-Intensive Tools)

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "2000m"
```

### CPU Only

```yaml
resources:
  requests:
    cpu: "100m"
  limits:
    cpu: "500m"
```

### Memory Only

```yaml
resources:
  requests:
    memory: "64Mi"
  limits:
    memory: "256Mi"
```

## Behavior

### When Resources Are Not Specified

If the `resources` field is omitted from CoDriverTool, the ephemeral container will be created without resource requests or limits, inheriting the cluster's default resource policies.

### Resource Enforcement

- **Requests**: Kubernetes scheduler uses these to find nodes with sufficient resources
- **Limits**: Kubernetes enforces these limits; containers exceeding memory limits will be OOM-killed

## Best Practices

1. **Always Set Limits**: Prevent runaway resource consumption
2. **Match Tool Requirements**: CPU-intensive tools need higher CPU limits
3. **Consider Node Resources**: Ensure limits don't exceed node capacity
4. **Test in Non-Production**: Verify resource settings before production use
5. **Monitor Usage**: Use metrics to tune resource settings

## Tool-Specific Recommendations

### Chaos Tools
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "1000m"
```

### Profiling Tools (perf, strace)
```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "200m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### Lightweight Monitoring Tools
```yaml
resources:
  requests:
    memory: "32Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

## Troubleshooting

### Pod Fails to Schedule

**Symptom**: CoDriverJob remains in pending state

**Cause**: Resource requests exceed available node capacity

**Solution**: Reduce resource requests or add nodes with more capacity

### Container OOM Killed

**Symptom**: Ephemeral container terminates with OOMKilled status

**Cause**: Memory usage exceeded limits

**Solution**: Increase memory limits or optimize tool usage

### CPU Throttling

**Symptom**: Tool runs slower than expected

**Cause**: CPU usage hitting limits

**Solution**: Increase CPU limits or reduce tool intensity

## Implementation Details

### CRD Changes

Added `ResourceSpec` type to `api/v1alpha1/common_types.go`:

```go
type ResourceSpec struct {
    Requests *ResourceList `json:"requests,omitempty"`
    Limits   *ResourceList `json:"limits,omitempty"`
}

type ResourceList struct {
    CPU    *string `json:"cpu,omitempty"`
    Memory *string `json:"memory,omitempty"`
}
```

### Controller Changes

The controller's `buildResourceRequirements()` function converts the CRD's `ResourceSpec` to Kubernetes `ResourceRequirements` and applies them to ephemeral containers.

### Testing

Unit tests in `resource_requirements_test.go` cover:
- Nil resources
- Requests only
- Limits only
- Both requests and limits
- CPU only
- Memory only
