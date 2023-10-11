#!/usr/bin/env bash

set -euo pipefail

TASK="$1"
ENVIRONMENT="$2"
SECRET_VERSION="${3:-latest}"

PROJECT="stackrox-infra"

check_not_empty() {
    for V in "$@"; do
        typeset -n VAR="$V"
        if [ -z "${VAR:-}" ]; then
            echo "ERROR: Variable $V is not set or empty"
            exit 1
        fi
    done
}

# Downloads secrets files for an ENVIRONMENT.
download_secrets() {
    mkdir -p chart/infra-server/configuration
    gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    > "chart/infra-server/configuration/${ENVIRONMENT}-values.yaml"

    gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${PROJECT}" \
    > "chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml"
}

# Uploads secrets files for an ENVIRONMENT.
upload_secrets() {
    gcloud secrets versions add \
    "infra-values-${ENVIRONMENT}" \
    --project "${PROJECT}" \
    --data-file "chart/infra-server/configuration/${ENVIRONMENT}-values.yaml"

   gcloud secrets versions add \
   "infra-values-from-files-${ENVIRONMENT}" \
    --project "${PROJECT}" \
    --data-file "chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml"
}

# Shows all available keys in a secrets file.
show_available_secret_files() {
    yq 'keys' "chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml"
}

# Downloads secrets, asks for which secret file to show, and displayed decoded value.
show() {
    download_secrets
    show_available_secret_files

    echo "> Secret file to show:"
    read -r secret_name

    echo "> Contents:"
    yq \
        ".${secret_name}" \
        "chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml" \
    | base64 --decode

}

# Downloads secrets, asks for which secret file to change and what to, and uploads new values.
edit() {
    download_secrets
    show_available_secret_files

    echo "> Secret file to change:"
    read -r secret_name

    echo "> Enter new value. Type 'EOF' on a line by itself to finish:"
    new_value=""

    while IFS= read -r line; do
        if [ "$line" = "EOF" ]; then
            break
        fi
        new_value+="$line\n"
    done

    yq eval \
        --inplace ".${secret_name} = \"$(echo -e -n "${new_value}" | base64)\"" \
        "chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml"
    upload_secrets
}

# Revert downloads a specific secrets version, and uploads it as the latest
revert() {
    download_secrets
    upload_secrets
}

check_not_empty TASK ENVIRONMENT
eval "$TASK"
