# CoDriverJob Built-in RBAC Roles

## Overview

KubeCoDriver provides three built-in ClusterRoles for managing access to CoDriverJob and CoDriverTool resources. These roles follow Kubernetes RBAC best practices and provide a secure, role-based approach to profiling operations.

## Built-in Roles

### 1. CoDriverJob Admin Role (`powertool-admin-role`)

**Purpose**: Full administrative access to all CoDriverJob resources.

**Permissions**:
- **CoDriverJob**: Full CRUD access (create, read, update, delete)
- **CoDriverTool**: Full CRUD access (create, read, update, delete)
- **Status resources**: Can read and update status for both resource types

**Intended for**:
- Platform administrators
- Security teams
- DevOps engineers with full cluster access

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-admin-role
rules:
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools", "powertoolconfigs"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status", "powertoolconfigs/status"]
  verbs: ["get", "patch", "update"]
```

### 2. CoDriverJob Editor Role (`powertool-editor-role`)

**Purpose**: Can create and manage profiling jobs using approved tools.

**Permissions**:
- **CoDriverJob**: Full CRUD access (create, read, update, delete)
- **CoDriverTool**: Read-only access (get, list, watch)
- **Status resources**: Can read and update CoDriverJob status, read-only access to CoDriverTool status

**Intended for**:
- Developers who need to run profiling
- Application teams
- QA engineers

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-editor-role
rules:
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status"]
  verbs: ["get", "patch", "update"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs/status"]
  verbs: ["get"]
```

### 3. CoDriverJob Viewer Role (`powertool-viewer-role`)

**Purpose**: Read-only access to all CoDriverJob resources for monitoring and auditing.

**Permissions**:
- **CoDriverJob**: Read-only access (get, list, watch)
- **CoDriverTool**: Read-only access (get, list, watch)
- **Status resources**: Read-only access to both resource types

**Intended for**:
- Monitoring systems
- Audit teams
- Read-only dashboard users
- Support teams

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-viewer-role
rules:
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools", "powertoolconfigs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status", "powertoolconfigs/status"]
  verbs: ["get"]
```

## Usage Examples

### Binding Roles to Users

#### Grant Admin Access to Platform Team

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: platform-team-powertool-admins
subjects:
- kind: Group
  name: platform-team
  apiGroup: rbac.authorization.k8s.io
- kind: User
  name: admin@company.com
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-admin-role
  apiGroup: rbac.authorization.k8s.io
```

#### Grant Editor Access to Development Teams

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: developers-powertool-editors
subjects:
- kind: Group
  name: developers
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: qa-engineers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-editor-role
  apiGroup: rbac.authorization.k8s.io
```

#### Grant Viewer Access to Monitoring Systems

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: monitoring-powertool-viewers
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: monitoring
- kind: Group
  name: support-team
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-viewer-role
  apiGroup: rbac.authorization.k8s.io
```

### Namespace-Scoped Access

For more granular control, you can create RoleBindings instead of ClusterRoleBindings:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: team-a-powertool-editors
  namespace: team-a-production
subjects:
- kind: Group
  name: team-a-developers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-editor-role
  apiGroup: rbac.authorization.k8s.io
```

## Security Model

### Separation of Concerns

1. **CoDriverTool Management** (Admin-only):
   - Define available profiling tools
   - Set security contexts and capabilities
   - Configure namespace restrictions
   - Control tool image sources

2. **CoDriverJob Operations** (Editor access):
   - Create profiling jobs using approved tools
   - Monitor profiling job status
   - Access profiling results
   - Cannot modify tool definitions or security settings

3. **Monitoring and Auditing** (Viewer access):
   - Read-only access to all resources
   - Monitor system usage
   - Generate reports and dashboards
   - Audit profiling activities

### Permission Matrix

| Action | Admin | Editor | Viewer |
|--------|-------|--------|--------|
| Create CoDriverTool | ✅ | ❌ | ❌ |
| Update CoDriverTool | ✅ | ❌ | ❌ |
| Delete CoDriverTool | ✅ | ❌ | ❌ |
| View CoDriverTool | ✅ | ✅ | ✅ |
| Create CoDriverJob | ✅ | ✅ | ❌ |
| Update CoDriverJob | ✅ | ✅ | ❌ |
| Delete CoDriverJob | ✅ | ✅ | ❌ |
| View CoDriverJob | ✅ | ✅ | ✅ |
| Update CoDriverJob Status | ✅ | ✅ | ❌ |
| View Status Resources | ✅ | ✅ | ✅ |

## Validation and Testing

### Check User Permissions

```bash
# Test admin permissions
kubectl auth can-i create powertoolconfigs --as=user:admin@company.com
kubectl auth can-i delete powertoolconfigs --as=user:admin@company.com

# Test editor permissions
kubectl auth can-i create powertools --as=user:developer@company.com
kubectl auth can-i create powertoolconfigs --as=user:developer@company.com  # Should be false

# Test viewer permissions
kubectl auth can-i get powertools --as=user:monitor@company.com
kubectl auth can-i delete powertools --as=user:monitor@company.com  # Should be false
```

### Verify Role Assignments

```bash
# List all ClusterRoleBindings for CoDriverJob roles
kubectl get clusterrolebindings -o json | jq -r '.items[] | select(.roleRef.name | contains("powertool")) | "\(.metadata.name): \(.roleRef.name)"'

# Check specific user's effective permissions
kubectl auth can-i --list --as=user:developer@company.com | grep powertool
```

## Best Practices

### 1. Principle of Least Privilege

- Start with viewer access and escalate as needed
- Use namespace-scoped RoleBindings when possible
- Regularly audit role assignments

### 2. Role Assignment Strategy

```yaml
# Example: Graduated access model
# Viewer -> Editor -> Admin based on experience and need

# New team members: Viewer access
- kind: Group
  name: new-developers
  # Bound to: powertool-viewer-role

# Experienced developers: Editor access  
- kind: Group
  name: senior-developers
  # Bound to: powertool-editor-role

# Platform team: Admin access
- kind: Group
  name: platform-team
  # Bound to: powertool-admin-role
```

### 3. Monitoring and Auditing

```bash
# Monitor CoDriverJob resource creation
kubectl get events --field-selector involvedObject.apiVersion=kubecodriver.codriverlabs.ai/v1alpha1

# Audit role usage
kubectl logs -n kube-system deployment/kube-apiserver | grep powertool
```

### 4. Integration with External Identity Providers

```yaml
# Example: OIDC integration
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-powertool-editors
subjects:
- kind: User
  name: oidc:developer@company.com
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-editor-role
  apiGroup: rbac.authorization.k8s.io
```

## Troubleshooting

### Common Issues

1. **Permission Denied Errors**:
   ```bash
   # Check current user's permissions
   kubectl auth whoami
   kubectl auth can-i create powertools
   ```

2. **Role Not Found**:
   ```bash
   # Verify roles are installed
   kubectl get clusterroles | grep powertool
   ```

3. **Binding Issues**:
   ```bash
   # Check role bindings
   kubectl describe clusterrolebinding <binding-name>
   ```

### Debugging Commands

```bash
# List all CoDriverJob-related RBAC resources
kubectl get clusterroles,clusterrolebindings,roles,rolebindings -A | grep powertool

# Check effective permissions for a user
kubectl auth can-i --list --as=user:example@company.com

# Verify role definitions
kubectl describe clusterrole powertool-admin-role
kubectl describe clusterrole powertool-editor-role
kubectl describe clusterrole powertool-viewer-role
```

## Migration from Custom Roles

If you're using custom RBAC roles, you can migrate to the built-in roles:

1. **Audit existing permissions**:
   ```bash
   kubectl get clusterroles -o yaml | grep -A 20 -B 5 powertool
   ```

2. **Create new bindings with built-in roles**:
   ```bash
   # Replace custom role names with built-in ones
   kubectl patch clusterrolebinding my-custom-binding --type='merge' -p='{"roleRef":{"name":"powertool-editor-role"}}'
   ```

3. **Clean up old roles**:
   ```bash
   kubectl delete clusterrole my-custom-powertool-role
   ```

The built-in roles provide a standardized, secure foundation for CoDriverJob access control while maintaining flexibility for organization-specific requirements.
