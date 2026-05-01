# KubeCoDriver Deployment Guide

This guide covers different ways to deploy the Tactical Observability Engine (KubeCoDriver) operator.

## 🚀 Quick Start

### Option 1: Direct YAML Installation (Recommended)

```bash
# Install the latest release
kubectl apply -f https://github.com/codriverlabs/KubeCoDriver/releases/latest/download/kubecodriver-operator-v1.1.0-public-preview.yaml

# Or install a specific version
kubectl apply -f https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.yaml
```

### Option 2: Helm Installation

```bash
# Install directly from GitHub release
helm install kubecodriver-operator \
  https://github.com/codriverlabs/KubeCoDriver/releases/latest/download/kubecodriver-operator-v1.1.0-public-preview.tgz

# Or with custom version
helm install kubecodriver-operator \
  https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.tgz \
  --set global.version=v1.1.0 \
  --set global.registry.repository=your-registry.com/kubecodriver
```

## 📦 Container Images

The following container images are published with each release:

- **Controller**: `ghcr.io/codriverlabs/ce/kubecodriver-controller:v1.1.0`
- **Collector**: `ghcr.io/codriverlabs/ce/kubecodriver-collector:v1.1.0`
- **Aperf Tool**: `ghcr.io/codriverlabs/ce/kubecodriver-aperf:v1.1.0`
- **Tcpdump Tool**: `ghcr.io/codriverlabs/ce/kubecodriver-tcpdump:v1.1.0`
- **Chaos Tool**: `ghcr.io/codriverlabs/ce/kubecodriver-chaos:v1.1.0`

## 🎯 What's Included in Each Release

### YAML Installer (`kubecodriver-operator-v1.1.0.yaml`)
- Complete operator deployment with CRDs
- Controller and RBAC configurations
- Webhook configurations
- Namespace setup (kubecodriver-system)

### Helm Chart (`kubecodriver-operator-v1.1.0.tgz`)
- Configurable Helm chart with values
- CoDriverJob configurations (aperf, tcpdump, chaos)
- ECR sync scripts for private registries
- Complete examples directory
- Support for different registry types (GHCR, ECR, local)

### Included Examples
- CoDriverJob configurations for all tools
- Target pod examples (StatefulSet, Pod with PVC)
- Testing configurations (multi-container, non-root)
- Output modes (ephemeral, PVC, collector)

## 🔧 Development Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/codriverlabs/KubeCoDriver.git
cd KubeCoDriver

# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/codriverlabs/ce/kubecodriver-controller:latest

# Deploy CoDriverJob configurations
make deploy-configs
```

### Local Development

```bash
# Generate configs with current images
make generate-configs

# Run locally (requires kubeconfig)
make run
```

## 🏗️ Build Your Own Release

```bash
# Generate all release artifacts
make github-release VERSION=v1.1.0

# Build and push all Docker images
make docker-build-all VERSION=v1.1.0
make docker-push-all VERSION=v1.1.0

# Generate Helm chart only
make helm-chart VERSION=v1.1.0
```

## 🎯 Deployment Options

### Standard Deployment (Controller + Collector + CoDriverJobs)
```bash
# Using YAML (includes both controller and collector)
kubectl apply -f https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.yaml

# Using Helm with CoDriverJobs enabled
helm install kubecodriver-operator \
  https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.tgz \
  --set powertools.enabled=true
```

### Without CoDriverJobs (Controller + Collector only)
```bash
helm install kubecodriver-operator \
  https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.tgz \
  --set powertools.enabled=false
```

### Custom Registry (ECR Example)
```bash
helm install kubecodriver-operator \
  https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.tgz \
  --set global.version=v1.1.0 \
  --set global.registry.repository=123456789012.dkr.ecr.us-west-2.amazonaws.com/codriverlabs/ce \
  --set ecr.accountId=123456789012 \
  --set ecr.region=us-west-2
```

## 🔍 Verification

```bash
# Check operator status
kubectl get pods -n kubecodriver-system

# Check CRDs
kubectl get crd | grep codriverlabs

# Check CoDriverJob configurations
kubectl get powertoolconfigs -n kubecodriver-system

# View controller logs
kubectl logs -n kubecodriver-system deployment/kubecodriver-operator-controller-manager

# View collector logs (if enabled)
kubectl logs -n kubecodriver-system deployment/kubecodriver-collector
```

## 🛠️ CoDriverJob Configuration

After deployment, CoDriverJob configurations are automatically created:

```bash
# List available tools
kubectl get powertoolconfigs -n kubecodriver-system

# Expected output:
# NAME           AGE
# aperf-config   1m
# chaos-config   1m
# tcpdump-config 1m
```

Create a CoDriverJob to use these configurations:

```yaml
apiVersion: kubecodriver.codriverlabs.ai/v1alpha1
kind: CoDriverJob
metadata:
  name: profile-my-app
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-application
  tool:
    name: "aperf"
    duration: "30s"
  output:
    mode: "ephemeral"
```

## 🗑️ Uninstallation

```bash
# Using Helm
helm uninstall kubecodriver-operator -n kubecodriver-system

# Using YAML (delete in reverse order)
kubectl delete -f https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.yaml

# Remove CRDs (this will delete all CoDriverJob resources!)
kubectl delete crd powertools.kubecodriver.codriverlabs.ai
kubectl delete crd powertoolconfigs.kubecodriver.codriverlabs.ai

# Remove namespace
kubectl delete namespace kubecodriver-system
```

## 📚 Next Steps

- See [DEPLOYMENT-EKS.md](DEPLOYMENT-EKS.md) for EKS-specific deployment
- Check [examples/](examples/) for CoDriverJob usage examples
- Review [docs/security/](docs/security/) for security considerations
