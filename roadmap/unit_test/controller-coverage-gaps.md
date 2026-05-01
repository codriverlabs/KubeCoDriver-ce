# Controller Coverage Gaps Analysis

## Current Status: 51.4% Coverage

### Functions with 0% Coverage (Critical Gaps)

#### 1. `NewCoDriverJobReconciler` (0%)
**What it does:** Constructor for CoDriverJobReconciler

**Missing tests:**
```go
TestNewCoDriverJobReconciler
- Valid client and scheme
- Verify fields are set correctly
```

**Priority:** Low (simple constructor)  
**Effort:** 5 minutes

---

#### 2. `getTokenDuration` (0%)
**What it does:** Calculates token duration with 10-minute minimum

**Missing tests:**
```go
TestGetTokenDuration
- Duration < 10 minutes → enforces 10-minute minimum
- Duration = 10 minutes → returns 10 minutes
- Duration > 10 minutes → returns duration + 60s buffer
- Duration = 30 seconds → returns 10 minutes (minimum)
```

**Priority:** HIGH (security-related logic)  
**Effort:** 15 minutes

---

#### 3. `findPVCVolumeName` (0%)
**What it does:** Finds PVC volume name in pod spec

**Missing tests:**
```go
TestFindPVCVolumeName
- PVC found → returns volume name
- PVC not found → returns "profiling-storage" default
- Multiple volumes, PVC in middle → returns correct name
- Pod with no volumes → returns default
```

**Priority:** MEDIUM (PVC mode functionality)  
**Effort:** 10 minutes

---

#### 4. `handleDeletion` (0%)
**What it does:** Cleanup logic when CoDriverJob is deleted

**Missing tests:**
```go
TestHandleDeletion
- Finalizer present → cleanup executed
- Finalizer removed → returns nil
- No finalizer → returns nil
```

**Priority:** MEDIUM (lifecycle management)  
**Effort:** 15 minutes

---

#### 5. `isContainerRunning` (0%)
**What it does:** Checks if ephemeral container is running

**Missing tests:**
```go
TestIsContainerRunning
- Container running → returns true
- Container not found → returns false
- Container terminated → returns false
- Container waiting → returns false
- Pod with no ephemeral containers → returns false
```

**Priority:** HIGH (status checking logic)  
**Effort:** 15 minutes

---

#### 6. `SetupWithManager` (0%)
**What it does:** Registers controller with manager

**Missing tests:**
```go
TestSetupWithManager
- Valid manager → returns nil
- Verify controller is registered
```

**Priority:** LOW (framework integration)  
**Effort:** 10 minutes

---

### Functions with Partial Coverage

#### 7. `buildSecurityContext` (38.5%)
**Current coverage:** Basic privileged flag

**Missing tests:**
```go
TestBuildSecurityContext_Capabilities
- Add capabilities (SYS_ADMIN, NET_ADMIN)
- Drop capabilities (ALL, CHOWN)
- Both add and drop capabilities
- Empty capabilities → returns empty SecurityContext
- Nil capabilities → no capabilities set
```

**Priority:** HIGH (security configuration)  
**Effort:** 15 minutes

---

#### 8. `validateNamespaceAccess` (33.3%)
**Current coverage:** Unknown

**Missing tests:**
```go
TestValidateNamespaceAccess
- No namespace restrictions → returns nil
- Allowed namespace → returns nil
- Disallowed namespace → returns error
- Empty allowed list → returns nil (allow all)
- Wildcard namespace → test behavior
```

**Priority:** HIGH (security/RBAC)  
**Effort:** 15 minutes

---

#### 9. `Reconcile` (56.7%)
**Current coverage:** Basic reconciliation

**Missing tests:**
```go
TestReconcile_ErrorCases
- CoDriverJob not found → returns nil (deleted)
- ToolConfig not found → returns error
- Invalid tool configuration → returns error
- No matching pods → updates status
- Pod selection error → returns error

TestReconcile_StatusUpdates
- Update selected pods count
- Update completed pods count
- Update phase transitions
- Update conditions

TestReconcile_RequeueIntervals
- Active running → 5s requeue
- Setup/teardown → 15s requeue
- Completed → 5m requeue
```

**Priority:** HIGH (core reconciliation logic)  
**Effort:** 30-45 minutes

---

#### 10. `checkForConflicts` (84.6%)
**Current coverage:** Most paths covered

**Missing tests:**
```go
TestCheckForConflicts_EdgeCases
- Multiple CoDriverJobs targeting same pod
- CoDriverJob with different containers
- Conflict resolution logic
```

**Priority:** MEDIUM (already mostly covered)  
**Effort:** 10 minutes

---

#### 11. `createEphemeralContainerForPod` (47.6%)
**Current coverage:** Basic creation

**Missing tests:**
```go
TestCreateEphemeralContainer_Collector
- Collector mode → token generated
- Collector env vars set correctly
- Token duration calculated
- Token generation error → returns error

TestCreateEphemeralContainer_PVC
- PVC volume mount added
- Correct mount path
- PVC not found in pod → uses default name

TestCreateEphemeralContainer_SecurityContext
- Privileged mode
- Capabilities added
- Capabilities dropped

TestCreateEphemeralContainer_Errors
- Pod update fails → returns error
- Invalid security context → returns error
```

**Priority:** HIGH (core functionality)  
**Effort:** 30 minutes

---

### CoDriverTool Controller (0% Coverage)

#### 12. CoDriverTool Reconcile (0%)
**What it does:** Manages CoDriverTool resources

**Missing tests:**
```go
TestCoDriverToolReconcile
- Valid config → updates status
- Invalid config → returns error
- Config not found → returns nil
- Status update → condition set
```

**Priority:** MEDIUM (separate controller)  
**Effort:** 20 minutes

---

## Priority Matrix

### HIGH Priority (Must Have - 70%+ coverage target)

1. **getTokenDuration** - Security-critical, 15 min
2. **isContainerRunning** - Status logic, 15 min
3. **buildSecurityContext** - Security config, 15 min
4. **validateNamespaceAccess** - RBAC, 15 min
5. **Reconcile error cases** - Core logic, 45 min
6. **createEphemeralContainer variants** - Core functionality, 30 min

**Total effort:** ~2.5 hours  
**Expected coverage gain:** 51.4% → 70%+

### MEDIUM Priority (Nice to Have)

7. **findPVCVolumeName** - PVC mode, 10 min
8. **handleDeletion** - Lifecycle, 15 min
9. **checkForConflicts edge cases** - 10 min
10. **CoDriverTool controller** - 20 min

**Total effort:** ~1 hour  
**Expected coverage gain:** 70% → 80%+

### LOW Priority (Optional)

11. **NewCoDriverJobReconciler** - Constructor, 5 min
12. **SetupWithManager** - Framework, 10 min

**Total effort:** 15 minutes  
**Expected coverage gain:** Minimal

---

## Implementation Plan

### Phase 1: Security & Core Logic (2.5 hours)
```
Day 1:
- getTokenDuration tests
- isContainerRunning tests
- buildSecurityContext tests
- validateNamespaceAccess tests

Day 2:
- Reconcile error cases
- createEphemeralContainer variants
```

**Target:** 51.4% → 70%+

### Phase 2: Lifecycle & Edge Cases (1 hour)
```
Day 3:
- findPVCVolumeName tests
- handleDeletion tests
- checkForConflicts edge cases
- CoDriverTool controller tests
```

**Target:** 70% → 80%+

### Phase 3: Polish (15 minutes)
```
- Constructor tests
- Framework integration tests
```

**Target:** 80% → 85%+

---

## Test Implementation Examples

### Example 1: getTokenDuration
```go
func TestGetTokenDuration(t *testing.T) {
	r := &CoDriverJobReconciler{}
	ctx := context.Background()

	tests := []struct {
		name               string
		collectionDuration time.Duration
		wantMinimum        time.Duration
	}{
		{
			name:               "below minimum",
			collectionDuration: 5 * time.Minute,
			wantMinimum:        10 * time.Minute,
		},
		{
			name:               "at minimum",
			collectionDuration: 10 * time.Minute,
			wantMinimum:        10 * time.Minute,
		},
		{
			name:               "above minimum",
			collectionDuration: 15 * time.Minute,
			wantMinimum:        16 * time.Minute, // 15 + 1 minute buffer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.getTokenDuration(ctx, tt.collectionDuration)
			if got < tt.wantMinimum {
				t.Errorf("getTokenDuration() = %v, want >= %v", got, tt.wantMinimum)
			}
		})
	}
}
```

### Example 2: isContainerRunning
```go
func TestIsContainerRunning(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name          string
		pod           *corev1.Pod
		containerName string
		want          bool
	}{
		{
			name: "container running",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "profiler",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			containerName: "profiler",
			want:          true,
		},
		{
			name: "container not found",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			containerName: "profiler",
			want:          false,
		},
		{
			name: "container terminated",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "profiler",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{},
							},
						},
					},
				},
			},
			containerName: "profiler",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.isContainerRunning(tt.pod, tt.containerName)
			if got != tt.want {
				t.Errorf("isContainerRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

---

## Success Criteria

### Coverage Targets
- **Phase 1:** 70%+ (HIGH priority items)
- **Phase 2:** 80%+ (MEDIUM priority items)
- **Phase 3:** 85%+ (LOW priority items)

### Quality Metrics
- All critical paths tested
- Security logic fully covered
- Error scenarios validated
- Edge cases handled
- Fast test execution (<10s)

---

## Estimated Total Effort

- **Phase 1 (HIGH):** 2.5 hours → 70% coverage
- **Phase 2 (MEDIUM):** 1 hour → 80% coverage
- **Phase 3 (LOW):** 15 minutes → 85% coverage

**Total:** ~4 hours for 85% controller coverage

**ROI:** High - covers all critical security and core logic paths
