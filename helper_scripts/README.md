# KubeCoDriver Helper Scripts

This directory contains utility scripts to help with KubeCoDriver deployment, configuration, and maintenance.

## Scripts Overview

### Image Management

#### `ecr/sync-images-from-ghcr-to-ecr.sh`
Synchronizes multi-architecture KubeCoDriver images from ghcr.io to Amazon ECR.

**Usage:**
```bash
./ecr/sync-images-from-ghcr-to-ecr.sh -a ACCOUNT_ID -r REGION --image-version VERSION
```

**Parameters:**
- `-a, --account-id`: AWS Account ID (required)
- `-r, --region`: AWS Region (required)
- `--image-version`: Image version tag (required)

**Example:**
```bash
./ecr/sync-images-from-ghcr-to-ecr.sh -a 123456789012 -r us-west-2 --image-version 1.1.4-beta
```

**Features:**
- Syncs all KubeCoDriver images (controller, collector, aperf)
- Preserves multi-architecture support (ARM64 and AMD64)
- Creates private ECR repositories automatically
- Cleans up local images after sync

### Kubernetes Secrets Management

#### `k8s-secrets/setup-namespace-docker-secrets.sh`
Sets up image pull secrets for target namespaces to access private ECR repositories.

**Usage:**
```bash
./k8s-secrets/setup-namespace-docker-secrets.sh <namespace> <ecr-registry> <region> [secret-name]
```

**Example:**
```bash
./k8s-secrets/setup-namespace-docker-secrets.sh default 123456789012.dkr.ecr.us-west-2.amazonaws.com us-west-2
```

#### `k8s-secrets/update-collector-signing-key.sh`
Generates and updates the collector authentication signing key.

**Usage:**
```bash
./k8s-secrets/update-collector-signing-key.sh
```

**Features:**
- Generates secure 32-byte signing key
- Updates collector-auth secret in kubecodriver-system namespace
- Automatically restarts collector deployment

### Debugging and Inspection

#### `collector/inspect-profiles.sh`
Inspects the contents of the collector PVC for debugging profiling data.

**Usage:**
```bash
./collector/inspect-profiles.sh
```

**Features:**
- Creates inspector pod if needed
- Lists PVC contents
- Provides interactive shell access to PVC data

## Prerequisites

- kubectl configured with cluster access
- Docker installed and configured
- AWS CLI configured with appropriate permissions
- KubeCoDriver operator deployed in cluster

## Common Workflows

## Customer Deployment Workflow

### For EKS with Private ECR (Recommended)

1. **Sync images to your ECR:**
   ```bash
   ./ecr/sync-images-from-ghcr-to-ecr.sh -a YOUR_ACCOUNT_ID -r YOUR_REGION --image-version 1.1.4-beta
   ```

2. **Configure Helm values for your ECR:**
   ```bash
   # Copy the EKS values template
   cp helm/kubecodriver-operator/values-eks.yaml my-values.yaml
   
   # Edit my-values.yaml and replace:
   # - ACCOUNT_ID with your AWS account ID
   # - REGION with your AWS region
   # - Update IAM role ARN for IRSA if using
   ```

3. **Deploy with Helm:**
   ```bash
   helm install kubecodriver-operator ./helm/kubecodriver-operator -f my-values.yaml
   ```

### For Non-EKS Clusters

1. **Sync images to ECR** (same as above)

2. **Setup image pull secrets:**
   ```bash
   ./k8s-secrets/setup-namespace-docker-secrets.sh kubecodriver-system YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com YOUR_REGION
   ```

3. **Configure Helm values:**
   ```yaml
   global:
     registry:
       type: ecr
       repository: "YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/codriverlabs/ce"
       imagePullSecrets: ["ecr-secret"]
   
   ecr:
     accountId: "YOUR_ACCOUNT_ID"
     region: "YOUR_REGION"
     useIRSA: false
     secretName: ecr-secret
   ```

4. **Deploy with Helm:**
   ```bash
   helm install kubecodriver-operator ./helm/kubecodriver-operator -f my-values.yaml
   ```

### Debugging Profiling Issues
1. Inspect collector PVC contents:
   ```bash
   ./collector/inspect-profiles.sh
   ```

## Notes

- All scripts include error handling and cleanup
- ECR repositories are created as private by default
- Multi-architecture images are properly preserved during sync
- Scripts follow KubeCoDriver security best practices
