# StackRox Setup Next App (UI)

This sub-project contains Web UI (SPA) for StackRox Setup Next (aka StackRox
Infra).

This project was bootstrapped with
[Create React App](https://github.com/facebook/create-react-app).

You can learn more in the
[Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started)
about the available scripts and the tooling behavior.

## Development

### Build Tooling

- [Docker](https://www.docker.com/)
- [Node.js](https://nodejs.org/en/) `10.15.3 LTS` or higher (it's highly
  recommended to use an LTS version, if you're managing multiple versions of
  Node.js on your machine, consider using
  [nvm](https://github.com/creationix/nvm))
- [Yarn](https://yarnpkg.com/en/)

### Development

_Before starting, make sure you have the above tools installed on your machine
and you've run `yarn install` to download dependencies._

By default UI dev server will be looking for APIs at `https://localhost:8443`.To
override it use `DEV_INFRA_API_ENDPOINT` env var. I.e. you can start dev server
via `export DEV_INFRA_API_ENDPOINT=<remote_endpoint>; yarn start`.
