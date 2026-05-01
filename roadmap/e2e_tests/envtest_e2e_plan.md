# KubeCoDriver Envtest-Based E2E Testing Plan

## Executive Summary

This document outlines a comprehensive plan for implementing E2E tests for the KubeCoDriver project using Kubebuilder's envtest framework. The plan focuses on testing controller logic, API interactions, and resource lifecycle management while acknowledging the limitations of envtest for actual pod profiling scenarios.

## Testing Strategy Overview

### Scope Definition

**✅ What We CAN Test with Envtest (High Confidence)**
- CoDriverJob CRD validation and schema enforcement
- Controller reconciliation logic and state transitions
- RBAC permissions and security policies
- Resource lifecycle management (create/update/delete)
- Status condition management and reporting
- Error handling and edge cases
- Webhook validation (if implemented)

**❌ What We CANNOT Test with Envtest (Requires Kind/Real Cluster)**
- Actual ephemeral container creation and execution
- Pod profiling and data collection
- Container runtime interactions
- Network policies and service mesh integration
- Persistent volume operations
- Resource consumption and limits enforcement

## Implementation Plan

### Phase 1: Foundation Setup (Week 1-2)
**Probability of Success: 95%**

#### 1.1 Test Infrastructure Setup
```go
// test/e2e/suite_test.go
var _ = BeforeSuite(func() {
    By("bootstrapping test environment")
    testEnv = &envtest.Environment{
        CRDDirectoryPaths: []string{
            filepath.Join("..", "..", "config", "crd", "bases"),
        },
        ErrorIfCRDPathMissing: true,
        WebhookInstallOptions: envtest.WebhookInstallOptions{
            Paths: []string{filepath.Join("..", "..", "config", "webhook")},
        },
    }
    
    cfg, err := testEnv.Start()
    Expect(err).NotTo(HaveOccurred())
    
    // Setup scheme and client
    err = v1alpha1.AddToScheme(scheme.Scheme)
    Expect(err).NotTo(HaveOccurred())
    
    k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
    Expect(err).NotTo(HaveOccurred())
    
    // Start controller manager
    mgr, err := ctrl.NewManager(cfg, ctrl.Options{
        Scheme: scheme.Scheme,
        Port:   testEnvWebhookPort,
        Host:   testEnvWebhookHost,
    })
    Expect(err).NotTo(HaveOccurred())
    
    // Setup CoDriverJob controller
    err = (&controller.CoDriverJobReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr)
    Expect(err).NotTo(HaveOccurred())
    
    go func() {
        defer GinkgoRecover()
        err = mgr.Start(ctx)
        Expect(err).NotTo(HaveOccurred())
    }()
})
```

#### 1.2 Test Utilities and Helpers
```go
// test/e2e/utils.go
func CreateTestNamespace() *corev1.Namespace {
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            GenerateName: "kubecodriver-e2e-",
        },
    }
    Expect(k8sClient.Create(ctx, ns)).To(Succeed())
    return ns
}

func CreateMockTargetPod(namespace string) *corev1.Pod {
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "target-pod",
            Namespace: namespace,
            Labels: map[string]string{
                "app": "test-app",
                "env": "testing",
            },
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{{
                Name:  "app",
                Image: "nginx:latest",
            }},
        },
        Status: corev1.PodStatus{
            Phase: corev1.PodRunning,
            ContainerStatuses: []corev1.ContainerStatus{{
                Name:  "app",
                Ready: true,
                State: corev1.ContainerState{
                    Running: &corev1.ContainerStateRunning{},
                },
            }},
        },
    }
    Expect(k8sClient.Create(ctx, pod)).To(Succeed())
    Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
    return pod
}
```

### Phase 2: Core Controller Testing (Week 3-4)
**Probability of Success: 90%**

#### 2.1 CoDriverJob Lifecycle Tests
```go
var _ = Describe("CoDriverJob Lifecycle", func() {
    var namespace *corev1.Namespace
    var targetPod *corev1.Pod
    
    BeforeEach(func() {
        namespace = CreateTestNamespace()
        targetPod = CreateMockTargetPod(namespace.Name)
    })
    
    AfterEach(func() {
        Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
    })
    
    Context("CoDriverJob Creation", func() {
        It("should create CoDriverJob with valid spec", func() {
            powerTool := &v1alpha1.CoDriverJob{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-powertool",
                    Namespace: namespace.Name,
                },
                Spec: v1alpha1.CoDriverJobSpec{
                    Targets: v1alpha1.TargetSpec{
                        LabelSelector: &metav1.LabelSelector{
                            MatchLabels: map[string]string{"app": "test-app"},
                        },
                    },
                    Tool: v1alpha1.ToolSpec{
                        Name:     "aperf",
                        Duration: "30s",
                    },
                    Output: v1alpha1.OutputSpec{
                        Mode: "ephemeral",
                    },
                },
            }
            
            Expect(k8sClient.Create(ctx, powerTool)).To(Succeed())
            
            Eventually(func() string {
                updated := &v1alpha1.CoDriverJob{}
                k8sClient.Get(ctx, client.ObjectKeyFromObject(powerTool), updated)
                return updated.Status.Phase
            }).Should(Equal("Pending"))
        })
        
        It("should reject CoDriverJob with invalid tool name", func() {
            powerTool := &v1alpha1.CoDriverJob{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "invalid-tool",
                    Namespace: namespace.Name,
                },
                Spec: v1alpha1.CoDriverJobSpec{
                    Tool: v1alpha1.ToolSpec{
                        Name: "nonexistent-tool",
                    },
                },
            }
            
            err := k8sClient.Create(ctx, powerTool)
            Expect(err).To(HaveOccurred())
        })
    })
})
```

#### 2.2 Controller Reconciliation Logic Tests
```go
var _ = Describe("Controller Reconciliation", func() {
    It("should handle missing target pods gracefully", func() {
        powerTool := &v1alpha1.CoDriverJob{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "no-targets",
                Namespace: namespace.Name,
            },
            Spec: v1alpha1.CoDriverJobSpec{
                Targets: v1alpha1.TargetSpec{
                    LabelSelector: &metav1.LabelSelector{
                        MatchLabels: map[string]string{"app": "nonexistent"},
                    },
                },
                Tool: v1alpha1.ToolSpec{Name: "aperf"},
            },
        }
        
        Expect(k8sClient.Create(ctx, powerTool)).To(Succeed())
        
        Eventually(func() string {
            updated := &v1alpha1.CoDriverJob{}
            k8sClient.Get(ctx, client.ObjectKeyFromObject(powerTool), updated)
            return updated.Status.Phase
        }).Should(Equal("Failed"))
        
        Eventually(func() []metav1.Condition {
            updated := &v1alpha1.CoDriverJob{}
            k8sClient.Get(ctx, client.ObjectKeyFromObject(powerTool), updated)
            return updated.Status.Conditions
        }).Should(ContainElement(HaveField("Type", "TargetsFound")))
    })
    
    It("should update status conditions correctly", func() {
        // Test status condition management
        // Verify Ready, TargetsFound, ToolConfigured conditions
    })
    
    It("should handle concurrent CoDriverJob conflicts", func() {
        // Test conflict detection between multiple CoDriverJobs
        // targeting the same pod
    })
})
```

### Phase 3: Advanced Scenarios (Week 5-6)
**Probability of Success: 80%**

#### 3.1 RBAC and Security Testing
```go
var _ = Describe("RBAC and Security", func() {
    It("should enforce namespace restrictions", func() {
        // Test namespace access controls
        // Verify CoDriverJob cannot target pods in restricted namespaces
    })
    
    It("should validate security context requirements", func() {
        powerTool := &v1alpha1.CoDriverJob{
            Spec: v1alpha1.CoDriverJobSpec{
                Security: v1alpha1.SecuritySpec{
                    Privileged: true,
                    Capabilities: v1alpha1.CapabilitiesSpec{
                        Add: []corev1.Capability{"SYS_ADMIN", "SYS_PTRACE"},
                    },
                },
            },
        }
        
        // Test security context validation
        // Verify proper capability handling
    })
})
```

#### 3.2 Output Mode Testing
```go
var _ = Describe("Output Modes", func() {
    Context("Ephemeral Mode", func() {
        It("should configure ephemeral output correctly", func() {
            // Test ephemeral mode configuration
            // Verify environment variables are set correctly
        })
    })
    
    Context("PVC Mode", func() {
        It("should validate PVC configuration", func() {
            // Test PVC mode setup
            // Verify volume mount configurations
        })
    })
    
    Context("Collector Mode", func() {
        It("should configure collector endpoint", func() {
            // Test collector mode configuration
            // Verify authentication token generation
        })
    })
})
```

### Phase 4: Error Handling and Edge Cases (Week 7-8)
**Probability of Success: 75%**

#### 4.1 Error Scenarios
```go
var _ = Describe("Error Handling", func() {
    It("should handle tool configuration not found", func() {
        // Test missing ToolConfig scenarios
    })
    
    It("should timeout gracefully on long-running operations", func() {
        // Test timeout handling
    })
    
    It("should clean up resources on deletion", func() {
        // Test finalizer logic and cleanup
    })
    
    It("should handle API server disconnections", func() {
        // Test resilience to temporary API failures
    })
})
```

#### 4.2 Resource Limits and Validation
```go
var _ = Describe("Resource Validation", func() {
    It("should validate tool duration limits", func() {
        // Test duration validation (min/max limits)
    })
    
    It("should enforce concurrent tool limits", func() {
        // Test maximum concurrent CoDriverJobs per namespace
    })
    
    It("should validate label selector complexity", func() {
        // Test complex label selector scenarios
    })
})
```

### Phase 5: Performance and Scale Testing (Week 9-10)
**Probability of Success: 70%**

#### 5.1 Scale Testing
```go
var _ = Describe("Scale Testing", func() {
    It("should handle multiple CoDriverJobs efficiently", func() {
        const numCoDriverJobs = 50
        
        for i := 0; i < numCoDriverJobs; i++ {
            powerTool := createTestCoDriverJob(fmt.Sprintf("scale-test-%d", i))
            Expect(k8sClient.Create(ctx, powerTool)).To(Succeed())
        }
        
        // Verify all CoDriverJobs are processed within reasonable time
        Eventually(func() int {
            powerTools := &v1alpha1.CoDriverJobList{}
            k8sClient.List(ctx, powerTools, client.InNamespace(namespace.Name))
            
            completed := 0
            for _, pt := range powerTools.Items {
                if pt.Status.Phase == "Completed" || pt.Status.Phase == "Failed" {
                    completed++
                }
            }
            return completed
        }).Should(Equal(numCoDriverJobs))
    })
})
```

#### 5.2 Memory and Resource Usage
```go
var _ = Describe("Resource Usage", func() {
    It("should not leak goroutines", func() {
        // Test for goroutine leaks during reconciliation
    })
    
    It("should handle large numbers of target pods", func() {
        // Test performance with many target pods
    })
})
```

## Test Organization Structure

```
test/
├── e2e/
│   ├── suite_test.go              # Test suite setup
│   ├── utils.go                   # Test utilities and helpers
│   ├── powertool_lifecycle_test.go # Core lifecycle tests
│   ├── controller_logic_test.go    # Reconciliation logic tests
│   ├── rbac_security_test.go      # Security and RBAC tests
│   ├── output_modes_test.go       # Output configuration tests
│   ├── error_handling_test.go     # Error scenarios
│   └── performance_test.go        # Scale and performance tests
└── fixtures/
    ├── powertools/               # Sample CoDriverJob manifests
    ├── toolconfigs/             # Sample ToolConfig manifests
    └── pods/                    # Sample target pod manifests
```

## Makefile Integration

```makefile
.PHONY: test-e2e
test-e2e: manifests generate fmt vet envtest ## Run E2E tests
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go test ./test/e2e/... -coverprofile e2e-cover.out -v -ginkgo.v

.PHONY: test-e2e-focus
test-e2e-focus: manifests generate fmt vet envtest ## Run focused E2E tests
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go test ./test/e2e/... -v -ginkgo.focus="$(FOCUS)"

.PHONY: test-e2e-parallel
test-e2e-parallel: manifests generate fmt vet envtest ## Run E2E tests in parallel
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go test ./test/e2e/... -v -ginkgo.procs=4
```

## Success Metrics and KPIs

### Coverage Targets
- **Controller Logic Coverage:** 85%+ (High Probability: 90%)
- **API Validation Coverage:** 95%+ (High Probability: 95%)
- **Error Handling Coverage:** 80%+ (Medium Probability: 75%)

### Performance Benchmarks
- **Single CoDriverJob Processing:** <5 seconds (High Probability: 85%)
- **50 Concurrent CoDriverJobs:** <30 seconds (Medium Probability: 70%)
- **Memory Usage:** <100MB during tests (High Probability: 80%)

### Test Reliability
- **Test Suite Success Rate:** 95%+ (High Probability: 85%)
- **Flaky Test Rate:** <5% (Medium Probability: 70%)
- **CI/CD Integration:** 100% automated (High Probability: 95%)

## Risk Assessment and Mitigation

### High Risk Areas (30-50% Probability of Issues)
1. **Timing-dependent tests** - Mitigate with proper Eventually/Consistently patterns
2. **Resource cleanup** - Implement robust teardown procedures
3. **Parallel test execution** - Ensure proper test isolation

### Medium Risk Areas (20-30% Probability of Issues)
1. **Complex reconciliation scenarios** - Break down into smaller test cases
2. **RBAC testing complexity** - Use dedicated test service accounts
3. **Performance test stability** - Use relative performance metrics

### Low Risk Areas (5-15% Probability of Issues)
1. **Basic CRUD operations** - Well-established patterns
2. **Status condition management** - Straightforward testing
3. **Configuration validation** - Clear success/failure criteria

## Timeline and Milestones

| Phase | Duration | Deliverables | Success Probability |
|-------|----------|--------------|-------------------|
| Phase 1 | 2 weeks | Test infrastructure, basic setup | 95% |
| Phase 2 | 2 weeks | Core controller tests, lifecycle | 90% |
| Phase 3 | 2 weeks | Advanced scenarios, security | 80% |
| Phase 4 | 2 weeks | Error handling, edge cases | 75% |
| Phase 5 | 2 weeks | Performance, scale testing | 70% |

**Total Timeline:** 10 weeks
**Overall Success Probability:** 82%

## Limitations and Future Considerations

### Envtest Limitations
- Cannot test actual pod profiling functionality
- No container runtime interactions
- Limited networking capabilities
- No persistent storage testing

### Future Enhancements
- **Kind-based E2E tests** for full integration testing
- **Chaos engineering** tests for resilience validation
- **Multi-cluster** testing scenarios
- **Performance profiling** of the controller itself

## Conclusion

This comprehensive envtest-based E2E testing plan provides robust coverage of the KubeCoDriver controller's core functionality while acknowledging the limitations of the envtest environment. The phased approach ensures incremental progress with measurable success criteria, while the probability assessments help set realistic expectations for each component.

The plan focuses on what can be effectively tested with envtest (controller logic, API interactions, resource management) while clearly identifying areas that require full cluster testing (actual pod profiling, container runtime interactions).

Implementation of this plan will significantly improve the reliability and maintainability of the KubeCoDriver project while providing a solid foundation for future testing enhancements.
