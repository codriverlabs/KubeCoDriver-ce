# Coverage Improvement Plan

## Current Coverage Status

### Overall Coverage: 16.8%

| Component | Current | Target | Gap |
|-----------|---------|--------|-----|
| Storage Manager | 72.2% | 80% | ✅ Close |
| Collector Server | 46.5% | 70% | ❌ 23.5% gap |
| Label Matching | 100% | 100% | ✅ Complete |
| Controller | 49.8% | 70% | ❌ 20.2% gap |
| Auth Module | 0% | 70% | ❌ 70% gap |

---

## 1. Collector Server - Missing Coverage (46.5% → 70%+)

### Current Gaps

#### A. Uncovered Functions
- `Start()` - 0% coverage
- `Shutdown()` - 0% coverage

#### B. Uncovered Paths in `handleProfile()` (41.4% coverage)

**Missing Test Cases:**

1. **Successful Profile Upload** ⭐ HIGH PRIORITY
   ```go
   TestHandleProfile_SuccessfulUpload
   - Valid token (need to mock auth.ValidateToken)
   - All headers present
   - Profile data saved successfully
   - Returns 200 OK
   ```
   **Why missing:** Requires mocking K8s token validation
   **Impact:** Covers ~30% of handleProfile logic

2. **Default Values Applied**
   ```go
   TestHandleProfile_DefaultValues
   - Missing X-CoDriverJob-Matching-Labels → defaults to "unknown"
   - Missing X-CoDriverJob-Filename → defaults to "{job-id}.profile"
   - Verify metadata has correct defaults
   ```
   **Why missing:** Requires successful auth flow
   **Impact:** Covers ~10% of handleProfile logic

3. **Storage Failure Handling**
   ```go
   TestHandleProfile_StorageError
   - Valid request
   - Storage.SaveProfile returns error
   - Returns 500 Internal Server Error
   ```
   **Why missing:** Requires mocking storage layer
   **Impact:** Covers ~5% of handleProfile logic

4. **Token Validation Failure**
   ```go
   TestHandleProfile_TokenValidationError
   - Valid Bearer token format
   - auth.ValidateToken returns error
   - Returns 401 Unauthorized
   ```
   **Why missing:** Currently tested but needs explicit case
   **Impact:** Already partially covered

### Implementation Strategy

**Option 1: Interface-Based Mocking (Recommended)**
```go
// Create interfaces for testing
type StorageManager interface {
    SaveProfile(io.Reader, storage.ProfileMetadata) error
}

type TokenValidator interface {
    ValidateToken(context.Context, string) (*authv1.UserInfo, error)
}

// Modify Server struct to use interfaces
type Server struct {
    storage StorageManager
    auth    TokenValidator
}
```

**Option 2: Test Helpers with Real Components**
```go
// Use real storage with temp directory
// Use fake K8s client with pre-configured responses
```

**Estimated Effort:** 2-3 hours  
**Expected Coverage:** 46.5% → 75%+

---

## 2. Auth Module - Zero Coverage (0% → 70%+)

### Files to Test

#### A. `pkg/collector/auth/validator.go`

**Test Cases Needed:**

1. **K8sTokenValidator Creation**
   ```go
   TestNewK8sTokenValidator
   - Valid K8s client
   - Verify audience is set correctly
   ```

2. **Token Validation - Success**
   ```go
   TestValidateToken_Success
   - Valid token
   - K8s TokenReview returns authenticated=true
   - Returns UserInfo with username
   ```

3. **Token Validation - Failures**
   ```go
   TestValidateToken_InvalidToken
   - K8s TokenReview returns authenticated=false
   - Returns error
   
   TestValidateToken_K8sAPIError
   - K8s API returns error
   - Returns error with context
   
   TestValidateToken_WrongAudience
   - Token for different audience
   - Returns error
   ```

4. **Token Manager - Generate Token**
   ```go
   TestGenerateToken_Success
   - Valid service account
   - Returns token string
   
   TestGenerateToken_MinimumDuration
   - Duration < 10 minutes
   - Enforces 10-minute minimum
   
   TestGenerateToken_ServiceAccountNotFound
   - Non-existent service account
   - Returns error
   ```

**Implementation Approach:**
```go
// Mock K8s client
type mockK8sClient struct {
    kubernetes.Interface
    authV1 *mockAuthV1
}

type mockAuthV1 struct {
    tokenReviews *mockTokenReviewInterface
}

// Or use fake.NewSimpleClientset() with reactors
```

**Estimated Effort:** 3-4 hours  
**Expected Coverage:** 0% → 75%+

---

## 3. Controller - Improve Coverage (49.8% → 70%+)

### Current Gaps

#### A. Uncovered Functions
- `handleDeletion()` - 0%
- `isContainerRunning()` - 0%
- `SetupWithManager()` - 0%

#### B. Partially Covered Functions
- `buildCoDriverJobEnvVars()` - 60%
- `createEphemeralContainerForPod()` - 47.6%
- `validateNamespaceAccess()` - 33.3%

### Missing Test Cases

1. **Tool Args Parsing**
   ```go
   TestBuildCoDriverJobEnvVars_ToolArgs
   - Valid JSON args
   - Invalid JSON (error handling)
   - Nested args
   - Special characters
   - Empty args
   ```
   **Impact:** +10% coverage

2. **PVC Configuration**
   ```go
   TestBuildCoDriverJobEnvVars_PVCPath
   - PVC mode with path
   - PVC mode without path
   - Non-PVC mode
   ```
   **Impact:** +5% coverage

3. **Ephemeral Container Creation**
   ```go
   TestCreateEphemeralContainer_Success
   - Valid configuration
   - Container created successfully
   
   TestCreateEphemeralContainer_WithPVC
   - PVC volume mount added
   - Correct mount path
   
   TestCreateEphemeralContainer_WithCollector
   - Collector env vars set
   - Token generated
   ```
   **Impact:** +15% coverage

4. **Container Status Checking**
   ```go
   TestIsContainerRunning
   - Container running
   - Container not found
   - Container terminated
   - Container waiting
   ```
   **Impact:** +5% coverage

5. **Deletion Handling**
   ```go
   TestHandleDeletion
   - Finalizer present
   - Cleanup logic
   - Finalizer removed
   ```
   **Impact:** +5% coverage

**Estimated Effort:** 4-5 hours  
**Expected Coverage:** 49.8% → 70%+

---

## 4. Storage Manager - Improve Coverage (72.2% → 85%+)

### Missing Test Cases

1. **Error Scenarios**
   ```go
   TestSaveProfile_ReadError
   - io.Reader returns error
   - Verify error handling
   
   TestSaveProfile_DiskFull
   - Simulate disk full
   - Verify error message
   
   TestSaveProfile_PermissionDenied
   - Read-only directory
   - Verify error handling
   ```
   **Impact:** +10% coverage

2. **Edge Cases**
   ```go
   TestSaveProfile_EmptyFile
   - Zero-byte file
   - Verify file created
   
   TestSaveProfile_ConcurrentWrites
   - Multiple goroutines
   - Verify no race conditions
   
   TestSaveProfile_SpecialCharactersInPath
   - Unicode in filename
   - Spaces in labels
   - Verify path sanitization
   ```
   **Impact:** +5% coverage

**Estimated Effort:** 1-2 hours  
**Expected Coverage:** 72.2% → 85%+

---

## 5. Integration Tests - New Coverage

### End-to-End Workflows

1. **Complete Profile Upload Flow**
   ```go
   TestE2E_ProfileUpload
   - Create CoDriverJob
   - Simulate power-tool upload
   - Verify file in correct path
   - Verify hierarchical structure
   ```

2. **Multi-Label Scenarios**
   ```go
   TestE2E_DifferentLabels
   - app label → app-nginx path
   - env label → env-prod path
   - tier label → tier-backend path
   ```

3. **Date Hierarchy**
   ```go
   TestE2E_DateStructure
   - Upload on different dates
   - Verify year/month/day folders
   - Verify date format from ConfigMap
   ```

**Estimated Effort:** 4-6 hours  
**Expected Coverage:** New dimension (workflow validation)

---

## Priority Order

### Week 1: High-Impact Tests (Target: 25% → 40%)

1. **Collector Server - Successful Upload** ⭐⭐⭐
   - Biggest coverage gap
   - Core functionality
   - 2-3 hours

2. **Auth Module - Basic Tests** ⭐⭐⭐
   - Zero coverage currently
   - Critical security component
   - 3-4 hours

### Week 2: Medium-Impact Tests (Target: 40% → 55%)

3. **Controller - Tool Args & PVC** ⭐⭐
   - Improve existing coverage
   - Common use cases
   - 2-3 hours

4. **Storage Manager - Error Cases** ⭐⭐
   - Edge case coverage
   - Robustness
   - 1-2 hours

### Week 3: Polish & Integration (Target: 55% → 65%+)

5. **Controller - Container Management** ⭐
   - Complete controller coverage
   - 2-3 hours

6. **Integration Tests** ⭐
   - Workflow validation
   - 4-6 hours

---

## Implementation Roadmap

### Phase 1: Enable Mocking (1 day)

**Create test interfaces:**
```go
// pkg/collector/server/interfaces.go
type StorageManager interface {
    SaveProfile(io.Reader, storage.ProfileMetadata) error
}

type TokenValidator interface {
    ValidateToken(context.Context, string) (*authv1.UserInfo, error)
}
```

**Refactor Server struct:**
```go
type Server struct {
    storage StorageManager  // was *storage.Manager
    auth    TokenValidator  // was *auth.K8sTokenValidator
}
```

**Benefits:**
- Easy mocking in tests
- Better separation of concerns
- No breaking changes to public API

### Phase 2: Implement High-Priority Tests (3-4 days)

1. Collector Server success cases
2. Auth module basic tests
3. Controller tool args tests

### Phase 3: Complete Coverage (2-3 days)

4. Storage error cases
5. Controller container management
6. Integration tests

---

## Expected Outcomes

### Coverage Targets

| Component | Current | After Phase 1 | After Phase 2 | After Phase 3 |
|-----------|---------|---------------|---------------|---------------|
| Storage | 72.2% | 72.2% | 85% | 85% |
| Server | 46.5% | 75% | 75% | 80% |
| Auth | 0% | 0% | 75% | 80% |
| Controller | 49.8% | 49.8% | 70% | 75% |
| **Overall** | **16.8%** | **25%** | **50%** | **65%** |

### Quality Metrics

- ✅ All critical paths tested
- ✅ Error scenarios covered
- ✅ Edge cases validated
- ✅ Integration workflows verified
- ✅ Fast test execution (<30s)
- ✅ Clear test documentation

---

## Key Insights

### Why Coverage is Low

1. **No mocking infrastructure** - Can't test success paths
2. **Auth module untested** - 0% coverage
3. **Missing error cases** - Only happy paths tested
4. **No integration tests** - Workflow gaps

### Quick Wins (Highest ROI)

1. **Add interfaces for mocking** → Enables 20%+ coverage gain
2. **Test auth module** → 0% → 75% in one component
3. **Test successful upload** → Covers main workflow

### Long-term Strategy

- **Maintain 70%+ coverage** for critical components
- **Add tests for new features** before merging
- **Review coverage** in CI/CD pipeline
- **Refactor for testability** when needed

---

## Tools & Resources

### Coverage Analysis
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View in browser
go tool cover -html=coverage.out

# Check specific package
go test ./pkg/collector/server -cover -v
```

### Mocking Libraries (Optional)
- `gomock` - Generate mocks from interfaces
- `testify/mock` - Manual mocking helpers
- Standard approach - Custom mock structs (current)

### CI/CD Integration
```yaml
# GitHub Actions
- name: Test Coverage
  run: |
    go test ./... -coverprofile=coverage.out
    go tool cover -func=coverage.out
    
- name: Coverage Gate
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 60" | bc -l) )); then
      echo "Coverage $coverage% is below 60%"
      exit 1
    fi
```

---

## Summary

**Current State:** 16.8% coverage, missing critical test cases

**Target State:** 65%+ coverage with comprehensive test suite

**Key Actions:**
1. Add mocking interfaces (1 day)
2. Test auth module (3-4 hours)
3. Test successful upload flow (2-3 hours)
4. Add error case tests (2-3 hours)
5. Integration tests (4-6 hours)

**Total Effort:** ~2 weeks for 65%+ coverage

**ROI:** High - Catches bugs early, enables confident refactoring, improves code quality
