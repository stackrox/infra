# Infra Deployment

## Production and Staging Clusters

To work with either of the clusters in `project=stackrox-infra` you will need to either be a member of the `team-automation` group or have someone add you as a project owner.

### [Staging (dev.infra.rox.systems)](https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-development?project=stackrox-infra&organizationId=847401270788)

When not in use, this cluster is resized to 0.

```
gcloud container clusters --project=stackrox-infra --zone=us-west2 resize infra-development --num-nodes=0
```

To bring it back for development:

```
gcloud container clusters --project=stackrox-infra --zone=us-west2 resize infra-development --num-nodes=4
```

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

Service configuration and secrets are stored in [GCP Secret Manager](https://console.cloud.google.com/security/secret-manager?project=stackrox-infra).

To view these, run:

`ENVIRONMENT=<development,production> SECRET_VERSION=<latest, 1,2,3,...> make secrets-download`.

This will download the secrets to `chart/infra-server/configuration/`.

- `<ENVIRONMENT>-values.yaml`: To show or edit a value, do it directly in this file, and use `ENVIRONMENT=<development,production> make secrets-upload` to upload the changes.
- `<ENVIRONMENT>-values-from-files.yaml`: To show or edit a value, use `ENVIRONMENT=<development,production> SECRET_VERSION=<latest,1,2,3> make secrets-<show, edit>` and follow the instructions. NOTE: This will download a fresh copy of the requested secret version and upload a new version after your changes. That ensures that your local secrets do not go stale.

## Regenerating the localhost certificates for the gRPC gateway

The connection for the gRPC gateway is secured by a self-generated "localhost" certificate.
To regenerate the certificate, run: `./scripts/cert/renew.sh <local|development|production>`.

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

Use the `deploy` Github action to update development or production environments with a new release.

### Argo Deployment

To install Argo workflow server, run:

`ENVIRONMENT=<development,production> make install-argo`

NOTE: This is a separate step and not a dependant chart for example to avoid too frequent Argo deployments.

### Manual deployment

To render a copy of the charts (for inspection), run:

`ENVIRONMENT=<development,production> SECRET_VERSION=<latest,1,2,3, ...> make helm-template`

To show the diff between the current Helm release and the charts, run:

`ENVIRONMENT=<development,production> SECRET_VERSION=<latest,1,2,3, ...> make helm-diff`

To then apply that chart to the cluster, run:

`ENVIRONMENT=<development,production> SECRET_VERSION=<latest,1,2,3, ...> make helm-deploy`

#### Test Mode

Use the environment variable `TEST_MODE` to disable certain infra service behavior, like:

`TEST_MODE=true ENVIRONMENT=development SECRET_VERSION=latest make helm-deploy`

## Verification

After deploying the service, browse to the appropriate endpoint to verify that you can login and view the UI.

| Environment | URL |
| --- | --- |
| Staging | http://dev.infra.rox.systems |
| Production | https://infra.rox.systems |

Download a copy of `infractl` and export your token. Verify API connectivity:

| Environment | Command |
| --- | --- |
| Staging | `infractl -e dev.infra.rox.systems:443 whoami` |
| Production | `infractl whoami` |

## Logging

The infra server logs are captured automatically by GCP.

- [Logs Explorer: Staging](https://cloudlogging.app.goo.gl/uSmEsjAmYR8Uyvyx9)
- [Logs Explorer: Production](https://cloudlogging.app.goo.gl/KqgSyE2mSq83M5Xs9)

Adding `jsonPayload."log-type"="audit"` to the query will filter for audit logs.
