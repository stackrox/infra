#!/usr/bin/env bash

set -euo pipefail

add_PR_comment_for_deploy_to_dev() {
    if [[ "$#" -ne 2 ]]; then
        die "missing args. usage: add_PR_comment_for_deploy_to_dev <PR URL> <dev cluster name>"
    fi

    # hub-comment is tied to Circle CI env and requires CIRCLE_PULL_REQUEST
    local url="$1"
    export CIRCLE_PULL_REQUEST="$url"

    export DEV_CLUSTER_NAME="$2"

    local tmpfile
    tmpfile=$(mktemp)
    cat > "$tmpfile" <<- EOT
A single node development cluster ({{.Env.DEV_CLUSTER_NAME}}) has been allocated in production infra for this PR.

You can connect to this cluster with:
\`\`\`
gcloud container clusters get-credentials {{.Env.DEV_CLUSTER_NAME}} --zone us-central1-a --project srox-temp-dev-test
\`\`\`

Once connected you can then deploy:
\`\`\`
make render-local
make install-local
\`\`\`

If this is
EOT

    hub-comment -type deploy -template-file "$tmpfile"
}

die() {
    echo >&2 "$@"
    exit 1
}

add_PR_comment_for_deploy_to_dev "$@"
