#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
TAG="$2"
ENVIRONMENT="$3"
SECRET_VERSION="$4"

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
		infra-server \
		chart/infra-server \
		--debug \
		--namespace infra \
		--set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
		--values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project stackrox-infra \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project stackrox-infra \
    )
}

deploy() {
    helm upgrade \
        infra-server \
        chart/infra-server \
        --install \
        --create-namespace \
        --namespace infra \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --values - \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project stackrox-infra \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project stackrox-infra \
    )
}

check_not_empty TAG ENVIRONMENT SECRET_VERSION
eval "$TASK"
