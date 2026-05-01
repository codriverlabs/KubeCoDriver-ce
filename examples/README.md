# CoDriverJob Examples

This directory contains example CoDriverJob CRD configurations demonstrating how to use various power tools.

## Directory Structure

```
examples/
├── configs/                    # General CoDriverTool examples
│   └── powertoolconfig-examples.yaml
├── aperf/                      # Aperf performance profiling examples
│   ├── powertool-aperf-ephemeral.yaml
│   ├── powertool-aperf-pvc.yaml
│   ├── powertool-aperf-collector.yaml
│   └── powertool-conflict-test.yaml
├── chaos/                      # Chaos engineering examples
│   ├── powertool-chaos-cpu.yaml
│   ├── powertool-chaos-memory.yaml
│   ├── powertool-chaos-network.yaml
│   ├── powertool-chaos-process.yaml
│   ├── powertool-chaos-storage.yaml
│   ├── powertool-chaos-terminate.yaml
│   └── powertool-chaos-workflow.yaml
├── tcpdump/                    # Network packet capture examples
│   ├── powertool-tcpdump-ephemeral.yaml
│   ├── powertool-tcpdump-pvc.yaml
│   └── powertool-tcpdump-collector.yaml
├── debug/                      # Debug profiles for kubectl debug
│   ├── debug-root-profile.yaml
│   ├── debug-root-perf-profile.yaml
│   └── debug-root-with-group-profile.yaml
├── testing/                    # Test resources for development
│   ├── test-nonroot-pod.yaml
│   ├── test-nonroot-powertool.yaml
│   ├── test-multicontainer-pod.yaml
│   └── test-multicontainer-powertool.yaml
└── targets/                    # Example target pods
    ├── target-pod-with-pvc.yaml
    ├── target-statefulset.yaml
    └── target-statefulset-spot.yaml
```

## CoDriverTool Location

CoDriverTool files (tool definitions) are located in the `power-tools/` directory:
- `power-tools/aperf/config/powertoolconfig-aperf.yaml`
- `power-tools/chaos/config/powertoolconfig-chaos.yaml`
- `power-tools/tcpdump/config/powertoolconfig-tcpdump.yaml`

See `power-tools/README.md` for more information about tool configurations.

## CoDriverJob Examples

### Aperf (Performance Profiling)
- `aperf/powertool-aperf-ephemeral.yaml` - Ephemeral execution
- `aperf/powertool-aperf-pvc.yaml` - Output to persistent volume
- `aperf/powertool-aperf-collector.yaml` - Output to collector service
- `aperf/powertool-conflict-test.yaml` - Conflict detection testing

### Chaos Engineering
- `chaos/powertool-chaos-cpu.yaml` - CPU stress testing
- `chaos/powertool-chaos-memory.yaml` - Memory stress testing
- `chaos/powertool-chaos-network.yaml` - Network chaos injection
- `chaos/powertool-chaos-process.yaml` - Process termination
- `chaos/powertool-chaos-storage.yaml` - Storage I/O stress
- `chaos/powertool-chaos-terminate.yaml` - Container termination
- `chaos/powertool-chaos-workflow.yaml` - Multi-step chaos workflow

### TCPDump (Network Capture)
- `tcpdump/powertool-tcpdump-ephemeral.yaml` - Ephemeral packet capture
- `tcpdump/powertool-tcpdump-pvc.yaml` - Capture to persistent volume
- `tcpdump/powertool-tcpdump-collector.yaml` - Capture to collector

## Usage

1. **Deploy CoDriverTool first (from power-tools directory):**
   ```bash
   kubectl apply -f power-tools/aperf/config/powertoolconfig-aperf.yaml
   ```

2. **Deploy CoDriverJob (from examples directory):**
   ```bash
   kubectl apply -f examples/aperf/powertool-aperf-ephemeral.yaml
   ```

3. **Check status:**
   ```bash
   kubectl get powertools
   kubectl describe powertool aperf-ephemeral-example
   ```

## Testing Resources

The `testing/` directory contains resources for development and testing:
- `test-nonroot-pod.yaml` - Non-root test pod (user 1001:1001)
- `test-multicontainer-pod.yaml` - Multi-container test pod
- `test-*-powertool.yaml` - CoDriverJobs for testing scenarios

## Debug Profiles

The `debug/` directory contains custom profiles for `kubectl debug` testing:
- `debug-root-profile.yaml` - Basic root access (runAsUser: 0)
- `debug-root-perf-profile.yaml` - Root with performance capabilities (SYS_PTRACE, PERFMON, SYS_ADMIN)
- `debug-root-with-group-profile.yaml` - Root with group inheritance

These profiles demonstrate the `runAsRoot` feature implementation.

## Target Examples

The `targets/` directory contains example target pods for testing:
- `target-pod-with-pvc.yaml` - Pod with PVC mount
- `target-statefulset.yaml` - StatefulSet example
- `target-statefulset-spot.yaml` - StatefulSet with spot instances
