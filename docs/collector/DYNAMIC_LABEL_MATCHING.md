# Dynamic Label Matching Implementation

## Problem

The initial implementation hardcoded the `app` label for path organization:
```go
appLabel := targetPod.Labels["app"]  // ❌ Assumes "app" label exists
```

This didn't align with the CoDriverJob CRD design, which allows ANY label selector:
```yaml
targets:
  labelSelector:
    matchLabels:
      env: production      # Could be any label
      tier: backend
      version: v2
```

## Solution

Extract the actual matching labels from the CoDriverJob's `labelSelector` that matched the target pod.

### Implementation

**Controller (`internal/controller/powertool_controller.go`):**
```go
func (r *CoDriverJobReconciler) extractMatchingLabels(selector *metav1.LabelSelector, podLabels map[string]string) string {
    if selector == nil || selector.MatchLabels == nil {
        return "unknown"
    }

    // Build compact representation: key-value (POSIX-compliant)
    var labels []string
    for key, value := range selector.MatchLabels {
        if podValue, exists := podLabels[key]; exists && podValue == value {
            labels = append(labels, fmt.Sprintf("%s-%s", key, value))
        }
    }

    if len(labels) == 0 {
        return "unknown"
    }

    return labels[0] // Use first matching label
}
```

**Environment Variable:**
```go
{Name: "POD_MATCHING_LABELS", Value: matchingLabels}
```

**HTTP Header:**
```bash
-H "X-CoDriverJob-Matching-Labels: ${POD_MATCHING_LABELS:-unknown}"
```

## Examples

### Scenario 1: App Label
```yaml
targets:
  labelSelector:
    matchLabels:
      app: nginx
```
**Path:** `/data/default/app-nginx/profile-job/2025-10-30/output.txt`

### Scenario 2: Environment Label
```yaml
targets:
  labelSelector:
    matchLabels:
      env: production
```
**Path:** `/data/production/env-production/profile-job/2025-10-30/output.txt`

### Scenario 3: Custom Label
```yaml
targets:
  labelSelector:
    matchLabels:
      tier: backend
      component: api
```
**Path:** `/data/default/tier-backend/profile-job/2025-10-30/output.txt`
(Uses first matching label)

### Scenario 4: No Match
```yaml
targets:
  labelSelector:
    matchLabels:
      nonexistent: label
```
**Path:** `/data/default/unknown/profile-job/2025-10-30/output.txt`

## Benefits

1. **Flexible:** Works with any label selector defined in CoDriverJob
2. **Semantic:** Path reflects actual targeting criteria
3. **Organized:** Groups profiles by the same selection criteria
4. **Discoverable:** Easy to find profiles by label used

## Path Organization Examples

### By Application
```
/data/
├── default/
│   ├── app-frontend/
│   │   ├── profile-frontend-1/2025-10-30/
│   │   └── profile-frontend-2/2025-10-30/
│   └── app-backend/
│       └── profile-backend/2025-10-30/
```

### By Environment
```
/data/
├── production/
│   └── env-production/
│       ├── profile-prod-1/2025-10-30/
│       └── profile-prod-2/2025-10-30/
└── staging/
    └── env-staging/
        └── profile-staging/2025-10-30/
```

### By Tier
```
/data/
└── default/
    ├── tier-frontend/
    │   └── profile-web/2025-10-30/
    ├── tier-backend/
    │   └── profile-api/2025-10-30/
    └── tier-database/
        └── profile-db/2025-10-30/
```

## Migration Notes

### From Hardcoded `app` Label

**Before:**
- Environment variable: `POD_APP_LABEL`
- HTTP header: `X-CoDriverJob-App-Label`
- Path: `/data/{namespace}/{app-value}/...`

**After:**
- Environment variable: `POD_MATCHING_LABELS`
- HTTP header: `X-CoDriverJob-Matching-Labels`
- Path: `/data/{namespace}/{key=value}/...`

### Backward Compatibility

If you were using `app` label:
```yaml
targets:
  labelSelector:
    matchLabels:
      app: my-app
```

**Old path:** `/data/default/my-app/...`  
**New path:** `/data/default/app-my-app/...`

The structure includes the label key for clarity and POSIX compliance.

## Future Enhancements

Potential improvements:
1. Support multiple labels in path: `/data/{namespace}/{label1=value1}/{label2=value2}/...`
2. Configurable label priority for path organization
3. Label-based retention policies
4. Label-based access control
