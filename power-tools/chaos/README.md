# Chaos Engineering Power Tool

A comprehensive chaos engineering tool for testing application resilience in Kubernetes environments. This tool runs as an ephemeral container sharing the process namespace with target containers, allowing for controlled chaos experiments.

## Overview

The chaos engineering power tool provides multiple types of chaos experiments:

- **Process Chaos**: Manipulate target processes (suspend, terminate, OOM)
- **CPU Chaos**: Generate CPU load and stress
- **Storage Chaos**: Fill storage to test disk pressure scenarios
- **Network Chaos**: Test network connectivity and latency patterns
- **Memory Chaos**: Create memory pressure and allocation patterns

## Chaos Experiment Types

### Process Chaos

Manipulates target processes within the shared process namespace.

**Actions:**
- `suspend`: Suspend and resume processes cyclically
- `terminate-graceful`: Send SIGTERM to process
- `terminate-force`: Send SIGKILL to process  
- `oom`: Trigger out-of-memory conditions

**Usage:**
```yaml
tool:
  name: "chaos"
  duration: "1m"
  args:
    - "process"
    - "suspend"  # action
```

### CPU Chaos

Generates CPU load to test application behavior under resource pressure.

**Parameters:**
- CPU percentage (default: 80%)
- Automatically scales with available cores

**Usage:**
```yaml
tool:
  name: "chaos"
  duration: "2m"
  args:
    - "cpu"
    - "90"  # CPU percentage
```

### Storage Chaos

Creates storage pressure by filling available disk space.

**Parameters:**
- Fill percentage (default: 80%)
- Target path (default: /tmp)

**Usage:**
```yaml
tool:
  name: "chaos"
  duration: "1m"
  args:
    - "storage"
    - "70"        # fill percentage
    - "/tmp"      # target path
```

### Network Chaos

Tests network connectivity and measures network behavior patterns.

**Actions:**
- `connectivity`: Test connection success/failure rates
- `dns`: Test DNS resolution patterns
- `latency`: Measure network latency over time
- `bandwidth`: Test download speeds and throughput
- `monitor`: Monitor network interfaces and connections

**Usage:**
```yaml
tool:
  name: "chaos"
  duration: "2m"
  args:
    - "network"
    - "connectivity"  # action
    - "google.com"    # target host
    - "80"            # port
```

### Memory Chaos

Creates memory pressure using various allocation patterns.

**Actions:**
- `pressure`: Allocate memory with different patterns
- `leak`: Simulate memory leak scenarios
- `fragmentation`: Create memory fragmentation
- `monitor`: Monitor memory usage patterns

**Patterns (for pressure action):**
- `linear`: Gradual memory allocation over time
- `spike`: Immediate large allocation
- `oscillating`: Cyclical allocation and release

**Usage:**
```yaml
tool:
  name: "chaos"
  duration: "1m"
  args:
    - "memory"
    - "pressure"  # action
    - "100"       # memory MB
    - "linear"    # pattern
```

## Environment Variables

The chaos tool supports several environment variables for configuration:

- `CHAOS_TYPE`: Type of chaos experiment (process, cpu, storage, network, memory)
- `DURATION`: Experiment duration (default: 30s)
- `INTERVAL`: Interval for cyclical operations (default: 5s)
- `TARGET_PID`: Target process ID for process chaos (default: 1)
- `OUTPUT_FILE`: Output file path (auto-generated if not specified)
- `COLLECTOR_URL`: Collector URL for result submission

## Example Configurations

### Basic Process Suspension
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
    mode: "pvc"
    pvcName: "chaos-results"
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

## Kubernetes Constraints

Since this tool runs as an ephemeral container in Kubernetes:

1. **Process Namespace Sharing**: Can see and interact with processes in the target container
2. **Network Namespace Sharing**: Shares network stack with target container
3. **Limited Privileges**: Cannot modify system-level network settings (iptables, tc) without additional capabilities
4. **File System**: Limited write access, primarily to /tmp and mounted volumes
5. **Resource Limits**: Constrained by the ephemeral container's resource limits

## Security Considerations

The chaos tool requires specific capabilities:

```yaml
securityContext:
  capabilities:
    add:
      - SYS_PTRACE  # For process manipulation
      - KILL        # For sending signals to processes
```

## Output and Monitoring

All chaos experiments generate detailed logs including:

- Experiment parameters and configuration
- Real-time progress and status updates
- Resource usage measurements where available
- Success/failure rates for network tests
- Cleanup confirmation

Results are automatically sent to the configured collector or stored according to the output mode specified in the CoDriverJob configuration.

## Best Practices

1. **Start Small**: Begin with short durations and low intensity
2. **Monitor Impact**: Watch application metrics during experiments
3. **Gradual Escalation**: Increase chaos intensity gradually
4. **Cleanup Verification**: Ensure experiments clean up properly
5. **Documentation**: Document experiment results and application behavior
6. **Safety Limits**: Set appropriate resource limits for the chaos container

## Troubleshooting

Common issues and solutions:

- **Permission Denied**: Ensure proper RBAC and security context capabilities
- **Process Not Found**: Verify target PID exists and is accessible
- **Network Tools Missing**: Some network tests require specific tools (nc, wget, ping)
- **Resource Limits**: Increase memory/CPU limits if chaos experiments are constrained
- **Cleanup Issues**: Check for remaining temporary files in /tmp after experiments
