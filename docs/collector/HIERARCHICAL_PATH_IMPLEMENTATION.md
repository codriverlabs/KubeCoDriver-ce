# Hierarchical Path Implementation

## Overview

Implemented hierarchical path structure for collector storage:
```
/data/{namespace}/{matching-labels}/{powertool-name}/{year}/{month}/{day}/{filename}
```

Example:
```
/data/default/app-nginx/profile-job-123/2025/10/30/aperf-output.txt
/data/production/env-prod/profile-prod/2025/10/30/perf-data.txt
```

The date structure uses separate folders for year/month/day for better organization and performance with large datasets.

The `matching-labels` component is dynamically extracted from the CoDriverJob's `labelSelector.matchLabels` - it uses the first label that matched the target pod in `key-value` format (POSIX-compliant).

## Changes Made

### 1. Storage Manager (`pkg/collector/storage/manager.go`)
- Added `ProfileMetadata` struct to capture namespace, app label, CoDriverJob name, and filename
- Added `dateFormat` field to Manager for configurable date formatting
- Updated `SaveProfile()` to build hierarchical paths and create directories

### 2. Collector Server (`pkg/collector/server/server.go`)
- Added `DateFormat` to Config struct
- Updated `handleProfile()` to extract metadata from HTTP headers:
  - `X-CoDriverJob-Namespace`
  - `X-CoDriverJob-Matching-Labels` (dynamic label from selector)
  - `X-CoDriverJob-Job-ID`
  - `X-CoDriverJob-Filename`
- Defaults `matching-labels` to "unknown" if not provided

### 3. Collector Main (`cmd/collector/main.go`)
- Added `DATE_FORMAT` environment variable support
- Defaults to "2006-01-02" (ISO 8601 format)

### 4. Controller (`internal/controller/powertool_controller.go`)
- Updated `buildCoDriverJobEnvVars()` to extract matching labels dynamically
- Added `extractMatchingLabels()` helper function
- Passes first matching label as `POD_MATCHING_LABELS` environment variable
- Format: `key=value` (e.g., `app=nginx`, `env=prod`)
- Defaults to "unknown" if no labels match

### 5. Power-Tools Scripts
- Updated `power-tools/aperf/send-profile.sh`
- Updated `power-tools/common/send-profile.sh`
- Added metadata headers to curl requests:
  - `X-CoDriverJob-Namespace: $TARGET_NAMESPACE`
  - `X-CoDriverJob-Matching-Labels: ${POD_MATCHING_LABELS:-unknown}`
- Added validation for `TARGET_NAMESPACE` environment variable

### 6. Deployment Configuration
- Created `deploy/collector/configmap.yaml` with date format configuration
- Updated `deploy/collector/deployment.yaml` to mount ConfigMap
- Updated `deploy/collector/kustomization.yaml` to include ConfigMap

## Configuration

### Date Format (Required)

ConfigMap must be created before deploying collector:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: collector-config
  namespace: kubecodriver-system
data:
  dateFormat: "2006/01/02"  # Default: creates year/month/day folders
```

Other examples:
- `"2006/01/02"` → 2025/10/30 (separate folders - recommended)
- `"2006-01-02"` → 2025-10-30 (single folder)
- `"2006/01/02/15"` → 2025/10/30/14 (with hour)

Uses Go time format strings.

**Why separate folders?**
- Better filesystem performance with many files
- Easier to navigate and query by date ranges
- Natural partitioning for retention policies
- Standard practice for time-series data

### Matching Labels

The collector uses the actual labels from the CoDriverJob's `labelSelector.matchLabels` that matched the target pod.

Examples:
```yaml
# Using app label
targets:
  labelSelector:
    matchLabels:
      app: nginx
# Path: /data/default/app-nginx/...

# Using environment label
targets:
  labelSelector:
    matchLabels:
      env: production
# Path: /data/production/env-production/...

# Multiple labels (uses first match)
targets:
  labelSelector:
    matchLabels:
      tier: backend
      version: v2
# Path: /data/default/tier-backend/...
```

Format: `key-value` (POSIX-compliant, DNS-safe)

If no labels match or selector is empty, defaults to "unknown".

## Deployment

1. Apply ConfigMap:
```bash
kubectl apply -f deploy/collector/configmap.yaml
```

2. Rebuild and deploy collector:
```bash
make docker-build-collector docker-push-collector
kubectl rollout restart deployment/kubecodriver-collector -n kubecodriver-system
```

3. Rebuild and deploy power-tools:
```bash
cd power-tools/aperf
docker build -t your-registry/aperf:latest .
docker push your-registry/aperf:latest
```

4. Rebuild and deploy controller:
```bash
make docker-build docker-push IMG=your-registry/kubecodriver:tag
make deploy IMG=your-registry/kubecodriver:tag
```

## Testing

1. Create a test pod with labels:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-app
  labels:
    app: my-application
    env: staging
spec:
  containers:
  - name: app
    image: nginx
```

2. Create CoDriverJob with collector output:
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: test-profile
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-application  # This will be used in path
  tool:
    name: aperf
    duration: 30s
  output:
    mode: collector
    collector:
      endpoint: https://kubecodriver-collector.kubecodriver-system.svc:8443
```

3. Verify hierarchical structure:
```bash
kubectl exec -n kubecodriver-system deployment/kubecodriver-collector -- ls -R /data
```

Expected output:
```
/data/default/app-my-application/test-profile/2025/10/30/aperf-output.txt
```

## Backward Compatibility

The implementation maintains backward compatibility:
- If headers are missing, uses sensible defaults
- ConfigMap is optional (defaults to ISO 8601 format)
- Existing deployments continue to work without changes
