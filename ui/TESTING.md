# Cypress E2E Testing

This directory contains Cypress E2E tests for the StackRox Infra UI.

## Testing Approaches

There are two ways to run the UI E2E tests:

1. **Against the Development Server** (Recommended - simplest option)
   - Extremely simple: just run `npm run test:e2e` (from the `ui/` directory)
   - Tests against real infrastructure at `dev.infra.rox.systems`
   - No local deployment needed
   - Requires access to the `team-automation` GCP project

2. **Against a Local Backend** (For offline work or testing local changes)
   - Self-contained, no external dependencies
   - Works offline
   - Tests your local code changes
   - Requires Docker/Podman and local Kubernetes (Colima, kind, etc.)

Choose the approach that best fits your development environment.

## Quick Start - Running E2E Tests Against Local Backend

**Note:** All commands in this section should be run from the repository root unless otherwise specified.

### Prerequisites

- Docker or Podman
- Local Kubernetes cluster (Colima, kind, minikube, Docker Desktop, etc.)
- Helm 3

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

## Alternative: Running E2E Tests Against Remote Server

If you have access to the development infra server at `dev.infra.rox.systems`,
you can run the UI tests against it instead of deploying locally. This approach is
simpler and mirrors how the Go e2e tests work.

### Prerequisites

- Access to the `team-automation` GCP project (or be added as a project owner)
- Valid authentication token (obtained by logging into https://dev.infra.rox.systems)

### Option 1: Using the Public Development Server (Simplest)

No kubectl or cluster access needed - just point to the public dev server:

```bash
cd ui
INFRA_API_ENDPOINT=https://dev.infra.rox.systems npm run test:e2e
```

**Note:** The UI dev server defaults to `https://dev.infra.rox.systems`, so you can
also omit the environment variable:

```bash
cd ui
npm run test:e2e
```

### Option 2: Using Port-Forward to Development Cluster

If you want to test against the actual cluster (e.g., to test pre-release changes):

1. **Connect to the development cluster**:

   ```bash
   gcloud container clusters get-credentials infra-development \
     --project acs-team-automation \
     --region us-west2
   ```

2. **Start port-forwarding**:

   ```bash
   kubectl -n infra port-forward svc/infra-server-service 8443:8443
   ```

   Keep this running in a separate terminal.

3. **Run the E2E tests**:

   ```bash
   cd ui
   INFRA_API_ENDPOINT=https://localhost:8443 npm run test:e2e
   ```

### Advantages of Remote Server Testing

- **Extremely simple** - Option 1 requires only `npm run test:e2e` (no setup!)
- **No local deployment needed** - Skip the `make image` and `make deploy-local` steps
- **Tests against real infrastructure** - Uses actual Argo Workflows and cloud resources
- **Consistent with Go e2e tests** - Same approach as existing test suite
- **Faster iteration** - No need to rebuild Docker images locally
- **Public endpoint available** - dev.infra.rox.systems is accessible without VPN

### Disadvantages

- **Requires team-automation access** - Need to be added to the GCP project
- **External dependencies** - Tests rely on remote services being available
- **Shared environment** - Other developers may be using the same dev server
- **May test older code** - Development server might not have your latest changes

## Interactive Mode

To run tests interactively with the Cypress UI (useful for debugging):

**Note:** This works with both local and remote backends. Just make sure you have
port-forwarding running to either your local deployment or remote server.

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
