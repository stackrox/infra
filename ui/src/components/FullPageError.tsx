import React, { ReactElement } from 'react';
import { EmptyState, EmptyStateBody } from '@patternfly/react-core';

type Props = {
  message: string;
};

export default function FullPageError({ message }: Props): ReactElement {
  return (
    <EmptyState
      status="danger"
      variant="lg"
      titleText="There was an unexpected error!"
      headingLevel="h4"
    >
      <EmptyStateBody>{message}</EmptyStateBody>
    </EmptyState>
  );
}
