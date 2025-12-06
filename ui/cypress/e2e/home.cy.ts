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
    cy.get('body').should('not.contain', 'access denied');
    cy.get('body').should('not.contain', 'There was an unexpected error');
  });
});
