# Infra Deployment

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

GitHub Actions will build and push the infra-server image based on `make tag` of
the most recent commit. Or you can build and push locally if you have the
correct tooling installed with:

`make image push`

### Development/Staging

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

