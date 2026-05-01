# TASK-05: Security Documentation Fixes

**Severity**: 🟢 LOW — Documentation only
**Files**: 5

## Replacement Rules

Apply these substitutions across all files listed below:
- `toe-k8s-operator` → `KubeCoDriver`
- `security.toe.run/` → `security.kubecodriver.codriverlabs.ai/`
- `tool.toe.run/` → `tool.kubecodriver.codriverlabs.ai/`
- `powertool.toe.run/` → `codriverjob.kubecodriver.codriverlabs.ai/`
- `toe-manager-role` → `kubecodriver-manager-role`
- `toe-manager-rolebinding` → `kubecodriver-manager-rolebinding`
- `localhost:32000/toe/` → `localhost:32000/codriverlabs/ce/kubecodriver-`

## Files & Lines

### 1. `docs/security/security-model.md` — Line 5
`toe-k8s-operator` → `KubeCoDriver`

### 2. `docs/security/README.md` — Line 3
`toe-k8s-operator` → `KubeCoDriver`

### 3. `docs/security/powertoolconfig-security.md`
- Line 54: image path
- Lines 150-152: `security.toe.run/` annotations
- Lines 265-267: `security.toe.run/` annotations
- Lines 281-283: `tool.toe.run/` annotations

### 4. `docs/security/controller-rbac.md`
- Line 82: `toe-manager-role`
- Line 170: `toe-manager-rolebinding`
- Line 174: `toe-manager-role`
- Lines 279-280: kubectl commands with old names

### 5. `docs/security/rbac-deployment.md`
- Line 187: `powertool.toe.run/access-level`
- Lines 435-437: `powertool.toe.run/` annotations

## Validation
```bash
grep -rn 'toe' docs/security/ --include='*.md'
# Expected: no output
```
