# KubeCoDriver Refactoring — Remaining Tasks

**Date**: 2026-03-22
**Baseline Tag**: `kubecodriver-1.0.0`
**Target Tag**: `kubecodriver-1.0.1`

## Status

The TOE → KubeCoDriver rename is ~80% complete. Go types, module path, CRD files, and Helm chart are done. The remaining issues are in Kubernetes manifests, runtime hardcoded strings, tests, docs, and examples.

## Task Breakdown

Tasks are designed for parallel execution by independent agents.

| Task | File | Severity | Parallelizable |
|------|------|----------|----------------|
| [TASK-01](TASK-01-critical-go-runtime.md) | Go source (2 files) | 🔴 CRITICAL | ✅ Yes |
| [TASK-02](TASK-02-critical-kustomize-config.md) | config/ manifests (10 files) | 🔴 CRITICAL | ✅ Yes |
| [TASK-03](TASK-03-critical-powertools-config.md) | power-tools/ configs (3 files) | 🔴 CRITICAL | ✅ Yes |
| [TASK-04](TASK-04-moderate-tests.md) | Test files (5 files) | 🟡 MODERATE | ✅ Yes |
| [TASK-05](TASK-05-low-docs-security.md) | docs/security/ (5 files) | 🟢 LOW | ✅ Yes |
| [TASK-06](TASK-06-low-docs-testing.md) | docs/testing/ (3 files) | 🟢 LOW | ✅ Yes |
| [TASK-07](TASK-07-low-examples.md) | examples/ (4 files) | 🟢 LOW | ✅ Yes |
| [TASK-08](TASK-08-low-roadmap.md) | roadmap/ (12+ files) | 🟢 LOW | ✅ Yes |

## Dependency Graph

```
TASK-01 ─┐
TASK-02 ─┤
TASK-03 ─┼─→ All independent, run in parallel
TASK-04 ─┤
TASK-05 ─┤
TASK-06 ─┤
TASK-07 ─┤
TASK-08 ─┘
         │
         ▼
    Validate (make build, grep -r 'toe')
         │
         ▼
    git tag kubecodriver-1.0.1
```

## Naming Convention

All replacements follow these rules:
- `toe-` prefix → `kubecodriver-` (resource names, kustomize prefix, image names)
- `toe.run` domain → `kubecodriver.codriverlabs.ai` (leader election, finalizers, annotations)
- `*.toe.run/` annotations → `*.kubecodriver.codriverlabs.ai/` (labels/annotations)
- `toe-k8s-operator` → `KubeCoDriver`
- `app.kubernetes.io/name: toe` → `app.kubernetes.io/name: kubecodriver`
- `toe-test` namespace → `kubecodriver-test`
- `toe-e2e-` prefix → `kubecodriver-e2e-`
- Image `toe-aperf` → `kubecodriver-aperf`, `toe-tcpdump` → `kubecodriver-tcpdump`, `toe-chaos` → `kubecodriver-chaos`
- Image path `codriverlabs/toe/` → `codriverlabs/ce/kubecodriver-`
