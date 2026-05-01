# TASK-06: Testing Documentation Fixes

**Severity**: 🟢 LOW — Documentation only
**Files**: 3

## Replacement Rules

- `toe-test` namespace → `kubecodriver-test`
- `toe-test-e2e` → `kubecodriver-test-e2e`
- `toe-e2e-` prefix → `kubecodriver-e2e-`
- `toev1alpha1` → `kubecodriverv1alpha1`
- `"^toe-e2e-"` → `"^kubecodriver-e2e-"`

## Files & Lines

### 1. `docs/testing/testing-setup.md`
- Line 23: `toev1alpha1` → `kubecodriverv1alpha1`
- Line 100: `toe-test-e2e` → `kubecodriver-test-e2e`

### 2. `docs/testing/e2e-kind-strategy.md`
- Line 13: `toe-e2e-` → `kubecodriver-e2e-`
- Line 26: `toe-e2e-` → `kubecodriver-e2e-`
- Line 359: `toe-e2e-` → `kubecodriver-e2e-`
- Line 441: `"^toe-e2e-"` → `"^kubecodriver-e2e-"`
- Lines 449, 452, 455: `toe-e2e-dev` → `kubecodriver-e2e-dev`

### 3. `docs/testing/test-nonroot-instructions.md`
- Lines 28, 29, 55, 58, 61, 67, 96, 152, 162, 165, 175, 178: all `toe-test` → `kubecodriver-test`
- Line 67: `get powertool` → `get codriverjob` (CRD kind fix)

## Validation
```bash
grep -rn 'toe' docs/testing/ --include='*.md'
# Expected: no output
```
