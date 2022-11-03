#!/usr/bin/env bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
source "$ROOT/scripts/lib.sh"

set -euo pipefail

add_PR_comment_for_deploy_to_dev() {
    if [[ "$#" -ne 2 ]]; then
        die "missing args. usage: add_PR_comment_for_deploy_to_dev <PR URL> <dev cluster name>"
    fi

    # hub-comment is tied to Circle CI env and requires CIRCLE_PULL_REQUEST
    local url="$1"
    export CIRCLE_PULL_REQUEST="$url"

    export DEV_CLUSTER_NAME="$2"

    IMAGE_NAME="$(make image-name)"
    export IMAGE_NAME

    local tmpfile
    tmpfile=$(mktemp)
    cat > "$tmpfile" <<- EOT
A single node development cluster ({{.Env.DEV_CLUSTER_NAME}}) was allocated in production infra for this PR.

CI will attempt to deploy {{.Env.IMAGE_NAME}} to it.

:electric_plug: You can **connect** to this cluster with:
\`\`\`
gcloud container clusters get-credentials {{.Env.DEV_CLUSTER_NAME}} --zone us-central1-a --project srox-temp-dev-test
\`\`\`

:hammer_and_wrench: And pull **infractl** from the deployed dev infra-server with:
\`\`\`
nohup kubectl -n infra port-forward svc/infra-server-service 8443:8443 &
make pull-infractl-from-dev-server
\`\`\`

:bike: You can then **use** the dev infra instance e.g.:
\`\`\`
bin/infractl -k -e localhost:8443 whoami
\`\`\`

:warning: ***Any clusters that you start using your dev infra instance should have a lifespan shorter then the development cluster instance. Otherwise they will not be destroyed when the dev infra instance ceases to exist when the development cluster is deleted.*** :warning:

### Further Development

:coffee: If you make changes, you can commit and push and CI will take care of updating the development cluster. 

:rocket: If you only modify configuration (chart/infra-server/configuration) or templates (chart/infra-server/{static,templates}), you can get a faster update with:

\`\`\`
make install-local
\`\`\`

### Logs

Logs for the development infra will be [here (authuser=1)](https://console.cloud.google.com/logs/query;query=resource.labels.cluster_name%3D%22{{.Env.DEV_CLUSTER_NAME}}%22%0Aresource.labels.container_name%3D%22infra-server%22?project=srox-temp-dev-test&authuser=1)

EOT

    hub-comment -type deploy -template-file "$tmpfile"
}

add_PR_comment_for_deploy_to_dev "$@"
