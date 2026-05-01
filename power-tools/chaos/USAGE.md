# Chaos Tool - Complete Usage Guide

## Overview

The chaos tool provides 5 types of chaos engineering experiments for testing application resilience in Kubernetes environments. It runs as an ephemeral container sharing the process namespace with target containers.

## Supported Chaos Types

### 1. Process Chaos

Manipulates target processes within the shared process namespace.

**Syntax:**
```yaml
args: ["process", "<action>"]
```

**Actions:**
- `suspend` - Suspend and resume processes cyclically ✅ **WORKING**
- `terminate-graceful` - Send SIGTERM to process ✅ **WORKING**
- `terminate-force` - Send SIGKILL to process ✅ **WORKING**
- `oom` - Trigger out-of-memory conditions ❌ **NOT SUPPORTED** (see limitations below)

**Important Notes:**
- **Suspend Action**: Cyclically sends SIGSTOP/SIGCONT signals to the target process, causing it to pause and resume. This simulates intermittent freezes and tests application resilience to process suspension.
- **Terminate Actions**: Send termination signals to the target process. The container will be restarted by Kubernetes if it has a restart policy.
- **OOM Limitation**: Due to Kubernetes cgroup isolation, ephemeral containers cannot trigger OOM on target containers. Memory allocated by the chaos tool counts against its own cgroup, not the target's. This is a fundamental limitation of the ephemeral container architecture.

**Examples:**
```yaml
# Suspend process cyclically
tool:
  name: "chaos"
  duration: "1m"
  args: ["process", "suspend"]

# Graceful termination
tool:
  name: "chaos"
  duration: "10s"
  args: ["process", "terminate-graceful"]

# Force kill
tool:
  name: "chaos"
  duration: "5s"
  args: ["process", "terminate-force"]
```

## Limitations

### OOM Action Not Supported

The OOM action cannot reliably trigger out-of-memory kills on target containers due to Kubernetes' cgroup isolation:

**Why it doesn't work:**
- Each container has its own memory cgroup with separate accounting
- Ephemeral containers cannot allocate memory that counts against the target container's limit
- Even with privileged mode and SYS_ADMIN, memory accounting happens at the kernel level per cgroup
- The chaos container will OOM itself before affecting the target

**Alternatives:**
- Use `terminate-force` to simulate sudden process death
- Use `suspend` to simulate process freezes
- Implement memory pressure at the application level if needed

---

### 2. CPU Chaos

Generates CPU load to test application behavior under resource pressure.

**Syntax:**
```yaml
args: ["cpu", "<cpu_percent>"]
```

**Parameters:**
- `cpu_percent` - Target CPU usage percentage (default: 80)
  - Automatically scales with available cores
  - Range: 1-100

**Examples:**
```yaml
# 80% CPU load (default)
tool:
  name: "chaos"
  duration: "2m"
  args: ["cpu", "80"]

# 95% CPU stress
tool:
  name: "chaos"
  duration: "1m"
  args: ["cpu", "95"]

# Light CPU load
tool:
  name: "chaos"
  duration: "5m"
  args: ["cpu", "30"]
```

---

### 3. Storage Chaos

Creates storage pressure by filling available disk space.

**Syntax:**
```yaml
args: ["storage", "<fill_percent>", "<target_path>"]
```

**Parameters:**
- `fill_percent` - Percentage of storage to fill (default: 80)
- `target_path` - Path to fill (default: /tmp)

**Examples:**
```yaml
# Fill 80% of /tmp (default)
tool:
  name: "chaos"
  duration: "1m"
  args: ["storage", "80", "/tmp"]

# Fill 70% of /tmp
tool:
  name: "chaos"
  duration: "2m"
  args: ["storage", "70", "/tmp"]

# Fill 90% of custom path
tool:
  name: "chaos"
  duration: "30s"
  args: ["storage", "90", "/var/data"]
```

---

### 4. Network Chaos

Tests network connectivity and measures network behavior patterns.

**Syntax:**
```yaml
args: ["network", "<action>", "<target_host>", "<port>"]
```

**Actions:**
- `connectivity` - Test connection success/failure rates (default)
- `dns` - Test DNS resolution patterns
- `latency` - Measure network latency over time
- `bandwidth` - Test download speeds and throughput
- `monitor` - Monitor network interfaces and connections

**Parameters:**
- `target_host` - Host to test (default: 8.8.8.8)
- `port` - Port number (default: 80)

**Examples:**
```yaml
# Test connectivity to external service
tool:
  name: "chaos"
  duration: "3m"
  args: ["network", "connectivity", "api.example.com", "443"]

# DNS resolution test
tool:
  name: "chaos"
  duration: "2m"
  args: ["network", "dns", "google.com", "80"]

# Latency measurement
tool:
  name: "chaos"
  duration: "5m"
  args: ["network", "latency", "10.0.0.1", "8080"]

# Bandwidth test
tool:
  name: "chaos"
  duration: "1m"
  args: ["network", "bandwidth", "speedtest.net", "80"]

# Network monitoring
tool:
  name: "chaos"
  duration: "2m"
  args: ["network", "monitor"]
```

---

### 5. Memory Chaos

Creates memory pressure using various allocation patterns.

**Syntax:**
```yaml
args: ["memory", "<action>", "<memory_mb>", "<pattern>"]
```

**Actions:**
- `pressure` - Allocate memory with different patterns (default)
- `leak` - Simulate memory leak scenarios
- `fragmentation` - Create memory fragmentation
- `monitor` - Monitor memory usage patterns

**Patterns (for pressure action):**
- `linear` - Gradual memory allocation over time (default)
- `spike` - Immediate large allocation
- `oscillating` - Cyclical allocation and release

**Parameters:**
- `memory_mb` - Memory to allocate in MB (default: 100)
- `pattern` - Allocation pattern (default: linear)

**Examples:**
```yaml
# Linear memory pressure (100MB over duration)
tool:
  name: "chaos"
  duration: "1m"
  args: ["memory", "pressure", "100", "linear"]

# Memory spike (immediate allocation)
tool:
  name: "chaos"
  duration: "90s"
  args: ["memory", "pressure", "200", "spike"]

# Oscillating memory pattern
tool:
  name: "chaos"
  duration: "2m"
  args: ["memory", "pressure", "150", "oscillating"]

# Memory leak simulation
tool:
  name: "chaos"
  duration: "3m"
  args: ["memory", "leak", "50"]

# Memory fragmentation
tool:
  name: "chaos"
  duration: "1m"
  args: ["memory", "fragmentation", "100"]

# Memory monitoring
tool:
  name: "chaos"
  duration: "2m"
  args: ["memory", "monitor"]
```

---

## Complete CoDriverJob Examples

### Process Suspension Test
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: chaos-process-suspend
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  tool:
    name: "chaos"
    duration: "1m"
    args: ["process", "suspend"]
  output:
    mode: "collector"
```

### CPU Stress Test
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: chaos-cpu-stress
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  tool:
    name: "chaos"
    duration: "2m"
    args: ["cpu", "95"]
  output:
    mode: "ephemeral"
```

### Storage Pressure Test
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: chaos-storage-fill
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  tool:
    name: "chaos"
    duration: "1m"
    args: ["storage", "70", "/tmp"]
  output:
    mode: "pvc"
    pvcName: "chaos-results"
```

### Network Connectivity Test
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: chaos-network-test
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  tool:
    name: "chaos"
    duration: "3m"
    args: ["network", "connectivity", "api.example.com", "443"]
  output:
    mode: "collector"
```

### Memory Pressure Test
```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: chaos-memory-pressure
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-app
  tool:
    name: "chaos"
    duration: "90s"
    args: ["memory", "pressure", "200", "spike"]
  output:
    mode: "collector"
```

---

## Environment Variables

The chaos tool supports these environment variables (set automatically by KubeCoDriver):

- `CHAOS_TYPE` - Type of chaos experiment (process, cpu, storage, network, memory)
- `DURATION` - Experiment duration (default: 30s)
- `INTERVAL` - Interval for cyclical operations (default: 5s)
- `TARGET_PID` - Target process ID for process chaos (default: 1)
- `OUTPUT_FILE` - Output file path (auto-generated)
- `COLLECTOR_ENDPOINT` - Collector URL for result submission

---

## Security Requirements

The chaos tool requires specific capabilities:

```yaml
securityContext:
  capabilities:
    add:
      - SYS_PTRACE  # For process manipulation
      - KILL        # For sending signals to processes
      - NET_ADMIN   # For network operations
```

---

## Kubernetes Constraints

Since this tool runs as an ephemeral container:

1. **Process Namespace Sharing** - Can see and interact with processes in target container
2. **Network Namespace Sharing** - Shares network stack with target container
3. **Limited Privileges** - Cannot modify system-level network settings without additional capabilities
4. **File System** - Limited write access, primarily to /tmp and mounted volumes
5. **Resource Limits** - Constrained by ephemeral container's resource limits

---

## Best Practices

1. **Start Small** - Begin with short durations and low intensity
2. **Monitor Impact** - Watch application metrics during experiments
3. **Gradual Escalation** - Increase chaos intensity gradually
4. **Cleanup Verification** - Ensure experiments clean up properly
5. **Documentation** - Document experiment results and application behavior
6. **Safety Limits** - Set appropriate resource limits for the chaos container

---

## Output and Results

All chaos experiments generate detailed logs including:

- Experiment parameters and configuration
- Real-time progress and status updates
- Resource usage measurements where available
- Success/failure rates for network tests
- Cleanup confirmation

Results are automatically sent to the configured collector or stored according to the output mode specified in the CoDriverJob configuration.

---

## Troubleshooting

**Permission Denied**
- Ensure proper RBAC and security context capabilities are configured

**Process Not Found**
- Verify target PID exists and is accessible
- Check process namespace sharing is enabled

**Network Tools Missing**
- Some network tests require specific tools (nc, wget, ping)
- Consider using a custom image with required tools

**Resource Limits**
- Increase memory/CPU limits if chaos experiments are constrained
- Check ephemeral container resource allocation

**Cleanup Issues**
- Check for remaining temporary files in /tmp after experiments
- Verify proper signal handling in cleanup routines
