#!/bin/bash

set -e

echo "=== Non-Root User Security Context Test ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Create namespace if not exists
echo -e "${YELLOW}Step 1: Ensuring kubecodriver-test namespace exists...${NC}"
kubectl create namespace kubecodriver-test --dry-run=client -o yaml | kubectl apply -f -
echo ""

# Step 2: Deploy test pod
echo -e "${YELLOW}Step 2: Deploying busybox pod with non-root user (1001:1001)...${NC}"
kubectl apply -f examples/test-nonroot-pod.yaml
echo ""

# Step 3: Wait for pod to be ready
echo -e "${YELLOW}Step 3: Waiting for pod to be ready...${NC}"
kubectl wait --for=condition=Ready pod/busybox-nonroot -n kubecodriver-test --timeout=60s
echo ""

# Step 4: Show pod security context
echo -e "${YELLOW}Step 4: Verifying pod security context...${NC}"
echo "Pod Security Context:"
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.securityContext}' | jq .
echo ""
echo "Container Security Context:"
kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.containers[0].securityContext}' | jq .
echo ""

# Step 5: Check current user in pod
echo -e "${YELLOW}Step 5: Checking current user in pod...${NC}"
kubectl exec busybox-nonroot -n kubecodriver-test -- id
echo ""

# Step 6: Apply CoDriverJob
echo -e "${YELLOW}Step 6: Applying CoDriverJob...${NC}"
kubectl apply -f examples/test-nonroot-powertool.yaml
echo ""

# Step 7: Wait a bit for controller to process
echo -e "${YELLOW}Step 7: Waiting for controller to process (10 seconds)...${NC}"
sleep 10
echo ""

# Step 8: Check CoDriverJob status
echo -e "${YELLOW}Step 8: Checking CoDriverJob status...${NC}"
kubectl get powertool aperf-nonroot-test -n kubecodriver-test -o yaml | grep -A 10 "status:"
echo ""

# Step 9: Check if ephemeral container was created
echo -e "${YELLOW}Step 9: Checking for ephemeral containers...${NC}"
EPHEMERAL_COUNT=$(kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers}' | jq '. | length')
echo "Number of ephemeral containers: $EPHEMERAL_COUNT"

if [ "$EPHEMERAL_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Ephemeral container(s) created!${NC}"
    echo ""
    echo "Ephemeral Container Details:"
    kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0]}' | jq .
    echo ""
    
    # Check ephemeral container security context
    echo "Ephemeral Container Security Context:"
    kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.spec.ephemeralContainers[0].securityContext}' | jq .
    echo ""
    
    # Check ephemeral container status
    echo "Ephemeral Container Status:"
    kubectl get pod busybox-nonroot -n kubecodriver-test -o jsonpath='{.status.ephemeralContainerStatuses[0]}' | jq .
    echo ""
else
    echo -e "${RED}✗ No ephemeral containers found${NC}"
    echo ""
fi

# Step 10: Show controller logs (last 50 lines)
echo -e "${YELLOW}Step 10: Controller logs (last 50 lines)...${NC}"
CONTROLLER_POD=$(kubectl get pods -n kubecodriver-system -l control-plane=controller-manager -o jsonpath='{.items[0].metadata.name}')
if [ -n "$CONTROLLER_POD" ]; then
    kubectl logs -n kubecodriver-system $CONTROLLER_POD --tail=50 | grep -E "(busybox-nonroot|aperf-nonroot-test|ephemeral)" || echo "No relevant logs found"
else
    echo -e "${RED}Controller pod not found${NC}"
fi
echo ""

# Step 11: Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"
echo "Pod: busybox-nonroot (running as user 1001:1001)"
echo "CoDriverJob: aperf-nonroot-test"
echo "Ephemeral Containers: $EPHEMERAL_COUNT"
echo ""
echo -e "${YELLOW}To clean up, run:${NC}"
echo "kubectl delete -f examples/test-nonroot-powertool.yaml"
echo "kubectl delete -f examples/test-nonroot-pod.yaml"
