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
A single node development cluster ({{.Env.DEV_CLUSTER_NAME}}) was allocated in production infra for this PR.

You can connect to this cluster with: :electric_plug:
\`\`\`
gcloud container clusters get-credentials {{.Env.DEV_CLUSTER_NAME}} --zone us-central1-a --project srox-temp-dev-test
\`\`\`

And then deploy your development infra-server with: :rocket:
\`\`\`
make render-local
make install-local
\`\`\`

And pull infractl from the deployed dev infra-server with: :hammer_and_wrench:
\`\`\`
nohup kubectl -n infra port-forward svc/infra-server-service 8443:8443 &
make pull-infractl-from-dev-server
\`\`\`

You can then use the dev infra instance e.g.: :bike:
\`\`\`
bin/infractl -k -e localhost:8443 whoami
\`\`\`

:warning: ***Any clusters that you start using your dev infra instance should have a lifespan shorter 
then the development cluster instance. Otherwise they will not be destroyed when the dev infra instance 
ceases to exist along when the development cluster is deleted.*** :warning:
EOT

    hub-comment -type deploy -template-file "$tmpfile"
}

die() {
    echo >&2 "$@"
    exit 1
}

add_PR_comment_for_deploy_to_dev "$@"
