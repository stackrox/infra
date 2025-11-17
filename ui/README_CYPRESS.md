# Cypress E2E Testing

This directory contains Cypress E2E tests for the StackRox Infra UI.

## Quick Start - Running E2E Tests Against Local Backend

### Prerequisites

1. **Deploy the local backend** (with authentication disabled):

   ```bash
   # From the repository root
   make deploy-local
   ```

   This deploys the infra-server to your local Kubernetes cluster with
   `TEST_MODE=true`, which disables authentication for local development.

2. **Start port-forwarding** to access the backend:

   ```bash
   kubectl port-forward -n infra svc/infra-server-service 8443:8443
   ```

   Keep this running in a separate terminal.

3. **Configure the UI to connect to local backend**:

   ```bash
   cd ui
   cp .env.example .env.local
   ```

   This creates a `.env.local` file. Note: The file contains `INFRA_API_ENDPOINT`
   but the environment variable must also be set when starting the dev server (see next step).

4. **Start the UI dev server** (in a separate terminal):

   ```bash
   cd ui
   BROWSER=none PORT=3001 INFRA_API_ENDPOINT=http://localhost:8443 npm start
   ```

   **Important:** The `INFRA_API_ENDPOINT` environment variable must be set when starting
   the dev server (not just in `.env.local`) because the proxy middleware reads it at startup.

   Keep this running. The dev server will:
   - Run on http://localhost:3001
   - Proxy API requests to http://localhost:8443 (your local backend)
   - Hot-reload when you make changes to the UI code

5. **Run the E2E tests** (in another terminal):

   ```bash
   cd ui
   npm run cypress:run:e2e
   ```

That's it! The tests will run against the UI dev server at http://localhost:3001,
which proxies API requests to your local backend at `https://localhost:8443`.

### Test Results

After the tests complete:
- **Videos** are saved to `ui/cypress/videos/` (one per test file)
- **Screenshots** (on failures only) are saved to `ui/cypress/screenshots/`

Review the videos to verify the tests are properly accessing the backend.

## Interactive Mode

To run tests interactively with the Cypress UI (useful for debugging):

**Prerequisites:** Make sure the UI dev server is running (step 4 above).

```bash
cd ui
npm run cypress:open
```

Then:
1. Select "E2E Testing"
2. Choose a browser
3. Click on any test file to run it

Interactive mode lets you see the tests run in real-time, inspect the DOM, and debug failures.

## Test Structure

- `cypress/e2e/home.cy.ts` - Basic home page tests
- `cypress/e2e/flavor-selection.cy.ts` - Tests for flavor API integration

## Configuration

Tests are configured in `cypress.config.ts` to:

- Run against the UI dev server at `http://localhost:3001` (which proxies to the backend)
- Accept self-signed certificates (`chromeWebSecurity: false`)
- Capture videos of all test runs
- Capture screenshots on failures only
- Retry failed tests 2 times in CI mode (run mode), 0 times in interactive mode

The UI dev server (configured via `ui/.env.local`) proxies API requests to your
local backend at `https://localhost:8443`.

## Adding More Tests

To add new E2E tests:

1. Create a new file in `cypress/e2e/` with the pattern `*.cy.ts`
2. Follow the existing test patterns for consistency
3. Run the tests locally before committing

## Troubleshooting

### Tests fail with "Cypress failed to verify that your server is running"

**Solution:** Make sure the UI dev server is running on port 3001 before running tests:
```bash
cd ui
BROWSER=none PORT=3001 npm start
```

### Tests show "access denied" or authentication errors

**Solution:** Verify that:
1. The backend was deployed with `TEST_MODE=true` (via `make deploy-local`)
2. Port-forwarding is active: `kubectl port-forward -n infra svc/infra-server-service 8443:8443`
3. The `.env.local` file points to the correct backend: `INFRA_API_ENDPOINT=https://localhost:8443`

You can check if TEST_MODE is enabled:
```bash
kubectl get deployment -n infra infra-server-deployment -o jsonpath='{.spec.template.spec.containers[0].env}' | grep TEST_MODE
```

### Port 3001 or 8443 already in use

**Solution:**
- Find and kill the process using the port: `lsof -i :3001` or `lsof -i :8443`
- Or use different ports by modifying `cypress.config.ts` and `.env.local`

## Documentation

- Full Cypress documentation: https://docs.cypress.io/
- Cypress Best Practices: https://docs.cypress.io/guides/references/best-practices
