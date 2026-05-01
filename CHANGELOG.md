# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.55] - 2025-11-02

### Added

#### New Power Tools
- **Chaos Engineering Tool**: Comprehensive chaos testing with process, CPU, storage, network, and memory experiments
  - Process chaos: suspend, terminate-graceful, terminate-force actions
  - CPU chaos: stress testing with configurable load
  - Storage chaos: disk pressure testing
  - Network chaos: connectivity, DNS, latency, bandwidth testing
  - Memory chaos: pressure testing with multiple patterns
- **tcpdump Tool**: Network packet capture for debugging and analysis
- CoDriverJob examples directory with ready-to-use manifests
- Chaos tool examples with detailed usage documentation

#### Resource Management
- Resource configuration support in CoDriverTool CRD
  - CPU and memory requests/limits for ephemeral containers
  - Configurable per power tool
- `runAsRoot` field in SecuritySpec for tools requiring root access
- Security context inheritance from CoDriverTool to ephemeral containers

#### Testing & Quality
- Kind-based E2E testing with complete cluster isolation
- Comprehensive unit tests improving controller coverage to 77.8%
- Collector server unit tests with mocking infrastructure
- E2E tests for ephemeral containers, conflict detection, and status management

#### Documentation
- Chaos tool comprehensive usage guide (USAGE.md)
- Examples directory with categorized CoDriverJob manifests
- Resource configuration documentation
- Roadmap for experimental OOM via nsenter approach
- Build tools documentation
- E2E testing plan and architecture details
- Unit test improvement roadmap

#### Infrastructure
- Centralized Go version management with `.go-version` file
- Reusable composite action for Docker manifest creation
- ECR sync script includes all power-tool images (aperf, tcpdump, chaos)
- Hierarchical path structure for collector storage

### Changed

#### Refactoring
- Reorganized documentation into categorized subdirectories
- Moved CoDriverTool files to `power-tools/*/config/` directories
- Restructured power-tools Docker setup with common scripts
- Separated build tools from output binaries
- Improved collector build structure
- Organized examples into categorized subdirectories

#### Updates
- Upgraded to Go 1.25.3 (fixes CVEs)
- Updated dependencies to Kubernetes 1.34
- Updated to stable Kubernetes v1.31 for E2E tests
- Chaos tool now uses al2023 base image with stress-ng
- Makefile bundles CoDriverTool files with image references

#### Improvements
- Enhanced CoDriverJob CRD argument handling (TOOL_ARG_* environment variables)
- Improved file naming for ephemeral containers
- Better container selection logic
- Enhanced collector storage with hierarchical paths
- Improved E2E test setup for better compatibility

### Fixed
- gofmt formatting in E2E tests
- Compilation errors and test failures
- Go toolchain version mismatch
- CVE vulnerabilities via Go 1.25.3 upgrade
- Docker build context for aperf tool
- Duplicate cache-from and cache-to in arm64 collector build
- Duplicate test-e2e target in Makefile
- runAsRoot override for pod-level runAsNonRoot
- Problematic utils.go in E2E tests
- Kind cluster port conflicts with host services

### Experimental

#### Chaos Tool OOM Action
- Attempted multiple approaches to trigger OOM on target containers:
  - stress-ng memory allocation (ephemeral container OOMs instead)
  - gdb-based memory injection (blocked by ptrace restrictions)
  - Privileged mode + SYS_ADMIN (still blocked by cgroup isolation)
- **Conclusion**: OOM action not supported due to Kubernetes cgroup isolation
- **Working alternatives**: suspend, terminate-graceful, terminate-force
- **Roadmap**: nsenter-based approach documented for future exploration

### Security
- Added SYS_ADMIN and SYS_PTRACE capabilities for chaos tool
- Privileged mode support in CoDriverTool
- Enhanced security context configuration
- Fixed CVEs through Go 1.25.3 upgrade

### Infrastructure
- Release workflow builds all 5 images (controller, collector, aperf, tcpdump, chaos)
- Conditional image naming: `ghcr.io/codriverlabs/ce/*` for public, `ghcr.io/codriverlabs/*` for private
- Helm chart packaging includes all CoDriverTool files
- ECR sync script supports all power-tool images

## [v1.0.48] - Previous Release

See previous releases for earlier changes.

---

## Notes

### Breaking Changes
None in this release.

### Deprecations
None in this release.

### Known Issues
- OOM chaos action cannot trigger OOM on target containers due to Kubernetes cgroup isolation
  - Use `terminate-force` as alternative to simulate sudden process death
  - See `docs/roadmap/oom-via-nsenter.md` for experimental approaches

### Migration Guide
No migration needed from v1.0.48 to v1.0.55.

To use new features:
1. Update CRDs: `kubectl apply -f config/crd/bases/`
2. Deploy new CoDriverTools: `kubectl apply -f power-tools/*/config/`
3. Update controller and collector images to v1.0.55
