# KubeCoDriver CICD Scripts

This directory contains scripts for building and deploying the KubeCoDriver (Tactical Observability Engine) components.

## Architecture Overview

KubeCoDriver consists of three main components:
- **Controller** - Manages CoDriverJob and CoDriverTool resources
- **Collector** - Receives and stores profiling data  
- **CoDriverJobs** - Profiling tool containers (aperf, etc.)

## Prerequisites

1. **Docker** - For building container images
2. **kubectl** - Configured for target Kubernetes cluster
3. **AWS CLI** - For ECR registry access (if using ECR)
4. **Configuration** - Update `config.env` with your settings

## Configuration

Edit `config.env` to set your environment:

```bash
# Registry configuration
REGISTRY=123456789012.dkr.ecr.us-west-2.amazonaws.com
TAG=latest

# Cluster configuration  
NAMESPACE=kubecodriver-system
KUBECONFIG_PATH=~/.kube/config
```

## Quick Start

For a complete automated deployment:

```bash
# Local registry (microk8s)
./deploy-all.sh local

# ECR registry  
./deploy-all.sh ecr
```

All configuration (version, registry settings, namespace) is managed via `config.env`.

## Deployment Flow

### Automated Deployment (Recommended)

Use the all-in-one script for complete deployment with automatic cleanup:

```bash
# Local registry deployment
./deploy-all.sh local

# ECR deployment
./deploy-all.sh ecr
```

**What deploy-all.sh does:**
1. **Loads config.env** - Uses configured registry settings and version
2. **Automatic cleanup** - Removes existing deployment cleanly (enforced)
3. **Build phase** - Builds all components first (controller, collector, powertools)
4. **Deploy phase** - Deploys all components only if builds succeed
5. **Verification** - Checks deployment status and provides next steps

**Configuration is managed via config.env:**
- Registry type (local/ecr)
- Version/tag to deploy
- Registry URLs and credentials
- Namespace settings

### Manual Deployment (Advanced)

#### 1. Controller Deployment (Required First)

The controller must be deployed first as it creates the collector's ServiceAccount and RBAC.

```bash
# Build controller image
./build-controller.sh

# Deploy controller (creates kubecodriver-system namespace, RBAC, CRDs)
./deploy-controller.sh
```

**What this creates:**
- `kubecodriver-system` namespace
- Controller deployment and ServiceAccount
- CoDriverJob and CoDriverTool CRDs
- `kubecodriver-collector` ServiceAccount with token creation permissions
- ClusterRole and ClusterRoleBinding for controller

#### 2. Collector Deployment (Optional)

Deploy collector only if you need centralized profile collection.

```bash
# Build collector image
./build-collector.sh

# Deploy collector (requires controller to be deployed first)
./deploy-collector.sh
```

**What this creates:**
- Collector deployment using `kubecodriver-collector` ServiceAccount
- Collector service (ClusterIP)
- TLS certificates for secure communication
- PVC for profile storage

#### 3. CoDriverJob Images (As Needed)

Build CoDriverJob images for the profiling tools you want to use.

```bash
# Build aperf CoDriverJob
./build-powertool-tool.sh aperf

# Build other tools
./build-powertool-tool.sh <tool-name>
```

## Complete Deployment Example

```bash
# 1. Configure environment
cp config.env.example config.env
# Edit config.env with your settings

# 2. Deploy everything (enforces cleanup automatically)
./deploy-all.sh local   # or 'ecr'

# 3. Verify deployment
kubectl get pods -n kubecodriver-system
kubectl get crds | grep kubecodriver
```

## Script Reference

### All-in-One Script

| Script | Purpose | Dependencies |
|--------|---------|--------------|
| `deploy-all.sh` | Complete automated deployment | All individual scripts |

### Controller Scripts

| Script | Purpose | Dependencies |
|--------|---------|--------------|
| `build-controller.sh` | Build controller Docker image | Docker, config.env |
| `deploy-controller.sh` | Deploy controller to cluster | kubectl, built image |
| `clean-controller.sh` | Remove controller from cluster | kubectl |

### Collector Scripts

| Script | Purpose | Dependencies |
|--------|---------|--------------|
| `build-collector.sh` | Build collector Docker image | Docker, config.env |
| `deploy-collector.sh` | Deploy collector to cluster | kubectl, controller deployed |

### CoDriverJob Scripts

| Script | Purpose | Dependencies |
|--------|---------|--------------|
| `build-powertool-tool.sh` | Build CoDriverJob images | Docker, config.env |

## Usage Patterns

### Development Workflow

```bash
# Quick complete redeploy
./deploy-all.sh local

# Individual component updates (manual approach)
./build-controller.sh local && ./deploy-controller.sh local
./build-collector.sh local && ./deploy-collector.sh local
```

### Production Deployment

```bash
# 1. Update config.env for production
REGISTRY_TYPE=ecr
VERSION=v1.0.0
ECR_ACCOUNT_ID=your-account-id
ECR_REGION=your-region

# 2. Deploy to production cluster
export KUBECONFIG=/path/to/prod-kubeconfig
./deploy-all.sh ecr
```

### Collector-Only Deployment

If you only want ephemeral profiling without centralized collection:

```bash
# Deploy only controller
./build-controller.sh
./deploy-controller.sh

# Build CoDriverJob images
./build-powertool-tool.sh aperf

# Skip collector deployment
# CoDriverJobs will use ephemeral or PVC output modes
```

## Troubleshooting

### Common Issues

1. **Controller deployment fails**:
   ```bash
   # Check cluster permissions
   kubectl auth can-i create clusterroles
   kubectl auth can-i create crds
   ```

2. **Collector deployment fails**:
   ```bash
   # Ensure controller is deployed first
   kubectl get serviceaccount kubecodriver-collector -n kubecodriver-system
   kubectl get clusterrole manager-role
   ```

3. **Image pull failures**:
   ```bash
   # Check registry authentication
   docker login $REGISTRY
   kubectl get secret -n kubecodriver-system | grep regcred
   ```

### Verification Commands

```bash
# Check controller status
kubectl get pods -n kubecodriver-system -l app.kubernetes.io/name=kubecodriver
kubectl logs -n kubecodriver-system deployment/kubecodriver-controller-manager

# Check collector status  
kubectl get pods -n kubecodriver-system -l app=kubecodriver-sdk-collector
kubectl logs -n kubecodriver-system deployment/kubecodriver-sdk-collector

# Check CRDs
kubectl get crds | grep kubecodriver.codriverlabs.ai

# Check RBAC
kubectl get clusterrole manager-role
kubectl get serviceaccount kubecodriver-collector -n kubecodriver-system
```

## Security Considerations

### RBAC Model

- **Controller**: ClusterRole with cross-namespace pod access for ephemeral containers
- **Collector**: Namespace-scoped Role in kubecodriver-system for token validation
- **ServiceAccounts**: Separate SAs for controller and collector with minimal permissions

### Token Flow

1. Controller generates TokenRequest for `kubecodriver-collector` ServiceAccount
2. Token has audience `kubecodriver-sdk-collector` and is bound to CoDriverJob object
3. CoDriverJob client sends token via `Authorization: Bearer` header
4. Collector validates token using Kubernetes TokenReview API

### Network Security

- Collector uses TLS for all communications
- ClusterIP service (internal access only)
- Network policies can restrict traffic to collector

## Environment Variables

### Required for All Scripts

- `REGISTRY` - Container registry URL
- `TAG` - Image tag to use
- `NAMESPACE` - Kubernetes namespace (default: kubecodriver-system)

### Optional

- `KUBECONFIG` - Path to kubeconfig file
- `DOCKER_BUILDKIT` - Enable Docker BuildKit (default: 1)
- `PUSH_IMAGES` - Whether to push images (default: true)

## File Structure

```
cicd-scripts/
├── README.md                    # This file
├── config.env                   # Environment configuration
├── deploy-all.sh                # Complete automated deployment
├── build-controller.sh          # Build controller image
├── deploy-controller.sh         # Deploy controller
├── clean-controller.sh          # Clean controller resources
├── build-collector.sh           # Build collector image  
├── deploy-collector.sh          # Deploy collector
└── build-powertool-tool.sh      # Build CoDriverJob images
```

## Next Steps

After successful deployment:

1. **Create CoDriverTool** - Define your profiling tools
2. **Deploy target applications** - Applications you want to profile
3. **Create CoDriverJob resources** - Start profiling jobs
4. **Monitor results** - Check collector logs and storage

See the main project README.md for usage examples and CoDriverJob configuration.
