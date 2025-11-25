#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
TAG="$2"
ENVIRONMENT="$3"
SECRET_VERSION="${4:-latest}"

# Enables LOCAL_DEPLOY mode in chart.
LOCAL_DEPLOY="${LOCAL_DEPLOY:-false}"

SECRETS_PROJECT="acs-team-automation"
RELEASE_NAMESPACE="infra"
RELEASE_NAME="infra-server"

check_not_empty() {
    for V in "$@"; do
        typeset -n VAR="$V"
        if [ -z "${VAR:-}" ]; then
            echo "ERROR: Variable $V is not set or empty"
            exit 1
        fi
    done
}

install_crds() {
    argo_chart_file=$(find "chart/infra-server/charts" -name "argo-workflows-*.tgz" 2>/dev/null | head -1)
    ARGO_WORKFLOWS_APP_VERSION="$(tar -xzOf "${argo_chart_file}" argo-workflows/Chart.yaml | yq eval '.appVersion' -)"
    echo "Using argo-workflows app version: ${ARGO_WORKFLOWS_APP_VERSION}" >&2
    kubectl apply --validate=false --kustomize \
        "https://github.com/argoproj/argo-workflows/manifests/base/crds/minimal?ref=${ARGO_WORKFLOWS_APP_VERSION}" >&2
}

template() {
    # Need to use helm upgrade --dry-run to have .Capabilities context available
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --dry-run \
        --namespace "${RELEASE_NAMESPACE}" \
        --values chart/infra-server/argo-values.yaml \
        --values chart/infra-server/monitoring-values.yaml \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set localDeploy="${LOCAL_DEPLOY}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    )
}

# Common helm upgrade arguments
HELM_COMMON_ARGS=(
    "${RELEASE_NAME}"
    chart/infra-server
    --install
    --create-namespace
    --namespace "${RELEASE_NAMESPACE}"
    --values chart/infra-server/argo-values.yaml
    --values chart/infra-server/monitoring-values.yaml
    --set tag="${TAG}"
    --set environment="${ENVIRONMENT}"
    --set localDeploy="${LOCAL_DEPLOY}"
)

# deploy upgrades the Helm release with secrets from GCP Secret Manager
deploy() {
    helm upgrade \
        "${HELM_COMMON_ARGS[@]}" \
        --timeout 5m \
        --wait \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    )
}

# diff renders the Helm chart and compares the deployed resources to show what would change on next deployment.
diff() {
    # Need to use helm upgrade --dry-run to have .Capabilities context available
    helm upgrade \
        "${HELM_COMMON_ARGS[@]}" \
        --dry-run \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    ) \
    | sed -n '/---/,$p' \
    | kubectl diff -R -f -
}

# deploy-local deploys to local cluster using local configuration files
deploy-local() {
    echo "Deploying infra-server to local cluster..." >&2
    echo "  Tag: ${TAG}" >&2
    echo "  Environment: ${ENVIRONMENT}" >&2
    echo "  Local Deploy: ${LOCAL_DEPLOY}" >&2
    echo "" >&2

    # Check for required configuration files
    if [ ! -f "chart/infra-server/configuration/${ENVIRONMENT}-values.yaml" ]; then
        echo "ERROR: Configuration files not found for environment: ${ENVIRONMENT}" >&2
        echo "Please run: ENVIRONMENT=${ENVIRONMENT} SECRET_VERSION=latest make secrets-download" >&2
        exit 1
    fi

    helm upgrade \
        "${HELM_COMMON_ARGS[@]}" \
        --timeout 5m \
        --values "chart/infra-server/configuration/${ENVIRONMENT}-values.yaml" \
        --values "chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml" >&2

    echo "" >&2
    echo "Deployment complete!" >&2
    echo "" >&2
    echo "To access the infra-server, run:" >&2
    echo "  kubectl port-forward -n infra svc/infra-server-service 8443:8443" >&2
    echo "" >&2
    echo "Then access at: https://localhost:8443" >&2
}

check_not_empty TASK TAG ENVIRONMENT
install_crds
eval "$TASK"
