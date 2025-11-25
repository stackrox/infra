describe('Flavor Selection', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it('should load the page without authentication errors', () => {
    // Verify no error messages (confirms LOCAL_DEPLOY mode is working)
    cy.get('body').should('not.contain', 'access denied');
    cy.get('body').should('not.contain', 'There was an unexpected error');
  });

  it('should display a list of available flavors', () => {
    // Wait for flavors to load (check for either "My Flavors" or "All Flavors" title)
    cy.contains('h2', /My Flavors|All Flavors/).should('be.visible');

    // Verify that the flavor gallery is not empty
    // Each flavor is rendered as a LinkCard inside a GalleryItem
    cy.get('.pf-v6-c-card').should('have.length.at.least', 1);
  });

  it('should display flavor details for each flavor card', () => {
    // Wait for flavors to load
    cy.contains('h2', /My Flavors|All Flavors/).should('be.visible');

    // Get the first flavor card and verify it has required elements
    cy.get('.pf-v6-c-card')
      .first()
      .within(() => {
        // Each flavor card should have a name (header text)
        cy.get('.pf-v6-c-card__title').should('exist').and('not.be.empty');

        // Each flavor card should have an availability label
        cy.get('.pf-v6-c-label').should('exist');
      });
  });

  it('should have clickable flavor cards that navigate to launch page', () => {
    // Wait for flavors to load
    cy.contains('h2', /My Flavors|All Flavors/).should('be.visible');

    // Click the first flavor card
    cy.get('.pf-v6-c-card').first().click();

    // Verify navigation to launch page (URL should contain /launch/)
    cy.url().should('include', '/launch/');

    // Verify the launch page loaded with a cluster launch form
    cy.contains('h1', /Launch/).should('be.visible');
  });

  it('should toggle between "My Flavors" and "All Flavors"', () => {
    // Verify initial state
    cy.contains('h2', 'My Flavors').should('be.visible');

    // Find and click the "Show All Flavors" toggle switch
    // Use force:true because PatternFly switch has a visual element covering the input
    cy.get('input[name="flavor-filter-toggle"]').click({ force: true });

    // Verify the title changed to "All Flavors"
    cy.contains('h2', 'All Flavors').should('be.visible');

    // Verify URL parameter was updated
    cy.url().should('include', 'showAllFlavors=true');

    // Toggle back
    cy.get('input[name="flavor-filter-toggle"]').click({ force: true });

    // Verify we're back to "My Flavors"
    cy.contains('h2', 'My Flavors').should('be.visible');
  });
});
