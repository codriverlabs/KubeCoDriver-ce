# RBAC Deployment Guide

## Overview

This guide defines how to deploy CoDriverJob with proper RBAC separation between administrators and users, including namespace restrictions for CoDriverTool.

## User Roles

### 1. CoDriverJob Administrators
- Can create/modify CoDriverTool CRDs
- Can define security contexts and capabilities
- Can restrict tools to specific namespaces
- Typically: Platform team, Security team

### 2. CoDriverJob Users
- Can create/modify CoDriverJob CRDs
- Cannot modify security settings
- Limited to using approved tools
- Typically: Application developers, DevOps engineers

### 3. Namespace Administrators
- Can create CoDriverJob in specific namespaces
- Cannot create CoDriverTool
- Scoped to their managed namespaces
- Typically: Team leads, Namespace owners

## RBAC Configuration

### CoDriverJob Administrators

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-admin
rules:
# CoDriverTool management (admin-only)
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs"]
  verbs: ["create", "update", "patch", "delete", "get", "list", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs/status"]
  verbs: ["get", "update", "patch"]

# CoDriverJob management (for testing/debugging)
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools"]
  verbs: ["create", "update", "patch", "delete", "get", "list", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status"]
  verbs: ["get", "update", "patch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: powertool-admins
subjects:
# Add admin users/groups here
- kind: User
  name: "admin@company.com"
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: "platform-team"
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: "security-team"
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-admin
  apiGroup: rbac.authorization.k8s.io
```

### CoDriverJob Users (Cluster-wide)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-user
rules:
# CoDriverJob management (user access)
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools"]
  verbs: ["create", "update", "patch", "delete", "get", "list", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status"]
  verbs: ["get", "list", "watch"]

# CoDriverTool read-only (to see available tools)
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: powertool-users
subjects:
# Add user groups here
- kind: Group
  name: "developers"
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: "devops-engineers"
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-user
  apiGroup: rbac.authorization.k8s.io
```

### Namespace-Scoped Users

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: production
  name: powertool-namespace-user
rules:
# CoDriverJob management in specific namespace
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools"]
  verbs: ["create", "update", "patch", "delete", "get", "list", "watch"]
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertools/status"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: powertool-production-users
  namespace: production
subjects:
- kind: Group
  name: "production-team"
  apiGroup: rbac.authorization.k8s.io
- kind: User
  name: "prod-lead@company.com"
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: powertool-namespace-user
  apiGroup: rbac.authorization.k8s.io

---
# Separate RoleBinding for CoDriverTool read access
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: powertool-production-config-readers
subjects:
- kind: Group
  name: "production-team"
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: powertool-config-reader
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-config-reader
rules:
- apiGroups: ["kubecodriver.codriverlabs.ai"]
  resources: ["powertoolconfigs"]
  verbs: ["get", "list", "watch"]
```

## CoDriverTool Examples with Namespace Restrictions

### Unrestricted Tool (Admin Use)

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: admin-debugger-config
  namespace: kubecodriver-system
  annotations:
    codriverjob.kubecodriver.codriverlabs.ai/access-level: "admin-only"
spec:
  name: "admin-debugger"
  image: "registry/admin-debugger:latest"
  security:
    allowPrivileged: true
    allowHostPID: true
    capabilities:
      add: ["SYS_ADMIN", "SYS_PTRACE"]
  # No allowedNamespaces = can be used anywhere (by admins)
  description: "Administrative debugging tool with full system access"
```

### Production-Only Tool

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: prod-profiler-config
  namespace: kubecodriver-system
spec:
  name: "prod-profiler"
  image: "registry/prod-profiler:latest"
  security:
    allowPrivileged: false
    capabilities:
      add: ["SYS_PTRACE"]
  allowedNamespaces:
    - "production"
    - "staging"
  description: "Production profiler - restricted to prod/staging environments"
```

### Development Tool

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: dev-analyzer-config
  namespace: kubecodriver-system
spec:
  name: "dev-analyzer"
  image: "registry/dev-analyzer:latest"
  security:
    allowPrivileged: false
    capabilities:
      add: ["SYS_PTRACE"]
  allowedNamespaces:
    - "development"
    - "testing"
    - "sandbox"
  description: "Development analyzer - safe for dev environments"
```

### Team-Specific Tool

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: team-a-profiler-config
  namespace: kubecodriver-system
spec:
  name: "team-a-profiler"
  image: "registry/team-a-profiler:latest"
  security:
    allowPrivileged: false
    capabilities:
      add: ["SYS_PTRACE"]
  allowedNamespaces:
    - "team-a-prod"
    - "team-a-staging"
    - "team-a-dev"
  description: "Team A specific profiler"
```

## Deployment Scenarios

### Scenario 1: Multi-Tenant Cluster

```yaml
# Platform team creates restricted tools per tenant
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: tenant-1-profiler
spec:
  name: "profiler"
  allowedNamespaces: ["tenant-1-prod", "tenant-1-dev"]
  # ... security config

---
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: tenant-2-profiler
spec:
  name: "profiler"  # Same tool name, different config
  allowedNamespaces: ["tenant-2-prod", "tenant-2-dev"]
  # ... different security config
```

### Scenario 2: Environment-Based Restrictions

```yaml
# High-privilege tool for production
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: prod-system-analyzer
spec:
  name: "system-analyzer"
  allowedNamespaces: ["production"]
  security:
    capabilities:
      add: ["SYS_PTRACE", "SYS_ADMIN"]

---
# Lower-privilege version for development
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverTool
metadata:
  name: dev-system-analyzer
spec:
  name: "system-analyzer-dev"
  allowedNamespaces: ["development", "testing"]
  security:
    capabilities:
      add: ["SYS_PTRACE"]
```

## User Groups Integration

### Active Directory/LDAP Integration

```yaml
# Map AD groups to Kubernetes groups
apiVersion: v1
kind: ConfigMap
metadata:
  name: group-mapping
  namespace: kube-system
data:
  # Platform team = CoDriverJob admins
  "CN=Platform-Team,OU=Groups,DC=company,DC=com": "platform-team"
  
  # Security team = CoDriverJob admins
  "CN=Security-Team,OU=Groups,DC=company,DC=com": "security-team"
  
  # Development teams = CoDriverJob users
  "CN=Developers,OU=Groups,DC=company,DC=com": "developers"
  "CN=DevOps,OU=Groups,DC=company,DC=com": "devops-engineers"
  
  # Production team = Namespace-scoped users
  "CN=Production-Team,OU=Groups,DC=company,DC=com": "production-team"
```

### OIDC Integration

```yaml
# OIDC configuration for API server
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver
spec:
  containers:
  - name: kube-apiserver
    command:
    - kube-apiserver
    - --oidc-issuer-url=https://auth.company.com
    - --oidc-client-id=kubernetes
    - --oidc-username-claim=email
    - --oidc-groups-claim=groups
    - --oidc-groups-prefix="oidc:"
```

## Validation and Testing

### Test RBAC Configuration

```bash
# Test admin permissions
kubectl auth can-i create powertoolconfigs --as=user:admin@company.com
kubectl auth can-i update powertoolconfigs --as=user:admin@company.com

# Test user permissions
kubectl auth can-i create powertools --as=user:developer@company.com
kubectl auth can-i create powertoolconfigs --as=user:developer@company.com  # Should be false

# Test namespace restrictions
kubectl auth can-i create powertools --as=user:prod-lead@company.com -n production
kubectl auth can-i create powertools --as=user:prod-lead@company.com -n development  # Should be false
```

### Validate CoDriverTool Restrictions

```bash
# Create test CoDriverJob in allowed namespace
kubectl apply -f - <<EOF
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: test-restricted-tool
  namespace: production
spec:
  tool:
    name: "prod-profiler"  # References restricted CoDriverTool
  # ... rest of spec
EOF

# Try to create in disallowed namespace (should fail)
kubectl apply -f - <<EOF
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: test-restricted-tool
  namespace: development  # Not in allowedNamespaces
spec:
  tool:
    name: "prod-profiler"
  # ... rest of spec
EOF
```

## Best Practices

### For Administrators

1. **Principle of Least Privilege**:
   - Create separate CoDriverTool for different security levels
   - Use namespace restrictions to limit tool scope
   - Regular review of tool permissions

2. **Tool Naming Convention**:
   ```
   {environment}-{tool-name}-config
   prod-profiler-config
   dev-analyzer-config
   team-a-debugger-config
   ```

3. **Documentation**:
   ```yaml
   metadata:
     annotations:
       codriverjob.kubecodriver.codriverlabs.ai/owner: "platform-team"
       codriverjob.kubecodriver.codriverlabs.ai/approved-by: "security-team"
       codriverjob.kubecodriver.codriverlabs.ai/allowed-users: "production-team"
   ```

### For Users

1. **Check Available Tools**:
   ```bash
   # List all available tools
   kubectl get powertoolconfigs -A
   
   # Check tool restrictions
   kubectl describe powertoolconfig prod-profiler-config
   ```

2. **Understand Namespace Restrictions**:
   ```bash
   # Check which tools are available in your namespace
   kubectl get powertoolconfigs -o json | jq '.items[] | select(.spec.allowedNamespaces == null or (.spec.allowedNamespaces[] | contains("'$(kubectl config view --minify -o jsonpath='{..namespace}')'"))) | .spec.name'
   ```

This RBAC model provides fine-grained control over who can create CoDriverTool, define security contexts, and restrict tools to specific namespaces while maintaining usability for different user roles.
