#!/usr/bin/env bash
# https://stack-rox.atlassian.net/wiki/spaces/ENG/pages/1868496931/Infra+Server+Release+Qualification+and+Rollout
set -eu
exit 1

export INFRA_TOKEN=$(pass show INFRA_TOKEN)
export PROD_INFRA_ENDPOINT="infra.rox.systems:443"
export DEV_INFRA_ENDPOINT="dev.infra.rox.systems:443"
export LOCAL_INFRA_ENDPOINT="localhost:8443"
export LOCAL_CENTRAL_ENDPOINT="localhost:8000"

export PROD_INFRA_KUBECONTEXT_NAME="gke_stackrox-infra_us-west2_infra-production"
export DEV_INFRA_KUBECONTEXT_NAME="gke_stackrox-infra_us-west2_infra-development"
export DEFINED_INFRA_COMMON=1

###############################################################################
#                           Helper functions                                  #
###############################################################################

function info { echo "INFO: $@"; }
function error { echo "ERROR: $@"; }
function warning { echo "WARNING: $@"; }
function join_by { local IFS="$1"; shift; echo "$*"; }

# bounce the pod (https://stack-rox.atlassian.net/browse/RS-142)
function bounce_infra_server_pod {
  info "Waiting for initial pod to launch" && sleep 10
  POD_NAME=$(kubectl get pods -n infra -l app=infra-server -o name)
  kubectl delete -n infra $POD_NAME
  info "Waiting for replacement pod to launch" && sleep 10
}

function assert_infra_server_version {
  local endpoint=${1:-none}
  local ACTUAL_VERSION=$(infractl -k -e $endpoint version --json | jq -r '.Server.Version')
  local EXPECTED_VERSION=$(cd $HOME/source/infra && git describe --tags)

  if [[ "$ACTUAL_VERSION" != "$EXPECTED_VERSION" ]]; then
    error "Infra server version mismatch:"
    error "  ACTUAL_VERSION   => $ACTUAL_VERSION"
    error "  EXPECTED_VERSION => $EXPECTED_VERSION"
    return 1
  fi

  info "Infra server version is $ACTUAL_VERSION"
}

function purge_dir {
  local dpath=$1
  rm -rf "$dpath"
  mkdir -p "$dpath"
  info "Created $dpath"
}

###############################################################################
#                           Infra Build & Deploy                              #
###############################################################################

function build_and_deploy_local {
  export INFRA_CLUSTER_NAME="shane-rs-156-infra"
  export INFRA_CLUSTER_FLAVOR="gke-default"
  export INFRA_ARTIFACTS_DIR="$HOME/artifacts/infra"

  export TEST_CLUSTER_NAME="shane-rs-156-demo-1"
  export TEST_CLUSTER_FLAVOR="openshift-4-demo"
  export TEST_ARTIFACTS_DIR="$HOME/artifacts/test"

  export KUBECONFIG=$(join_by ':' \
		      "$HOME/.kube/config" \
		      "${INFRA_ARTIFACTS_DIR}/kubeconfig" \
		      "${TEST_ARTIFACTS_DIR}/kubeconfig")

  # Bring up cluster to host 'Infra Server'
  infractl -e $PROD_INFRA_ENDPOINT create $INFRA_CLUSTER_FLAVOR $INFRA_CLUSTER_NAME \
    --description $INFRA_CLUSTER_NAME --lifespan="3h" --arg "nodes=1" \
    --arg "machine-type=e2-standard-2" --slack-me --wait
  infractl -e $PROD_INFRA_ENDPOINT get $INFRA_CLUSTER_NAME
  infractl -e $PROD_INFRA_ENDPOINT lifespan $INFRA_CLUSTER_NAME '+24h'
  purge_dir $INFRA_ARTIFACTS_DIR
  INFRA_ARTIFACTS_INFO=$(infractl -e "$PROD_INFRA_ENDPOINT" artifacts \
    "$INFRA_CLUSTER_NAME" --download-dir="$INFRA_ARTIFACTS_DIR" --json)
  INFRA_CLUSTER_CONTEXT=$(yq e '.contexts[0].name' $INFRA_ARTIFACTS_DIR/kubeconfig)
  kubectl --context "$INFRA_CLUSTER_CONTEXT" config view --raw --minify

  function infra_server_build_deploy {
    cd ~/source/infra
    make configuration-download
    kubectl --context "$INFRA_CLUSTER_CONTEXT" delete deployments/infra-server-deployment -n infra
    kubectl config use-context "$INFRA_CLUSTER_CONTEXT"
    make deploy-local
    bounce_infra_server_pod

    # Setup port forwarding
    pkill -f "$INFRA_CLUSTER_CONTEXT.*port-forward.*8443"
    kubectl --context "$INFRA_CLUSTER_CONTEXT" -n infra \
      port-forward deployment/infra-server-deployment 8443:8443 &>/tmp/infra-server.log &
    sleep 2 # wait for port forwarding to be established

    # Check logs and nodes
    assert_infra_server_version $LOCAL_INFRA_ENDPOINT
    kubectl --context "$INFRA_CLUSTER_CONTEXT" -n infra logs deployment/infra-server-deployment -f
    kubectl --context "$INFRA_CLUSTER_CONTEXT" top nodes
  }

  # Shell into pod for debugging
  function shell_into_infra_server_pod {
    kubectx "$INFRA_CLUSTER_CONTEXT"; kubens "infra"
    pod=$(kubectl get pods -l "app=infra-server" -o json | jq -r '.items[0].metadata.name')
    kubectl exec -it $pod -- ls -l ./configuration/
    kubectl exec -it $pod -- cat ./configuration/workflow-openshift-4-demo.yaml
  }

  # Launch a cluster via 'Infra Server' (local)
  function launch_gui {
    # Safari allows bad cert with warning and user acknowledgement
    open -a Safari "https://$LOCAL_INFRA_ENDPOINT/"
    # Chrome also works if "chrome://flags/#allow-insecure-localhost" is enabled
    open -a "Google Chrome" "https://$LOCAL_INFRA_ENDPOINT/"
  }

  function launch_small_cluster {
    infractl -k -e $LOCAL_INFRA_ENDPOINT create $TEST_CLUSTER_FLAVOR $TEST_CLUSTER_NAME \
      --description $TEST_CLUSTER_NAME --lifespan="6h" --arg "nodes=3" \
      --arg "machine-type=n1-standard-4" --slack-me
  }

  function launch_default_cluster {  # <---- Verified this works now
    infractl -k -e $LOCAL_INFRA_ENDPOINT create $TEST_CLUSTER_FLAVOR $TEST_CLUSTER_NAME \
      --description $TEST_CLUSTER_NAME --lifespan="3h" --slack-me
  }

  function modify_cluster_lifespan {
    infractl -k -e $LOCAL_INFRA_ENDPOINT get $TEST_CLUSTER_NAME
    infractl -k -e $LOCAL_INFRA_ENDPOINT lifespan $TEST_CLUSTER_NAME '+24h'
  }

  function observe_cluster_management_workflows {
    argo --context "$INFRA_CLUSTER_CONTEXT" list
    argo --context "$INFRA_CLUSTER_CONTEXT" logs "$TEST_CLUSTER_NAME"
    infractl -k -e "$LOCAL_INFRA_ENDPOINT" list --all --expired
    kubectl --context "$INFRA_CLUSTER_CONTEXT" get workflows -n default
    kubectl --context "$INFRA_CLUSTER_CONTEXT" describe workflow/$TEST_CLUSTER_NAME -n default
  }

  function observe_cluster_resources {
    # Observe cluster resources for GKE clusters (and OpenShift on GCP):
    open -a Safari "https://console.cloud.google.com/kubernetes/list?project=srox-temp-dev-test"
    # Observe cluster resources for sales demo clusters
    open -a Safari "https://console.cloud.google.com/kubernetes/list?project=srox-temp-sales-demos"
    # Observe cluster resources for static demo clusters
    open -a Safari "https://console.cloud.google.com/kubernetes/list?project=ultra-current-825"
  }

  function gcloud_list_clusters_all_projects {  # TODO: Where is my demo-1 cluster?
    gcp_project_ids=$(gcloud projects list --format=json | jq -r '.[] | .projectId')
    for project_id in $gcp_project_ids; do
      echo "==== $project_id ===="
      gcloud container clusters list --project "$project_id" --format=json \
	| jq -r '.[] | {"name":.name, "zone":.zone, "currentNodeCount":.currentNodeCount}'
    done
  }

  # Check access to the child cluster (used to setup config.env for ansible-demo)
  purge_dir $TEST_ARTIFACTS_DIR
  TEST_ARTIFACTS_INFO=$(infractl -k -e "$LOCAL_INFRA_ENDPOINT" artifacts "$TEST_CLUSTER_NAME" \
    --download-dir="$TEST_ARTIFACTS_DIR" --json)
  TEST_CLUSTER_CONTEXT=$(yq e '.contexts[0].name' $TEST_ARTIFACTS_DIR/kubeconfig)
  kubectl config use-context "$TEST_CLUSTER_CONTEXT"
  kubectl --context "$TEST_CLUSTER_CONTEXT" config view --raw --minify | base64 | pbcopy
  kubectl --context "$TEST_CLUSTER_CONTEXT" top nodes

  # Manually access the k8s api-server
  # https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/
  kubectl --context "$TEST_CLUSTER_CONTEXT" proxy --port=8080 &
  curl http://localhost:8080/api/
  pkill -f 'kubectl.*proxy.*port=8080'

  # Test StackRox deployment
  docker login stackrox.io
  cd ~/source/workflow && echo yes | ./bin/teardown
  cd ~/source/rox && ./deploy/k8s/deploy-local.sh
  ps | grep "kubectl.*port-forward" | grep -v grep
  STACKROX_CENTRAL_USER="admin"
  STACKROX_CENTRAL_PASSWORD=$(cat deploy/k8s/central-deploy/password)
  echo "Log in to central with: $STACKROX_CENTRAL_USER/$STACKROX_CENTRAL_PASSWORD"
  open "https://$LOCAL_CENTRAL_ENDPOINT/"

  # Test ansible-demo installer (and troubleshooting config.env)
  cd ~/source/ansible-demo
  DOCKERCONFIG_BASE64=$(pass show RS_156_DOCKERCONFIG_BASE64)
  echo $DOCKERCONFIG_BASE64| base64 -d
  echo $DOCKERCONFIG_BASE64| base64 -d | jq -r '.auths."gcr.io".auth' | base64 -d
  TEST_CLUSTER_PUBLIC_IP=$(kubectl --context "$TEST_CLUSTER_CONTEXT" \
    config view --raw --minify -o json | jq -r '.clusters[0].cluster.server' | sed -e 's#https://##')
  echo -e "TEST_CLUSTER_KUBE_VERSION:" && curl -ks https://$TEST_CLUSTER_PUBLIC_IP/version
  vim config.env  # <--- AS NEEDED
  yq e '.stackroxLicense' ~/source/infra/chart/infra-server/configuration/development-values.yaml
  cat config.env | sed -ne 's/KUBECONFIG_BASE64=//p' | base64 -d | yq e
  cat config.env | sed -ne 's/DOCKERCONFIG_BASE64=//p' | base64 -d | jq
  cat config.env | sed -ne 's/DOCKERCONFIG_BASE64=//p' | base64 -d | jq -r '.auths."gcr.io".auth' | base64 -d
  kubectl config use-context "$TEST_CLUSTER_CONTEXT"
  docker-compose run ansible-demo-build

  # troubleshooting ansible-demo installer
  docker run -it --rm --entrypoint '' --env-file config.env \
    -v "$(pwd)/playbooks/main.yml:/ansible/playbooks/main.yml" \
    us.gcr.io/rox-se/ansible-demo:latest \
    sh -c "ansible-playbook -i inventory.yml main.yml --skip-tags skip \
      && echo -e '\n----' && cat /ansible/playbooks/files/config.json \
      && echo -e '\n----' && cat /ansible/playbooks/files/kubeconfig"

  function teardown_clusters {
    infractl -k -e "$LOCAL_INFRA_ENDPOINT" delete "$TEST_CLUSTER_NAME"
    infractl -e "$PROD_INFRA_ENDPOINT" delete "$INFRA_CLUSTER_NAME"
  }
}

# Deploy to development cluster (dev.infra.rox.systems)
function build_and_deploy_development {
  gcloud container clusters get-credentials "infra-development" --project stackrox-infra --region us-west2
  cd ~/source/infra
  make configuration-download
  kubectl --context "$DEV_INFRA_KUBECONTEXT_NAME" delete deployments/infra-server-deployment -n infra
  make deploy-development
  bounce_infra_server_pod
  assert_infra_server_version
}

# Deploy to development cluster (infra.rox.systems)
function build_and_deploy_production {
  gcloud container clusters get-credentials "infra-production" --project stackrox-infra --region us-west2
  cd ~/source/infra
  make configuration-download
  kubectl --context "$PROD_INFRA_KUBECONTEXT_NAME" delete deployments/infra-server-deployment -n infra
  make deploy-production
  bounce_infra_server_pod
  assert_infra_server_version
}

function manual_cleanup {
  open "https://console.cloud.google.com/kubernetes/list?authuser=1&project=stackrox-hub"
  open "https://console.cloud.google.com/kubernetes/list?authuser=1&project=srox-temp-dev-test"
  cat <<EOF
Delete these resources via the web console:

* gcp/PROJECT=srox-temp-dev-test/cloud_dns/ZONE=openshift-infra-rox-systems/
     A_RECORD=*.apps.shane-rs-156-demo-1.openshift.infra.rox.systems.
* gcp/PROJECT=srox-temp-dev-test/cloud_dns/ZONE=openshift-infra-rox-systems/
     A_RECORD=api.shane-rs-156-demo-1.openshift.infra.rox.systems.
* gcp/PROJECT=srox-temp-dev-test/cloud_dns/ZONE=shane-rs-156-demo-1-qrkzc-private-zone/
     A_RECORD=*.apps.shane-rs-156-demo-1.openshift.infra.rox.systems.
* gcp/PROJECT=srox-temp-dev-test/cloud_dns/ZONE=shane-rs-156-demo-1-qrkzc-private-zone/
     A_RECORD=api-int.shane-rs-156-demo-1.openshift.infra.rox.systems.
* gcp/PROJECT=srox-temp-dev-test/cloud_dns/ZONE=shane-rs-156-demo-1-qrkzc-private-zone/
     A_RECORD=api.shane-rs-156-demo-1.openshift.infra.rox.systems.
* gcp/PROJECT=srox-temp-dev-test/cloud_dns/ZONE=shane-rs-156-demo-1-qrkzc-private-zone
* gcp/PROJECT=srox-temp-dev-test/compute_engine/shane-rs-156-demo-1-*
EOF
}
