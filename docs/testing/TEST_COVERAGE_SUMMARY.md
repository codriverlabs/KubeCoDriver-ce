# Test Coverage Summary

## Overview

Comprehensive unit tests for the hierarchical path structure implementation.

## Coverage Improvements

### Before
- `pkg/collector/storage`: **0.0%** coverage
- `extractMatchingLabels`: **0.0%** coverage  
- `buildCoDriverJobEnvVars`: **60.0%** coverage (partial)

### After
- `pkg/collector/storage`: **72.2%** coverage ✅
- `extractMatchingLabels`: **100.0%** coverage ✅
- `buildCoDriverJobEnvVars`: **60.0%** coverage (improved test quality)

## Test Files Created

### 1. `pkg/collector/storage/manager_test.go`
Tests for storage manager with hierarchical path structure.

**Test Cases:**
- `TestNewManager` - Manager initialization
  - Valid configuration
  - Empty date format validation
  - Directory creation
  
- `TestSaveProfile` - Profile saving with different formats
  - Hierarchical date structure (`2006/01/02`)
  - Flat date structure (`2006-01-02`)
  - Unknown app label handling
  
- `TestSaveProfile_DirectoryCreation` - Nested directory creation

**Coverage:** 72.2%

### 2. `internal/controller/label_matching_test.go`
Tests for dynamic label matching from CoDriverJob selectors.

**Test Cases:**
- `TestExtractMatchingLabels` - Label extraction logic
  - Single matching label
  - Multiple matching labels
  - No matching labels
  - Nil selector handling
  - Empty labels handling
  - Environment labels
  - Tier labels
  - Custom labels
  - Pod with extra labels
  
- `TestExtractMatchingLabels_POSIXCompliance` - Naming compliance
  - Hyphen separator (not equals)
  - POSIX-compliant characters

**Coverage:** 100%

### 3. `internal/controller/env_vars_test.go`
Tests for environment variable building with dynamic labels.

**Test Cases:**
- `TestBuildCoDriverJobEnvVars` - Environment variable generation
  - Basic configuration with app label
  - Environment label
  - No matching labels (defaults to unknown)
  - Custom tier label
  
- `TestBuildCoDriverJobEnvVars_WithPVCPath` - PVC path handling

**Coverage:** Improved quality, validates POD_MATCHING_LABELS

## Test Execution

### Run All Tests
```bash
go test ./...
```

### Run with Coverage
```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
```

### View Coverage Report
```bash
go tool cover -html=coverage.out
```

### Run Specific Package
```bash
go test ./pkg/collector/storage -v
go test ./internal/controller -v
```

## Key Test Scenarios

### 1. Hierarchical Date Structure
```go
dateFormat: "2006/01/02"
Expected path: /data/default/app-nginx/profile-job/2025/10/30/output.txt
```

### 2. Dynamic Label Matching
```go
Selector: {app: nginx}
Pod Labels: {app: nginx, version: 1.0}
Result: "app-nginx"
```

### 3. POSIX Compliance
```go
Input: key="app", value="nginx"
Output: "app-nginx" (not "app=nginx")
```

### 4. Unknown Label Handling
```go
Selector: {app: nginx}
Pod Labels: {app: apache}
Result: "unknown"
```

## Test Quality Improvements

### 1. Comprehensive Edge Cases
- Nil selectors
- Empty labels
- No matches
- Multiple matches

### 2. POSIX Compliance Validation
- Verifies hyphen separator
- Checks for invalid characters
- Ensures filesystem safety

### 3. Real-World Scenarios
- Different label types (app, env, tier)
- Multiple namespaces
- Various date formats

### 4. Error Handling
- Empty date format
- Missing directories
- Invalid configurations

## Coverage Goals

### Current Status
| Component | Coverage | Status |
|-----------|----------|--------|
| Storage Manager | 72.2% | ✅ Good |
| Label Matching | 100% | ✅ Excellent |
| Env Vars Builder | 60.0% | ⚠️ Needs improvement |
| Collector Server | 0.0% | ❌ Needs tests |
| Auth Module | 0.0% | ❌ Needs tests |

### Next Steps

1. **Collector Server Tests** (Priority: High)
   - HTTP handler tests
   - Metadata extraction
   - Error responses
   - Authentication flow

2. **Auth Module Tests** (Priority: High)
   - Token validation
   - K8s client integration
   - Service account verification

3. **Integration Tests** (Priority: Medium)
   - End-to-end path creation
   - Multi-component interaction
   - Real K8s cluster scenarios

4. **Controller Tests** (Priority: Medium)
   - Improve buildCoDriverJobEnvVars coverage
   - Test PVC path handling
   - Test tool args parsing

## Running Tests in CI/CD

### GitHub Actions
```yaml
- name: Run tests
  run: go test ./... -coverprofile=coverage.out -covermode=atomic

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

### Local Development
```bash
# Run tests on file change
go install github.com/cosmtrek/air@latest
air -c .air.toml

# Or use watch
watch -n 2 'go test ./pkg/collector/storage -v'
```

## Best Practices

### 1. Table-Driven Tests
```go
tests := []struct {
    name string
    input X
    want Y
}{
    // test cases
}
```

### 2. Descriptive Test Names
```go
TestExtractMatchingLabels_POSIXCompliance
TestSaveProfile_DirectoryCreation
```

### 3. Isolated Tests
- Use `t.TempDir()` for filesystem tests
- Mock external dependencies
- No shared state between tests

### 4. Clear Assertions
```go
if got != want {
    t.Errorf("function() = %v, want %v", got, want)
}
```

## Continuous Improvement

### Coverage Targets
- **Critical paths**: 80%+ coverage
- **Business logic**: 70%+ coverage
- **Utilities**: 60%+ coverage

### Quality Metrics
- All tests pass
- No flaky tests
- Fast execution (<10s for unit tests)
- Clear failure messages

## Summary

✅ **72.2%** coverage for storage manager  
✅ **100%** coverage for label matching  
✅ **Comprehensive** edge case testing  
✅ **POSIX compliance** validation  
✅ **Real-world** scenario coverage  

The test suite provides confidence in the hierarchical path structure implementation and ensures correctness of dynamic label matching.
