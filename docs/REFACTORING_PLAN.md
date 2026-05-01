# Refactoring Plan: KubeCoDriver to KubeCoDriver

This document outlines the comprehensive analysis and detailed plan for refactoring the project from **KubeCoDriver** to **KubeCoDriver**.

## 1. Project Identity & Domains
- **Current Product Name:** KubeCoDriver
- **New Product Name:** KubeCoDriver
- **Current Group/Domain:** `kubecodriver.codriverlabs.ai`
- **New Group/Domain:** `kubecodriver.codriverlabs.ai`
- **Current Module Name:** `toe`
- **New Module Name:** `github.com/codriverlabs/KubeCoDriver`

## 2. CRD Kind Renames
All job and tool definitions are rebranded under the KubeCoDriver identity:

| Current Kind     | New Kind        | Purpose                                      |
|------------------|-----------------|----------------------------------------------|
| `CoDriverJob`      | `CoDriverJob`   | Declares a profiling/diagnostic/chaos job    |
| `CoDriverTool`| `CoDriverTool`  | Admin-defined tool template                  |

Full API version after refactor: `kubecodriver.codriverlabs.ai/v1alpha1`

## 3. Kubernetes Architecture
- **Namespace:** `kubecodriver-system` → `kubecodriver-system`
- **Controller Name:** `kubecodriver-controller` → `kubecodriver-controller`
- **Collector Name:** `kubecodriver-collector` → `kubecodriver-collector`
- **Token Audience:** `kubecodriver-sdk-collector` → `kubecodriver-sdk-collector`
- **API Resources:** `powertools.kubecodriver.codriverlabs.ai` → `codriverjobs.kubecodriver.codriverlabs.ai`

## 4. Container Images (ghcr.io/codriverlabs/ce/)
- `kubecodriver-controller` → `kubecodriver-controller`
- `kubecodriver-collector` → `kubecodriver-collector`
- `kubecodriver-aperf` → `kubecodriver-aperf`
- `kubecodriver-tcpdump` → `kubecodriver-tcpdump`
- `kubecodriver-chaos` → `kubecodriver-chaos`

## 5. File Structure Impact
- **Helm Chart:** `helm/kubecodriver-operator/` → `helm/kubecodriver-operator/` (directory rename + contents)
- **CRD Bases:** `config/crd/bases/kubecodriver.codriverlabs.ai_*.yaml` → `config/crd/bases/kubecodriver.codriverlabs.ai_*.yaml`
- **Kind Manifests:** `test/e2e-kind/manifests/kubecodriver-controller.yaml` → `test/e2e-kind/manifests/kubecodriver-controller.yaml`
- **Helper YAML:** `helper_scripts/collector/kubecodriver-collector-pvc-inspector.yaml` → `kubecodriver-collector-pvc-inspector.yaml`

## 6. Git Baseline
The codebase has been tagged at `kubecodriver-1.0.0` before this refactor begins. No rollback is planned; all features will be branded as KubeCoDriver going forward.

---

## Detailed Refactoring Plan

### Phase 1: Codebase & Modules
1. Update `go.mod`: Change `module toe` to `module github.com/codriverlabs/KubeCoDriver`.
2. Update all `.go` files: Replace all imports starting with `"toe/"` with `"github.com/codriverlabs/KubeCoDriver/"`.
3. Update hardcoded strings in Go source files:
   - `pkg/collector/server/server.go`: `"kubecodriver-sdk-collector"` → `"kubecodriver-sdk-collector"`
   - `internal/controller/powertool_controller.go`: `"kubecodriver-system"` and `"kubecodriver-sdk-collector"` → new values
   - Any other `.go` files referencing `kubecodriver-system` or `kubecodriver-sdk-collector` directly
4. Update `PROJECT` (Kubebuilder config): Change `domain`, `group`, `projectName`, `repo`, and `path` fields under `resources`.
5. Update `api/v1alpha1/groupversion_info.go`: Update the `+groupName` marker and `Group` in `GroupVersion`.
6. Rename Go types: `CoDriverJob` → `CoDriverJob`, `CoDriverTool` → `CoDriverTool` across all `.go` files, including generated deepcopy and test files.
7. Run `go mod tidy` to regenerate `go.sum` after the module rename.

### Phase 2: Configuration & Manifests
1. **Search & Replace (Case-Insensitive):** Replace `toe` with `kubecodriver` in all YAML and script files, ensuring appropriate case preservation (e.g., `KubeCoDriver` → `KubeCoDriver`).
2. **Kustomize:** Update `config/` directory — namespaces, image names, service account references, and namespace references in `config/certmanager/collector-certificate.yaml`.
3. **config/samples/:** Update `apiVersion` from `kubecodriver.codriverlabs.ai/v1alpha1` to `kubecodriver.codriverlabs.ai/v1alpha1` and kind names to `CoDriverJob` / `CoDriverTool`.
4. **deploy/collector/:** Update `deployment.yaml`, `configmap.yaml`, `pvc.yaml`, `debug-pod.yaml` for new image names and namespace.
5. **Helm:** Rename the chart directory (`helm/kubecodriver-operator/` → `helm/kubecodriver-operator/`) and update `Chart.yaml`, `values.yaml`, `values-eks.yaml`, and all templates.
6. **examples/:** Update all YAML files — `apiVersion`, kind names (`CoDriverJob` → `CoDriverJob`, `CoDriverTool` → `CoDriverTool`), and any namespace references.

### Phase 3: Scripts & Build Tools
1. **Makefile:** Update all image variables (`CONTROLLER_IMG`, `COLLECTOR_IMG`, etc.), build targets, and GitHub URLs.
2. **cicd-scripts/config.env:** Update all image name variables and any `toe`-prefixed values.
3. **CI/CD:** Update `.github/workflows/` to use new image names and cluster prefixes (`kubecodriver-e2e-`).
4. **Helper Scripts:** Update `cicd-scripts/` and `helper_scripts/` to reflect new names and paths.
5. **Root-level scripts:** Update `render_helm_chart.sh`, `configure-image-pull-secrets.sh`, `switch-to-test-namespace.sh`, `switch-to-controllers-namespace.sh`.

### Phase 4: Documentation & Metadata
1. Update `README.md`, `DEPLOYMENT.md`, `DEPLOYMENT-EKS.md`, `SECURITY.md`, `CHANGELOG.md`, and `AGENTS.md`.
2. Update the entire `docs/` directory tree, including:
   - `docs/architecture/`, `docs/collector/`, `docs/controller/`, `docs/security/`, `docs/testing/`
   - Security docs reference `kubecodriver-system` namespace and `kubecodriver-sdk-collector` audience extensively — update all occurrences.
3. Update the entire `roadmap/` directory — all markdown files reference `toe`-branded names.
4. Update all architectural diagrams (Mermaid) within documentation files.
5. Update the `.agents/summary/` directory to reflect the new project identity and CRD kind names.

### Phase 5: File Operations & Validation
1. Rename all directories and files containing `toe` or `KubeCoDriver` (including `helm/kubecodriver-operator/` directory).
2. Rename `helper_scripts/collector/kubecodriver-collector-pvc-inspector.yaml`.
3. Rename `test/e2e-kind/manifests/kubecodriver-controller.yaml` and any other manifests in that directory.
4. Execute `make generate` and `make manifests` to synchronize generated code (`zz_generated.deepcopy.go`, CRD YAML bases) with the new API group and kind names.
5. Run `go mod tidy` to ensure `go.sum` is consistent with the renamed module.
6. Run `go test ./...` to verify that all imports and logic remain consistent.
7. Run `make docker-build-all` to confirm all images build cleanly under the new names.
