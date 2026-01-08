const ERROR_MESSAGES = {
  ACCESS_DENIED: 'access denied',
  UNEXPECTED_ERROR: 'There was an unexpected error',
};

const SELECTORS = {
  FLAVOR_CARD: '.pf-v6-c-card',
  CARD_TITLE: '.pf-v6-c-card__title',
  LABEL: '.pf-v6-c-label',
  FLAVOR_TOGGLE: 'input[name="flavor-filter-toggle"]',
  PAGE_HEADING: 'h2',
};

describe('Flavor Selection', () => {
  beforeEach(() => {
    // Authenticate for local development before visiting the page
    cy.loginForLocalDev();
    cy.visit('/');
  });

  it('should load the page without authentication errors', () => {
    // Verify no error messages
    cy.get('body').should('not.contain', ERROR_MESSAGES.ACCESS_DENIED);
    cy.get('body').should('not.contain', ERROR_MESSAGES.UNEXPECTED_ERROR);
  });

  it('should display a list of available flavors', () => {
    // Wait for the page to load, then check if flavors are available
    // In environments without flavors (e.g., PR clusters), this test will be skipped
    cy.get('body', { timeout: 30000 }).then(($body) => {
      if ($body.find(SELECTORS.PAGE_HEADING).length === 0) {
        cy.log('No flavors available in this environment - skipping test');
        // @ts-ignore - skip is a valid function
        return cy.skip();
      }

      // Page heading exists, verify it's visible
      cy.get(SELECTORS.PAGE_HEADING).should('be.visible');

      // Verify that the flavor gallery is not empty
      // Each flavor is rendered as a LinkCard inside a GalleryItem
      cy.get(SELECTORS.FLAVOR_CARD).should('have.length.at.least', 1);
    });
  });

  it('should display flavor details for each flavor card', () => {
    // Wait for the page to load, then check if flavors are available
    // In environments without flavors (e.g., PR clusters), this test will be skipped
    cy.get('body', { timeout: 30000 }).then(($body) => {
      if ($body.find(SELECTORS.PAGE_HEADING).length === 0) {
        cy.log('No flavors available in this environment - skipping test');
        // @ts-ignore - skip is a valid function
        return cy.skip();
      }

      // Page heading exists, verify it's visible
      cy.get(SELECTORS.PAGE_HEADING).should('be.visible');

      // Get the first flavor card and verify it has required elements
      cy.get(SELECTORS.FLAVOR_CARD, { timeout: 30000 })
        .first()
        .within(() => {
          // Each flavor card should have a name (header text)
          cy.get(SELECTORS.CARD_TITLE).should('exist').and('not.be.empty');

          // Each flavor card should have an availability label
          cy.get(SELECTORS.LABEL).should('exist');
        });
    });
  });

  it('should have clickable flavor cards that navigate to launch page', () => {
    // Wait for the page to load, then check if flavors are available
    // In environments without flavors (e.g., PR clusters), this test will be skipped
    cy.get('body', { timeout: 30000 }).then(($body) => {
      if ($body.find(SELECTORS.PAGE_HEADING).length === 0) {
        cy.log('No flavors available in this environment - skipping test');
        // @ts-ignore - skip is a valid function
        return cy.skip();
      }

      // Page heading exists, verify it's visible
      cy.get(SELECTORS.PAGE_HEADING).should('be.visible');

      // Click the first flavor card
      cy.get(SELECTORS.FLAVOR_CARD, { timeout: 30000 }).first().click();

      // Verify navigation to launch page (URL should contain /launch/)
      cy.url().should('include', '/launch/');

      // Verify the launch page loaded with a cluster launch form
      cy.contains('h1', /Launch/, { timeout: 30000 }).should('be.visible');
    });
  });

  it('should toggle between flavor filter states', () => {
    // Wait for the page to load, then check if flavors are available
    // In environments without flavors (e.g., PR clusters), this test will be skipped
    cy.get('body', { timeout: 30000 }).then(($body) => {
      if ($body.find(SELECTORS.PAGE_HEADING).length === 0) {
        cy.log('No flavors available in this environment - skipping test');
        // @ts-ignore - skip is a valid function
        return cy.skip();
      }

      // Page heading exists, get the initial heading text
      cy.get(SELECTORS.PAGE_HEADING)
        .should('be.visible')
        .invoke('text')
        .then((initialHeading) => {
          // Find and click the flavor filter toggle switch
          // Use force:true because PatternFly switch has a visual element covering the input
          cy.get(SELECTORS.FLAVOR_TOGGLE, { timeout: 30000 }).click({ force: true });

          // Verify the heading text changed
          cy.get(SELECTORS.PAGE_HEADING, { timeout: 30000 })
            .invoke('text')
            .should('not.equal', initialHeading);

          // Verify URL parameter was updated
          cy.url().should('include', 'showAllFlavors=true');

          // Toggle back
          cy.get(SELECTORS.FLAVOR_TOGGLE, { timeout: 30000 }).click({ force: true });

          // Verify we're back to the original heading
          cy.get(SELECTORS.PAGE_HEADING, { timeout: 30000 })
            .invoke('text')
            .should('equal', initialHeading);
        });
    });
  });
});
