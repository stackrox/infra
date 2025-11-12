# Cypress E2E Testing

This directory contains Cypress E2E tests for the StackRox Infra UI.

## Quick Start - Running E2E Tests Against Local Backend

### Prerequisites

1. **Deploy the local backend** (with authentication disabled):
   ```bash
   # From the repository root
   ./scripts/deploy-local.sh
   ```

   This deploys the infra-server to your local Colima Kubernetes cluster with `TEST_MODE=true`, which disables authentication for local development.

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

   This creates a `.env.local` file that tells the UI dev server to proxy API requests to your local backend at `https://localhost:8443`.

4. **Run the E2E tests**:
   ```bash
   npm run cypress:run:e2e
   ```

That's it! The tests will run against the local backend at `https://localhost:8443`.

## Interactive Mode

To run tests interactively with the Cypress UI:

```bash
cd ui
npm run cypress:open
```

Then select "E2E Testing" and choose which tests to run.

## Test Structure

- `cypress/e2e/home.cy.ts` - Basic home page tests
- `cypress/e2e/flavor-selection.cy.ts` - Tests for flavor API integration

## Configuration

Tests are configured in `cypress.config.ts` to:
- Run against the local backend at `https://localhost:8443`
- Accept self-signed certificates
- Capture videos and screenshots on failures
- Retry failed tests in CI mode

## Adding More Tests

To add new E2E tests:

1. Create a new file in `cypress/e2e/` with the pattern `*.cy.ts`
2. Follow the existing test patterns for consistency
3. Run the tests locally before committing

## Documentation

- Full Cypress documentation: https://docs.cypress.io/
