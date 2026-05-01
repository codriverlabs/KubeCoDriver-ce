# KubeCoDriver Documentation

This directory contains comprehensive documentation for the KubeCoDriver project.

## Directory Structure

```
docs/
├── architecture/           # System architecture and design
│   ├── auth-flow.md       # Authentication flow between components
│   └── tls-setup.md       # TLS configuration guide
├── collector/             # Collector component documentation
│   ├── DYNAMIC_LABEL_MATCHING.md
│   ├── HIERARCHICAL_DATE_STRUCTURE.md
│   ├── HIERARCHICAL_PATH_IMPLEMENTATION.md
│   └── POSIX_COMPLIANCE.md
├── controller/            # Controller component documentation
│   ├── container-selection-logic.md
│   └── non-root-user-analysis.md
├── security/              # Security model and RBAC
│   ├── README.md
│   ├── collector-rbac.md
│   ├── controller-rbac.md
│   ├── powertool-rbac-roles.md
│   ├── powertoolconfig-security.md
│   ├── rbac-deployment.md
│   └── security-model.md
├── testing/               # Testing guides and coverage
│   ├── e2e-kind-strategy.md
│   ├── test-nonroot-instructions.md
│   ├── testing-setup.md
│   └── TEST_COVERAGE_SUMMARY.md
└── version-management.md  # Go version management guide
```

## Quick Links

### Getting Started
- [Security Model](security/README.md) - Overview of KubeCoDriver security architecture
- [TLS Setup](architecture/tls-setup.md) - Configure TLS for secure communication
- [Version Management](version-management.md) - Managing Go versions across the project

### Component Documentation
- **Controller**: [Container Selection](controller/container-selection-logic.md), [Non-Root Analysis](controller/non-root-user-analysis.md)
- **Collector**: [Label Matching](collector/DYNAMIC_LABEL_MATCHING.md), [Hierarchical Paths](collector/HIERARCHICAL_PATH_IMPLEMENTATION.md)

### Testing
- [Testing Setup](testing/testing-setup.md) - Overview of testing infrastructure
- [E2E Strategy](testing/e2e-kind-strategy.md) - End-to-end testing with Kind
- [Test Coverage](testing/TEST_COVERAGE_SUMMARY.md) - Current test coverage status

### Security
- [RBAC Deployment](security/rbac-deployment.md) - Role-based access control setup
- [CoDriverJob Security](security/powertoolconfig-security.md) - Tool security configuration

## Contributing to Documentation

When adding new documentation:

1. **Choose the right directory:**
   - `architecture/` - System design, flows, setup guides
   - `collector/` - Collector-specific implementation details
   - `controller/` - Controller-specific implementation details
   - `security/` - Security model, RBAC, authentication
   - `testing/` - Test guides, strategies, coverage reports

2. **Use descriptive filenames:**
   - Use kebab-case: `my-feature-guide.md`
   - Be specific: `container-selection-logic.md` not `containers.md`

3. **Include in this README:**
   - Add your document to the appropriate section
   - Update the directory structure tree

4. **Cross-reference:**
   - Link to related documents
   - Keep documentation interconnected

## Documentation Standards

- Use Markdown format
- Include code examples where applicable
- Add diagrams for complex flows (use Mermaid or ASCII)
- Keep documents focused and concise
- Update documentation when code changes
