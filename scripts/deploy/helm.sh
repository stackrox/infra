#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
TAG="$2"
ENVIRONMENT="$3"
SECRET_VERSION="${4:-latest}"

# Enables TEST_MODE in chart.
# Cannot use CI, because then CD with GHA would not be possible.
TEST_MODE="${TEST_MODE:-false}"

PROJECT="stackrox-infra"
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

template() {
    # Need to use helm upgrade --dry-run to have .Capabilities context available
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --dry-run \
        --namespace "${RELEASE_NAMESPACE}" \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
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
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
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
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    ) \
    | sed -n '/---/,$p' \
    | kubectl diff -R -f -
}

check_not_empty TASK TAG ENVIRONMENT
eval "$TASK"
