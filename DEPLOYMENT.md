# Infra Deployment

## Clusters

### [Development](https://console.cloud.google.com/kubernetes/clusters/details/us-west2/infra-development?project=stackrox-infra&organizationId=847401270788)

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

## Configuration

Service configuration is [stored in a GCS bucket](https://console.cloud.google.com/storage/browser/infra-configuration?organizationId=847401270788&project=stackrox-infra).

You will need to download this configuration if you plan to do a deployment update.

To download the configuration locally to `chart/infra-server/configuration`, run:

`make configuration-download`

To upload the local configuration back to the bucket, run:

`make configuration-upload`

## Deployment

Deployments consist of an installation of Argo, as well as the various service/flavor components.

to build and push an image, run:

`make push`

### Development

To render a copy of the charts (for inspection), run:

`make render-development`

To then apply that chart to the development cluster, run:

`make install-development`

To do everything in one command, run:

`make deploy-development`

Note: The deployment will not execute the latest image if the version string
does not change. See the output from `make tag` versus the version reported by
the `infra-server`.

### Production

To render a copy of the charts (for inspection), run:

`make render-production`

To then apply that chart to the development cluster, run:

`make install-production`

To do everything in one command, run:

`make deploy-production`

## Verification

After deploying the service, browse to the appropriate endpoint to verify that you  can login and view the UI.

| Environment | URL |
| --- | --- |
| Development | http://dev.infra.rox.systems |
| Production | https://infra.rox.systems |

Download a copy of `infractl` and export your token. Verify API connectivity:

| Environment | Command |
| --- | --- |
| Development | `infractl -e dev.infra.rox.systems:443 whoami` |
| Production | `infractl whoami` |

