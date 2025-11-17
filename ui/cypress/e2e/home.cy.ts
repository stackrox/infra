describe('Home Page', () => {
  beforeEach(() => {
    cy.visit('/').wait();
  });

  it('should load the home page', () => {
    cy.get('body').should('be.visible');
  });

  it('should not show error messages', () => {
    cy.get('body').should('not.contain', 'access denied');
    cy.get('body').should('not.contain', 'There was an unexpected error');
  });
});
