[![CircleCI][circleci-badge]][circleci-link]
[![Dev][dev-badge]][dev-link]
[![Prod][prod-badge]][prod-link]

# Infra

🌧️ Automated infrastructure and demo provisioning

## Development

Infra (the server) and infractl (the cli) are written in Go, and use gRPC for client-server communication.

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

`make image` or `make push`

## Deployment

For additional information on how this service is deployed, please refer to the [deployment instructions](https://github.com/stackrox/infra/blob/master/DEPLOYMENT.md).

## Runbook

For additional information on how to debug and remediate issues with the deployed service, please refer to the [runbook instructions](https://github.com/stackrox/infra/blob/master/TROUBLESHOOTING.md).

[circleci-badge]: https://circleci.com/gh/stackrox/infra.svg?style=shield&circle-token=afa342906b658b5349c68b70fa82fd85d1422212
[circleci-link]:  https://circleci.com/gh/stackrox/infra
[dev-badge]:      https://img.shields.io/badge/infra-development-green
[dev-link]:       https://infra.rox.systems
[prod-badge]:     https://img.shields.io/badge/infra-production-green
[prod-link]:      https://infra.rox.systems
