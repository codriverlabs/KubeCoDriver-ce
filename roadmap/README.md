# KubeCoDriver Roadmap

This folder tracks planned features, improvements, and bug fixes for the KubeCoDriver project.

## Active Items

### 🔴 High Priority

1. **[Non-Root Security Context Fix](non-root-security-context-fix.md)** - Status: Not Started
   - Enable ephemeral containers to work with non-root pods
   - Target: v1.0.52
   - Effort: 2-3 hours

## Completed Items

See [UNIT_TEST_SUMMARY.md](UNIT_TEST_SUMMARY.md) for completed test coverage improvements.

## Test Results

- [Non-Root User Test Results](test-results-nonroot.md) - Test execution showing permission issues with non-root pods

## Documentation

- [Non-Root User Analysis](../docs/non-root-user-analysis.md) - Detailed analysis of the security context issue
- [Testing Setup](../docs/testing-setup.md) - Overview of testing infrastructure
- [Test Instructions](../docs/test-nonroot-instructions.md) - Manual testing guide

## Status Legend

- 🔴 Not Started
- 🟡 In Progress
- 🟢 Completed
- ⚪ On Hold

## Priority Levels

- **High**: Critical for production use or blocking other work
- **Medium**: Important but not blocking
- **Low**: Nice to have, can be deferred

## How to Use This Roadmap

1. Each feature/fix has its own markdown file with detailed implementation guide
2. Files include:
   - Problem statement
   - Proposed solution
   - Implementation checklist
   - Success criteria
   - Timeline estimate
3. Update status as work progresses
4. Move completed items to "Completed Items" section

## Use Cases & Test Results

- [Test Plan](use-cases/test-plan.md) - Comprehensive test scenarios
- [Test Results](use-cases/test-results-final.md) - Final test execution results

## Implementation Summaries

- [Container Selection Summary](container-selection-summary.md) - Container selection implementation
- [Non-Root Implementation Summary](implementation-summary.md) - Security context inheritance implementation
