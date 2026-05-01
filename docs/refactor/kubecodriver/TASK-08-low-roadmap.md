# TASK-08: Roadmap Documentation Fixes

**Severity**: 🟢 LOW — Historical documentation
**Files**: 12+

## Replacement Rules

- `toe-test` namespace → `kubecodriver-test`
- `localhost:32000/codriverlabs/toe/` → `localhost:32000/codriverlabs/ce/kubecodriver-`
- `toev1alpha1` → `kubecodriverv1alpha1`
- `toe/internal/controller` → `github.com/codriverlabs/KubeCoDriver/internal/controller`
- `"toe-e2e-"` → `"kubecodriver-e2e-"`
- `powertool` (kubectl resource kind) → `codriverjob`

## Files

### 1. `roadmap/use-cases/aperf-root-requirement.md`
- Lines 18-19, 28-30: `toe-test` namespace, `toe/aperf` image
- Lines 61, 147: `toe/aperf` image

### 2. `roadmap/use-cases/test-results-final.md`
- Lines 217, 222, 227, 232, 237: `toe-test` namespace

### 3. `roadmap/use-cases/test-plan.md`
- Line 7: `toe-test` namespace
- Lines 20, 23, 35, 38, 41, 44, 75, 78, 82, 86, 99, 102, 105, 135, 143, 162, 165, 178, 208, 217, 228: `toe-test` namespace
- Line 44: double `toe-test` reference

### 4. `roadmap/test-results-nonroot.md`
- Line 10: `toe-test` namespace
- Line 25: `toe/aperf` image

### 5. `roadmap/unit_test/test-improvement-plan.md`
- Line 368: `toev1alpha1` type alias

### 6. `roadmap/runasroot-implementation-summary.md`
- Line 61: `toe/internal/controller` module path
- Line 76: `toe/aperf` image

### 7. `roadmap/e2e_tests/envtest_e2e_plan.md`
- Line 87: `"toe-e2e-"` prefix

### 8. `roadmap/run-as-root-feature.md`
- Lines 45, 169: `toev1alpha1` type references
- Line 147: `toe/aperf` image

### 9. `roadmap/container-selection-fix.md`
- Lines 66, 147: `toev1alpha1` type references
- Lines 314, 340: `toe-test` namespace

### 10. `roadmap/non-root-security-context-fix.md`
- Line 47: `toev1alpha1` type reference

### 11. `roadmap/container-selection-summary.md`
- Lines 147, 156: `toe-test` namespace

### 12. `roadmap/TODO-LIST.md`, `roadmap/README.md`
- Check for any remaining `toe` references

## Validation
```bash
grep -rn 'toe' roadmap/ --include='*.md'
# Expected: no output
```
