#!/bin/bash

# Generate a random 32-byte signing key and update the secret
SIGNING_KEY=$(openssl rand -base64 32)
NAMESPACE="kubecodriver-system"

# Create or update the secret
kubectl create secret generic collector-auth \
  --from-literal=signing-key="$SIGNING_KEY" \
  --namespace="$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart the collector deployment to pick up the new key
kubectl rollout restart deployment kubecodriver-sdk-collector -n "$NAMESPACE"

echo "Signing key updated successfully"
