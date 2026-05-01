# KubeCoDriver Project - TODO List

**Last Updated**: 2025-11-01  
**Current Version**: v1.0.47

## ✅ Recently Completed

### Container Selection & Security Context (v1.0.50-v1.0.51)
- ✅ Container selection feature for multi-container pods
- ✅ Security context inheritance from target pods/containers
- ✅ runAsRoot feature for tools requiring root access
- ✅ Unit tests for all new features (coverage: 72.4%)
- ✅ CoDriverTool examples updated with runAsRoot
- ✅ Examples folder reorganized into categorized subdirectories
- ✅ CoDriverTool files moved to power-tools/*/config/
- ✅ Makefile updated to bundle configs with versioned images

## 🔴 High Priority - Pending

### 1. E2E Testing Infrastructure
**Status**: Not Started  
**Effort**: 2-3 weeks  
**Files**: `roadmap/e2e_tests/envtest_e2e_plan.md`, `roadmap/e2e_tests/kind_e2e_implementation.md`

**Tasks**:
- [ ] Set up envtest-based E2E tests for controller logic
- [ ] Implement Kind-based E2E tests for actual pod profiling
- [ ] Add test cases for:
  - [ ] CoDriverJob lifecycle (create/update/delete)
  - [ ] Multi-container pod targeting
  - [ ] Non-root pod profiling
  - [ ] runAsRoot feature validation
  - [ ] Conflict detection
  - [ ] Output modes (ephemeral, PVC, collector)
- [ ] Integrate E2E tests into CI/CD pipeline

**Why Important**: Ensures features work end-to-end in real cluster scenarios

---

### 2. Test Coverage Improvements
**Status**: Partially Complete  
**Effort**: 1-2 weeks  
**Files**: `roadmap/unit_test/coverage-improvement-plan.md`, `roadmap/unit_test/controller-coverage-gaps.md`

**Current Coverage**: 72.4% (Controller)  
**Target Coverage**: 80%+

**Gaps to Address**:
- [ ] Collector Server (46.5% → 70%+)
  - [ ] `Start()` function coverage
  - [ ] `Shutdown()` function coverage
  - [ ] `handleProfile()` success paths
  - [ ] Token validation mocking
- [ ] Auth Module (0% → 70%+)
  - [ ] Token generation tests
  - [ ] Token validation tests
  - [ ] K8s client mocking
- [ ] Controller edge cases
  - [ ] Namespace selector logic
  - [ ] Conflict detection scenarios
  - [ ] Error handling paths

**Why Important**: Higher coverage = fewer production bugs

---

### 3. Documentation Updates
**Status**: Partially Complete  
**Effort**: 1-2 days

**Tasks**:
- [ ] Update main README.md with runAsRoot feature
- [ ] Document new examples folder structure
- [ ] Add user guide for multi-container pod targeting
- [ ] Document security context inheritance behavior
- [ ] Add troubleshooting guide for common issues
- [ ] Update architecture diagrams
- [ ] Create video tutorials for key features

**Why Important**: Users need clear documentation to use new features

---

## 🟡 Medium Priority - Pending

### 4. Performance Optimization
**Status**: Not Started  
**Effort**: 1 week

**Tasks**:
- [ ] Profile controller memory usage
- [ ] Optimize reconciliation loops
- [ ] Add caching for frequently accessed resources
- [ ] Implement rate limiting for API calls
- [ ] Add metrics for performance monitoring

**Why Important**: Improves scalability for large clusters

---

### 5. Enhanced Observability
**Status**: Not Started  
**Effort**: 1 week

**Tasks**:
- [ ] Add Prometheus metrics for:
  - [ ] CoDriverJob execution duration
  - [ ] Success/failure rates
  - [ ] Resource usage
  - [ ] Queue depth
- [ ] Create Grafana dashboards
- [ ] Add structured logging with levels
- [ ] Implement distributed tracing

**Why Important**: Better visibility into system behavior

---

### 6. Security Enhancements
**Status**: Not Started  
**Effort**: 1-2 weeks

**Tasks**:
- [ ] Implement Pod Security Standards compliance
- [ ] Add security context validation webhook
- [ ] Implement RBAC policy generator
- [ ] Add audit logging for sensitive operations
- [ ] Security scanning in CI/CD
- [ ] Vulnerability assessment

**Why Important**: Production-grade security requirements

---

## 🔵 Low Priority - Future Enhancements

### 7. Advanced Features
**Status**: Not Started  
**Effort**: 2-4 weeks

**Tasks**:
- [ ] Scheduled profiling with cron expressions
- [ ] Multi-pod profiling (fleet-wide)
- [ ] Profile comparison and diff tools
- [ ] Historical data analysis
- [ ] Integration with APM tools (Datadog, New Relic)
- [ ] Custom tool plugin system

**Why Important**: Advanced use cases for power users

---

### 8. UI/Dashboard
**Status**: Not Started  
**Effort**: 4-6 weeks

**Tasks**:
- [ ] Web-based dashboard for CoDriverJob management
- [ ] Real-time profiling status visualization
- [ ] Profile data viewer
- [ ] Configuration management UI
- [ ] User access control

**Why Important**: Improves user experience for non-CLI users

---

### 9. Multi-Cluster Support
**Status**: Not Started  
**Effort**: 2-3 weeks

**Tasks**:
- [ ] Central controller for multiple clusters
- [ ] Cross-cluster profiling coordination
- [ ] Unified data collection
- [ ] Multi-cluster RBAC

**Why Important**: Enterprise deployments with multiple clusters

---

## 📋 Maintenance Tasks

### Ongoing
- [ ] Keep dependencies up to date
- [ ] Monitor and fix security vulnerabilities
- [ ] Review and merge community contributions
- [ ] Update documentation as features evolve
- [ ] Maintain compatibility with new Kubernetes versions

---

## 🎯 Next Sprint Focus (Recommended)

**Priority Order**:
1. **E2E Testing Infrastructure** - Critical for release confidence
2. **Test Coverage Improvements** - Reduce production bugs
3. **Documentation Updates** - Enable users to adopt new features

**Estimated Timeline**: 4-6 weeks for all three

---

## 📊 Progress Tracking

| Category | Completed | In Progress | Not Started | Total |
|----------|-----------|-------------|-------------|-------|
| Core Features | 3 | 0 | 0 | 3 |
| Testing | 1 | 0 | 2 | 3 |
| Documentation | 1 | 0 | 1 | 2 |
| Performance | 0 | 0 | 1 | 1 |
| Security | 0 | 0 | 1 | 1 |
| Advanced | 0 | 0 | 3 | 3 |
| **Total** | **5** | **0** | **8** | **13** |

**Completion Rate**: 38%

---

## 📝 Notes

- All completed items have been moved to implementation summary documents
- Roadmap files remain for reference and detailed implementation guides
- This TODO list focuses only on pending work
- Update this file as items are completed or priorities change
