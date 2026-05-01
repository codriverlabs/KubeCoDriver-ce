package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
	"github.com/codriverlabs/KubeCoDriver/pkg/collector/auth"
)

// Reconciliation timing constants
const (
	ActiveRunningInterval   = 5 * time.Second
	SetupTeardownInterval   = 15 * time.Second
	CompletedJobInterval    = 5 * time.Minute
	EphemeralStatusInterval = 3 * time.Second
)

// Output mode constants
const (
	OutputModePVC = "pvc"
)

// Phase constants
const (
	PhaseCompleted = "Completed"
)

//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codriverjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codriverjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codriverjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codrivertools,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods/ephemeralcontainers,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create

type CoDriverJobReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	K8sClient kubernetes.Interface
}

func NewCoDriverJobReconciler(c client.Client, scheme *runtime.Scheme, k8sClient kubernetes.Interface) *CoDriverJobReconciler {
	return &CoDriverJobReconciler{
		Client:    c,
		Scheme:    scheme,
		K8sClient: k8sClient,
	}
}

func (r *CoDriverJobReconciler) getToolConfig(ctx context.Context, toolName string) (*kubecodriverv1alpha1.CoDriverTool, error) {
	// Look for CoDriverTool in the same namespace first, then kubecodriver-system
	namespaces := []string{"kubecodriver-system", "default"}

	for _, namespace := range namespaces {
		var toolConfig kubecodriverv1alpha1.CoDriverTool
		configKey := client.ObjectKey{
			Name:      toolName + "-config",
			Namespace: namespace,
		}

		if err := r.Get(ctx, configKey, &toolConfig); err == nil {
			return &toolConfig, nil
		}
	}

	return nil, fmt.Errorf("CoDriverTool not found for tool: %s", toolName)
}

func (r *CoDriverJobReconciler) getTokenDuration(ctx context.Context, collectionDuration time.Duration) time.Duration {
	logger := log.FromContext(ctx)

	// Simple calculation: collection duration + 60 seconds buffer for overhead
	buffer := 60 * time.Second
	tokenDuration := collectionDuration + buffer

	// Kubernetes minimum requirement: 10 minutes (600 seconds)
	minDuration := 10 * time.Minute
	if tokenDuration < minDuration {
		logger.V(1).Info("Token duration below minimum, using 10 minutes",
			"calculated", tokenDuration,
			"minimum", minDuration,
			"collectionDuration", collectionDuration)
		tokenDuration = minDuration
	}

	logger.V(1).Info("Token duration calculated",
		"collectionDuration", collectionDuration,
		"buffer", buffer,
		"finalTokenDuration", tokenDuration)

	return tokenDuration
}

// buildCoDriverJobEnvVars builds environment variables from CoDriverJob spec
func (r *CoDriverJobReconciler) buildCoDriverJobEnvVars(job *kubecodriverv1alpha1.CoDriverJob, targetPod corev1.Pod) []corev1.EnvVar {
	// Extract matching labels from the CoDriverJob's label selector
	matchingLabels := r.extractMatchingLabels(job.Spec.Targets.LabelSelector, targetPod.Labels)

	// Determine target container name
	targetContainerName := "default"
	if job.Spec.Targets.Container != nil && *job.Spec.Targets.Container != "" {
		targetContainerName = *job.Spec.Targets.Container
	} else if len(targetPod.Spec.Containers) > 0 {
		targetContainerName = targetPod.Spec.Containers[0].Name
	}

	envVars := []corev1.EnvVar{
		{Name: "PROFILER_TOOL", Value: job.Spec.Tool.Name},
		{Name: "PROFILER_DURATION", Value: job.Spec.Tool.Duration},
		{Name: "TARGET_POD_NAME", Value: targetPod.Name},
		{Name: "TARGET_NAMESPACE", Value: targetPod.Namespace},
		{Name: "TARGET_CONTAINER_NAME", Value: targetContainerName},
		{Name: "POD_MATCHING_LABELS", Value: matchingLabels},
		{Name: "OUTPUT_MODE", Value: job.Spec.Output.Mode},
	}

	// Add tool-specific arguments as environment variables
	if job.Spec.Tool.Args != nil && len(job.Spec.Tool.Args) > 0 {
		// Convert args slice to a single environment variable
		argsStr := strings.Join(job.Spec.Tool.Args, " ")
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TOOL_ARGS",
			Value: argsStr,
		})

		// Also add individual args as numbered environment variables
		for i, arg := range job.Spec.Tool.Args {
			envVars = append(envVars, corev1.EnvVar{
				Name:  fmt.Sprintf("TOOL_ARG_%d", i),
				Value: arg,
			})
		}
	}

	// Add PVC path if specified
	if job.Spec.Output.Mode == "pvc" && job.Spec.Output.PVC != nil && job.Spec.Output.PVC.Path != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "PVC_PATH",
			Value: *job.Spec.Output.PVC.Path,
		})
	}

	return envVars
}

// extractMatchingLabels extracts the labels that matched the selector
func (r *CoDriverJobReconciler) extractMatchingLabels(selector *metav1.LabelSelector, podLabels map[string]string) string {
	if selector == nil || selector.MatchLabels == nil {
		return "unknown"
	}

	// Build a compact representation of matching labels: key-value
	var labels []string
	for key, value := range selector.MatchLabels {
		if podValue, exists := podLabels[key]; exists && podValue == value {
			labels = append(labels, fmt.Sprintf("%s-%s", key, value))
		}
	}

	if len(labels) == 0 {
		return "unknown"
	}

	return labels[0] // Use first matching label for path organization
}

// findPVCVolumeName finds the volume name for a given PVC claim name in the pod
func (r *CoDriverJobReconciler) findPVCVolumeName(pod corev1.Pod, claimName string) string {
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == claimName {
			return volume.Name
		}
	}
	// Return a default name if not found
	return "profiling-storage"
}

// getTargetContainer returns the target container from the pod
// If targetContainerName is specified, it finds that container
// Otherwise, it returns the first container
func (r *CoDriverJobReconciler) getTargetContainer(pod corev1.Pod, targetContainerName *string) *corev1.Container {
	// If no container specified, use first container
	if targetContainerName == nil || *targetContainerName == "" {
		if len(pod.Spec.Containers) > 0 {
			return &pod.Spec.Containers[0]
		}
		return nil
	}

	// Find the specified container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == *targetContainerName {
			return &pod.Spec.Containers[i]
		}
	}

	// Container not found, fallback to first
	if len(pod.Spec.Containers) > 0 {
		return &pod.Spec.Containers[0]
	}
	return nil
}

// buildSecurityContext converts SecuritySpec to SecurityContext
func (r *CoDriverJobReconciler) buildSecurityContext(securitySpec kubecodriverv1alpha1.SecuritySpec) *corev1.SecurityContext {
	securityContext := &corev1.SecurityContext{}

	if securitySpec.AllowPrivileged != nil {
		securityContext.Privileged = securitySpec.AllowPrivileged
	}

	if securitySpec.Capabilities != nil {
		capabilities := &corev1.Capabilities{}

		if securitySpec.Capabilities.Add != nil {
			for _, cap := range securitySpec.Capabilities.Add {
				capabilities.Add = append(capabilities.Add, corev1.Capability(cap))
			}
		}

		if securitySpec.Capabilities.Drop != nil {
			for _, cap := range securitySpec.Capabilities.Drop {
				capabilities.Drop = append(capabilities.Drop, corev1.Capability(cap))
			}
		}

		securityContext.Capabilities = capabilities
	}

	return securityContext
}

func (r *CoDriverJobReconciler) validateNamespaceAccess(job *kubecodriverv1alpha1.CoDriverJob, toolConfig *kubecodriverv1alpha1.CoDriverTool) error {
	// If no namespace restrictions, allow all
	if len(toolConfig.Spec.AllowedNamespaces) == 0 {
		return nil
	}

	// Check if CoDriverJob namespace is in allowed list
	for _, allowedNS := range toolConfig.Spec.AllowedNamespaces {
		if job.Namespace == allowedNS {
			return nil
		}
	}

	return fmt.Errorf("CoDriverJob namespace '%s' is not allowed for tool '%s'. Allowed namespaces: %v",
		job.Namespace, toolConfig.Spec.Name, toolConfig.Spec.AllowedNamespaces)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CoDriverJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CoDriverJob instance
	var coDriverJob kubecodriverv1alpha1.CoDriverJob
	if err := r.Get(ctx, req.NamespacedName, &coDriverJob); err != nil {
		logger.Error(err, "unable to fetch CoDriverJob")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(1).Info("Reconciling CoDriverJob", "name", coDriverJob.Name, "namespace", coDriverJob.Namespace)

	// Handle deletion
	if coDriverJob.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &coDriverJob)
	}

	// Initialize status if needed
	if coDriverJob.Status.Phase == nil {
		phase := "Pending"
		coDriverJob.Status.Phase = &phase
		now := metav1.Now()
		coDriverJob.Status.StartedAt = &now
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionReady, "False", kubecodriverv1alpha1.ReasonTargetsSelected, "Initializing CoDriverJob")
		if err := r.Status().Update(ctx, &coDriverJob); err != nil {
			logger.Error(err, "unable to update CoDriverJob status")
			return ctrl.Result{}, err
		}
	}

	// Get tool configuration
	toolConfig, err := r.getToolConfig(ctx, coDriverJob.Spec.Tool.Name)
	if err != nil {
		logger.Error(err, "failed to get tool configuration")
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionFailed, "True", kubecodriverv1alpha1.ReasonFailed, fmt.Sprintf("Tool configuration error: %v", err))
		if updateErr := r.Status().Update(ctx, &coDriverJob); updateErr != nil {
			logger.Error(updateErr, "failed to update CoDriverJob status")
		}
		return ctrl.Result{}, err
	}

	// Validate namespace access
	if err := r.validateNamespaceAccess(&coDriverJob, toolConfig); err != nil {
		logger.Error(err, "namespace access denied")
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionFailed, "True", kubecodriverv1alpha1.ReasonFailed, fmt.Sprintf("Namespace access denied: %v", err))
		if updateErr := r.Status().Update(ctx, &coDriverJob); updateErr != nil {
			logger.Error(updateErr, "failed to update CoDriverJob status")
		}
		return ctrl.Result{}, err
	}

	// Get target pods
	var podList corev1.PodList
	selector, err := metav1.LabelSelectorAsSelector(coDriverJob.Spec.Targets.LabelSelector)
	if err != nil {
		logger.Error(err, "unable to convert label selector")
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionFailed, "True", kubecodriverv1alpha1.ReasonFailed, fmt.Sprintf("Invalid label selector: %v", err))
		if updateErr := r.Status().Update(ctx, &coDriverJob); updateErr != nil {
			logger.Error(updateErr, "failed to update CoDriverJob status")
		}
		return ctrl.Result{}, err
	}

	if err := r.List(ctx, &podList, &client.ListOptions{
		Namespace:     coDriverJob.Namespace,
		LabelSelector: selector,
	}); err != nil {
		logger.Error(err, "unable to list target pods")
		return ctrl.Result{}, err
	}

	selectedPods := int32(len(podList.Items))
	coDriverJob.Status.SelectedPods = &selectedPods

	// Check for conflicts with other active CoDriverJobs
	if conflict, conflictMsg := r.checkForConflicts(ctx, &coDriverJob, podList.Items); conflict {
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionConflicted, "True", kubecodriverv1alpha1.ReasonConflictDetected, conflictMsg)
		phase := "Conflicted"
		coDriverJob.Status.Phase = &phase
		if err := r.Status().Update(ctx, &coDriverJob); err != nil {
			logger.Error(err, "unable to update CoDriverJob status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Initialize ActivePods map if needed
	if coDriverJob.Status.ActivePods == nil {
		coDriverJob.Status.ActivePods = make(map[string]string)
	}

	// Process pods for profiling
	for _, pod := range podList.Items {
		containerName := fmt.Sprintf("codriverjob-%s-%s", coDriverJob.Name, string(coDriverJob.UID)[:8])

		// Check if we already have a container for this pod
		if existingContainer, exists := coDriverJob.Status.ActivePods[pod.Name]; exists {
			if r.isContainerRunning(pod, existingContainer) {
				continue // Still running
			} else {
				// Container finished, move to completed
				delete(coDriverJob.Status.ActivePods, pod.Name)
				if coDriverJob.Status.CompletedPods == nil {
					coDriverJob.Status.CompletedPods = new(int32)
				}
				*coDriverJob.Status.CompletedPods++
				continue // Don't process this pod further
			}
		}

		// Check if container already exists in pod spec
		containerExists := false
		for _, ec := range pod.Spec.EphemeralContainers {
			if ec.Name == containerName {
				containerExists = true
				coDriverJob.Status.ActivePods[pod.Name] = containerName
				break
			}
		}

		if containerExists {
			continue
		}

		// Create new ephemeral container
		if err := r.createEphemeralContainerForPod(ctx, &coDriverJob, toolConfig, pod, containerName); err != nil {
			logger.Error(err, "failed to create ephemeral container", "pod", pod.Name)
			continue
		}

		coDriverJob.Status.ActivePods[pod.Name] = containerName
	}

	// Update status based on active containers
	completedPods := selectedPods - int32(len(coDriverJob.Status.ActivePods))
	coDriverJob.Status.CompletedPods = &completedPods

	if len(coDriverJob.Status.ActivePods) > 0 {
		phase := "Running"
		coDriverJob.Status.Phase = &phase
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionRunning, "True", kubecodriverv1alpha1.ReasonRunning, fmt.Sprintf("Running on %d pods", len(coDriverJob.Status.ActivePods)))
	} else if selectedPods > 0 {
		phase := PhaseCompleted
		coDriverJob.Status.Phase = &phase
		now := metav1.Now()
		coDriverJob.Status.FinishedAt = &now
		r.setCondition(&coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionCompleted, "True", kubecodriverv1alpha1.ReasonCompleted, "All containers completed")
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version
		latest := &kubecodriverv1alpha1.CoDriverJob{}
		if err := r.Get(ctx, req.NamespacedName, latest); err != nil {
			return err
		}
		// Preserve our status changes
		latest.Status = coDriverJob.Status
		return r.Status().Update(ctx, latest)
	}); err != nil {
		logger.Error(err, "unable to update CoDriverJob status")
		return ctrl.Result{}, err
	}

	// Determine requeue interval
	interval := r.getRequeueInterval(&coDriverJob)
	return ctrl.Result{RequeueAfter: interval}, nil
}

func (r *CoDriverJobReconciler) getRequeueInterval(job *kubecodriverv1alpha1.CoDriverJob) time.Duration {
	if job.Status.Phase == nil {
		return SetupTeardownInterval
	}

	switch *job.Status.Phase {
	case "Running":
		return ActiveRunningInterval
	case "Completed", "Failed":
		return CompletedJobInterval
	default:
		return SetupTeardownInterval
	}
}

// handleDeletion handles CoDriverJob deletion with proper cleanup
func (r *CoDriverJobReconciler) handleDeletion(ctx context.Context, coDriverJob *kubecodriverv1alpha1.CoDriverJob) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Handling CoDriverJob deletion", "name", coDriverJob.Name)

	// Note: Ephemeral containers cannot be removed from pods once created
	// They will be cleaned up when the pod is deleted
	// We just need to ensure proper status reporting

	return ctrl.Result{}, nil
}

// setCondition sets or updates a condition in the CoDriverJob status
func (r *CoDriverJobReconciler) setCondition(coDriverJob *kubecodriverv1alpha1.CoDriverJob, conditionType, status, reason, message string) {
	now := metav1.Now()

	// Find existing condition
	for i, condition := range coDriverJob.Status.Conditions {
		if condition.Type == conditionType {
			if condition.Status != status {
				coDriverJob.Status.Conditions[i].Status = status
				coDriverJob.Status.Conditions[i].LastTransitionTime = now
			}
			coDriverJob.Status.Conditions[i].Reason = reason
			coDriverJob.Status.Conditions[i].Message = message
			return
		}
	}

	// Add new condition
	coDriverJob.Status.Conditions = append(coDriverJob.Status.Conditions, kubecodriverv1alpha1.CoDriverJobCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	})
}

// checkForConflicts checks if there are conflicting CoDriverJobs targeting the same pods
func (r *CoDriverJobReconciler) checkForConflicts(ctx context.Context, currentTool *kubecodriverv1alpha1.CoDriverJob, targetPods []corev1.Pod) (bool, string) {
	var allCoDriverJobs kubecodriverv1alpha1.CoDriverJobList
	if err := r.List(ctx, &allCoDriverJobs); err != nil {
		return false, ""
	}

	for _, tool := range allCoDriverJobs.Items {
		// Skip self and completed tools
		if tool.Name == currentTool.Name || tool.Namespace != currentTool.Namespace {
			continue
		}
		if tool.Status.Phase != nil && (*tool.Status.Phase == "Completed" || *tool.Status.Phase == "Failed") {
			continue
		}

		// Check if this tool has active pods that overlap with our targets
		if tool.Status.ActivePods != nil {
			for _, targetPod := range targetPods {
				if _, exists := tool.Status.ActivePods[targetPod.Name]; exists {
					return true, fmt.Sprintf("Pod %s is already being profiled by CoDriverJob %s", targetPod.Name, tool.Name)
				}
			}
		}
	}

	return false, ""
}

// isContainerRunning checks if the specified ephemeral container is still running
func (r *CoDriverJobReconciler) isContainerRunning(pod corev1.Pod, containerName string) bool {
	// Check if container exists in ephemeral containers
	for _, ec := range pod.Spec.EphemeralContainers {
		if ec.Name == containerName {
			// Container exists, check its status
			for _, status := range pod.Status.EphemeralContainerStatuses {
				if status.Name == containerName {
					// Check if container is running
					if status.State.Running != nil {
						return true
					}
					// Check if container is terminated (completed or failed)
					if status.State.Terminated != nil {
						return false
					}
					// Container is waiting or unknown state
					return true
				}
			}
			// Container exists but no status yet, assume running
			return true
		}
	}
	return false
}

// createEphemeralContainerForPod creates an ephemeral container for a specific pod
func (r *CoDriverJobReconciler) createEphemeralContainerForPod(ctx context.Context, coDriverJob *kubecodriverv1alpha1.CoDriverJob, toolConfig *kubecodriverv1alpha1.CoDriverTool, pod corev1.Pod, containerName string) error {
	logger := log.FromContext(ctx)

	// Get target container
	targetContainer := r.getTargetContainer(pod, coDriverJob.Spec.Targets.Container)
	if targetContainer != nil {
		logger.V(1).Info("Target container identified", "container", targetContainer.Name)
	}

	// Build environment variables
	envVars := r.buildCoDriverJobEnvVars(coDriverJob, pod)

	// Add collector configuration if specified
	if coDriverJob.Spec.Output.Collector != nil {
		collectionDuration, err := time.ParseDuration(coDriverJob.Spec.Tool.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}

		tokenDuration := r.getTokenDuration(ctx, collectionDuration)

		// Create a token manager for the collector
		collectorTokenManager := auth.NewK8sTokenManager(r.K8sClient, "kubecodriver-system", "kubecodriver-sdk-collector")
		token, err := collectorTokenManager.GenerateToken(ctx, coDriverJob.Name, tokenDuration)
		if err != nil {
			return fmt.Errorf("failed to generate collection token: %w", err)
		}

		envVars = append(envVars,
			corev1.EnvVar{Name: "COLLECTOR_ENDPOINT", Value: coDriverJob.Spec.Output.Collector.Endpoint},
			corev1.EnvVar{Name: "COLLECTOR_TOKEN", Value: token},
			corev1.EnvVar{Name: "CODRIVERJOB_JOB_ID", Value: coDriverJob.Name},
		)
	}

	// Build base security context from toolConfig
	securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

	// Check if runAsRoot is enabled
	runAsRoot := toolConfig.Spec.SecurityContext.RunAsRoot != nil && *toolConfig.Spec.SecurityContext.RunAsRoot

	if runAsRoot {
		// Override user to root
		rootUser := int64(0)
		securityContext.RunAsUser = &rootUser
		runAsNonRootFalse := false
		securityContext.RunAsNonRoot = &runAsNonRootFalse
		logger.V(1).Info("Running as root due to runAsRoot=true")

		// Inherit group from target container or pod for file compatibility
		if targetContainer != nil && targetContainer.SecurityContext != nil && targetContainer.SecurityContext.RunAsGroup != nil {
			securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
			logger.V(1).Info("Inherited runAsGroup from target container for root user",
				"container", targetContainer.Name,
				"group", *targetContainer.SecurityContext.RunAsGroup)
		} else if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsGroup != nil {
			securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
			logger.V(1).Info("Inherited runAsGroup from pod for root user", "group", *pod.Spec.SecurityContext.RunAsGroup)
		}
	} else {
		// Normal inheritance: pod-level first, then container-level override
		if pod.Spec.SecurityContext != nil {
			if pod.Spec.SecurityContext.RunAsUser != nil {
				securityContext.RunAsUser = pod.Spec.SecurityContext.RunAsUser
				logger.V(1).Info("Inherited runAsUser from pod", "user", *pod.Spec.SecurityContext.RunAsUser)
			}
			if pod.Spec.SecurityContext.RunAsGroup != nil {
				securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
				logger.V(1).Info("Inherited runAsGroup from pod", "group", *pod.Spec.SecurityContext.RunAsGroup)
			}
			if pod.Spec.SecurityContext.RunAsNonRoot != nil {
				securityContext.RunAsNonRoot = pod.Spec.SecurityContext.RunAsNonRoot
				logger.V(1).Info("Inherited runAsNonRoot from pod", "nonRoot", *pod.Spec.SecurityContext.RunAsNonRoot)
			}
		}

		// Override with target container's security context if available
		if targetContainer != nil && targetContainer.SecurityContext != nil {
			if targetContainer.SecurityContext.RunAsUser != nil {
				securityContext.RunAsUser = targetContainer.SecurityContext.RunAsUser
				logger.V(1).Info("Inherited runAsUser from target container",
					"container", targetContainer.Name,
					"user", *targetContainer.SecurityContext.RunAsUser)
			}
			if targetContainer.SecurityContext.RunAsGroup != nil {
				securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
				logger.V(1).Info("Inherited runAsGroup from target container",
					"container", targetContainer.Name,
					"group", *targetContainer.SecurityContext.RunAsGroup)
			}
			if targetContainer.SecurityContext.RunAsNonRoot != nil {
				securityContext.RunAsNonRoot = targetContainer.SecurityContext.RunAsNonRoot
				logger.V(1).Info("Inherited runAsNonRoot from target container",
					"container", targetContainer.Name,
					"nonRoot", *targetContainer.SecurityContext.RunAsNonRoot)
			}
		}
	}

	// Create ephemeral container
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:            containerName,
			Image:           toolConfig.Spec.Image,
			ImagePullPolicy: corev1.PullAlways,
			Env:             envVars,
			SecurityContext: securityContext,
			Resources:       r.buildResourceRequirements(toolConfig),
		},
	}

	// Add PVC volume mount if specified
	if coDriverJob.Spec.Output.Mode == OutputModePVC && coDriverJob.Spec.Output.PVC != nil {
		ec.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      r.findPVCVolumeName(pod, coDriverJob.Spec.Output.PVC.ClaimName),
				MountPath: "/mnt/profiling-storage",
			},
		}
	}

	// Update pod with ephemeral container
	podCopy := pod.DeepCopy()
	podCopy.Spec.EphemeralContainers = append(podCopy.Spec.EphemeralContainers, *ec)
	if err := r.SubResource("ephemeralcontainers").Update(ctx, podCopy); err != nil {
		return fmt.Errorf("failed to add ephemeral container to pod %s: %w", pod.Name, err)
	}

	logger.Info("Successfully added ephemeral container",
		"pod", pod.Name,
		"container", containerName,
		"image", toolConfig.Spec.Image)

	return nil
}

// buildResourceRequirements converts ResourceSpec to Kubernetes ResourceRequirements
func (r *CoDriverJobReconciler) buildResourceRequirements(toolConfig *kubecodriverv1alpha1.CoDriverTool) corev1.ResourceRequirements {
	if toolConfig.Spec.Resources == nil {
		return corev1.ResourceRequirements{}
	}

	requirements := corev1.ResourceRequirements{}

	if toolConfig.Spec.Resources.Requests != nil {
		requirements.Requests = corev1.ResourceList{}
		if toolConfig.Spec.Resources.Requests.CPU != nil {
			requirements.Requests[corev1.ResourceCPU] = resource.MustParse(*toolConfig.Spec.Resources.Requests.CPU)
		}
		if toolConfig.Spec.Resources.Requests.Memory != nil {
			requirements.Requests[corev1.ResourceMemory] = resource.MustParse(*toolConfig.Spec.Resources.Requests.Memory)
		}
	}

	if toolConfig.Spec.Resources.Limits != nil {
		requirements.Limits = corev1.ResourceList{}
		if toolConfig.Spec.Resources.Limits.CPU != nil {
			requirements.Limits[corev1.ResourceCPU] = resource.MustParse(*toolConfig.Spec.Resources.Limits.CPU)
		}
		if toolConfig.Spec.Resources.Limits.Memory != nil {
			requirements.Limits[corev1.ResourceMemory] = resource.MustParse(*toolConfig.Spec.Resources.Limits.Memory)
		}
	}

	return requirements
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoDriverJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubecodriverv1alpha1.CoDriverJob{}).
		Complete(r)
}
