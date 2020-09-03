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

You will need to download this configuration if you plan to make a change to infra. Configuration changes are baked in to the `infra-server` image at build time. 

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

