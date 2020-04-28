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

### UI Dev Server

_Before starting, make sure you have the above tools installed on your machine
and you've run `yarn install` to download dependencies._

`yarn start` command will start the UI dev server and open UI in a browser
window that will auto-refresh on any source code or CSS changes.

By default UI dev server will try to proxy API requests to
`https://dev.infra.stackrox.com`. To override the API endpoint use
`INFRA_API_ENDPOINT` env var. I.e. you can start the dev server via
`export INFRA_API_ENDPOINT=<api_endpoint>; yarn start`.
