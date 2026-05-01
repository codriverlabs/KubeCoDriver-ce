#!/bin/bash

# Script to configure image pull secrets for the KubeCoDriver operator deployment
# Usage: ./configure-image-pull-secrets.sh <secret-name>

SECRET_NAME=${1:-"ecr-secret"}

if [ -z "$SECRET_NAME" ]; then
    echo "Usage: $0 <secret-name>"
    echo "Example: $0 my-ecr-secret"
    exit 1
fi

echo "Configuring image pull secrets with secret name: $SECRET_NAME"

# Update the patch file with the provided secret name
cat > config/default/manager_image_pull_secrets_patch.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      imagePullSecrets:
      - name: $SECRET_NAME
EOF

echo "Updated config/default/manager_image_pull_secrets_patch.yaml"
echo "You can now run 'make deploy IMG=<your-image>' to deploy with image pull secrets"
