const ERROR_MESSAGES = {
  ACCESS_DENIED: 'access denied',
  UNEXPECTED_ERROR: 'There was an unexpected error',
};

describe('Home Page', () => {
  beforeEach(() => {
    // Authenticate for local development before visiting the page
    cy.loginForLocalDev();
    cy.visit('/');
  });

  it('should load the home page', () => {
    cy.get('body').should('be.visible');
  });

  it('should not show error messages', () => {
    cy.get('body').should('not.contain', ERROR_MESSAGES.ACCESS_DENIED);
    cy.get('body').should('not.contain', ERROR_MESSAGES.UNEXPECTED_ERROR);
  });
});
