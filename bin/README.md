# Build Tools Directory

This directory contains build and development tools used by the Kubebuilder-based KubeCoDriver project.

## Tool Overview

### `controller-gen` - Kubernetes Controller Code Generator
**Purpose:** Core Kubebuilder tool for generating Kubernetes-specific code
**Tasks:**
- Generate CRD YAML manifests from Go struct tags
- Generate RBAC rules from `//+kubebuilder:rbac` comments
- Generate DeepCopy methods for API types
- Generate webhook configurations

**Usage in KubeCoDriver:**
```bash
# Generate CRDs and RBAC
controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate DeepCopy methods
controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
```

### `kustomize` - Kubernetes Configuration Management
**Purpose:** Declarative configuration management for Kubernetes
**Tasks:**
- Overlay configurations for different environments
- Patch and transform Kubernetes manifests
- Generate final deployment YAML from base + overlays
- Manage image tags and resource names

**Usage in KubeCoDriver:**
```bash
# Build final manifests
kustomize build config/default > dist/install.yaml

# Update image references
cd config/manager && kustomize edit set image controller=myregistry/kubecodriver:v1.0.0
```

### `golangci-lint` - Go Code Linter
**Purpose:** Comprehensive Go code quality and style checking
**Tasks:**
- Static code analysis (deadcode, unused variables, etc.)
- Style enforcement (gofmt, goimports)
- Security vulnerability detection
- Performance issue identification
- Code complexity analysis

**Usage in KubeCoDriver:**
```bash
# Run all configured linters
golangci-lint run

# Run with specific configuration
golangci-lint run --config .golangci.yml
```

### `setup-envtest` - Test Environment Manager
**Purpose:** Manages local Kubernetes API server for testing
**Tasks:**
- Download and install etcd + kube-apiserver binaries
- Start/stop local Kubernetes control plane for tests
- Manage multiple Kubernetes versions for compatibility testing
- Provide KUBEBUILDER_ASSETS environment variable

**Usage in KubeCoDriver:**
```bash
# Install test environment
setup-envtest use 1.33.x

# Get assets path for tests
export KUBEBUILDER_ASSETS=$(setup-envtest use -p path 1.33.x)
```

## Kubebuilder Best Practices

### 1. Tool Version Management
- **Pin versions:** Use specific tool versions for reproducible builds
- **Symlinks:** Create version-agnostic symlinks (e.g., `controller-gen` → `controller-gen-v0.18.0`)
- **Makefile integration:** Download tools automatically if missing

### 2. Code Generation Workflow
```bash
# Standard Kubebuilder workflow
make generate  # Generate DeepCopy methods
make manifests # Generate CRDs and RBAC
make fmt       # Format code
make vet       # Static analysis
make test      # Run tests with envtest
```

### 3. Directory Structure (Kubebuilder Standard)
```
api/v1alpha1/     # API type definitions
config/           # Kustomize configurations
├── crd/          # CRD base configurations
├── manager/      # Controller manager deployment
├── rbac/         # RBAC configurations
└── default/      # Default overlay
internal/controller/ # Controller implementations
```

### 4. Annotation-Driven Development
- Use `//+kubebuilder:` comments for code generation
- Keep annotations close to relevant code
- Validate generated output regularly

### 5. Testing Strategy
- **Unit tests:** Test business logic with standard Go testing
- **Integration tests:** Use envtest for controller testing
- **E2E tests:** Test against real clusters with Kind/minikube

## Tool Installation & Updates

Tools are automatically downloaded by the Makefile when needed:
```makefile
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen-$(CONTROLLER_TOOLS_VERSION)
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))
```

## Output Binaries Location
Project output binaries are located in `build/bin/`:
- `build/bin/manager` - KubeCoDriver controller binary
- `build/bin/collector` - KubeCoDriver collector binary
