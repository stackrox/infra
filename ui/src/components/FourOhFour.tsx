/* eslint-disable react/jsx-props-no-spreading */
import React, { ReactElement } from 'react';
import { EmptyState, EmptyStateActions, Button, EmptyStateBody } from '@patternfly/react-core';
import { OutlinedFlushedIcon } from '@patternfly/react-icons';
import { Link } from 'react-router-dom';

export default function FourOhFour(): ReactElement {
  return (
    <EmptyState
      className="pf-v6-u-my-2xl"
      icon={OutlinedFlushedIcon}
      variant="lg"
      titleText="You have found a page that does not seem to exist!"
      headingLevel="h4"
    >
      <EmptyStateBody>
        <EmptyStateActions>
          <Button variant="primary" component={(props) => <Link {...props} to="/" />}>
            Back to home
          </Button>
        </EmptyStateActions>
      </EmptyStateBody>
    </EmptyState>
  );
}
/* eslint-enable react/jsx-props-no-spreading */
