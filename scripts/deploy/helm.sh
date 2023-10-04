#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
TAG="$2"
SECRET_VERSION="$3"

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
    helm template \
		"${RELEASE_NAME}" \
		chart/infra-server \
		--debug \
		--namespace "${RELEASE_NAMESPACE}" \
		--set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
		--values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    )
}

deploy() {
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --namespace "${RELEASE_NAMESPACE}" \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    )
}

diff() {
    helm template \
		"${RELEASE_NAME}" \
		chart/infra-server \
		--debug \
		--namespace "${RELEASE_NAMESPACE}" \
		--set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
		--values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    ) | \
	kubectl diff -R -f -
}

check_not_empty TAG ENVIRONMENT SECRET_VERSION
eval "$TASK"
