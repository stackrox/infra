[![Dev][dev-badge]][dev-link]
[![Prod][prod-badge]][prod-link]

# Infra

🌧️ Automated infrastructure and demo provisioning

## Development

Infra (the server) and infractl (the cli) are written in Go, and use gRPC for
client-server communication. The UI uses a React/Typescript/Yarn toolchain (see
(ui/README.md)[ui/README.md]).

While a development workflow can be achieved using a locally installed
toolchain, it is also possible to rely on CI. CI will lint, build and push the
infra server. And then deploy it to a development cluster created using the
production infra deployment. A
(comment)[https://github.com/stackrox/infra/pull/711#issuecomment-1270457578]
will appear on PRs with more detail.

### Debugging

If opened in VS Code, you will find configurations for debugging CLI and Server.

#### CLI

To debug a CLI command, update the VS Code launch configuration "Debug CLI" with the desired command and update the `INFRA_TOKEN` value if necessary.
You may also point to a different infra server with the `--endpoint` flag.

#### Server

To debug the server, you need to fulfil the prerequisites first.

1. Have an authenticated `gcloud` CLI and GNU `sed` installed.
1. Have your `KUBECONFIG` point to a cluster where Argo Workflows and the ConfigMaps and Secrets for infra are deployed. This is most easily achieved by connecting to a PR cluster or deploying infra with `ENVIRONMENT=<DEVELOPMENT,PRODUCTION> make install-argo clean-argo-config helm-deploy` to a new or local cluster. This cluster will only be used to run workflows.
1. Run `make prepare-local-server-debugging` to set the contents of the `configuration` directory and compile the UI + CLI (for downloads).

Then, you can use the "Debug Server" launch configuration.

### Regenerate Go bindings from protos

To regenerate the Go proto bindings, run:

`make proto-generated-srcs`

### Building the server and cli

To compile a server and client binary, run:

`make server cli-local`

### Building or pushing images

GitHub Actions will build and push the infra-server image based on `make tag` of
the most recent commit. Or you can build and push locally if you have the
correct tooling installed with:

`make image push`

## Deployment

For additional information on how this service is deployed, please refer to the [deployment instructions](DEPLOYMENT.md).

## Runbook

For additional information on how to debug and remediate issues with the deployed service, please refer to the [runbook instructions](TROUBLESHOOTING.md).

[dev-badge]:      https://img.shields.io/badge/infra-development-green
[dev-link]:       https://infra.rox.systems
[prod-badge]:     https://img.shields.io/badge/infra-production-green
[prod-link]:      https://infra.rox.systems

