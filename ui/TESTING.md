# Cypress E2E Testing

This directory contains Cypress E2E tests for the StackRox Infra UI.

## Testing Approach

The UI E2E tests require a local deployment because they use custom JWT
authentication that only works in `LOCAL_DEPLOY=true` mode. The tests cannot
share cookies from your browser session, so they generate test JWTs using a
hardcoded local development secret.

**Requirements:**

- Docker/Podman
- Local Kubernetes cluster (Colima, kind, minikube, Docker Desktop, etc.)
- Helm 3

**Why local deployment?**

- Tests need controlled authentication (can't use real OIDC in Cypress)
- Tests run in isolated browser context (can't share browser cookies)
- Enables offline testing and testing of local UI changes

## Quick Start

**Note:** All commands should be run from the repository root unless otherwise
specified.

### Steps

0. **Build** images:

   ```bash
   make image
   ```

   This builds the infra-server into a Docker image for local deployment.

1. **Deploy the local backend**:

   ```bash
   make deploy-local
   ```

2. **Start port-forwarding** to access the backend:

   ```bash
   kubectl port-forward -n infra svc/infra-server-service 8443:8443
   ```

   Keep this running in a separate terminal.

3. **Run the E2E tests**:

   ```bash
   cd ui
   INFRA_API_ENDPOINT=https://localhost:8443 npm run test:e2e
   ```

That's it! The `test:e2e` command will:

- Automatically start the UI dev server on http://localhost:3001
- Proxy API requests to your local backend at https://localhost:8443
- Run all Cypress E2E tests
- Shut down the dev server when tests complete

The tests run against the UI dev server at http://localhost:3001, which proxies
API requests to your local backend at `https://localhost:8443`.

### Test Results

After the tests complete:

- **Videos** are saved to `ui/cypress/videos/` (one per test file)
- **Screenshots** (on failures only) are saved to `ui/cypress/screenshots/`

Review the videos to verify the tests are properly accessing the backend.

## Interactive Mode

To run tests interactively with the Cypress UI (useful for debugging):

**Prerequisites:** Start the UI dev server with the backend endpoint configured:

```bash
cd ui
INFRA_API_ENDPOINT=https://localhost:8443 npm start
```

Keep this running, then in another terminal:

```bash
cd ui
npm run cypress:open
```

Then:

1. Select "E2E Testing"
2. Choose a browser
3. Click on any test file to run it

Interactive mode lets you see the tests run in real-time, inspect the DOM, and
debug failures.

## Test Structure

- `cypress/e2e/home.cy.ts` - Basic home page tests
- `cypress/e2e/flavor-selection.cy.ts` - Tests for flavor API integration

## Configuration

Tests are configured in `cypress.config.ts` to:

- Run against the UI dev server at `http://localhost:3001` (which proxies to the
  backend)
- Accept self-signed certificates (`chromeWebSecurity: false`)
- Capture videos of all test runs
- Capture screenshots on failures only
- Retry failed tests 2 times in CI mode (run mode), 0 times in interactive mode

The UI dev server (configured via `INFRA_API_ENDPOINT` environment variable)
proxies API requests to your local backend at `https://localhost:8443`.

## Adding More Tests

To add new E2E tests:

1. Create a new file in `cypress/e2e/` with the pattern `*.cy.ts`
2. Follow the existing test patterns for consistency
3. Run the tests locally before committing

## Troubleshooting

### Tests fail with "Cypress failed to verify that your server is running"

**Solution:** Make sure you're using `npm run test:e2e` which automatically
starts the dev server. If you want to run the dev server manually, use:

```bash
cd ui
INFRA_API_ENDPOINT=https://localhost:8443 npm start
```

### Port 3001 or 8443 already in use

**Solution:**

- Find and kill the process using the port: `lsof -i :3001` or `lsof -i :8443`
- Or use different ports by modifying `cypress.config.ts` and the
  `INFRA_API_ENDPOINT` environment variable

## Documentation

- Full Cypress documentation: https://docs.cypress.io/
- Cypress Best Practices:
  https://docs.cypress.io/guides/references/best-practices
