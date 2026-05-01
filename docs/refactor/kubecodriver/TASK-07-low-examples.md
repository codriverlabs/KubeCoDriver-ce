# TASK-07: Examples YAML Fixes

**Severity**: 🟢 LOW — Example files only
**Files**: 4

## Replacement Rules

- `codriverjob.toe.run/` → `codriverjob.kubecodriver.codriverlabs.ai/`
- `chaos.toe.run/` → `chaos.kubecodriver.codriverlabs.ai/`
- `test.toe.run/` → `test.kubecodriver.codriverlabs.ai/`
- `localhost:32000/codriverlabs/toe/` → `localhost:32000/codriverlabs/ce/kubecodriver-`
- `localhost:32000/toe/` → `localhost:32000/codriverlabs/ce/kubecodriver-`
- `toe-test` namespace → `kubecodriver-test`

## Files & Lines

### 1. `examples/configs/powertoolconfig-examples.yaml`
- Lines 8-9: `codriverjob.toe.run/` annotations
- Line 12: image `localhost:32000/codriverlabs/toe/aperf:v1.0.12`
- Lines 34-35: `codriverjob.toe.run/` annotations
- Line 38: image `localhost:32000/codriverlabs/toe/aperf:v1.1.0`
- Lines 60-62: `codriverjob.toe.run/` annotations
- Line 65: image `localhost:32000/codriverlabs/toe/debugger:v1.0.0`
- Lines 83-84: `codriverjob.toe.run/` annotations
- Line 87: image `localhost:32000/codriverlabs/toe/team-profiler:v1.1.0`

### 2. `examples/chaos/powertool-chaos-workflow.yaml`
- Lines 7-8: `chaos.toe.run/` annotations

### 3. `examples/aperf/powertool-conflict-test.yaml`
- Line 13: image `localhost:32000/toe/aperf:v1.0.5`
- Lines 29, 55: namespace `toe-test`
- Lines 31-32, 57-58: `test.toe.run/` annotations

## Validation
```bash
grep -rn 'toe' examples/ --include='*.yaml'
# Expected: no output
```
