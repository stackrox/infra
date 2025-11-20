// eslint-disable-next-line import/no-extraneous-dependencies
import { defineConfig } from 'cypress';

export default defineConfig({
  e2e: {
    // UI dev server runs on :3001
    // API calls are proxied to backend at :8443 (configured via INFRA_API_ENDPOINT in setupProxy.js)
    // Routes proxied: /v1, /login, /callback, /logout, /downloads/infractl-*
    baseUrl: 'http://localhost:3001',
    specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
    supportFile: 'cypress/support/e2e.ts',
    fixturesFolder: 'cypress/fixtures',
    screenshotsFolder: 'cypress/screenshots',
    videosFolder: 'cypress/videos',
    viewportWidth: 1280,
    viewportHeight: 720,
    video: true,
    screenshotOnRunFailure: true,
    chromeWebSecurity: false, // Allow self-signed certificates for local dev
    setupNodeEvents(on, config) {
      // implement node event listeners here
      return config;
    },
    env: {
      // Default environment variables for tests
      // Can be overridden via CLI or cypress.env.json
      API_URL: 'https://localhost:8000',
    },
  },
  component: {
    devServer: {
      framework: 'react',
      bundler: 'vite',
    },
    specPattern: 'src/**/*.cy.{js,jsx,ts,tsx}',
    supportFile: 'cypress/support/component.ts',
  },
  retries: {
    runMode: 2, // Retry failed tests in CI
    openMode: 0, // Don't retry in interactive mode
  },
  defaultCommandTimeout: 10000,
  requestTimeout: 10000,
  responseTimeout: 10000,
});
