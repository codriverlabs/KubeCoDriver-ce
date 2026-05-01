# Security Model

## Profiler Image Security

The KubeCoDriver operator implements a secure image resolution mechanism to prevent users from specifying arbitrary Docker images in CoDriverJob resources.

### How It Works

1. **No Direct Image Specification**: Users cannot specify Docker images directly in CoDriverJob CRDs
2. **Tool-Based Resolution**: Users specify a `tool` name (e.g., "aperf", "aprof") 
3. **ConfigMap Mapping**: The operator resolves tool names to Docker images using a ConfigMap
4. **Operator-Only Access**: Only the operator has access to the `profiler-images` ConfigMap

### ConfigMap Structure

The `profiler-images` ConfigMap in the `kubecodriver-system` namespace contains tool-to-image mappings:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: profiler-images
  namespace: kubecodriver-system
data:
  # Production images
  aperf: "your-registry/aperf:v1.0.0"
  aprof: "your-registry/aprof:v1.0.0"
  
  # Development/snapshot images
  aperf-dev: "your-registry/aperf:dev-latest"
  aperf-snapshot: "your-registry/aperf:snapshot-20241015"
```

### Managing Profiler Images

Use the provided script to update profiler image mappings:

```bash
# Add or update a profiler image
./update-profiler-images.sh aperf-dev your-registry/aperf:dev-latest

# Add a snapshot version
./update-profiler-images.sh aperf-snapshot your-registry/aperf:snapshot-$(date +%Y%m%d)
```

### Security Benefits

1. **Controlled Image Access**: Only pre-approved images can be used
2. **Version Control**: Administrators control which image versions are available
3. **Audit Trail**: All image changes are tracked through ConfigMap updates
4. **Namespace Isolation**: ConfigMap is isolated in the operator namespace
5. **RBAC Protection**: Regular users cannot modify the image mappings

### Example Usage

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: secure-profiling
spec:
  profiler:
    tool: "aperf"  # Resolved to image from ConfigMap
    duration: "30s"
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  output:
    mode: "ephemeral"
```

The operator will automatically resolve `tool: "aperf"` to the corresponding Docker image from the ConfigMap.
