# Test Plan - Container Selection & Security Context Inheritance

## Prerequisites

- ✅ Controller rebuilt and redeployed
- ✅ Collector rebuilt and redeployed (if needed)
- ✅ `kubecodriver-test` namespace exists

## Test 1: Non-Root Single Container Pod

**Purpose**: Verify security context inheritance from pod-level

### Steps

```bash
# 1. Deploy non-root pod
kubectl apply -f examples/test-nonroot-pod.yaml

# 2. Wait for ready
kubectl wait --for=condition=Ready pod/busybox-nonroot -n kubecodriver-test --timeout=60s

# 3. Verify pod is running as user 1001
kubectl exec busybox-nonroot -n kubecodriver-test -- id

# 4. Apply CoDriverJob
kubectl apply -f examples/test-nonroot-powertool.yaml

# 5. Wait a few seconds
sleep 10

# 6. Check controller logs
kubectl logs -n kubecodriver-system -l control-plane=controller-manager --tail=50 | grep -E "busybox-nonroot|Inherited"

# 7. Check ephemeral container
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0]}' | jq .

# 8. Check security context
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .

# 9. Check container status
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.status.ephemeralContainerStatuses[0]}' | jq .

# 10. Check container logs (should not have permission errors)
kubectl logs busybox-nonroot -n kubecodriver-test -c $(kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].name}')
```

### Expected Results

- ✅ Ephemeral container created
- ✅ Security context shows `runAsUser: 1001`, `runAsGroup: 1001`
- ✅ Controller logs show "Inherited runAsUser from pod"
- ✅ Container runs successfully (no permission errors)
- ✅ Container status is "running" or "terminated" with exit code 0

### Cleanup

```bash
kubectl delete -f examples/test-nonroot-powertool.yaml
kubectl delete -f examples/test-nonroot-pod.yaml
```

---

## Test 2: Multi-Container Pod with Target Container

**Purpose**: Verify container selection and security context inheritance from target container

### Steps

```bash
# 1. Deploy multi-container pod
kubectl apply -f examples/test-multicontainer-pod.yaml

# 2. Wait for ready
kubectl wait --for=condition=Ready pod/multi-container-test -n kubecodriver-test --timeout=60s

# 3. Verify containers are running
kubectl get pod multi-container-test -n kubecodriver-test -o jsonpath='{.spec.containers[*].name}'
# Should show: sidecar main-app

# 4. Check sidecar user
kubectl exec multi-container-test -n kubecodriver-test -c sidecar -- id
# Should show: uid=2000 gid=2000

# 5. Check main-app user
kubectl exec multi-container-test -n kubecodriver-test -c main-app -- id
# Should show: uid=1001 gid=1001

# 6. Apply CoDriverJob (targets main-app)
kubectl apply -f examples/test-multicontainer-powertool.yaml

# 7. Wait a few seconds
sleep 10

# 8. Check controller logs
kubectl logs -n kubecodriver-system -l control-plane=controller-manager --tail=50 | grep -E "multi-container|Target container|Inherited"

# 9. Check TARGET_CONTAINER_NAME env var
kubectl get pod multi-container-test -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].env[?(@.name=="TARGET_CONTAINER_NAME")]}' | jq .

# 10. Check ephemeral container security context
kubectl get pod multi-container-test -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .

# 11. Check container status
kubectl get pod multi-container-test -n kubecodriver-test -o jsonpath='{.status.ephemeralContainerStatuses[0]}' | jq .
```

### Expected Results

- ✅ Ephemeral container created
- ✅ Controller logs show "Target container identified: main-app"
- ✅ Controller logs show "Inherited runAsUser from target container: 1001"
- ✅ Security context shows `runAsUser: 1001`, `runAsGroup: 1001` (NOT 2000!)
- ✅ `TARGET_CONTAINER_NAME` env var is "main-app"
- ✅ Container runs successfully

### Cleanup

```bash
kubectl delete -f examples/test-multicontainer-powertool.yaml
kubectl delete -f examples/test-multicontainer-pod.yaml
```

---

## Test 3: Multi-Container Pod WITHOUT Target Container

**Purpose**: Verify fallback to first container when no target specified

### Steps

```bash
# 1. Deploy multi-container pod (reuse from Test 2)
kubectl apply -f examples/test-multicontainer-pod.yaml
kubectl wait --for=condition=Ready pod/multi-container-test -n kubecodriver-test --timeout=60s

# 2. Create CoDriverJob WITHOUT container field
cat <<EOF | kubectl apply -f -
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: aperf-multicontainer-nospec
  namespace: kubecodriver-test
spec:
  targets:
    labelSelector:
      matchLabels:
        app: multi-test
    # NO container field - should use first (sidecar)
  tool:
    name: "aperf"
    duration: "10s"
  output:
    mode: "ephemeral"
EOF

# 3. Wait and check logs
sleep 10
kubectl logs -n kubecodriver-system -l control-plane=controller-manager --tail=50 | grep -E "multi-container|Target container|Inherited"

# 4. Check security context
kubectl get pod multi-container-test -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .

# 5. Check TARGET_CONTAINER_NAME
kubectl get pod multi-container-test -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].env[?(@.name=="TARGET_CONTAINER_NAME")]}' | jq .
```

### Expected Results

- ✅ Ephemeral container created
- ✅ Controller logs show "Target container identified: sidecar"
- ✅ Security context shows `runAsUser: 2000`, `runAsGroup: 2000` (first container)
- ✅ `TARGET_CONTAINER_NAME` env var is "sidecar"

### Cleanup

```bash
kubectl delete codriverjob aperf-multicontainer-nospec -n kubecodriver-test
kubectl delete -f examples/test-multicontainer-pod.yaml
```

---

## Test 4: Automated Test Script

**Purpose**: Run all tests automatically

```bash
# Run the non-root test script
./test-nonroot-scenario.sh
```

### Expected Results

- ✅ All steps complete successfully
- ✅ Ephemeral container created
- ✅ No permission errors
- ✅ Controller logs show inheritance

---

## Troubleshooting

### Issue: Ephemeral container not created

**Check**:
```bash
kubectl describe codriverjob <name> -n kubecodriver-test
kubectl logs -n kubecodriver-system -l control-plane=controller-manager --tail=100
```

### Issue: Permission denied errors

**Check**:
```bash
# Verify security context was inherited
kubectl get pod <pod-name> -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .

# Check controller logs for inheritance messages
kubectl logs -n kubecodriver-system -l control-plane=controller-manager | grep "Inherited"
```

### Issue: Wrong container targeted

**Check**:
```bash
# Verify TARGET_CONTAINER_NAME
kubectl get pod <pod-name> -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].env[?(@.name=="TARGET_CONTAINER_NAME")]}' | jq .

# Check controller logs
kubectl logs -n kubecodriver-system -l control-plane=controller-manager | grep "Target container"
```

---

## Success Criteria Summary

### Test 1: Non-Root Single Container
- [x] Ephemeral container runs as user 1001
- [x] No permission errors
- [x] Controller logs show inheritance

### Test 2: Multi-Container with Target
- [x] Targets correct container (main-app)
- [x] Inherits security context from main-app (1001)
- [x] TARGET_CONTAINER_NAME is "main-app"

### Test 3: Multi-Container without Target
- [x] Falls back to first container (sidecar)
- [x] Inherits security context from sidecar (2000)
- [x] TARGET_CONTAINER_NAME is "sidecar"

### Test 4: Automated Script
- [x] All automated tests pass

---

## Notes

- All tests should be run in order
- Clean up between tests to avoid conflicts
- Controller logs are crucial for debugging
- Security context inheritance is logged at INFO level
l
s

---

## Notes

- All tests should be run in order
- Clean up between tests to avoid conflicts
- Controller logs are crucial for debugging
- Security context inheritance is logged at INFO level
