# Infra Troubleshooting

## Useful Links

Cloudflare rox.systems Zone

https://dash.cloudflare.com/78bd9269af1a43e82ef944523786fd80/rox.systems/dns

GCP infra.rox.systems Zone

https://console.cloud.google.com/net-services/dns/zones/infra-rox-systems?project=stackrox-infra&organizationId=847401270788

Argo Releases and CLI

https://github.com/argoproj/argo/releases

GKE Production Cluster

https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-production?project=stackrox-infra&organizationId=847401270788

Production Address

http://infra.rox.systems

GKE Development Cluster

https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-development?project=stackrox-infra&organizationId=847401270788

Development Address

http://infra.rox.systems

## Troubleshooting with Argo

Argo is used as the underlying workflow orchestrator; Argo Workflow specs are
submitted when a user requests a cluster, and Argo then consumes that spec and
wires together a sequence of pods with volumes/secrets/logs/etc.

`brew install argo`

By using the Argo CLI (`argo`) we can get a low-level view of the world, and
debug potential issues.

To list all workflows, run:

```
argo list
NAME         STATUS    AGE   DURATION   PRIORITY
demo-mxgf9   Running   18s   18s        0
```

To view details and DAG status for a specific workflow, run:

```
argo get demo-mxgf9 -o wide
Name:                demo-mxgf9
Namespace:           default
ServiceAccount:      default
Status:              Running
Created:             Mon Jun 01 13:43:12 -0700 (42 seconds ago)
Started:             Mon Jun 01 13:43:12 -0700 (42 seconds ago)
Duration:            42 seconds
Parameters:
  name:              june1demo1
  main-image:        stackrox.io/main:3.0.43.1
  scanner-image:     stackrox.io/scanner:2.2.6
  scanner-db-image:  stackrox.io/scanner-db:2.2.6

STEP                    PODNAME                DURATION  ARTIFACTS  MESSAGE
 ● demo-mxgf9 (start)
 ├---✔ roxctl (roxctl)  demo-mxgf9-522422286   9s        roxctl
 └---● create (create)  demo-mxgf9-3875809567  32s
```

To get logs from a step, run:

```
argo logs demo-mxgf9-3875809567 | head -n 20
[PASS] /tmp/google-credentials.json (GCP service account credential file)
[PASS] /tmp/google-scanner-credentials.json (GCP service account credential file for scanner to pull images from GCR)
[PASS] /usr/bin/roxctl (A copy of the roxctl binary)
[PASS] AUTH_CLIENT_ID (Auth0 integration client ID)
[PASS] AUTH_DOMAIN (Auth0 tenant)
```

The workflow steps are just pod executions, so `kubectl logs ...` also works:
```
MY_CLUSTER_NAME=demo-mxgf9
WORKFLOW_STEP_POD_ID=$(argo get oc-init-bundle-test -o json | jq -r '.status.nodes[] | select(.type=="Pod") | select(.displayName=="create") | .id')
kubectl logs $WORKFLOW_STEP_POD_ID -c 'main'
# Retrieving logs via argo cli obviates the need to specify the container
argo logs $MY_CLUSTER_NAME $WORKFLOW_STEP_POD_ID --follow
```

If a step is stuck in a pending state (e.x. a referenced secret doesn't exist),
`argo get` should show that information. Otherwise `kubectl describe po ...` can
be used to see exact reasons.

Note: The containers used to execute an argo workflow are created in the
`default` k8s namespace.

## Submit an Argo workflow directly

For a faster development iteration it can be useful to submit workflow changes
directly to Argo and bypass `infra`. For example, if you only change the
automation build image or the parameters for a workflow.

```
# make sure you are using the development cluster
$ kubectl config current-context
gke_stackrox-infra_us-west2_infra-development

# submit the eks workflow with some parameters
$ argo submit -p 'name=your-unique-name' -p 'user-arns=arn:aws:iam::051999192406:user/you@stackrox.com' chart/infra-server/static/workflow-eks.yaml
Name:                eks-xrhx4
Namespace:           default
ServiceAccount:      default
...

# check on progress
$ argo logs eks-xrhx4
...

# if it fails you may need to manually delete your cluster!

# a typical cluster build workflow will wait at its wait step
$ argo list
NAME        STATUS                AGE   DURATION   PRIORITY
eks-2h8gb   Running (Suspended)   24m   24m        0

# and can be resumed to test steps after that i.e. delete
$ argo resume eks-2h8gb
```

## View the Argo Workflows web interface

```
kubectl -n argo port-forward deployment/argo-server 2746:2746
open http://localhost:2746/workflows/
```
