describe('Flavor Selection', () => {
  it('should load the page without authentication errors', () => {
    cy.visit('/');

    // Verify no error messages (confirms TEST_MODE is working)
    cy.get('body').should('not.contain', 'access denied');
    cy.get('body').should('not.contain', 'There was an unexpected error');
  });
});
