# Security Model

## Overview

The KubeCoDriver implements a multi-layered security model that separates administrative control from user execution, ensuring that security policies are enforced consistently across all power tool executions.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Administrator │    │      User        │    │   Controller    │
│                 │    │                  │    │                 │
│ Creates         │    │ Creates          │    │ Enforces        │
│ CoDriverTool │───▶│ CoDriverJob        │───▶│ Security        │
│ (Security)      │    │ (Execution)      │    │ (Runtime)       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Security Layers

### 1. Administrative Layer (CoDriverTool)
- **Who**: Cluster administrators only
- **What**: Define security contexts, capabilities, and privileges
- **Where**: CoDriverTool CRDs in system namespaces
- **Control**: RBAC restricts creation/modification

### 2. User Layer (CoDriverJob)
- **Who**: Application developers and operators
- **What**: Define execution parameters (duration, targets, output)
- **Where**: CoDriverJob CRDs in user namespaces
- **Restriction**: Cannot override security settings

### 3. Runtime Layer (Controller)
- **Who**: System controller
- **What**: Enforce security policies during execution
- **Where**: Ephemeral containers in target pods
- **Guarantee**: Only CoDriverTool security is applied

## Security Flow

1. **Admin Phase**:
   ```yaml
   # Administrator creates CoDriverTool
   apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
   kind: CoDriverTool
   spec:
     name: "aperf"
     security:
       allowPrivileged: false
       capabilities:
         add: ["SYS_PTRACE"]  # Minimal required capabilities
   ```

2. **User Phase**:
   ```yaml
   # User creates CoDriverJob (no security field allowed)
   apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
   kind: CoDriverJob
   spec:
     tool:
       name: "aperf"  # References admin-controlled config
     # security: FORBIDDEN - would be ignored/rejected
   ```

3. **Runtime Phase**:
   - Controller looks up CoDriverTool by tool name
   - Applies ONLY the security context from CoDriverTool
   - Creates ephemeral container with enforced security

## Threat Model

### Threats Mitigated

| Threat | Mitigation | Implementation |
|--------|------------|----------------|
| Privilege Escalation | Security only in CoDriverTool | API validation, controller enforcement |
| Capability Abuse | Minimal required capabilities | Admin-defined capability sets |
| Resource Access | RBAC boundaries | Service accounts, role bindings |
| Data Exfiltration | Controlled output modes | Collector authentication, PVC restrictions |
| Lateral Movement | Namespace isolation | RBAC, network policies |

### Attack Scenarios

1. **Malicious CoDriverJob Creation**:
   - User tries to add dangerous capabilities
   - System ignores user security settings
   - Only CoDriverTool security is applied

2. **CoDriverTool Tampering**:
   - RBAC prevents non-admin modification
   - Audit logs track all changes
   - Validation webhooks can enforce policies

3. **Controller Compromise**:
   - Minimal RBAC permissions
   - No cluster-admin privileges
   - Scoped to specific resources only

## Security Boundaries

### Namespace Boundaries
- CoDriverTool: System namespaces (kubecodriver-system)
- CoDriverJob: User namespaces (default, app-*)
- Collector: System namespace (kubecodriver-system)

### RBAC Boundaries
- Admins: Full CoDriverTool access
- Users: CoDriverJob creation only
- Controller: Minimal required permissions
- Collector: Token validation only

### Runtime Boundaries
- Ephemeral containers: Restricted capabilities
- Network access: Controlled endpoints
- File system: Read-only or specific mounts
- Process isolation: Container boundaries

## Compliance Considerations

### Audit Requirements
- All CoDriverTool changes logged
- CoDriverJob executions tracked
- Security violations recorded
- Access patterns monitored

### Policy Enforcement
- Admission controllers for validation
- OPA/Gatekeeper for advanced policies
- Network policies for traffic control
- Pod security standards compliance

## Best Practices

### For Administrators
1. Define minimal required capabilities per tool
2. Use separate CoDriverTool per security profile
3. Regular security reviews of tool configurations
4. Monitor for unusual capability requests

### For Users
1. Use existing CoDriverTool when possible
2. Request new tools through proper channels
3. Follow principle of least privilege
4. Report security concerns promptly

### For Operations
1. Regular RBAC audits
2. Monitor controller logs for violations
3. Implement network segmentation
4. Maintain security documentation current
