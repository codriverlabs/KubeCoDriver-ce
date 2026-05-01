# Non-Root User Test Results

## Test Execution Summary

**Date**: 2025-10-31  
**Test**: Ephemeral container creation with non-root pod (user 1001:1001)

## Test Setup

- **Pod**: `busybox-nonroot` in `kubecodriver-test` namespace
- **Pod Security Context**: 
  - `runAsUser: 1001`
  - `runAsGroup: 1001`
  - `fsGroup: 1001`
  - `runAsNonRoot: true`
- **CoDriverJob**: `aperf-nonroot-test`
- **Tool**: aperf with collector output mode

## Results

### ✅ Ephemeral Container Created Successfully

The controller successfully created the ephemeral container:
- Container name: `powertool-aperf-nonroot-test-7af41f70`
- Image: `localhost:32000/codriverlabs/ce/kubecodriver-aperf:v1.0.47`
- Security context includes capabilities (SYS_PTRACE, PERFMON, SYS_ADMIN)

### ❌ Container Failed with Permission Error

**Exit Code**: 1  
**Error**: `Permission denied (os error 13)`

**Container Logs**:
```
Starting AWS aperf profiling...
Target Pod: busybox-nonroot
Target Container: default
Duration: 10s
Warmup: 0s
Profile Type: cpu
Run Name: profile-busybox-nonroot-busybox-nonroot-20251031-223011
Output Directory: /tmp/
Error: Permission denied (os error 13)
```

## Analysis

### Security Context Inheritance

**Pod-level security context**:
```yaml
securityContext:
  runAsUser: 1001
  runAsGroup: 1001
  fsGroup: 1001
  runAsNonRoot: true
```

**Ephemeral container security context** (from controller):
```yaml
securityContext:
  capabilities:
    add: [SYS_PTRACE, PERFMON, SYS_ADMIN]
    drop: [ALL]
  privileged: false
```

**Missing**: `runAsUser`, `runAsGroup`, `runAsNonRoot` fields

### Root Cause

According to Kubernetes behavior:
1. ✅ Pod-level security context SHOULD be inherited by ephemeral containers
2. ❌ However, the aperf tool is failing with permission denied

**Possible causes**:
1. The aperf tool image may have files/directories owned by root
2. The tool may be trying to access resources that require root
3. The tool may need explicit user context set in the container spec

## Conclusion

**Status**: ⚠️ **Partially Working - Needs Fix**

The controller successfully creates ephemeral containers, but they fail when the target pod runs as non-root user.

**Recommendation**: Implement the hybrid approach from `docs/non-root-user-analysis.md`:
1. Extend `SecuritySpec` to include `runAsUser`, `runAsGroup`, `runAsNonRoot`
2. Auto-inherit from target container when not specified
3. Update `buildSecurityContext()` to set these fields explicitly

This will ensure the ephemeral container runs with the same user context as the target pod.
