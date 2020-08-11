# StackRox Infra App (UI)

This sub-project contains Web UI (SPA) for StackRox Infra.

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

### GitHub Packages NPM Registry

This project depends on packages with `@stackrox` scope accessible from GitHub
Packages NPM registry. Setup your dev env by following
[these instructions](https://stack-rox.atlassian.net/wiki/spaces/ENGKB/pages/1411515467/Using+GitHub+Packages+with+NPM#Setting-Up-Dev-Env)
to authenticate with the registry.

### UI Dev Server

_If you're going to use `yarn` instead of `make` targets, make sure you've run
`yarn install` to download dependencies._

`make start-dev-server` OR `yarn start` will start the UI dev server and open UI
in a browser window that will auto-refresh on any source code or CSS changes.

By default UI dev server will try to proxy API requests to
`https://dev.infra.rox.systems`. To override the API endpoint use
`INFRA_API_ENDPOINT` env var. I.e. you can start the dev server via
`export INFRA_API_ENDPOINT=<api_endpoint>; yarn start`.

### Generated Sources

Some of the UI code has been generated automatically and checked in, like API
client and models. To re-generate the sources run `make gen-src` or
`yarn gen:src`.

_Hint: for the API client to generate new Swagger definitions from protos in the
parent dir run `make proto-generated-srcs`._
