# Non-Root User Security Context Test

## Quick Start

Run the automated test script:

```bash
./test-nonroot-scenario.sh
```

## Manual Testing Steps

### 1. Deploy Test Pod

```bash
kubectl apply -f examples/test-nonroot-pod.yaml
```

This creates a busybox pod running as user 1001:1001 with:
- `runAsUser: 1001`
- `runAsGroup: 1001`
- `fsGroup: 1001`
- `runAsNonRoot: true`

### 2. Verify Pod is Running

```bash
kubectl get pod busybox-nonroot -n kubecodriver-test
kubectl exec busybox-nonroot -n kubecodriver-test -- id
```

Expected output: `uid=1001 gid=1001`

### 3. Apply CoDriverJob

```bash
kubectl apply -f examples/test-nonroot-powertool.yaml
```

### 4. Monitor Controller Logs

```bash
kubectl logs -n kubecodriver-system -l control-plane=controller-manager -f
```

Look for:
- "Reconciling CoDriverJob" messages
- "Successfully added ephemeral container" messages
- Any error messages

### 5. Check Ephemeral Container

```bash
# Check if ephemeral container was created
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers}' | jq .

# Check ephemeral container security context
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .

# Check ephemeral container status
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.status.ephemeralContainerStatuses}' | jq .
```

### 6. Check CoDriverJob Status

```bash
kubectl get codriverjob aperf-nonroot-test -n kubecodriver-test -o yaml
```

## What to Look For

### ✅ Success Indicators

1. **Ephemeral container created**:
   ```json
   {
     "name": "aperf-nonroot-test-xxxxx",
     "image": "...",
     "securityContext": { ... }
   }
   ```

2. **Ephemeral container running**:
   ```json
   {
     "state": {
       "running": {
         "startedAt": "..."
       }
     }
   }
   ```

3. **Security context inherited** (if pod-level):
   - Ephemeral container should run as user 1001
   - Check with: `kubectl exec busybox-nonroot -n kubecodriver-test -c <ephemeral-container-name> -- id`

### ❌ Failure Indicators

1. **No ephemeral container created**:
   - Check controller logs for errors
   - Check CoDriverJob status for error messages

2. **Ephemeral container in waiting state**:
   ```json
   {
     "state": {
       "waiting": {
         "reason": "...",
         "message": "..."
       }
     }
   }
   ```

3. **Permission errors in logs**:
   - "permission denied"
   - "operation not permitted"
   - "user mismatch"

## Expected Behavior

### Current Implementation

Since the pod has **pod-level** security context:
- ✅ Ephemeral container SHOULD inherit `runAsUser: 1001`
- ✅ Ephemeral container SHOULD inherit `runAsGroup: 1001`
- ✅ Ephemeral container SHOULD inherit `fsGroup: 1001`

### If It Fails

If the ephemeral container fails to start or has permission issues:
- The controller needs to be updated to explicitly set security context
- See `docs/non-root-user-analysis.md` for implementation options

## Cleanup

```bash
kubectl delete -f examples/test-nonroot-powertool.yaml
kubectl delete -f examples/test-nonroot-pod.yaml
```

## Troubleshooting

### Controller Not Creating Ephemeral Container

```bash
# Check controller logs
kubectl logs -n kubecodriver-system -l control-plane=controller-manager --tail=100

# Check CoDriverJob events
kubectl describe powertool aperf-nonroot-test -n kubecodriver-test

# Check if aperf-config exists
kubectl get powertoolconfig -n kubecodriver-system
```

### Ephemeral Container Stuck in Waiting

```bash
# Check container status
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.status.ephemeralContainerStatuses[0].state}' | jq .

# Check pod events
kubectl describe pod busybox-nonroot -n kubecodriver-test
```

### Permission Denied Errors

This indicates the ephemeral container is running as a different user than expected.

Check:
```bash
# Pod security context
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.securityContext}' | jq .

# Ephemeral container security context
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .
```
