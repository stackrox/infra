// eslint-disable-next-line import/no-extraneous-dependencies
import { defineConfig } from 'cypress';
import * as crypto from 'crypto';

interface JWTPayload {
  user: {
    Name: string;
    Email: string;
    Picture: string;
    Expiry: {
      seconds: number;
    };
  };
  exp: number;
  nbf: number;
  iat: number;
}

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
      // Task to generate JWT tokens for local dev authentication
      on('task', {
        generateJWT({ payload, secret }: { payload: JWTPayload; secret: string }): string {
          // Create JWT header
          const header = {
            alg: 'HS256',
            typ: 'JWT',
          };

          // Base64url encode header and payload
          const base64UrlEncode = (obj: object) =>
            Buffer.from(JSON.stringify(obj))
              .toString('base64')
              .replace(/\+/g, '-')
              .replace(/\//g, '_')
              .replace(/=/g, '');

          const encodedHeader = base64UrlEncode(header);
          const encodedPayload = base64UrlEncode(payload);

          // Create signature
          const signatureInput = `${encodedHeader}.${encodedPayload}`;
          const signature = crypto
            .createHmac('sha256', secret)
            .update(signatureInput)
            .digest('base64')
            .replace(/\+/g, '-')
            .replace(/\//g, '_')
            .replace(/=/g, '');

          // Return complete JWT
          return `${encodedHeader}.${encodedPayload}.${signature}`;
        },
      });

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
