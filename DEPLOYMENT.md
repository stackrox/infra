# Infra Deployment

## Deploy to an adhoc development cluster

For example one created with `infractl create gke-default`.

To deploy to such a cluster simply:

```
make deploy-local
```

The infra server should start and argo should deploy.

```
$ kubectl -n infra get pods
NAME                                      READY   STATUS    RESTARTS   AGE
infra-server-deployment-5c6cfb69c-54k6x   1/1     Running   0          11s
$ kubectl -n argo get pods
NAME                                        READY   STATUS    RESTARTS   AGE
argo-server-58bf6d4f79-cc96j                1/1     Running   1          95s
argo-workflow-controller-6487cc4688-cdbfz   1/1     Running   0          95s
```

To connect to the infra-server run a proxy:

```
kubectl -n infra port-forward svc/infra-server-service 8443:8443
```

Then use *safari* to connect to the UI if needed. (note: chrome will not accept
the infra self-signed cert).

Or the locally compiled infractl binary:

```
bin/infractl-darwin-amd64 -k -e localhost:8443 whoami
```

### Notes

For clusters created in the `srox-temp-dev-test` to be able to pull images from
the `stackrox-infra` `us.gcr.io` and `gcr.io` registries, the
`srox-temp-dev-test` default compute service account requires *Storage Object Viewer* access to
`artifacts.stackrox-infra.appspot.com` and
`us.artifacts.stackrox-infra.appspot.com`.

For other clusters e.g. `docker-desktop` image pull secrets will work after the
deployment has created the namespaces. e.g.

```
kubectl create secret docker-registry infra-us-gcr-access --docker-server=us.gcr.io --docker-username=_json_key \
    --docker-password="$(cat chart/infra-server/configuration/production/gke/gke-credentials.json)" --docker-email=infra@stackrox.com
kubectl create secret docker-registry infra-gcr-access --docker-server=gcr.io --docker-username=_json_key \
    --docker-password="$(cat chart/infra-server/configuration/production/gke/gke-credentials.json)" --docker-email=infra@stackrox.com
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "infra-gcr-access"},{"name": "infra-us-gcr-access"}]}'

kubectl -n infra create secret docker-registry infra-us-gcr-access --docker-server=us.gcr.io --docker-username=_json_key \
    --docker-password="$(cat chart/infra-server/configuration/production/gke/gke-credentials.json)" --docker-email=infra@stackrox.com
kubectl -n infra create secret docker-registry infra-gcr-access --docker-server=gcr.io --docker-username=_json_key \
    --docker-password="$(cat chart/infra-server/configuration/production/gke/gke-credentials.json)" --docker-email=infra@stackrox.com
kubectl -n infra patch serviceaccount default -p '{"imagePullSecrets": [{"name": "infra-gcr-access"},{"name": "infra-us-gcr-access"}]}'
```

## Production and Staging Clusters

To work with either of the clusters in `project=stackrox-infra` you will need to either be a member of the `team-automation` group or have someone add you as a project owner.

### [Development (Staging)](https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-development?project=stackrox-infra&organizationId=847401270788)

To connect to this cluster using kubectl, run:

```
gcloud container clusters get-credentials infra-development \
    --project stackrox-infra \
    --region us-west2
```

### [Production](https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-production?project=stackrox-infra&organizationId=847401270788)

To connect to this cluster using kubectl, run:

```
gcloud container clusters get-credentials infra-production \
    --project stackrox-infra \
    --region us-west2
```

## Ingress

Infra uses GKE `Ingress` and `ManagedCertificate` CRDs to handle ingress. Plus two global static IPs:

```
$ gcloud compute addresses list --project stackrox-infra
NAME                       ADDRESS/RANGE   TYPE      PURPOSE  NETWORK  REGION  SUBNET  STATUS
infra-address-development  35.227.221.195  EXTERNAL                                    IN_USE
infra-address-production   35.227.207.252  EXTERNAL                                    IN_USE
```

## Configuration

Service configuration is [stored in a GCS bucket](https://console.cloud.google.com/storage/browser/infra-configuration?organizationId=847401270788&project=stackrox-infra).

You will need to download this configuration if you plan to make a change to infra. Configuration changes are baked in to the `infra-server` image at build time.

To download the configuration locally to `chart/infra-server/configuration`, run:

`make configuration-download`

To upload the local configuration back to the bucket, run:

`make configuration-upload`

## Creating a Tag for Release

To create a full GitHub release, draft a new release from the console.
Edit the release to include a summary of key features, changes, deprecations,
etc since the last full release.

```bash
# find the next tag
git fetch --tags
git tag -l

# review commits between last release tag and head of mainline branch
git log --decorate --graph --abbrev-commit --date=relative 0.2.13..master
```

We often deploy Infra from a tag without creating a full GitHub release.
To create a tag for deployment under this scenario:

```bash
cd $GOPATH/src/github.com/stackrox/infra
git tag 0.2.14  # for example
git push origin --tags
```

Prior to deployment make note of the current version of infra in case a rollback is needed.
A rollback consists of checking out the previously deployed tag and redeploying.

    https://infra.rox.systems/      => version 0.2.13
    https://dev.infra.rox.systems/  => version 0.2.13

Once the tag is ready for deployment &mdash; via full release or manually pushing a
new tag &mdash; the next step is to deploy to target environments.

## Deployment

Deployments consist of an installation of Argo, as well as the various service/flavor components.

To build and push an image, run:

`make push`

### Development

To render a copy of the charts (for inspection), run:

`make render-development`

To then apply that chart to the development cluster, run:

`make install-development`

To do everything in one command, run:

`make deploy-development`

Note: If the development server version does not change then `make deploy-development`
will not result in a running `infra-server` that reflects
your local changes. A brute force way to ensure an update is to delete the
`infra-server-deployment` deployment in the `infra` namespace. It will be recreated by
`make deploy-development`.

### Production

To render a copy of the charts (for inspection), run:

`make render-production`

To then apply that chart to the development cluster, run:

`make install-production`

To do everything in one command, run:

`make deploy-production`

## Verification

After deploying the service, browse to the appropriate endpoint to verify that you can login and view the UI.

| Environment | URL |
| --- | --- |
| Development | http://dev.infra.rox.systems |
| Production | https://infra.rox.systems |

Download a copy of `infractl` and export your token. Verify API connectivity:

| Environment | Command |
| --- | --- |
| Development | `infractl -e dev.infra.rox.systems:443 whoami` |
| Production | `infractl whoami` |

