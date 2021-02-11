#!/usr/bin/env bash
set -eu
exit 1

###############################################################################
#                           Troubleshoting Examples                           #
###############################################################################

export INFRA_TOKEN=$(pass show INFRA_TOKEN)
export PROD_INFRA_ENDPOINT="infra.rox.systems:443"
export DEV_INFRA_ENDPOINT="dev.infra.rox.systems:443"


# infra-server auth check
infractl -e $PROD_INFRA_ENDPOINT whoami

# infra-server view logs
kubectl logs deployments/infra-server-deployment -n infra -f

# launch cluster -- demo
MY_CLUSTER_NAME="shane-infra-demo-test";
MY_CLUSTER_FLAVOR="demo"
infractl -e $PROD_INFRA_ENDPOINT create $MY_CLUSTER_FLAVOR $MY_CLUSTER_NAME --description $MY_CLUSTER_NAME --lifespan="90m" --slack-me
infractl -e $PROD_INFRA_ENDPOINT get $MY_CLUSTER_NAME

# launch cluster -- aks
MY_CLUSTER_NAME="shane-infra-aks-test";
MY_CLUSTER_FLAVOR="aks"
infractl -e $PROD_INFRA_ENDPOINT create $MY_CLUSTER_FLAVOR $MY_CLUSTER_NAME --description $MY_CLUSTER_NAME --lifespan="30m" --slack-me
infractl -e $PROD_INFRA_ENDPOINT get $MY_CLUSTER_NAME

# launch cluster -- eks
MY_CLUSTER_NAME="shane-infra-eks-test";
MY_CLUSTER_FLAVOR="eks"
infractl -e $PROD_INFRA_ENDPOINT create $MY_CLUSTER_FLAVOR $MY_CLUSTER_NAME \
  --description $MY_CLUSTER_NAME --lifespan="90m" --slack-me \
  --arg user-arns="arn:aws:iam::051999192406:user/automation/setup_automation"
infractl -e $PROD_INFRA_ENDPOINT get $MY_CLUSTER_NAME

# launch cluster -- gke-default
MY_CLUSTER_NAME="shane-infra-gke-test";
MY_CLUSTER_FLAVOR="gke-default"
infractl -e $PROD_INFRA_ENDPOINT create $MY_CLUSTER_FLAVOR $MY_CLUSTER_NAME --description $MY_CLUSTER_NAME --lifespan="30m" --slack-me

# inspect argo workflows and workflow step logs
kubectl get workflow/$MY_CLUSTER_NAME -o json | jq -r '.status.nodes[] | select(.displayName=="create") | .id'
WORKFLOW_STEP_POD_ID=$(argo get $MY_CLUSTER_NAME -o json | jq -r '.status.nodes[] | select(.type=="Pod") | select(.displayName=="create") | .id')
argo logs $MY_CLUSTER_NAME $WORKFLOW_STEP_POD_ID --follow
kubectl logs $WORKFLOW_STEP_POD_ID -c 'main' -f
kubectl get workflows -n default
kubectl describe workflow/$MY_CLUSTER_NAME

# using argo cli
argo list
argo list -l workflows.argoproj.io/workflow-template=wait --running -o name
argo get $MY_CLUSTER_NAME
argo logs $MY_CLUSTER_NAME --follow
argo "$ACTION" $MY_CLUSTER_NAME  # stop | terminate | delete

# viewing argo web console
kubectl -n argo port-forward deployment/argo-server 2746:2746
open -a Safari http://localhost:2746/workflows/

# switching kube context to new cluster
# prefer long-lived clusters merged into ~/.kube/config
# prefer short-lived clusters in ~/artifacts/kubeconfig (via infractl artifacts download)
export KUBECONFIG="~/.kube/config:~/artifacts/kubeconfig"
infractl -e $INFRA_ENDPOINT artifacts $CLUSTER_NAME --download-dir="~/artifacts"
kubectl port-forward -n stackrox svc/central-loadbalancer 8443:443
open -a Safari "https://localhost/"
open -a Safari "https://shane1.demo.stackrox.com/main/network"
kubectl get secrets/central-default-tls-cert -n stackrox -o json | jq -r '.data."tls.crt"' | base64 -d | openssl x509 -in /tmp/cert -text -noout | head -n15

# bump lifespan of cluster
infractl -e $PROD_INFRA_ENDPOINT lifespan $MY_CLUSTER_NAME '+1h'

# apply a json patch to a k8s resource with kubectl
kubectl get workflow shane-1 --output json | jq '.status.phase'
kubectl patch workflow shane-1 --type=json --patch='[{"op": "replace", "path": "/status/phase", "value": "Failed"}]'

# use roxctl against a demo cluster
export ROX_API_TOKEN=$(cat /tmp/shane-1.token)
export ROX_CENTRAL_ADDRESS="shane1.demo.stackrox.com:443"
roxctl -e $ROX_CENTRAL_ADDRESS central db backup
