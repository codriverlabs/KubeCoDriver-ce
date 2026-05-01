# KubeCoDriver - Community Edition

A Kubernetes operator for running performance profiling and diagnostic tools on target pods in a secure, controlled manner.

## Description

KubeCoDriver provides a declarative way to run profiling tools like `perf`, `strace`, and other diagnostic utilities against running pods in Kubernetes clusters. It uses a secure architecture with separate controller and collector components to ensure proper isolation and access control.

**Use Cases:**
- **Performance Profiling**: CPU, memory, and I/O analysis of running applications
- **Network Diagnostics**: Testing network policies, connectivity, and packet capture
- **Observability Testing**: Validating OTEL collectors and monitoring infrastructure running alongside applications
- **Security Analysis**: Runtime security scanning and compliance validation
- **Chaos Engineering**: Controlled failure injection and resilience testing
- **Disaster Recovery**: Transparent data migration from ephemeral storage to EBS RWO volumes during Spot reclamation or continuous data migration scenarios

## Key Features

- **Declarative Profiling**: Define profiling jobs using Kubernetes custom resources
- **Security First**: Secure token-based authentication between components
- **Multiple Output Modes**: Support for ephemeral, PVC, and collector-based output storage
- **Intelligent Data Organization**: Hierarchical storage with automatic date-based partitioning
- **Tool Management**: Centralized configuration of available profiling tools
- **RBAC Integration**: Full Kubernetes RBAC support for access control

## Architecture

KubeCoDriver consists of two main components that work together:

### Controller
- Manages CoDriverJob resources and orchestrates profiling jobs
- Handles target pod selection and container identification
- Provides secure token-based authentication
- Supports non-root user analysis and container selection logic

### Collector
- Securely receives and stores profiling data from tools
- **Hierarchical Storage**: Organizes files by namespace, labels, date, and CoDriverJob name
- **Configurable Date Formats**: Uses ConfigMap for i18n and date structure customization
- **Path Structure**: `/data/{namespace}/{matching-labels}/{powertool-name}/{year}/{month}/{day}/{filename}`
- **Performance Optimized**: Handles thousands of profiles efficiently with proper directory structure

## Getting Started

### Prerequisites
- go version v1.24.0+
- docker version 25.0+
- kubectl version v1.33.0+
- Access to a Kubernetes v1.33+ cluster

Check your versions:
```sh
go version
docker --version
kubectl version --client
```

### Quick Install

Install using the pre-built YAML manifests:

```sh
kubectl apply -f https://github.com/codriverlabs/KubeCoDriver/releases/latest/download/kubecodriver-operator-v1.1.0-public-preview.yaml
```

Or using Helm with centralized version management:

```sh
helm install kubecodriver-operator \
  https://github.com/codriverlabs/KubeCoDriver/releases/latest/download/kubecodriver-operator-v1.1.0-public-preview.tgz \
  --set global.version=v1.1.0 \
  --set powertools.enabled=true
```

### Building from Source

**Build and push your image:**

```sh
make docker-build-all docker-push-all VERSION=v1.1.0
```

**Install the CRDs:**

```sh
make install
```

**Deploy the operator:**

```sh
make deploy IMG=<your-registry>/kubecodriver:tag
```

**Generate and deploy CoDriverJob configurations:**

```sh
make generate-configs
make deploy-configs
```

### Using Private Container Registries

For private registries like ECR, configure image pull secrets:

```sh
# Configure ECR access for target namespace
./helper_scripts/setup-namespace-docker-secrets.sh default 123456789012.dkr.ecr.us-west-2.amazonaws.com us-west-2
```

### Example Usage

Create a CoDriverJob to profile a running application:

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
    mode: "collector"  # Saves to hierarchical collector storage
```

Apply the configuration:

```sh
kubectl apply -f examples/powertool-aperf-collector.yaml
```

### Accessing Collected Data

The collector organizes data in a hierarchical structure for efficient querying:

```bash
# View profiles by application
kubectl exec -n kubecodriver-system deployment/kubecodriver-collector -- ls /data/default/app-my-application/

# View today's profiles
kubectl exec -n kubecodriver-system deployment/kubecodriver-collector -- find /data -path "*/$(date +%Y/%m/%d)/*"

# View specific CoDriverJob results
kubectl exec -n kubecodriver-system deployment/kubecodriver-collector -- ls /data/default/app-my-application/profile-my-app/
```

## Container Images

All KubeCoDriver components use centralized version management:

- **Controller**: `ghcr.io/codriverlabs/ce/kubecodriver-controller:v1.1.0`
- **Collector**: `ghcr.io/codriverlabs/ce/kubecodriver-collector:v1.1.0`
- **Aperf Tool**: `ghcr.io/codriverlabs/ce/kubecodriver-aperf:v1.1.0`
- **Tcpdump Tool**: `ghcr.io/codriverlabs/ce/kubecodriver-tcpdump:v1.1.0`
- **Chaos Tool**: `ghcr.io/codriverlabs/ce/kubecodriver-chaos:v1.1.0`

## Documentation

### Getting Started
- [Deployment Guide](DEPLOYMENT.md) - Complete deployment instructions
- [EKS Deployment](DEPLOYMENT-EKS.md) - Amazon EKS specific deployment
- [Examples](examples/README.md) - CoDriverJob usage examples

### Architecture & Components
- [Architecture Overview](docs/architecture/) - System design and TLS setup
- [Controller Documentation](docs/controller/) - Container selection and non-root analysis
- [Collector Documentation](docs/collector/) - Storage structure and path organization
- [Security Model](docs/security/README.md) - RBAC and security architecture

### Advanced Topics
- [Hierarchical Date Structure](docs/collector/HIERARCHICAL_DATE_STRUCTURE.md) - Date-based partitioning for performance
- [Hierarchical Path Implementation](docs/collector/HIERARCHICAL_PATH_IMPLEMENTATION.md) - Technical implementation details
- [Dynamic Label Matching](docs/collector/DYNAMIC_LABEL_MATCHING.md) - Label-based file organization
- [Version Management](docs/version-management.md) - Go version management across the project

### Testing & Development
- [Testing Setup](docs/testing/testing-setup.md) - Development and testing guides
- [Test Coverage](docs/testing/TEST_COVERAGE_SUMMARY.md) - Current test coverage status
- [Roadmap](ROADMAP.md) - Future development plans

## Collector Configuration

The collector uses a ConfigMap for internationalization and date formatting:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: collector-config
  namespace: kubecodriver-system
data:
  dateFormat: "2006/01/02"  # Creates hierarchical year/month/day structure
```

This enables:
- **Performance**: Efficient directory structure for thousands of files
- **Organization**: Natural date-based partitioning
- **Querying**: Easy to find profiles by date, namespace, or application
- **Retention**: Simple cleanup policies by date range

## Contributing

We welcome contributions! Please see our contributing guidelines for more information.

## Uninstall

```sh
# Using Helm
helm uninstall kubecodriver-operator -n kubecodriver-system

# Using YAML
kubectl delete -f https://github.com/codriverlabs/KubeCoDriver/releases/download/v1.1.0/kubecodriver-operator-v1.1.0.yaml

# Remove CRDs (this will delete all CoDriverJob resources!)
make uninstall
```

## License

See [LICENSE.md](LICENSE.md) for license information.
