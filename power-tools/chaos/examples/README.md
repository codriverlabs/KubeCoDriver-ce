# Chaos Tool Examples

This directory contains example CoDriverJob manifests for the chaos engineering tool.

## Supported Process Chaos Actions

### 1. Process Suspend (`process-suspend.yaml`)

Cyclically suspends and resumes the target process to test resilience to intermittent freezes.

**Use case:** Test how your application handles temporary unresponsiveness

```bash
kubectl apply -f process-suspend.yaml
```

**What it does:**
- Sends SIGSTOP to pause the process
- Waits 5 seconds
- Sends SIGCONT to resume the process
- Repeats for the specified duration

### 2. Graceful Termination (`process-terminate-graceful.yaml`)

Sends SIGTERM to allow graceful shutdown of the target process.

**Use case:** Test graceful shutdown procedures and restart recovery

```bash
kubectl apply -f process-terminate-graceful.yaml
```

**What it does:**
- Sends SIGTERM to the target process
- Process can perform cleanup (close connections, flush buffers, etc.)
- Kubernetes restarts the container based on restart policy

### 3. Force Termination (`process-terminate-force.yaml`)

Sends SIGKILL to immediately terminate the target process without cleanup.

**Use case:** Test crash recovery and data consistency after sudden failures

```bash
kubectl apply -f process-terminate-force.yaml
```

**What it does:**
- Sends SIGKILL to immediately terminate the process
- No cleanup or graceful shutdown
- Simulates sudden crashes or power failures
- Kubernetes restarts the container based on restart policy

## Customization

All examples can be customized by modifying:

- `metadata.namespace`: Target namespace
- `spec.targets.labelSelector`: Pod selection criteria
- `spec.tool.duration`: How long the chaos runs
- `spec.output.mode`: Where to store results (ephemeral, pvc, collector)

## Monitoring Results

### Ephemeral Mode (default)
Results are stored in the ephemeral container and lost when it terminates:
```bash
kubectl logs <pod-name> -c <chaos-container-name>
```

### Check CoDriverJob Status
```bash
kubectl get powertool <name> -o yaml
```

### Check Pod Restarts
```bash
kubectl get pods -l app=my-app
```

Look for the `RESTARTS` column to see if containers were terminated and restarted.

## Safety Considerations

1. **Start in non-production**: Test chaos experiments in development/staging first
2. **Monitor impact**: Watch application metrics during experiments
3. **Short durations**: Start with short durations (30s-2m) and increase gradually
4. **Restart policies**: Ensure pods have appropriate restart policies configured
5. **Backup data**: For terminate actions, ensure data is backed up or replicated

## Not Supported

### OOM Action
The `oom` action is **not supported** with ephemeral containers due to Kubernetes cgroup isolation. See [USAGE.md](../USAGE.md#limitations) for details.

**Alternatives:**
- Use `terminate-force` to simulate sudden process death
- Use `suspend` to simulate process freezes
- Implement memory pressure testing at the application level
