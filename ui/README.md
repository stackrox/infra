# StackRox Infra App (UI)

This sub-project contains Web UI (SPA) for StackRox Infra.

This project was bootstrapped with
[Create React App](https://github.com/facebook/create-react-app).

You can learn more in the
[Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started)
about the available scripts and the tooling behavior.

## Development

### Build Tooling

One easy route to current stackrox build tooling is to use
https://github.com/stackrox/stackrox-env.

- [Node.js](https://nodejs.org/en/) `10.15.3 LTS` or higher (it's highly
  recommended to use an LTS version, if you're managing multiple versions of
  Node.js on your machine, consider using
  [nvm](https://github.com/creationix/nvm))
- [Yarn](https://yarnpkg.com/en/)

### GitHub Packages NPM Registry

This project depends on packages with `@stackrox` scope accessible from GitHub
Packages NPM registry. Get access with: 
```
npm login --auth-type=legacy --registry=https://npm.pkg.github.com
```
Use your github username and a token with `repo` and `read:packages` rights.
More details can be found
[here](https://docs.engineering.redhat.com/display/StackRox/Using+GitHub+Packages+with+NPM).

### UI Dev Server

To avoid a connection error with node v1.17+ set:
```
export NODE_OPTIONS=--openssl-legacy-provider
```

_If you're going to use `yarn` instead of `make` targets, make sure you've run
`yarn install` to download dependencies._

`make start-dev-server` OR `yarn start` will start the UI dev server and open UI
in a browser window that will auto-refresh on any source code or CSS changes.

By default UI dev server will try to proxy API requests to
`https://dev.infra.rox.systems`. To override the API endpoint use
`INFRA_API_ENDPOINT` env var. For example if you are only changing `ui/` code
you can interact with the production infra instance via:
```
INFRA_API_ENDPOINT=https://infra.rox.systems yarn start
```

To access the API you need to copy a `token` cookie from a session with the
infra instance you are using to the browser window that appears when you execute
`yarn start`.

### Generated Sources

Some of the UI code has been generated automatically and checked in, like API
client and models. To re-generate the sources run `make gen-src` or
`yarn gen:src`.

_Hint: for the API client to generate new Swagger definitions from protos in the
parent dir run `make proto-generated-srcs`._
