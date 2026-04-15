#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
TAG="$2"
ENVIRONMENT="$3"
SECRET_VERSION="${4:-latest}"

# Enables TEST_MODE in chart.
# Cannot use CI, because then CD with GHA would not be possible.
TEST_MODE="${TEST_MODE:-false}"

# Session secret for JWT signing (used in local and test deployments)
# Randomly generated at deployment time for security
SESSION_SECRET="${SESSION_SECRET:-}"

# Helm debug mode (optional)
HELM_DEBUG="${HELM_DEBUG:-}"

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

deploy() {
    echo "Deploy.." >&2
    echo "  RELEASE_NAME: ${RELEASE_NAME}" >&2
    echo "  Tag: ${TAG}" >&2
    echo "  Environment: ${ENVIRONMENT}" >&2
    echo "  Test Mode: ${TEST_MODE}" >&2
    echo "  Local Values: ${local_values:-false}" >&2
    echo "  Session Secret: ${SESSION_SECRET:+<set>}" >&2

    # Delete existing secret in TEST_MODE to force recreation with correct template
    # Previous deployments may have created the secret with the wrong template
    if [[ "${TEST_MODE}" == "true" ]]; then
        echo "Deleting existing infra-server-secrets to force recreation..." >&2
        kubectl delete secret infra-server-secrets -n "${RELEASE_NAMESPACE}" --ignore-not-found >&2

        # Generate self-signed cert for TEST_MODE deployments (similar to deploy-local)
        local cert_dir='chart/infra-server/configuration'
        local cert_file="${cert_dir}/local-cert.pem"
        local key_file="${cert_dir}/local-key.pem"

        mkdir -p "${cert_dir}"

        if [[ ! -f "${cert_file}" ]] || [[ ! -f "${key_file}" ]]; then
            echo "Generating self-signed certificate for TEST_MODE..." >&2
            # Create a temporary config file for SAN extension
            local san_config
            san_config=$(mktemp) || { echo "Failed to create temporary config file" >&2; return 1; }
            cat > "${san_config}" <<EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF
            openssl req -x509 -newkey rsa:2048 -nodes \
                -keyout "${key_file}" \
                -out "${cert_file}" \
                -days 36500 \
                -config "${san_config}" >&2
            rm -f "${san_config}"
            echo "Certificate generated: ${cert_file}" >&2
        else
            echo "Using existing certificate: ${cert_file}" >&2
        fi
    fi

    # Build helm command
    local helm_cmd=(
        helm upgrade
        "${RELEASE_NAME}"
        chart/infra-server
        --install
        --create-namespace
        --timeout 5m
        --wait
        --namespace "${RELEASE_NAMESPACE}"
        --values chart/infra-server/argo-values.yaml
        --values chart/infra-server/monitoring-values.yaml
        --values -
    )

    # Add test-mode values file if TEST_MODE is enabled
    # This ensures testMode=true overrides any GCloud secret values
    if [[ "${TEST_MODE}" == "true" ]]; then
        helm_cmd+=(--values chart/infra-server/test-mode-values.yaml)
    fi

    # Add runtime values as --set flags
    helm_cmd+=(
        --set tag="${TAG}"
        --set environment="${ENVIRONMENT}"
    )

    # Add sessionSecret if set
    if [[ -n "${SESSION_SECRET}" ]]; then
        helm_cmd+=(--set sessionSecret="${SESSION_SECRET}")
    fi

    # Enable debug in TEST_MODE to see actual values
    if [[ -n "${HELM_DEBUG}" ]] || [[ "${TEST_MODE}" == "true" ]]; then
        helm_cmd+=(--debug)
    fi

    # Debug: Show the helm command that will be executed
    echo "=== Helm Command ===" >&2
    printf '%q ' "${helm_cmd[@]}" >&2
    echo >&2
    echo "===================" >&2

    # Check if test-mode-values.yaml exists
    if [[ "${TEST_MODE}" == "true" ]]; then
        if [[ -f "chart/infra-server/test-mode-values.yaml" ]]; then
            echo "test-mode-values.yaml exists" >&2
            cat chart/infra-server/test-mode-values.yaml >&2
        else
            echo "ERROR: test-mode-values.yaml NOT FOUND" >&2
        fi
    fi

    # Temporarily disable error exit to capture helm failure and run debugging
    set +e
    "${helm_cmd[@]}" < <(
    if [[ ${local_values:-} != 'true' ]]; then
      gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
      && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}";
    else
      cat 'chart/infra-server/configuration/local-values.yaml'
    fi
    )
    local helm_exit=$?
    set -e

    if [[ $helm_exit -ne 0 ]]; then
        echo "Helm deployment failed with exit code $helm_exit" >&2
        echo "=== Pod Status ===" >&2
        kubectl get pods -n "${RELEASE_NAMESPACE}" >&2 || true
        echo "=== Pod Descriptions ===" >&2
        kubectl describe pods -n "${RELEASE_NAMESPACE}" >&2 || true
        echo "=== Pod Logs ===" >&2
        kubectl logs -n "${RELEASE_NAMESPACE}" -l app=infra-server --tail=100 >&2 || true
        return $helm_exit
    fi
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

# deploy-local deploys without secrets
deploy-local() {
    echo 'Deploying for testing without secrets...' >&2

    # Generate random session secret for JWT signing
    # This is used by both the server and Cypress tests
    if [[ -z "${SESSION_SECRET}" ]]; then
        SESSION_SECRET=$(openssl rand -base64 32 | tr -d '\n')
        echo "Generated random session secret for this deployment" >&2
    fi

    # Generate self-signed cert for local development if it doesn't exist
    local cert_dir='chart/infra-server/configuration'
    local cert_file="${cert_dir}/local-cert.pem"
    local key_file="${cert_dir}/local-key.pem"

    if [[ ! -f "${cert_file}" ]] || [[ ! -f "${key_file}" ]]; then
        echo "Generating self-signed certificate for local development..." >&2
        # Create a temporary config file for SAN extension
        # SAN (Subject Alternative Name) is required for gRPC-Gateway TLS validation in modern Go versions
        # Without SAN, you'll get: "x509: certificate relies on legacy Common Name field"
        local san_config
        san_config=$(mktemp) || { echo "Failed to create temporary config file" >&2; return 1; }
        cat > "${san_config}" <<EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF
        openssl req -x509 -newkey rsa:2048 -nodes \
            -keyout "${key_file}" \
            -out "${cert_file}" \
            -days 36500 \
            -config "${san_config}" >&2
        rm -f "${san_config}"
        echo "Certificate generated: ${cert_file}" >&2
    else
        echo "Using existing self-signed certificate: ${cert_file}" >&2
    fi

    ENVIRONMENT='local'
    local_values='true'
    deploy
    echo -e '\nDeployment complete!' >&2
}

check_not_empty TASK TAG ENVIRONMENT
install_crds
eval "$TASK"
