#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
TAG="$2"
ENVIRONMENT="$3"
SECRET_VERSION="${4:-latest}"

# Enables TEST_MODE in chart.
# Cannot use CI, because then CD with GHA would not be possible.
TEST_MODE="${TEST_MODE:-false}"

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
    kubectl apply --kustomize \
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
        --set testMode="${TEST_MODE}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    )
}

# deploy upgrades the Helm release with
deploy() {
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --timeout 5m \
        --wait \
        --namespace "${RELEASE_NAMESPACE}" \
        --values chart/infra-server/argo-values.yaml \
        --values chart/infra-server/monitoring-values.yaml \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
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
        --set testMode="${TEST_MODE}" \
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

check_not_empty TASK TAG ENVIRONMENT
install_crds
eval "$TASK"
