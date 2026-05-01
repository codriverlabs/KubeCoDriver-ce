# Controller RBAC Configuration

## Overview

The CoDriverJob controller requires specific RBAC permissions to manage CoDriverJob and CoDriverTool resources, create ephemeral containers, and interact with the Kubernetes API.

## Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubecodriver-controller-manager
  namespace: kubecodriver-system
```

## ClusterRole Permissions

### CoDriverJob Resources

```yaml
# CoDriverJob management
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# CoDriverJob status updates
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status"]
  verbs: ["get", "update", "patch"]

# CoDriverJob finalizers
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/finalizers"]
  verbs: ["update"]
```

### CoDriverTool Resources

```yaml
# CoDriverTool lookup (read-only)
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs"]
  verbs: ["get", "list", "watch"]

# CoDriverTool status updates
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs/status"]
  verbs: ["get", "update", "patch"]

# CoDriverTool finalizers
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs/finalizers"]
  verbs: ["update"]
```

### Core Kubernetes Resources

```yaml
# Pod management for ephemeral containers
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "update", "patch"]

# Ephemeral container management
- apiGroups: [""]
  resources: ["pods/ephemeralcontainers"]
  verbs: ["get", "list", "watch", "update", "patch"]

# ConfigMap access for configuration
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
```

## Complete RBAC Manifest

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubecodriver-manager-role
rules:
# CoDriverJob resources
- apiGroups:
  - kubecodriver.codriverlabs.ai
  resources:
  - powertools
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - kubecodriver.codriverlabs.ai
  resources:
  - powertools/finalizers
  verbs:
  - update
- apiGroups:
  - kubecodriver.codriverlabs.ai
  resources:
  - powertools/status
  verbs:
  - get
  - patch
  - update

# CoDriverTool resources
- apiGroups:
  - kubecodriver.codriverlabs.ai
  resources:
  - powertoolconfigs
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - kubecodriver.codriverlabs.ai
  resources:
  - powertoolconfigs/finalizers
  verbs:
  - update
- apiGroups:
  - kubecodriver.codriverlabs.ai
  resources:
  - powertoolconfigs/status
  verbs:
  - get
  - patch
  - update

# Core resources
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - pods/ephemeralcontainers
  verbs:
  - get
  - list
  - patch
  - update
  - watch

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubecodriver-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubecodriver-manager-role
subjects:
- kind: ServiceAccount
  name: kubecodriver-controller-manager
  namespace: kubecodriver-system
```

## Security Considerations

### Minimal Permissions
- **No cluster-admin**: Controller has only required permissions
- **No secret access**: Cannot read arbitrary secrets
- **No node access**: Cannot modify node resources
- **Scoped resources**: Only specific CRDs and core resources

### Permission Justification

| Permission | Justification | Risk Level |
|------------|---------------|------------|
| powertools/* | Core functionality - manage CoDriverJob lifecycle | Low |
| powertoolconfigs/get,list,watch | Tool configuration lookup - read-only | Low |
| pods/update,patch | Ephemeral container creation | Medium |
| pods/ephemeralcontainers/* | Direct ephemeral container management | Medium |
| configmaps/get,list,watch | Token configuration - read-only | Low |

### Risk Mitigation

1. **Pod Access Limitation**:
   - Only update/patch permissions (no create/delete)
   - Scoped to ephemeral containers only
   - No access to pod secrets or volumes

2. **ConfigMap Restriction**:
   - Read-only access
   - No write permissions to prevent configuration tampering
   - Limited to specific ConfigMaps via controller logic

3. **CoDriverTool Security**:
   - Read-only access prevents privilege escalation
   - Cannot modify security contexts
   - Enforces admin-defined security policies

## Monitoring and Auditing

### RBAC Violations
Monitor for these potential security events:

```bash
# Unauthorized resource access attempts
kubectl get events --field-selector reason=Forbidden

# Controller permission denials
kubectl logs -n kubecodriver-system deployment/kubecodriver-controller-manager | grep "forbidden\|unauthorized"

# RBAC policy violations
kubectl get events --field-selector involvedObject.kind=ClusterRole
```

### Audit Queries

```bash
# Review controller permissions
kubectl auth can-i --list --as=system:serviceaccount:kubecodriver-system:kubecodriver-controller-manager

# Check CoDriverTool access
kubectl auth can-i create powertoolconfigs --as=system:serviceaccount:kubecodriver-system:kubecodriver-controller-manager

# Verify pod access scope
kubectl auth can-i delete pods --as=system:serviceaccount:kubecodriver-system:kubecodriver-controller-manager
```

## Troubleshooting

### Common Permission Issues

1. **CoDriverTool Not Found**:
   ```
   Error: failed to get tool configuration: powertoolconfigs.kubecodriver.codriverlabs.ai "aperf-config" is forbidden
   ```
   - Check ClusterRole includes powertoolconfigs get/list/watch
   - Verify ClusterRoleBinding is correct

2. **Ephemeral Container Creation Failed**:
   ```
   Error: pods/ephemeralcontainers is forbidden
   ```
   - Ensure pods/ephemeralcontainers permissions are granted
   - Check Kubernetes version supports ephemeral containers

3. **ConfigMap Access Denied**:
   ```
   Error: configmaps "tools-configuration" is forbidden
   ```
   - Verify configmaps get permission in ClusterRole
   - Check ConfigMap exists in expected namespace

### Validation Commands

```bash
# Test controller permissions
kubectl auth can-i get powertools --as=system:serviceaccount:kubecodriver-system:kubecodriver-controller-manager
kubectl auth can-i update pods/ephemeralcontainers --as=system:serviceaccount:kubecodriver-system:kubecodriver-controller-manager
kubectl auth can-i get powertoolconfigs --as=system:serviceaccount:kubecodriver-system:kubecodriver-controller-manager

# Check effective permissions
kubectl describe clusterrolebinding kubecodriver-manager-rolebinding
kubectl describe clusterrole kubecodriver-manager-role
```
