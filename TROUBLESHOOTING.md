# Infra Troubleshooting

## Useful Links

Cloudflare rox.systems Zone

https://dash.cloudflare.com/78bd9269af1a43e82ef944523786fd80/rox.systems/dns

GCP infra.rox.systems Zone

https://console.cloud.google.com/net-services/dns/zones/infra-rox-systems?project=stackrox-infra&organizationId=847401270788

Auth0 Application

https://manage.auth0.com/dashboard/us/sr-dev/applications/AsyLUYxwV2GX2oG0PjwTXhMlxHuI7qmE/settings

Argo Releases and CLI

https://github.com/argoproj/argo/releases

GKE Production Cluster

https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-production?project=stackrox-infra&organizationId=847401270788

Production Address

http://infra.stackrox.com

GKE Development Cluster

https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-development?project=stackrox-infra&organizationId=847401270788

Development Address

http://infra.stackrox.com

## Troubleshooting with Argo

Argo is used as the underlying workflow orchestrator; Argo Workflow specs are submitted when a user requests a cluster, and Argo then consumes that spec and wires together a sequence of pods with volumes/secrets/logs/etc.
By using the Argo CLI (`argo`) we can get a low-level view of the world, and debug potential issues.

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

(it's just a pod, so `kubectl logs ...` would also work)

```
argo logs demo-mxgf9-3875809567 | head -n 20
[PASS] /tmp/google-credentials.json (GCP service account credential file)
[PASS] /tmp/google-scanner-credentials.json (GCP service account credential file for scanner to pull images from GCR)
[PASS] /usr/bin/roxctl (A copy of the roxctl binary)
[PASS] AUTH_CLIENT_ID (Auth0 integration client ID)
[PASS] AUTH_DOMAIN (Auth0 tenant)
[PASS] DOCKER_IO_PASSWORD (password for Docker Hub)
[PASS] DOCKER_IO_USERNAME (username for Docker Hub)
```

If a step is stuck in a pending state (e.x. a referenced secret doesn't exist), `argo get` should show that information.
Otherwise `kubectl describe po ...` can be used to see exact reasons.

