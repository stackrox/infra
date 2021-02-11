#!/usr/bin/env bash
set -eu
exit 1

###############################################################################
#                           Helper functions                                  #
###############################################################################
cd ~/source/infra && git pull && git checkout 0.0.8

function infra_devenv_setup {
  test -e ~/.docker/config.json || gcloud auth configure-docker
}

function error { echo "ERROR: $@"; }

# bounce the pod (https://stack-rox.atlassian.net/browse/RS-142)
function bounce_pod {
  echo "Waiting for pod to launch" && sleep 10
  POD_NAME=$(kubectl get pods -n infra -l app=infra-server -o name)
  kubectl delete -n infra $POD_NAME
}

function check_version {
  echo "Waiting for pod to launch" && sleep 10
  EXPECTED_VERSION=$(git describe --tags)
  ACTUAL_VERSION=$(infractl -e $PROD_INFRA_ENDPOINT version --json | jq -r '.Server.Version')
  [[ $EXPECTED_VERSION == $ACTUAL_VERSION ]] && echo "Version OK [$ACTUAL_VERSION]" || {
    error "ERROR: server version mismatch $EXPECTED_VERSION == $ACTUAL_VERSION"
  }
}

function purge_dir {
  local dpath=$1
  rm -rf "$dpath"
  mkdir -p "$dpath"
}

###############################################################################
#                           Infra Build & Deploy                              #
###############################################################################

# Bring up a 1-node cluster to host the infra stack (infra-server, argo-workflows, etc)
# Access the infra-server web console via port-forward
# Access the infra-server gRPC interface via `infractl` client
# Access argo workflows in the cluster via `argo` client
function build_and_deploy_local {
  export INFRA_TOKEN=$(pass show INFRA_TOKEN)
  export PROD_INFRA_ENDPOINT="infra.rox.systems:443"  # Creating a cluster for testing
  export MY_CLUSTER_NAME="shane-infra-1"
  export MY_CLUSTER_FLAVOR="gke-default"
  export MY_CLUSTER_ARTIFACTS_DIR="$HOME/artifacts"
  export KUBECONFIG="$HOME/.kube/config:$MY_CLUSTER_ARTIFACTS_DIR"

  infractl -e $PROD_INFRA_ENDPOINT create $MY_CLUSTER_FLAVOR $MY_CLUSTER_NAME --description $MY_CLUSTER_NAME \
    --lifespan="3h" --arg "nodes=1" --arg "machine-type=e2-standard-2" --slack-me --wait
  purge_dir $MY_CLUSTER_ARTIFACTS_DIR
  infractl -e "$PROD_INFRA_ENDPOINT" artifacts "$MY_CLUSTER_NAME" --download-dir="$MY_CLUSTER_ARTIFACTS_DIR"
  MY_CLUSTER_CONTEXT=$(yq e '.users[].name' $MY_CLUSTER_ARTIFACTS_DIR/kubeconfig)
  kubectl config use-context "$MY_CLUSTER_CONTEXT"
  [[ "$(kubectl config current-context)" == "$MY_CLUSTER_CONTEXT" ]] || error "context mismatch"
  cd ~/source/infra
  make configuration-download
  make server cli-local image
  kubectl get deployments/infra-server-deployment -n infra
  kubectl delete deployments/infra-server-deployment -n infra
  make deploy-local
  bounce_pod
  check_version
  kubectl -n infra port-forward deployment/infra-server-deployment 8443:8443 &>/tmp/infra-server.log &

  export TEST_INFRA_ENDPOINT="localhost:8443"
  export TEST_CLUSTER_FLAVOR="gke-default"
  export TEST_CLUSTER_NAME="shane-1"
  open -a Safari "https://$TEST_INFRA_ENDPOINT/"  # (Safari allows bad cert with warning and user acknowledgement)
  infractl -e $TEST_INFRA_ENDPOINT create $TEST_CLUSTER_FLAVOR $TEST_CLUSTER_NAME --description $TEST_CLUSTER_NAME \
    --lifespan="10m" --arg "nodes=1" --arg "machine-type=e2-standard-2" --slack-me --wait --insecure
  argo list
  infractl -e $TEST_INFRA_ENDPOINT list --all --expired --insecure
  open -a Safari "https://console.cloud.google.com/kubernetes/list?project=srox-temp-dev-test"     # gke clusters
  open -a Safari "https://console.cloud.google.com/kubernetes/list?project=srox-temp-sales-demos"  # demo clusters
  infractl -e $TEST_INFRA_ENDPOINT delete $TEST_CLUSTER_NAME --insecure
  kubectl get workflows -n default
  kubectl describe workflow/$TEST_CLUSTER_NAME -n default

  infractl -e $PROD_INFRA_ENDPOINT delete $MY_CLUSTER_NAME
  purge_dir $MY_CLUSTER_ARTIFACTS_DIR
}

# Deploy to development cluster (dev.infra.rox.systems)
function build_and_deploy_development {
  export INFRA_TOKEN=$(pass show INFRA_TOKEN)
  export PROD_INFRA_ENDPOINT="dev.infra.rox.systems:443"
  gcloud container clusters get-credentials "infra-development" --project stackrox-infra --region us-west2
  kubectl config use-context gke_stackrox-infra_us-west2_infra-development
  cd ~/source/infra
  make configuration-download
  kubectl delete deployments/infra-server-deployment -n infra
  make deploy-development
  bounce_pod
  check_version
}

# Deploy to development cluster (infra.rox.systems)
function build_and_deploy_production {
  gcloud container clusters get-credentials "infra-production" --project stackrox-infra --region us-west2
  export INFRA_TOKEN=$(pass show INFRA_TOKEN)
  export PROD_INFRA_ENDPOINT="infra.rox.systems:443"
  kubectl config use-context gke_stackrox-infra_us-west2_infra-production
  cd ~/source/infra
  make configuration-download
  kubectl delete deployments/infra-server-deployment -n infra
  make deploy-production
  bounce_pod
  check_version
}
