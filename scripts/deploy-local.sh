#!/usr/bin/env bash

set -euo pipefail

# This script deploys the infra-server to a local Kubernetes cluster (like Colima)
# without requiring GCP Secret Manager access.

RELEASE_NAMESPACE="infra"
RELEASE_NAME="infra-server"
TAG="${TAG:-$(make tag)}"
ENVIRONMENT="${ENVIRONMENT:-local}"
TEST_MODE="${TEST_MODE:-true}"

echo "Deploying infra-server to local cluster..."
echo "  Tag: ${TAG}"
echo "  Environment: ${ENVIRONMENT}"
echo "  Test Mode: ${TEST_MODE}"
echo ""

# Update helm dependencies
echo "Updating Helm dependencies..."
helm dependency update chart/infra-server

# Create required namespaces
echo "Creating namespaces..."
kubectl create namespace argo 2>/dev/null || echo "namespace/argo already exists"
kubectl create namespace monitoring 2>/dev/null || echo "namespace/monitoring already exists"

# Install Argo CRDs
echo "Installing Argo CRDs..."
argo_chart_file=$(find "chart/infra-server/charts" -name "argo-workflows-*.tgz" 2>/dev/null | head -1)
ARGO_WORKFLOWS_APP_VERSION="$(tar -xzOf "${argo_chart_file}" argo-workflows/Chart.yaml | yq eval '.appVersion' -)"
echo "Using argo-workflows app version: ${ARGO_WORKFLOWS_APP_VERSION}"
kubectl apply --validate=false --kustomize \
    "https://github.com/argoproj/argo-workflows/manifests/base/crds/minimal?ref=${ARGO_WORKFLOWS_APP_VERSION}"

# Check if development configuration files exist
if [ ! -f "chart/infra-server/configuration/development-values.yaml" ]; then
    echo "ERROR: Development configuration files not found."
    echo "Please run: ENVIRONMENT=development SECRET_VERSION=latest make secrets-download"
    exit 1
fi

# Deploy using Helm
echo "Deploying with Helm..."
helm upgrade \
    "${RELEASE_NAME}" \
    chart/infra-server \
    --install \
    --create-namespace \
    --timeout 5m \
    --namespace "${RELEASE_NAMESPACE}" \
    --values chart/infra-server/argo-values.yaml \
    --values chart/infra-server/monitoring-values.yaml \
    --values chart/infra-server/configuration/development-values.yaml \
    --values chart/infra-server/configuration/development-values-from-files.yaml \
    --values chart/infra-server/configuration/local-values.yaml \
    --values chart/infra-server/configuration/local-values-from-files.yaml \
    --set tag="${TAG}" \
    --set environment="${ENVIRONMENT}" \
    --set testMode="${TEST_MODE}"

echo ""
echo "Deployment complete!"
echo ""
echo "To access the infra-server, run:"
echo "  kubectl port-forward -n infra svc/infra-server-service 8443:8443"
echo ""
echo "Then access at: https://localhost:8443"
