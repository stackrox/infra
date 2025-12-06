// Custom Cypress commands for authentication

/**
 * Logs in by setting a valid JWT token cookie for local development.
 * This uses the known session secret from local-deploy oidc.yaml.
 */
Cypress.Commands.add('loginForLocalDev', () => {
  // The session secret from chart/infra-server/templates/secrets.yaml for local deploy
  const sessionSecret = 'local-dev-secret-min-32-chars-long';

  // Create a test user matching the backend's expected structure
  // Note: Fields are capitalized to match Go's JSON serialization of protobuf structs
  const testUser = {
    Name: 'Test User',
    Email: 'test@redhat.com',
    Picture: '',
    Expiry: {
      seconds: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
    },
  };

  // Create JWT payload matching the backend's userClaims structure
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    user: testUser,
    exp: now + 3600, // 1 hour expiry
    nbf: now,
    iat: now,
  };

  // Generate JWT token using HS256
  cy.task('generateJWT', { payload, secret: sessionSecret }).then((token) =>
    // Set the token cookie
    cy.setCookie('token', token as string, {
      path: '/',
      httpOnly: false,
      secure: false,
      sameSite: 'lax',
    })
  );
});

// TypeScript declaration for custom command
declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      /**
       * Custom command to log in for local development by setting a valid JWT token cookie.
       * @example cy.loginForLocalDev()
       */
      loginForLocalDev(): Chainable<void>;
    }
  }
}

export {};
